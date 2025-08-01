package destination_transformer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"
	backendconfig "github.com/rudderlabs/rudder-server/backend-config"
	"github.com/rudderlabs/rudder-server/processor/internal/transformer/destination_transformer/embedded/warehouse"
	transformerfs "github.com/rudderlabs/rudder-server/services/transformer"

	"github.com/cenkalti/backoff"
	"github.com/samber/lo"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/filemanager"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/stats"

	transformerclient "github.com/rudderlabs/rudder-server/internal/transformer-client"
	"github.com/rudderlabs/rudder-server/processor/integrations"
	transformerutils "github.com/rudderlabs/rudder-server/processor/internal/transformer"
	"github.com/rudderlabs/rudder-server/processor/internal/transformer/destination_transformer/embedded/kafka"
	"github.com/rudderlabs/rudder-server/processor/internal/transformer/destination_transformer/embedded/pubsub"
	"github.com/rudderlabs/rudder-server/processor/types"
	"github.com/rudderlabs/rudder-server/utils/httputil"
	reportingtypes "github.com/rudderlabs/rudder-server/utils/types"
	warehouseutils "github.com/rudderlabs/rudder-server/warehouse/utils"
)

type warehouseClient interface {
	Transform(ctx context.Context, clientEvents []types.TransformerEvent) types.Response
	CompareResponsesAndUpload(ctx context.Context, events []types.TransformerEvent, legacyResponse types.Response)
}

type Opt func(*Client)

func WithClient(client transformerclient.Client) Opt {
	return func(s *Client) {
		s.client = client
	}
}

// WithFeatureService is used to set the feature service for the transformer client.
// It is used to check if the destination transformer supports compacted payloads.
// If this option is omitted, the transformer client will not be able to use compacted payloads.
func WithFeatureService(featureService transformerfs.FeaturesService) Opt {
	return func(s *Client) {
		if featureService == nil {
			return
		}
		go func() {
			// Wait for the feature service to be ready
			<-featureService.Wait()
			s.config.compactionSupported = featureService.SupportDestTransformCompactedPayloadV1()
		}()
	}
}

func New(conf *config.Config, log logger.Logger, stat stats.Stats, opts ...Opt) *Client {
	handle := &Client{}
	handle.conf = conf
	handle.log = log
	handle.stat = stat
	handle.client = transformerclient.NewClient(transformerutils.TransformerClientConfig(conf, "DestinationTransformer"))
	handle.config.destTransformationURL = handle.conf.GetString("DEST_TRANSFORM_URL", "http://localhost:9090")
	handle.config.timeoutDuration = conf.GetDuration("HttpClient.procTransformer.timeout", 600, time.Second)
	handle.config.maxRetry = conf.GetReloadableIntVar(30, 1, "Processor.DestinationTransformer.maxRetry", "Processor.maxRetry")
	handle.config.failOnError = conf.GetReloadableBoolVar(false, "Processor.DestinationTransformer.failOnError", "Processor.Transformer.failOnError")
	handle.config.maxRetryBackoffInterval = conf.GetReloadableDurationVar(30, time.Second, "Processor.DestinationTransformer.maxRetryBackoffInterval", "Processor.maxRetryBackoffInterval")
	handle.config.batchSize = conf.GetReloadableIntVar(100, 1, "Processor.DestinationTransformer.batchSize", "Processor.transformBatchSize")

	handle.config.maxLoggedEvents = conf.GetReloadableIntVar(100, 1, "Processor.DestinationTransformer.maxLoggedEvents")

	handle.stats.comparisonTime = handle.stat.NewStat("embedded_destination_transform_comparison_time", stats.TimerType)
	handle.stats.matchedEvents = handle.stat.NewStat("embedded_destination_transform_matched_events", stats.CountType)
	handle.stats.mismatchedEvents = handle.stat.NewStat("embedded_destination_transform_mismatched_events", stats.CountType)

	var err error
	handle.samplingFileManager, err = getSamplingUploader(conf, log)
	if err != nil {
		log.Errorn("failed to create dt sampling file manager", obskit.Error(err))
		handle.samplingFileManager = nil
	}

	handle.warehouseClient = warehouse.New(conf, log, stat)
	handle.config.warehouseTransformations.enable = conf.GetReloadableBoolVar(false, "Processor.enableWarehouseTransformations")
	handle.config.warehouseTransformations.verify = conf.GetReloadableBoolVar(true, "Processor.verifyWarehouseTransformations")

	handle.config.compactionEnabled = conf.GetReloadableBoolVar(false, "Processor.DestinationTransformer.compactionEnabled", "Transformer.compactionEnabled")

	for _, opt := range opts {
		opt(handle)
	}

	return handle
}

type Client struct {
	config struct {
		destTransformationURL   string
		maxRetry                config.ValueLoader[int]
		failOnError             config.ValueLoader[bool]
		maxRetryBackoffInterval config.ValueLoader[time.Duration]
		timeoutDuration         time.Duration
		batchSize               config.ValueLoader[int]

		maxLoggedEvents config.ValueLoader[int]

		compactionEnabled        config.ValueLoader[bool]
		warehouseTransformations struct {
			enable config.ValueLoader[bool]
			verify config.ValueLoader[bool]
		}
		compactionSupported bool
	}
	conf            *config.Config
	log             logger.Logger
	stat            stats.Stats
	client          transformerclient.Client
	warehouseClient warehouseClient

	stats struct {
		comparisonTime   stats.Timer
		matchedEvents    stats.Counter
		mismatchedEvents stats.Counter
	}

	loggedEvents        atomic.Int64
	samplingFileManager filemanager.S3Manager
}

func (d *Client) transform(ctx context.Context, clientEvents []types.TransformerEvent) types.Response {
	batchSize := d.config.batchSize.Load()
	if len(clientEvents) == 0 {
		return types.Response{}
	}

	destinationType := clientEvents[0].Destination.DestinationDefinition.Name
	destURL := d.destTransformURL(destinationType)

	labels := types.TransformerMetricLabels{
		Endpoint:        transformerutils.GetEndpointFromURL(destURL),
		Stage:           "dest_transformer",
		DestinationType: destinationType,
		DestinationID:   clientEvents[0].Destination.ID,
		SourceID:        clientEvents[0].Metadata.SourceID,
		WorkspaceID:     clientEvents[0].Metadata.WorkspaceID,
		SourceType:      clientEvents[0].Metadata.SourceType,
	}

	var trackWg sync.WaitGroup
	defer trackWg.Wait()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	trackWg.Add(1)
	go func() {
		l := d.log.Withn(labels.ToLoggerFields()...)
		transformerutils.TrackLongRunningTransformation(ctx, labels.Stage, d.config.timeoutDuration, l)
		trackWg.Done()
	}()

	batches := lo.Chunk(clientEvents, batchSize)

	d.stat.NewTaggedStat(
		"processor_transformer_request_batch_count",
		stats.HistogramType,
		labels.ToStatsTag(),
	).Observe(float64(len(batches)))

	transformResponse := make([][]types.TransformerResponse, len(batches))

	var wg sync.WaitGroup
	wg.Add(len(batches))
	var err error
	var foundError bool
	lo.ForEach(
		batches,
		func(batch []types.TransformerEvent, i int) {
			go func() {
				transformResponse[i], err = d.sendBatch(ctx, destURL, labels, batch)
				if err != nil {
					foundError = true
				}
				wg.Done()
			}()
		},
	)
	wg.Wait()

	if foundError {
		panic(err)
	}
	var outClientEvents []types.TransformerResponse
	var failedEvents []types.TransformerResponse

	for _, batch := range transformResponse {
		// Transform is one to many mapping so returned
		// response for each is an array. We flatten it out
		for _, transformerResponse := range batch {
			switch transformerResponse.StatusCode {
			case http.StatusOK:
				outClientEvents = append(outClientEvents, transformerResponse)
			default:
				failedEvents = append(failedEvents, transformerResponse)
			}
		}
	}

	d.stat.NewStat("processor_transformer_sent", stats.CountType).Count(len(clientEvents))
	d.stat.NewStat("processor_transformer_received", stats.CountType).Count(len(outClientEvents))

	return types.Response{
		Events:       outClientEvents,
		FailedEvents: failedEvents,
	}
}

func (d *Client) sendBatch(ctx context.Context, url string, labels types.TransformerMetricLabels, data []types.TransformerEvent) ([]types.TransformerResponse, error) {
	if len(data) == 0 {
		return nil, nil
	}
	start := time.Now()
	compactRequestPayloads := d.compactRequestPayloads() // consistent state for the entire request
	// Call remote transformation
	rawJSON, err := d.getRequestPayload(data, compactRequestPayloads)
	if err != nil {
		panic(err)
	}

	var extraHeaders map[string]string
	if compactRequestPayloads {
		extraHeaders = map[string]string{"X-Content-Format": "json+compactedv1"}
	}
	respData, statusCode, err := d.doPost(ctx, rawJSON, url, labels, extraHeaders)
	if err != nil {
		return nil, err
	}

	switch statusCode {
	case http.StatusOK,
		http.StatusBadRequest,
		http.StatusNotFound,
		http.StatusRequestEntityTooLarge:
	default:
		d.log.Errorn("Transformer returned status code", logger.NewStringField("statusCode", strconv.Itoa(statusCode)))
	}

	var transformerResponses []types.TransformerResponse
	switch statusCode {
	case http.StatusOK:
		integrations.CollectIntgTransformErrorStats(respData)

		err = jsonrs.Unmarshal(respData, &transformerResponses)
		// This is returned by our JS engine so should  be parsable
		// Panic the processor to avoid replays
		if err != nil {
			return nil, err
		}
	default:
		for i := range data {
			transformEvent := &data[i]
			resp := types.TransformerResponse{StatusCode: statusCode, Error: string(respData), Metadata: transformEvent.Metadata}
			transformerResponses = append(transformerResponses, resp)
		}
	}
	d.stat.NewTaggedStat("transformer_client_request_total_events", stats.CountType, labels.ToStatsTag()).Count(len(data))
	d.stat.NewTaggedStat("transformer_client_response_total_events", stats.CountType, labels.ToStatsTag()).Count(len(transformerResponses))
	d.stat.NewTaggedStat("transformer_client_total_time", stats.TimerType, labels.ToStatsTag()).SendTiming(time.Since(start))

	return transformerResponses, nil
}

func (d *Client) doPost(ctx context.Context, rawJSON []byte, url string, labels types.TransformerMetricLabels, extraHeaders map[string]string) ([]byte, int, error) {
	var (
		retryCount int
		resp       *http.Response
		respData   []byte
	)
	retryStrategy := backoff.NewExponentialBackOff()
	// MaxInterval caps the RetryInterval
	retryStrategy.MaxInterval = d.config.maxRetryBackoffInterval.Load()

	err := backoff.RetryNotify(
		func() error {
			var reqErr error
			requestStartTime := time.Now()

			var req *http.Request
			req, reqErr = http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(rawJSON))
			if reqErr != nil {
				return reqErr
			}

			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			req.Header.Set("X-Feature-Gzip-Support", "?1")
			// Header to let transformer know that the client understands event filter code
			req.Header.Set("X-Feature-Filter-Code", "?1")
			for k, v := range extraHeaders {
				req.Header.Set(k, v)
			}

			resp, reqErr = d.client.Do(req)
			defer func() { httputil.CloseResponse(resp) }()
			// Record metrics with labels
			tags := labels.ToStatsTag()
			duration := time.Since(requestStartTime)
			d.stat.NewTaggedStat("transformer_client_request_total_bytes", stats.CountType, tags).Count(len(rawJSON))

			d.stat.NewTaggedStat("transformer_client_total_durations_seconds", stats.CountType, tags).Count(int(duration.Seconds()))
			d.stat.NewTaggedStat("processor_transformer_request_time", stats.TimerType, labels.ToStatsTag()).SendTiming(duration)

			if reqErr != nil {
				return reqErr
			}

			if !transformerutils.IsJobTerminated(resp.StatusCode) {
				return fmt.Errorf("transformer returned status code: %v", resp.StatusCode)
			}

			respData, reqErr = io.ReadAll(resp.Body)
			if reqErr == nil {
				d.stat.NewTaggedStat("transformer_client_response_total_bytes", stats.CountType, tags).Count(len(respData))
				// We'll count response events after unmarshaling in the request method
			}
			return reqErr
		},
		backoff.WithMaxRetries(retryStrategy, uint64(d.config.maxRetry.Load())),
		func(err error, t time.Duration) {
			retryCount++
			d.log.Warnn(
				"JS HTTP connection error",
				logger.NewErrorField(err),
				logger.NewIntField("attempts", int64(retryCount)),
			)
		},
	)
	if err != nil {
		if d.config.failOnError.Load() {
			return []byte(fmt.Sprintf("transformer request failed: %s", err)), transformerutils.TransformerRequestFailure, nil
		} else {
			return nil, 0, err
		}
	}

	// perform version compatibility check only on success
	if resp.StatusCode == http.StatusOK {
		transformerAPIVersion, _ := strconv.Atoi(resp.Header.Get("apiVersion"))
		if reportingtypes.SupportedTransformerApiVersion != transformerAPIVersion {
			unexpectedVersionError := fmt.Errorf("incompatible transformer version: Expected: %d Received: %s, URL: %v", reportingtypes.SupportedTransformerApiVersion, resp.Header.Get("apiVersion"), url)
			d.log.Errorn("Unexpected version", obskit.Error(unexpectedVersionError))
			return nil, 0, unexpectedVersionError
		}
	}

	return respData, resp.StatusCode, nil
}

func (d *Client) destTransformURL(destType string) string {
	destinationEndPoint := fmt.Sprintf("%s/v0/destinations/%s", d.config.destTransformationURL, strings.ToLower(destType))

	if _, ok := warehouseutils.PseudoWarehouseDestinationMap[destType]; ok {
		whSchemaVersionQueryParam := fmt.Sprintf("whIDResolve=%t", d.conf.GetBool("Warehouse.enableIDResolution", false))
		switch destType {
		case warehouseutils.CLICKHOUSE:
			enableArraySupport := fmt.Sprintf("chEnableArraySupport=%s", fmt.Sprintf("%v", d.conf.GetBool("Warehouse.clickhouse.enableArraySupport", false)))
			return destinationEndPoint + "?" + whSchemaVersionQueryParam + "&" + enableArraySupport
		default:
			return destinationEndPoint + "?" + whSchemaVersionQueryParam
		}
	}
	return destinationEndPoint
}

type transformer func(ctx context.Context, clientEvents []types.TransformerEvent) types.Response

var embeddedTransformerImpls = map[string]transformer{
	"GOOGLEPUBSUB": pubsub.Transform,
	"KAFKA":        kafka.Transform,
}

func (c *Client) Transform(ctx context.Context, clientEvents []types.TransformerEvent) types.Response {
	if len(clientEvents) == 0 {
		return types.Response{}
	}

	destType := clientEvents[0].Destination.DestinationDefinition.Name
	if c.canRunWarehouseTransformations(destType) {
		if c.config.warehouseTransformations.verify.Load() {
			legacyResponse := c.transform(ctx, clientEvents)
			c.warehouseClient.CompareResponsesAndUpload(ctx, deepCopy(clientEvents), deepCopy(legacyResponse))
			return legacyResponse
		}
		return c.warehouseClient.Transform(ctx, clientEvents)
	}

	impl, ok := embeddedTransformerImpls[destType]
	if !ok {
		return c.transform(ctx, clientEvents)
	}
	if !c.conf.GetBoolVar(false, "Processor.Transformer.Embedded."+destType+".Enabled") {
		return c.transform(ctx, clientEvents)
	}
	if c.conf.GetBoolVar(true, "Processor.Transformer.Embedded."+destType+".Verify") {
		legacyTransformerResponse := c.transform(ctx, clientEvents)
		embeddedTransformerResponse := impl(ctx, clientEvents)
		c.CompareAndLog(ctx, embeddedTransformerResponse, legacyTransformerResponse)
		return legacyTransformerResponse
	}
	return impl(ctx, clientEvents)
}

func deepCopy[T any](src T) T {
	var dst T
	if data, err := jsonrs.Marshal(src); err == nil {
		_ = jsonrs.Unmarshal(data, &dst)
	}
	return dst
}

func (c *Client) canRunWarehouseTransformations(destType string) bool {
	if _, ok := warehouseutils.PseudoWarehouseDestinationMap[destType]; ok {
		return c.config.warehouseTransformations.enable.Load()
	}
	return false
}

func (d *Client) compactRequestPayloads() bool {
	return (d.config.compactionSupported && d.config.compactionEnabled.Load())
}

func (d *Client) getRequestPayload(data []types.TransformerEvent, compactRequestPayloads bool) ([]byte, error) {
	if compactRequestPayloads {
		ctr := types.CompactedTransformRequest{
			Input:        make([]types.CompactedTransformerEvent, 0, len(data)),
			Connections:  make(map[string]backendconfig.Connection),
			Destinations: make(map[string]backendconfig.DestinationT),
		}
		for i := range data {
			ctr.Input = append(ctr.Input, types.CompactedTransformerEvent{
				Message:     data[i].Message,
				Metadata:    data[i].Metadata,
				Libraries:   data[i].Libraries,
				Credentials: data[i].Credentials,
			})
			if _, ok := ctr.Destinations[data[i].Metadata.DestinationID]; !ok {
				ctr.Destinations[data[i].Metadata.DestinationID] = data[i].Destination
			}
			connectionKey := data[i].Metadata.SourceID + ":" + data[i].Metadata.DestinationID
			if _, ok := ctr.Connections[connectionKey]; !ok {
				ctr.Connections[connectionKey] = data[i].Connection
			}
		}
		return jsonrs.Marshal(&ctr)

	}
	return jsonrs.Marshal(data)
}

func getSamplingUploader(conf *config.Config, log logger.Logger) (filemanager.S3Manager, error) {
	var (
		bucket           = conf.GetString("DTSampling.Bucket", "processor-dt-sampling")
		endpoint         = conf.GetString("DTSampling.Endpoint", "")
		accessKeyID      = conf.GetStringVar("", "DTSampling.AccessKeyId", "AWS_ACCESS_KEY_ID")
		accessKey        = conf.GetStringVar("", "DTSampling.AccessKey", "AWS_SECRET_ACCESS_KEY")
		s3ForcePathStyle = conf.GetBool("DTSampling.S3ForcePathStyle", false)
		disableSSL       = conf.GetBool("DTSampling.DisableSsl", false)
		enableSSE        = conf.GetBoolVar(false, "DTSampling.EnableSse", "AWS_ENABLE_SSE")
		useGlue          = conf.GetBool("DTSampling.UseGlue", false)
		region           = conf.GetStringVar("us-east-1", "DTSampling.Region", "AWS_DEFAULT_REGION")
	)
	s3Config := map[string]any{
		"bucketName":       bucket,
		"endpoint":         endpoint,
		"accessKeyID":      accessKeyID,
		"accessKey":        accessKey,
		"s3ForcePathStyle": s3ForcePathStyle,
		"disableSSL":       disableSSL,
		"enableSSE":        enableSSE,
		"useGlue":          useGlue,
		"region":           region,
	}

	return filemanager.NewS3Manager(conf, s3Config, log.Withn(logger.NewStringField("component", "dt-uploader")), func() time.Duration {
		return conf.GetDuration("DTSampling.Timeout", 120, time.Second)
	})
}
