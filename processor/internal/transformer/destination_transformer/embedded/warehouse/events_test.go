package warehouse_test

import (
	"context"
	"net/http"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-server/internal/enricher"
	"github.com/rudderlabs/rudder-server/processor/internal/transformer/destination_transformer"
	"github.com/rudderlabs/rudder-server/processor/internal/transformer/destination_transformer/embedded/warehouse"
	"github.com/rudderlabs/rudder-server/processor/internal/transformer/destination_transformer/embedded/warehouse/internal/response"
	"github.com/rudderlabs/rudder-server/processor/internal/transformer/destination_transformer/embedded/warehouse/testhelper"
	whutils "github.com/rudderlabs/rudder-server/warehouse/utils"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"github.com/rudderlabs/rudder-server/processor/types"
	"github.com/rudderlabs/rudder-server/utils/misc"

	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/stats"
	transformertest "github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/transformer"

	backendconfig "github.com/rudderlabs/rudder-server/backend-config"
)

func TestEvents(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	transformerResource, err := transformertest.Setup(pool, t)
	require.NoError(t, err)

	t.Run("Basic events", func(t *testing.T) {
		identifyDefaultOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"anonymous_id":             "anonymousId",
					"channel":                  "web",
					"context_destination_id":   "destinationID",
					"context_destination_type": "POSTGRES",
					"context_ip":               "1.2.3.4",
					"context_passed_ip":        "1.2.3.4",
					"context_request_ip":       "5.6.7.8",
					"context_source_id":        "sourceID",
					"context_source_type":      "sourceType",
					"context_traits_email":     "rhedricks@example.com",
					"context_traits_logins":    float64(2),
					"context_traits_name":      "Richard Hendricks",
					"email":                    "rhedricks@example.com",
					"id":                       "messageId",
					"logins":                   float64(2),
					"name":                     "Richard Hendricks",
					"original_timestamp":       "2021-09-01T00:00:00.000Z",
					"product_id":               "9578257311",
					"rating":                   3.0,
					"received_at":              "2021-09-01T00:00:00.000Z",
					"review_body":              "OK for the price. It works but the material feels flimsy.",
					"review_id":                "86ac1cd43",
					"sent_at":                  "2021-09-01T00:00:00.000Z",
					"timestamp":                "2021-09-01T00:00:00.000Z",
					"user_id":                  "userId",
				},
				"metadata": map[string]any{
					"columns": map[string]any{
						"anonymous_id":             "string",
						"channel":                  "string",
						"context_destination_id":   "string",
						"context_destination_type": "string",
						"context_ip":               "string",
						"context_passed_ip":        "string",
						"context_request_ip":       "string",
						"context_source_id":        "string",
						"context_source_type":      "string",
						"context_traits_email":     "string",
						"context_traits_logins":    "int",
						"context_traits_name":      "string",
						"email":                    "string",
						"id":                       "string",
						"logins":                   "int",
						"name":                     "string",
						"original_timestamp":       "datetime",
						"product_id":               "string",
						"rating":                   "int",
						"received_at":              "datetime",
						"review_body":              "string",
						"review_id":                "string",
						"sent_at":                  "datetime",
						"timestamp":                "datetime",
						"user_id":                  "string",
						"uuid_ts":                  "datetime",
					},
					"receivedAt": "2021-09-01T00:00:00.000Z",
					"table":      "identifies",
				},
				"userId": "",
			}
		}
		userDefaultOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"context_destination_id":   "destinationID",
					"context_destination_type": "POSTGRES",
					"context_ip":               "1.2.3.4",
					"context_passed_ip":        "1.2.3.4",
					"context_request_ip":       "5.6.7.8",
					"context_source_id":        "sourceID",
					"context_source_type":      "sourceType",
					"context_traits_email":     "rhedricks@example.com",
					"context_traits_logins":    float64(2),
					"context_traits_name":      "Richard Hendricks",
					"email":                    "rhedricks@example.com",
					"id":                       "userId",
					"logins":                   float64(2),
					"name":                     "Richard Hendricks",
					"original_timestamp":       "2021-09-01T00:00:00.000Z",
					"product_id":               "9578257311",
					"rating":                   3.0,
					"received_at":              "2021-09-01T00:00:00.000Z",
					"review_body":              "OK for the price. It works but the material feels flimsy.",
					"review_id":                "86ac1cd43",
					"sent_at":                  "2021-09-01T00:00:00.000Z",
					"timestamp":                "2021-09-01T00:00:00.000Z",
				},
				"metadata": map[string]any{
					"columns": map[string]any{
						"context_destination_id":   "string",
						"context_destination_type": "string",
						"context_ip":               "string",
						"context_passed_ip":        "string",
						"context_request_ip":       "string",
						"context_source_id":        "string",
						"context_source_type":      "string",
						"context_traits_email":     "string",
						"context_traits_logins":    "int",
						"context_traits_name":      "string",
						"email":                    "string",
						"id":                       "string",
						"logins":                   "int",
						"name":                     "string",
						"original_timestamp":       "datetime",
						"product_id":               "string",
						"rating":                   "int",
						"received_at":              "datetime",
						"review_body":              "string",
						"review_id":                "string",
						"sent_at":                  "datetime",
						"timestamp":                "datetime",
						"uuid_ts":                  "datetime",
					},
					"receivedAt": "2021-09-01T00:00:00.000Z",
					"table":      "users",
				},
				"userId": "",
			}
		}
		identifyDefaultMergeOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"merge_property_1_type":  "anonymous_id",
					"merge_property_1_value": "anonymousId",
					"merge_property_2_type":  "user_id",
					"merge_property_2_value": "userId",
				},
				"metadata": map[string]any{
					"columns": map[string]any{
						"merge_property_1_type":  "string",
						"merge_property_1_value": "string",
						"merge_property_2_type":  "string",
						"merge_property_2_value": "string",
					},
					"isMergeRule":  true,
					"mergePropOne": "anonymousId",
					"mergePropTwo": "userId",
					"receivedAt":   "2021-09-01T00:00:00.000Z",
					"table":        "rudder_identity_merge_rules",
				},
				"userId": "",
			}
		}
		aliasDefaultOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"anonymous_id":             "anonymousId",
					"channel":                  "web",
					"context_ip":               "1.2.3.4",
					"context_passed_ip":        "1.2.3.4",
					"context_request_ip":       "5.6.7.8",
					"context_traits_email":     "rhedricks@example.com",
					"context_traits_logins":    float64(2),
					"id":                       "messageId",
					"original_timestamp":       "2021-09-01T00:00:00.000Z",
					"received_at":              "2021-09-01T00:00:00.000Z",
					"sent_at":                  "2021-09-01T00:00:00.000Z",
					"timestamp":                "2021-09-01T00:00:00.000Z",
					"title":                    "Home | RudderStack",
					"url":                      "https://www.rudderstack.com",
					"user_id":                  "userId",
					"previous_id":              "previousId",
					"context_destination_id":   "destinationID",
					"context_destination_type": "POSTGRES",
					"context_source_id":        "sourceID",
					"context_source_type":      "sourceType",
				},
				"metadata": map[string]any{
					"columns": map[string]any{
						"anonymous_id":             "string",
						"channel":                  "string",
						"context_destination_id":   "string",
						"context_destination_type": "string",
						"context_source_id":        "string",
						"context_source_type":      "string",
						"context_ip":               "string",
						"context_passed_ip":        "string",
						"context_request_ip":       "string",
						"context_traits_email":     "string",
						"context_traits_logins":    "int",
						"id":                       "string",
						"original_timestamp":       "datetime",
						"received_at":              "datetime",
						"sent_at":                  "datetime",
						"timestamp":                "datetime",
						"title":                    "string",
						"url":                      "string",
						"user_id":                  "string",
						"previous_id":              "string",
						"uuid_ts":                  "datetime",
					},
					"receivedAt": "2021-09-01T00:00:00.000Z",
					"table":      "aliases",
				},
				"userId": "",
			}
		}
		aliasMergeDefaultOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"merge_property_1_type":  "user_id",
					"merge_property_1_value": "userId",
					"merge_property_2_type":  "user_id",
					"merge_property_2_value": "previousId",
				},
				"metadata": map[string]any{
					"table":        "rudder_identity_merge_rules",
					"columns":      map[string]any{"merge_property_1_type": "string", "merge_property_1_value": "string", "merge_property_2_type": "string", "merge_property_2_value": "string"},
					"isMergeRule":  true,
					"receivedAt":   "2021-09-01T00:00:00.000Z",
					"mergePropOne": "userId",
					"mergePropTwo": "previousId",
				},
				"userId": "",
			}
		}
		extractDefaultOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"name":                     "Home",
					"context_ip":               "1.2.3.4",
					"context_traits_email":     "rhedricks@example.com",
					"context_traits_logins":    float64(2),
					"context_traits_name":      "Richard Hendricks",
					"id":                       "recordID",
					"event":                    "event",
					"received_at":              "2021-09-01T00:00:00.000Z",
					"title":                    "Home | RudderStack",
					"url":                      "https://www.rudderstack.com",
					"context_destination_id":   "destinationID",
					"context_destination_type": "POSTGRES",
					"context_source_id":        "sourceID",
					"context_source_type":      "sourceType",
				},
				"metadata": map[string]any{
					"columns": map[string]any{
						"name":                     "string",
						"context_destination_id":   "string",
						"context_destination_type": "string",
						"context_source_id":        "string",
						"context_source_type":      "string",
						"context_ip":               "string",
						"context_traits_email":     "string",
						"context_traits_logins":    "int",
						"context_traits_name":      "string",
						"id":                       "string",
						"event":                    "string",
						"received_at":              "datetime",
						"title":                    "string",
						"url":                      "string",
						"uuid_ts":                  "datetime",
					},
					"receivedAt": "2021-09-01T00:00:00.000Z",
					"table":      "event",
				},
				"userId": "",
			}
		}
		pageDefaultOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"name":                     "Home",
					"anonymous_id":             "anonymousId",
					"channel":                  "web",
					"context_ip":               "1.2.3.4",
					"context_passed_ip":        "1.2.3.4",
					"context_request_ip":       "5.6.7.8",
					"context_traits_email":     "rhedricks@example.com",
					"context_traits_logins":    float64(2),
					"context_traits_name":      "Richard Hendricks",
					"id":                       "messageId",
					"original_timestamp":       "2021-09-01T00:00:00.000Z",
					"received_at":              "2021-09-01T00:00:00.000Z",
					"sent_at":                  "2021-09-01T00:00:00.000Z",
					"timestamp":                "2021-09-01T00:00:00.000Z",
					"title":                    "Home | RudderStack",
					"url":                      "https://www.rudderstack.com",
					"user_id":                  "userId",
					"context_destination_id":   "destinationID",
					"context_destination_type": "POSTGRES",
					"context_source_id":        "sourceID",
					"context_source_type":      "sourceType",
				},
				"metadata": map[string]any{
					"columns": map[string]any{
						"name":                     "string",
						"anonymous_id":             "string",
						"channel":                  "string",
						"context_destination_id":   "string",
						"context_destination_type": "string",
						"context_source_id":        "string",
						"context_source_type":      "string",
						"context_ip":               "string",
						"context_passed_ip":        "string",
						"context_request_ip":       "string",
						"context_traits_email":     "string",
						"context_traits_logins":    "int",
						"context_traits_name":      "string",
						"id":                       "string",
						"original_timestamp":       "datetime",
						"received_at":              "datetime",
						"sent_at":                  "datetime",
						"timestamp":                "datetime",
						"title":                    "string",
						"url":                      "string",
						"user_id":                  "string",
						"uuid_ts":                  "datetime",
					},
					"receivedAt": "2021-09-01T00:00:00.000Z",
					"table":      "pages",
				},
				"userId": "",
			}
		}
		pageMergeDefaultOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"merge_property_1_type":  "anonymous_id",
					"merge_property_1_value": "anonymousId",
					"merge_property_2_type":  "user_id",
					"merge_property_2_value": "userId",
				},
				"metadata": map[string]any{
					"table":        "rudder_identity_merge_rules",
					"columns":      map[string]any{"merge_property_1_type": "string", "merge_property_1_value": "string", "merge_property_2_type": "string", "merge_property_2_value": "string"},
					"isMergeRule":  true,
					"receivedAt":   "2021-09-01T00:00:00.000Z",
					"mergePropOne": "anonymousId",
					"mergePropTwo": "userId",
				},
				"userId": "",
			}
		}
		screenDefaultOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"name":                     "Main",
					"anonymous_id":             "anonymousId",
					"channel":                  "web",
					"context_ip":               "1.2.3.4",
					"context_passed_ip":        "1.2.3.4",
					"context_request_ip":       "5.6.7.8",
					"context_traits_email":     "rhedricks@example.com",
					"context_traits_logins":    float64(2),
					"context_traits_name":      "Richard Hendricks",
					"id":                       "messageId",
					"original_timestamp":       "2021-09-01T00:00:00.000Z",
					"received_at":              "2021-09-01T00:00:00.000Z",
					"sent_at":                  "2021-09-01T00:00:00.000Z",
					"timestamp":                "2021-09-01T00:00:00.000Z",
					"title":                    "Home | RudderStack",
					"url":                      "https://www.rudderstack.com",
					"user_id":                  "userId",
					"context_destination_id":   "destinationID",
					"context_destination_type": "POSTGRES",
					"context_source_id":        "sourceID",
					"context_source_type":      "sourceType",
				},
				"metadata": map[string]any{
					"columns": map[string]any{
						"name":                     "string",
						"anonymous_id":             "string",
						"channel":                  "string",
						"context_destination_id":   "string",
						"context_destination_type": "string",
						"context_source_id":        "string",
						"context_source_type":      "string",
						"context_ip":               "string",
						"context_passed_ip":        "string",
						"context_request_ip":       "string",
						"context_traits_email":     "string",
						"context_traits_logins":    "int",
						"context_traits_name":      "string",
						"id":                       "string",
						"original_timestamp":       "datetime",
						"received_at":              "datetime",
						"sent_at":                  "datetime",
						"timestamp":                "datetime",
						"title":                    "string",
						"url":                      "string",
						"user_id":                  "string",
						"uuid_ts":                  "datetime",
					},
					"receivedAt": "2021-09-01T00:00:00.000Z",
					"table":      "screens",
				},
				"userId": "",
			}
		}
		screenMergeDefaultOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"merge_property_1_type":  "anonymous_id",
					"merge_property_1_value": "anonymousId",
					"merge_property_2_type":  "user_id",
					"merge_property_2_value": "userId",
				},
				"metadata": map[string]any{
					"table":        "rudder_identity_merge_rules",
					"columns":      map[string]any{"merge_property_1_type": "string", "merge_property_1_value": "string", "merge_property_2_type": "string", "merge_property_2_value": "string"},
					"isMergeRule":  true,
					"receivedAt":   "2021-09-01T00:00:00.000Z",
					"mergePropOne": "anonymousId",
					"mergePropTwo": "userId",
				},
				"userId": "",
			}
		}
		groupDefaultOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"anonymous_id":             "anonymousId",
					"channel":                  "web",
					"context_ip":               "1.2.3.4",
					"context_passed_ip":        "1.2.3.4",
					"context_request_ip":       "5.6.7.8",
					"context_traits_email":     "rhedricks@example.com",
					"context_traits_logins":    float64(2),
					"id":                       "messageId",
					"original_timestamp":       "2021-09-01T00:00:00.000Z",
					"received_at":              "2021-09-01T00:00:00.000Z",
					"sent_at":                  "2021-09-01T00:00:00.000Z",
					"timestamp":                "2021-09-01T00:00:00.000Z",
					"title":                    "Home | RudderStack",
					"url":                      "https://www.rudderstack.com",
					"user_id":                  "userId",
					"group_id":                 "groupId",
					"context_destination_id":   "destinationID",
					"context_destination_type": "POSTGRES",
					"context_source_id":        "sourceID",
					"context_source_type":      "sourceType",
				},
				"metadata": map[string]any{
					"columns": map[string]any{
						"anonymous_id":             "string",
						"channel":                  "string",
						"context_destination_id":   "string",
						"context_destination_type": "string",
						"context_source_id":        "string",
						"context_source_type":      "string",
						"context_ip":               "string",
						"context_passed_ip":        "string",
						"context_request_ip":       "string",
						"context_traits_email":     "string",
						"context_traits_logins":    "int",
						"id":                       "string",
						"original_timestamp":       "datetime",
						"received_at":              "datetime",
						"sent_at":                  "datetime",
						"timestamp":                "datetime",
						"title":                    "string",
						"url":                      "string",
						"user_id":                  "string",
						"group_id":                 "string",
						"uuid_ts":                  "datetime",
					},
					"receivedAt": "2021-09-01T00:00:00.000Z",
					"table":      "groups",
				},
				"userId": "",
			}
		}
		groupMergeDefaultOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"merge_property_1_type":  "anonymous_id",
					"merge_property_1_value": "anonymousId",
					"merge_property_2_type":  "user_id",
					"merge_property_2_value": "userId",
				},
				"metadata": map[string]any{
					"table":        "rudder_identity_merge_rules",
					"columns":      map[string]any{"merge_property_1_type": "string", "merge_property_1_value": "string", "merge_property_2_type": "string", "merge_property_2_value": "string"},
					"isMergeRule":  true,
					"receivedAt":   "2021-09-01T00:00:00.000Z",
					"mergePropOne": "anonymousId",
					"mergePropTwo": "userId",
				},
				"userId": "",
			}
		}
		trackDefaultOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"anonymous_id":             "anonymousId",
					"channel":                  "web",
					"context_destination_id":   "destinationID",
					"context_destination_type": "POSTGRES",
					"context_ip":               "1.2.3.4",
					"context_passed_ip":        "1.2.3.4",
					"context_request_ip":       "5.6.7.8",
					"context_source_id":        "sourceID",
					"context_source_type":      "sourceType",
					"context_traits_email":     "rhedricks@example.com",
					"context_traits_logins":    float64(2),
					"context_traits_name":      "Richard Hendricks",
					"event":                    "event",
					"event_text":               "event",
					"id":                       "messageId",
					"original_timestamp":       "2021-09-01T00:00:00.000Z",
					"received_at":              "2021-09-01T00:00:00.000Z",
					"sent_at":                  "2021-09-01T00:00:00.000Z",
					"timestamp":                "2021-09-01T00:00:00.000Z",
					"user_id":                  "userId",
				},
				"metadata": map[string]any{
					"columns": map[string]any{
						"anonymous_id":             "string",
						"channel":                  "string",
						"context_destination_id":   "string",
						"context_destination_type": "string",
						"context_ip":               "string",
						"context_passed_ip":        "string",
						"context_request_ip":       "string",
						"context_source_id":        "string",
						"context_source_type":      "string",
						"context_traits_email":     "string",
						"context_traits_logins":    "int",
						"context_traits_name":      "string",
						"event":                    "string",
						"event_text":               "string",
						"id":                       "string",
						"original_timestamp":       "datetime",
						"received_at":              "datetime",
						"sent_at":                  "datetime",
						"timestamp":                "datetime",
						"user_id":                  "string",
						"uuid_ts":                  "datetime",
					},
					"receivedAt": "2021-09-01T00:00:00.000Z",
					"table":      "tracks",
				},
				"userId": "",
			}
		}
		trackEventDefaultOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"anonymous_id":             "anonymousId",
					"channel":                  "web",
					"context_destination_id":   "destinationID",
					"context_destination_type": "POSTGRES",
					"context_ip":               "1.2.3.4",
					"context_passed_ip":        "1.2.3.4",
					"context_request_ip":       "5.6.7.8",
					"context_source_id":        "sourceID",
					"context_source_type":      "sourceType",
					"context_traits_email":     "rhedricks@example.com",
					"context_traits_logins":    float64(2),
					"context_traits_name":      "Richard Hendricks",
					"event":                    "event",
					"event_text":               "event",
					"id":                       "messageId",
					"original_timestamp":       "2021-09-01T00:00:00.000Z",
					"product_id":               "9578257311",
					"rating":                   3.0,
					"received_at":              "2021-09-01T00:00:00.000Z",
					"review_body":              "OK for the price. It works but the material feels flimsy.",
					"review_id":                "86ac1cd43",
					"sent_at":                  "2021-09-01T00:00:00.000Z",
					"timestamp":                "2021-09-01T00:00:00.000Z",
					"user_id":                  "userId",
				},
				"metadata": map[string]any{
					"columns": map[string]any{
						"anonymous_id":             "string",
						"channel":                  "string",
						"context_destination_id":   "string",
						"context_destination_type": "string",
						"context_ip":               "string",
						"context_passed_ip":        "string",
						"context_request_ip":       "string",
						"context_source_id":        "string",
						"context_source_type":      "string",
						"context_traits_email":     "string",
						"context_traits_logins":    "int",
						"context_traits_name":      "string",
						"event":                    "string",
						"event_text":               "string",
						"id":                       "string",
						"original_timestamp":       "datetime",
						"product_id":               "string",
						"rating":                   "int",
						"received_at":              "datetime",
						"review_body":              "string",
						"review_id":                "string",
						"sent_at":                  "datetime",
						"timestamp":                "datetime",
						"user_id":                  "string",
						"uuid_ts":                  "datetime",
					},
					"receivedAt": "2021-09-01T00:00:00.000Z",
					"table":      "event",
				},
				"userId": "",
			}
		}
		trackMergeDefaultOutput := func() testhelper.OutputBuilder {
			return testhelper.OutputBuilder{
				"data": map[string]any{
					"merge_property_1_type":  "anonymous_id",
					"merge_property_1_value": "anonymousId",
					"merge_property_2_type":  "user_id",
					"merge_property_2_value": "userId",
				},
				"metadata": map[string]any{
					"columns": map[string]any{
						"merge_property_1_type":  "string",
						"merge_property_1_value": "string",
						"merge_property_2_type":  "string",
						"merge_property_2_value": "string",
					},
					"isMergeRule":  true,
					"mergePropOne": "anonymousId",
					"mergePropTwo": "userId",
					"receivedAt":   "2021-09-01T00:00:00.000Z",
					"table":        "rudder_identity_merge_rules",
				},
				"userId": "",
			}
		}

		testCases := []struct {
			name             string
			configOverride   map[string]any
			eventPayload     string
			metadata         types.Metadata
			destination      backendconfig.DestinationT
			expectedResponse types.Response
		}{
			{
				name:         "identify (POSTGRES)",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("identify", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     identifyDefaultOutput(),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
						{
							Output:     userDefaultOutput(),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (POSTGRES) Empty userID",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("identify", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     identifyDefaultOutput().RemoveDataFields("user_id").RemoveColumnFields("user_id"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (POSTGRES) userID with spaces",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"   ","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("identify", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     identifyDefaultOutput().SetDataField("user_id", "   "),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (S3_DATALAKE)",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("identify", "S3_DATALAKE"),
				destination: getDestination("S3_DATALAKE", map[string]any{
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: identifyDefaultOutput().
								SetDataField("_timestamp", "2021-09-01T00:00:00.000Z").
								SetColumnField("_timestamp", "datetime").
								RemoveDataFields("timestamp").
								RemoveColumnFields("timestamp").
								SetDataField("context_destination_type", "S3_DATALAKE"),
							Metadata:   getMetadata("identify", "S3_DATALAKE"),
							StatusCode: http.StatusOK,
						},
						{
							Output: userDefaultOutput().
								RemoveDataFields("timestamp", "original_timestamp", "sent_at").
								RemoveColumnFields("timestamp", "original_timestamp", "sent_at").
								SetDataField("context_destination_type", "S3_DATALAKE"),
							Metadata:   getMetadata("identify", "S3_DATALAKE"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (POSTGRES) without traits",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("identify", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: identifyDefaultOutput().
								RemoveDataFields("product_id", "review_id").
								RemoveColumnFields("product_id", "review_id"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
						{
							Output: userDefaultOutput().
								RemoveDataFields("product_id", "review_id").
								RemoveColumnFields("product_id", "review_id"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (POSTGRES) without userProperties",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("identify", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: identifyDefaultOutput().
								RemoveDataFields("rating", "review_body").
								RemoveColumnFields("rating", "review_body"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
						{
							Output: userDefaultOutput().
								RemoveDataFields("rating", "review_body").
								RemoveColumnFields("rating", "review_body"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (POSTGRES) without context.traits",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("identify", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: identifyDefaultOutput().
								RemoveDataFields("context_traits_email", "context_traits_logins", "context_traits_name", "email", "logins", "name").
								RemoveColumnFields("context_traits_email", "context_traits_logins", "context_traits_name", "email", "logins", "name"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
						{
							Output: userDefaultOutput().
								RemoveDataFields("context_traits_email", "context_traits_logins", "context_traits_name", "email", "logins", "name").
								RemoveColumnFields("context_traits_email", "context_traits_logins", "context_traits_name", "email", "logins", "name"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (POSTGRES) without context",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."}}`,
				metadata:     getMetadata("identify", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: identifyDefaultOutput().
								SetDataField("context_ip", "5.6.7.8"). // overriding the default value
								RemoveDataFields("context_passed_ip", "context_traits_email", "context_traits_logins", "context_traits_name", "email", "logins", "name").
								RemoveColumnFields("context_passed_ip", "context_traits_email", "context_traits_logins", "context_traits_name", "email", "logins", "name"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
						{
							Output: userDefaultOutput().
								SetDataField("context_ip", "5.6.7.8"). // overriding the default value
								RemoveDataFields("context_passed_ip", "context_traits_email", "context_traits_logins", "context_traits_name", "email", "logins", "name").
								RemoveColumnFields("context_passed_ip", "context_traits_email", "context_traits_logins", "context_traits_name", "email", "logins", "name"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (POSTGRES) not allowUsersContextTraits",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("identify", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: identifyDefaultOutput().
								RemoveDataFields("email", "logins", "name").
								RemoveColumnFields("email", "logins", "name"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
						{
							Output: userDefaultOutput().
								RemoveDataFields("email", "logins", "name").
								RemoveColumnFields("email", "logins", "name"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (POSTGRES) user_id already exists",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"user_id":"user_id","rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("identify", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     identifyDefaultOutput(),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
						{
							Output:     userDefaultOutput(),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (POSTGRES) store rudder event",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("identify", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"storeFullEvent":          true,
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: identifyDefaultOutput().
								SetDataField("rudder_event", "{\"anonymousId\":\"anonymousId\",\"channel\":\"web\",\"context\":{\"ip\":\"1.2.3.4\",\"traits\":{\"email\":\"rhedricks@example.com\",\"logins\":2,\"name\":\"Richard Hendricks\"},\"sourceId\":\"sourceID\",\"sourceType\":\"sourceType\",\"destinationId\":\"destinationID\",\"destinationType\":\"POSTGRES\"},\"messageId\":\"messageId\",\"originalTimestamp\":\"2021-09-01T00:00:00.000Z\",\"receivedAt\":\"2021-09-01T00:00:00.000Z\",\"request_ip\":\"5.6.7.8\",\"sentAt\":\"2021-09-01T00:00:00.000Z\",\"timestamp\":\"2021-09-01T00:00:00.000Z\",\"traits\":{\"product_id\":\"9578257311\",\"review_id\":\"86ac1cd43\"},\"type\":\"identify\",\"userId\":\"userId\",\"userProperties\":{\"rating\":3,\"review_body\":\"OK for the price. It works but the material feels flimsy.\"}}").
								SetColumnField("rudder_event", "json"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
						{
							Output:     userDefaultOutput(),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (POSTGRES) partial rules",
				eventPayload: `{"type":"identify","messageId":"messageId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("identify", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: identifyDefaultOutput().
								RemoveDataFields("anonymous_id", "channel", "context_request_ip").
								RemoveColumnFields("anonymous_id", "channel", "context_request_ip"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
						{
							Output: userDefaultOutput().
								RemoveDataFields("context_request_ip").
								RemoveColumnFields("context_request_ip"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (POSTGRES) no userID",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("identify", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: identifyDefaultOutput().
								RemoveDataFields("user_id").
								RemoveColumnFields("user_id"),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (POSTGRES) skipUsersTable (destOpts)",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("identify", "POSTGRES"),
				destination: backendconfig.DestinationT{
					Name: "POSTGRES",
					Config: map[string]any{
						"allowUsersContextTraits": true,
						"skipUsersTable":          true,
					},
					DestinationDefinition: backendconfig.DestinationDefinitionT{
						Name: "POSTGRES",
					},
				},
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     identifyDefaultOutput(),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (POSTGRES) skipUsersTable (intrOpts)",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"},"integrations":{"POSTGRES":{"options":{"skipUsersTable":true}}}}`,
				metadata:     getMetadata("identify", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     identifyDefaultOutput(),
							Metadata:   getMetadata("identify", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name: "identify (BQ) merge event",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"},"integrations":{"POSTGRES":{"options":{"skipUsersTable":true}}}}`,
				metadata:     getMetadata("identify", "BQ"),
				destination: getDestination("BQ", map[string]any{
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: identifyDefaultOutput().
								SetDataField("context_destination_type", "BQ").
								SetColumnField("loaded_at", "datetime"),
							Metadata:   getMetadata("identify", "BQ"),
							StatusCode: http.StatusOK,
						},
						{
							Output:     identifyDefaultMergeOutput(),
							Metadata:   getMetadata("identify", "BQ"),
							StatusCode: http.StatusOK,
						},
						{
							Output: userDefaultOutput().
								SetDataField("context_destination_type", "BQ").
								SetColumnField("loaded_at", "datetime"),
							Metadata:   getMetadata("identify", "BQ"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "identify (Snowflake)",
				eventPayload: `{"type":"identify","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("identify", "SNOWFLAKE"),
				destination: getDestination("SNOWFLAKE", map[string]any{
					"allowUsersContextTraits": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: identifyDefaultOutput().
								SetDataField("context_destination_type", "SNOWFLAKE").
								BuildForSnowflake(),
							Metadata:   getMetadata("identify", "SNOWFLAKE"),
							StatusCode: http.StatusOK,
						},
						{
							Output: userDefaultOutput().
								SetDataField("context_destination_type", "SNOWFLAKE").
								BuildForSnowflake(),
							Metadata:   getMetadata("identify", "SNOWFLAKE"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},

			{
				name:         "alias (Postgres)",
				eventPayload: `{"type":"alias","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","previousId":"previousId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("alias", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     aliasDefaultOutput(),
							Metadata:   getMetadata("alias", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "alias (Postgres) without traits",
				eventPayload: `{"type":"alias","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","previousId":"previousId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("alias", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: aliasDefaultOutput().
								RemoveDataFields("title", "url").
								RemoveColumnFields("title", "url"),
							Metadata:   getMetadata("alias", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "alias (Postgres) without context",
				eventPayload: `{"type":"alias","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","previousId":"previousId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"}}`,
				metadata:     getMetadata("alias", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: aliasDefaultOutput().
								SetDataField("context_ip", "5.6.7.8"). // overriding the default value
								RemoveDataFields("context_passed_ip", "context_traits_email", "context_traits_logins").
								RemoveColumnFields("context_passed_ip", "context_traits_email", "context_traits_logins"),
							Metadata:   getMetadata("alias", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "alias (Postgres) store rudder event",
				eventPayload: `{"type":"alias","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","previousId":"previousId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("alias", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"storeFullEvent": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: aliasDefaultOutput().
								SetDataField("rudder_event", "{\"type\":\"alias\",\"anonymousId\":\"anonymousId\",\"channel\":\"web\",\"context\":{\"destinationId\":\"destinationID\",\"destinationType\":\"POSTGRES\",\"ip\":\"1.2.3.4\",\"sourceId\":\"sourceID\",\"sourceType\":\"sourceType\",\"traits\":{\"email\":\"rhedricks@example.com\",\"logins\":2}},\"messageId\":\"messageId\",\"originalTimestamp\":\"2021-09-01T00:00:00.000Z\",\"previousId\":\"previousId\",\"receivedAt\":\"2021-09-01T00:00:00.000Z\",\"request_ip\":\"5.6.7.8\",\"sentAt\":\"2021-09-01T00:00:00.000Z\",\"timestamp\":\"2021-09-01T00:00:00.000Z\",\"traits\":{\"title\":\"Home | RudderStack\",\"url\":\"https://www.rudderstack.com\"},\"userId\":\"userId\"}").
								SetColumnField("rudder_event", "json"),
							Metadata:   getMetadata("alias", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "alias (Postgres) partial rules",
				eventPayload: `{"type":"alias","messageId":"messageId","userId":"userId","previousId":"previousId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("alias", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: aliasDefaultOutput().
								RemoveDataFields("anonymous_id", "channel", "context_request_ip").
								RemoveColumnFields("anonymous_id", "channel", "context_request_ip"),
							Metadata:   getMetadata("alias", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name: "alias (BQ) merge event",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"alias","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","previousId":"previousId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("alias", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: aliasDefaultOutput().
								SetDataField("context_destination_type", "BQ").
								SetColumnField("loaded_at", "datetime"),
							Metadata:   getMetadata("alias", "BQ"),
							StatusCode: http.StatusOK,
						},
						{
							Output:     aliasMergeDefaultOutput(),
							Metadata:   getMetadata("alias", "BQ"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "alias (Snowflake)",
				eventPayload: `{"type":"alias","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","previousId":"previousId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("alias", "SNOWFLAKE"),
				destination:  getDestination("SNOWFLAKE", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: aliasDefaultOutput().
								SetDataField("context_destination_type", "SNOWFLAKE").
								BuildForSnowflake(),
							Metadata:   getMetadata("alias", "SNOWFLAKE"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},

			{
				name:         "extract (Postgres)",
				eventPayload: `{"type":"extract","recordId":"recordID","messageId":"messageId","event":"event","receivedAt":"2021-09-01T00:00:00.000Z","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("extract", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     extractDefaultOutput(),
							Metadata:   getMetadata("extract", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "extract (Postgres) Empty event",
				eventPayload: `{"type":"extract","recordId":"recordID","messageId":"messageId","event":"","receivedAt":"2021-09-01T00:00:00.000Z","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("extract", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					FailedEvents: []types.TransformerResponse{
						{
							Error:      response.ErrExtractEventNameEmpty.Error(),
							StatusCode: response.ErrExtractEventNameEmpty.StatusCode(),
							Metadata:   getMetadata("extract", "POSTGRES"),
						},
					},
				},
			},
			{
				name:         "extract (Postgres) Event with spaces",
				eventPayload: `{"type":"extract","recordId":"recordID","messageId":"messageId","event":"","receivedAt":"2021-09-01T00:00:00.000Z","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("extract", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					FailedEvents: []types.TransformerResponse{
						{
							Error:      response.ErrExtractEventNameEmpty.Error(),
							StatusCode: response.ErrExtractEventNameEmpty.StatusCode(),
							Metadata:   getMetadata("extract", "POSTGRES"),
						},
					},
				},
			},
			{
				name:         "extract (Postgres) without properties",
				eventPayload: `{"type":"extract","recordId":"recordID","messageId":"messageId","event":"event","receivedAt":"2021-09-01T00:00:00.000Z","context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("extract", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: extractDefaultOutput().
								RemoveDataFields("name", "title", "url").
								RemoveColumnFields("name", "title", "url"),
							Metadata:   getMetadata("extract", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "extract (Postgres) without context",
				eventPayload: `{"type":"extract","recordId":"recordID","messageId":"messageId","event":"event","receivedAt":"2021-09-01T00:00:00.000Z","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"}}`,
				metadata:     getMetadata("extract", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: extractDefaultOutput().
								SetDataField("context_ip", "5.6.7.8"). // overriding the default value
								RemoveDataFields("context_ip", "context_traits_email", "context_traits_logins", "context_traits_name").
								RemoveColumnFields("context_ip", "context_traits_email", "context_traits_logins", "context_traits_name"),
							Metadata:   getMetadata("extract", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "extract (Postgres) RudderCreatedTable",
				eventPayload: `{"type":"extract","recordId":"recordID","messageId":"messageId","event":"accounts","receivedAt":"2021-09-01T00:00:00.000Z","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("extract", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"storeFullEvent": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: extractDefaultOutput().
								SetDataField("event", "accounts").
								SetTableName("_accounts"),
							Metadata:   getMetadata("extract", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "extract (Postgres) RudderCreatedTable with skipReservedKeywordsEscaping",
				eventPayload: `{"type":"extract","recordId":"recordID","messageId":"messageId","event":"accounts","receivedAt":"2021-09-01T00:00:00.000Z","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"},"integrations":{"POSTGRES":{"options":{"skipReservedKeywordsEscaping":true}}}}`,
				metadata:     getMetadata("extract", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: extractDefaultOutput().
								SetDataField("event", "accounts").
								SetTableName("accounts"),
							Metadata:   getMetadata("extract", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "extract (Postgres) RudderIsolatedTable",
				eventPayload: `{"type":"extract","recordId":"recordID","messageId":"messageId","event":"users","receivedAt":"2021-09-01T00:00:00.000Z","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("extract", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: extractDefaultOutput().
								SetDataField("event", "users").
								SetTableName("_users"),
							Metadata:   getMetadata("extract", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "extract (Snowflake)",
				eventPayload: `{"type":"extract","recordId":"recordID","messageId":"messageId","event":"event","receivedAt":"2021-09-01T00:00:00.000Z","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("extract", "SNOWFLAKE"),
				destination:  getDestination("SNOWFLAKE", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: extractDefaultOutput().
								SetDataField("context_destination_type", "SNOWFLAKE").
								BuildForSnowflake(),
							Metadata:   getMetadata("extract", "SNOWFLAKE"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},

			{
				name:         "page (Postgres)",
				eventPayload: `{"type":"page","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("page", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     pageDefaultOutput(),
							Metadata:   getMetadata("page", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "page (Postgres) without properties",
				eventPayload: `{"type":"page","name":"Home","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("page", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: pageDefaultOutput().
								RemoveDataFields("title", "url").
								RemoveColumnFields("title", "url"),
							Metadata:   getMetadata("page", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "page (Postgres) without context",
				eventPayload: `{"type":"page","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"}}`,
				metadata:     getMetadata("page", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: pageDefaultOutput().
								SetDataField("context_ip", "5.6.7.8"). // overriding the default value
								RemoveDataFields("context_passed_ip", "context_traits_email", "context_traits_logins", "context_traits_name").
								RemoveColumnFields("context_passed_ip", "context_traits_email", "context_traits_logins", "context_traits_name"),
							Metadata:   getMetadata("page", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "page (Postgres) store rudder event",
				eventPayload: `{"type":"page","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("page", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"storeFullEvent": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: pageDefaultOutput().
								SetDataField("rudder_event", "{\"anonymousId\":\"anonymousId\",\"channel\":\"web\",\"context\":{\"ip\":\"1.2.3.4\",\"traits\":{\"email\":\"rhedricks@example.com\",\"logins\":2,\"name\":\"Richard Hendricks\"},\"sourceId\":\"sourceID\",\"sourceType\":\"sourceType\",\"destinationId\":\"destinationID\",\"destinationType\":\"POSTGRES\"},\"messageId\":\"messageId\",\"originalTimestamp\":\"2021-09-01T00:00:00.000Z\",\"properties\":{\"name\":\"Home\",\"title\":\"Home | RudderStack\",\"url\":\"https://www.rudderstack.com\"},\"receivedAt\":\"2021-09-01T00:00:00.000Z\",\"request_ip\":\"5.6.7.8\",\"sentAt\":\"2021-09-01T00:00:00.000Z\",\"timestamp\":\"2021-09-01T00:00:00.000Z\",\"type\":\"page\",\"userId\":\"userId\"}").
								SetColumnField("rudder_event", "json"),
							Metadata:   getMetadata("page", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "page (Postgres) partial rules",
				eventPayload: `{"type":"page","messageId":"messageId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("page", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: pageDefaultOutput().
								RemoveDataFields("anonymous_id", "channel", "context_request_ip").
								RemoveColumnFields("anonymous_id", "channel", "context_request_ip"),
							Metadata:   getMetadata("page", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name: "page (BQ) merge event",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"page","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("page", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: pageDefaultOutput().
								SetDataField("context_destination_type", "BQ").
								SetColumnField("loaded_at", "datetime"),
							Metadata:   getMetadata("page", "BQ"),
							StatusCode: http.StatusOK,
						},
						{
							Output:     pageMergeDefaultOutput(),
							Metadata:   getMetadata("page", "BQ"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "page (Snowflake)",
				eventPayload: `{"type":"page","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("page", "SNOWFLAKE"),
				destination:  getDestination("SNOWFLAKE", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: pageDefaultOutput().
								SetDataField("context_destination_type", "SNOWFLAKE").
								BuildForSnowflake(),
							Metadata:   getMetadata("page", "SNOWFLAKE"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},

			{
				name:         "screen (Postgres)",
				eventPayload: `{"type":"screen","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","properties":{"name":"Main","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("screen", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     screenDefaultOutput(),
							Metadata:   getMetadata("screen", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "screen (Postgres) without properties",
				eventPayload: `{"type":"screen","name":"Main","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("screen", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: screenDefaultOutput().
								RemoveDataFields("title", "url").
								RemoveColumnFields("title", "url"),
							Metadata:   getMetadata("screen", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "screen (Postgres) without context",
				eventPayload: `{"type":"screen","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","properties":{"name":"Main","title":"Home | RudderStack","url":"https://www.rudderstack.com"}}`,
				metadata:     getMetadata("screen", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: screenDefaultOutput().
								SetDataField("context_ip", "5.6.7.8"). // overriding the default value
								RemoveDataFields("context_passed_ip", "context_traits_email", "context_traits_logins", "context_traits_name").
								RemoveColumnFields("context_passed_ip", "context_traits_email", "context_traits_logins", "context_traits_name"),
							Metadata:   getMetadata("screen", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "screen (Postgres) store rudder event",
				eventPayload: `{"type":"screen","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","properties":{"name":"Main","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("screen", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"storeFullEvent": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: screenDefaultOutput().
								SetDataField("rudder_event", "{\"anonymousId\":\"anonymousId\",\"channel\":\"web\",\"context\":{\"ip\":\"1.2.3.4\",\"traits\":{\"email\":\"rhedricks@example.com\",\"logins\":2,\"name\":\"Richard Hendricks\"},\"sourceId\":\"sourceID\",\"sourceType\":\"sourceType\",\"destinationId\":\"destinationID\",\"destinationType\":\"POSTGRES\"},\"messageId\":\"messageId\",\"originalTimestamp\":\"2021-09-01T00:00:00.000Z\",\"properties\":{\"name\":\"Main\",\"title\":\"Home | RudderStack\",\"url\":\"https://www.rudderstack.com\"},\"receivedAt\":\"2021-09-01T00:00:00.000Z\",\"request_ip\":\"5.6.7.8\",\"sentAt\":\"2021-09-01T00:00:00.000Z\",\"timestamp\":\"2021-09-01T00:00:00.000Z\",\"type\":\"screen\",\"userId\":\"userId\"}").
								SetColumnField("rudder_event", "json"),
							Metadata:   getMetadata("screen", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "screen (Postgres) partial rules",
				eventPayload: `{"type":"screen","messageId":"messageId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","properties":{"name":"Main","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("screen", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: screenDefaultOutput().
								RemoveDataFields("anonymous_id", "channel", "context_request_ip").
								RemoveColumnFields("anonymous_id", "channel", "context_request_ip"),
							Metadata:   getMetadata("screen", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name: "screen (BQ) merge event",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"screen","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","properties":{"name":"Main","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("screen", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: screenDefaultOutput().
								SetDataField("context_destination_type", "BQ").
								SetColumnField("loaded_at", "datetime"),
							Metadata:   getMetadata("screen", "BQ"),
							StatusCode: http.StatusOK,
						},
						{
							Output:     screenMergeDefaultOutput(),
							Metadata:   getMetadata("screen", "BQ"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "screen (Snowflake)",
				eventPayload: `{"type":"screen","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","properties":{"name":"Main","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("screen", "SNOWFLAKE"),
				destination:  getDestination("SNOWFLAKE", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: screenDefaultOutput().
								SetDataField("context_destination_type", "SNOWFLAKE").
								BuildForSnowflake(),
							Metadata:   getMetadata("screen", "SNOWFLAKE"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},

			{
				name:         "group (Postgres)",
				eventPayload: `{"type":"group","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","groupId":"groupId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("group", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     groupDefaultOutput(),
							Metadata:   getMetadata("group", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "group (Postgres) without traits",
				eventPayload: `{"type":"group","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","groupId":"groupId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("group", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: groupDefaultOutput().
								RemoveDataFields("title", "url").
								RemoveColumnFields("title", "url"),
							Metadata:   getMetadata("group", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "group (Postgres) without context",
				eventPayload: `{"type":"group","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","groupId":"groupId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"}}`,
				metadata:     getMetadata("group", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: groupDefaultOutput().
								SetDataField("context_ip", "5.6.7.8"). // overriding the default value
								RemoveDataFields("context_passed_ip", "context_traits_email", "context_traits_logins").
								RemoveColumnFields("context_passed_ip", "context_traits_email", "context_traits_logins"),
							Metadata:   getMetadata("group", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "group (Postgres) store rudder event",
				eventPayload: `{"type":"group","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","groupId":"groupId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("group", "POSTGRES"),
				destination: getDestination("POSTGRES", map[string]any{
					"storeFullEvent": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: groupDefaultOutput().
								SetDataField("rudder_event", "{\"type\":\"group\",\"anonymousId\":\"anonymousId\",\"channel\":\"web\",\"context\":{\"destinationId\":\"destinationID\",\"destinationType\":\"POSTGRES\",\"ip\":\"1.2.3.4\",\"sourceId\":\"sourceID\",\"sourceType\":\"sourceType\",\"traits\":{\"email\":\"rhedricks@example.com\",\"logins\":2}},\"groupId\":\"groupId\",\"messageId\":\"messageId\",\"originalTimestamp\":\"2021-09-01T00:00:00.000Z\",\"receivedAt\":\"2021-09-01T00:00:00.000Z\",\"request_ip\":\"5.6.7.8\",\"sentAt\":\"2021-09-01T00:00:00.000Z\",\"timestamp\":\"2021-09-01T00:00:00.000Z\",\"traits\":{\"title\":\"Home | RudderStack\",\"url\":\"https://www.rudderstack.com\"},\"userId\":\"userId\"}").
								SetColumnField("rudder_event", "json"),
							Metadata:   getMetadata("group", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "group (Postgres) partial rules",
				eventPayload: `{"type":"group","messageId":"messageId","userId":"userId","groupId":"groupId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("group", "POSTGRES"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: groupDefaultOutput().
								RemoveDataFields("anonymous_id", "channel", "context_request_ip").
								RemoveColumnFields("anonymous_id", "channel", "context_request_ip"),
							Metadata:   getMetadata("group", "POSTGRES"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name: "group (BQ) merge event",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"group","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","groupId":"groupId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("group", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: groupDefaultOutput().
								SetDataField("context_destination_type", "BQ").
								SetColumnField("loaded_at", "datetime").
								SetTableName("_groups"),
							Metadata:   getMetadata("group", "BQ"),
							StatusCode: http.StatusOK,
						},
						{
							Output:     groupMergeDefaultOutput(),
							Metadata:   getMetadata("group", "BQ"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "group (Snowflake)",
				eventPayload: `{"type":"group","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","groupId":"groupId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("group", "SNOWFLAKE"),
				destination:  getDestination("SNOWFLAKE", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: groupDefaultOutput().
								SetDataField("context_destination_type", "SNOWFLAKE").
								BuildForSnowflake(),
							Metadata:   getMetadata("group", "SNOWFLAKE"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},

			{
				name:         "track (POSTGRES)",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     trackDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output:     trackEventDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) without properties",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     trackDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								RemoveDataFields("product_id", "review_id").
								RemoveColumnFields("product_id", "review_id"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) without userProperties",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     trackDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								RemoveDataFields("rating", "review_body").
								RemoveColumnFields("rating", "review_body"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) without context",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								SetDataField("context_ip", "5.6.7.8"). // overriding the default value
								RemoveDataFields("context_passed_ip", "context_traits_email", "context_traits_logins", "context_traits_name").
								RemoveColumnFields("context_passed_ip", "context_traits_email", "context_traits_logins", "context_traits_name"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("context_ip", "5.6.7.8"). // overriding the default value
								RemoveDataFields("context_passed_ip", "context_traits_email", "context_traits_logins", "context_traits_name").
								RemoveColumnFields("context_passed_ip", "context_traits_email", "context_traits_logins", "context_traits_name"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) RudderCreatedTable",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"accounts","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								SetDataField("event", "accounts").
								SetDataField("event_text", "accounts"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("event", "accounts").
								SetDataField("event_text", "accounts").
								SetTableName("_accounts"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) RudderCreatedTable with skipReservedKeywordsEscaping",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"accounts","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"},"integrations":{"POSTGRES":{"options":{"skipReservedKeywordsEscaping":true}}}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								SetDataField("event", "accounts").
								SetDataField("event_text", "accounts"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("event", "accounts").
								SetDataField("event_text", "accounts").
								SetTableName("accounts"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) RudderIsolatedTable",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"users","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								SetDataField("event", "users").
								SetDataField("event_text", "users"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("event", "users").
								SetDataField("event_text", "users").
								SetTableName("_users"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) empty event",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								SetDataField("event", "").
								RemoveDataFields("event_text").
								RemoveColumnFields("event_text"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) no event",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								SetDataField("event", "").
								RemoveDataFields("event_text").
								RemoveColumnFields("event_text"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) store rudder event",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination: getDestination("POSTGRES", map[string]any{
					"storeFullEvent": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								SetDataField("rudder_event", "{\"type\":\"track\",\"anonymousId\":\"anonymousId\",\"channel\":\"web\",\"context\":{\"destinationId\":\"destinationID\",\"destinationType\":\"POSTGRES\",\"ip\":\"1.2.3.4\",\"sourceId\":\"sourceID\",\"sourceType\":\"sourceType\",\"traits\":{\"email\":\"rhedricks@example.com\",\"logins\":2,\"name\":\"Richard Hendricks\"}},\"event\":\"event\",\"messageId\":\"messageId\",\"originalTimestamp\":\"2021-09-01T00:00:00.000Z\",\"properties\":{\"product_id\":\"9578257311\",\"review_id\":\"86ac1cd43\"},\"receivedAt\":\"2021-09-01T00:00:00.000Z\",\"request_ip\":\"5.6.7.8\",\"sentAt\":\"2021-09-01T00:00:00.000Z\",\"timestamp\":\"2021-09-01T00:00:00.000Z\",\"userId\":\"userId\",\"userProperties\":{\"rating\":3,\"review_body\":\"OK for the price. It works but the material feels flimsy.\"}}").
								SetColumnField("rudder_event", "json"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output:     trackEventDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) partial rules",
				eventPayload: `{"type":"track","messageId":"messageId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","event":"event","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								RemoveDataFields("anonymous_id", "channel", "context_request_ip").
								RemoveColumnFields("anonymous_id", "channel", "context_request_ip"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								RemoveDataFields("anonymous_id", "channel", "context_request_ip").
								RemoveColumnFields("anonymous_id", "channel", "context_request_ip"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) skipTracksTable (destOpts)",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination: getDestination("POSTGRES", map[string]any{
					"skipTracksTable": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     trackEventDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) skipTracksTable (intrOpts)",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"},"integrations":{"POSTGRES":{"options":{"skipTracksTable":true}}}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     trackEventDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) jsonPaths (escape characters for &, <, and >)",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"location": {"radiation_level": 0.000000006, "city":"Palo Alto \b <p>Ampersand: &;</p> Palo Alto","state":"California","country":"USA","coordinates":{"latitude":37.4419,"longitude":-122.143,"geo":{"altitude":30.5,"accuracy":5,"details":{"altitudeUnits":"meters","accuracyUnits":"meters"}}}},"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination: getDestination("POSTGRES", map[string]any{
					"jsonPaths": "location",
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     trackDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("location", "{\"city\":\"Palo Alto \\b <p>Ampersand: &;</p> Palo Alto\",\"coordinates\":{\"geo\":{\"accuracy\":5,\"altitude\":30.5,\"details\":{\"accuracyUnits\":\"meters\",\"altitudeUnits\":\"meters\"}},\"latitude\":37.4419,\"longitude\":-122.143},\"country\":\"USA\",\"radiation_level\":6e-9,\"state\":\"California\"}").
								SetColumnField("location", "json"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) jsonPaths (nested object level limits to 3 when source category is cloud with escape characters for &, <, and >)",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2,"location":{"coordinates":{"geo":{"description":"Palo Alto <p>Ampersand: &;</p> Palo Alto","altitude":30.5,"accuracy":5,"details":{"altitudeUnits":"meters","accuracyUnits":"meters"}}}}},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "cloud"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								SetDataField("context_traits_location_coordinates_geo", `{"accuracy":5,"altitude":30.5,"description":"Palo Alto <p>Ampersand: &;</p> Palo Alto","details":{"accuracyUnits":"meters","altitudeUnits":"meters"}}`).
								SetColumnField("context_traits_location_coordinates_geo", "string"),
							Metadata:   getTrackMetadata("POSTGRES", "cloud"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("context_traits_location_coordinates_geo", `{"accuracy":5,"altitude":30.5,"description":"Palo Alto <p>Ampersand: &;</p> Palo Alto","details":{"accuracyUnits":"meters","altitudeUnits":"meters"}}`).
								SetColumnField("context_traits_location_coordinates_geo", "string"),
							Metadata:   getTrackMetadata("POSTGRES", "cloud"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) store rudder event with escape characters for &, <, and >",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. <p>Ampersand: &;</p>. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination: getDestination("POSTGRES", map[string]any{
					"storeFullEvent": true,
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								SetDataField("rudder_event", "{\"type\":\"track\",\"anonymousId\":\"anonymousId\",\"channel\":\"web\",\"context\":{\"destinationId\":\"destinationID\",\"destinationType\":\"POSTGRES\",\"ip\":\"1.2.3.4\",\"sourceId\":\"sourceID\",\"sourceType\":\"sourceType\",\"traits\":{\"email\":\"rhedricks@example.com\",\"logins\":2,\"name\":\"Richard Hendricks\"}},\"event\":\"event\",\"messageId\":\"messageId\",\"originalTimestamp\":\"2021-09-01T00:00:00.000Z\",\"properties\":{\"product_id\":\"9578257311\",\"review_id\":\"86ac1cd43\"},\"receivedAt\":\"2021-09-01T00:00:00.000Z\",\"request_ip\":\"5.6.7.8\",\"sentAt\":\"2021-09-01T00:00:00.000Z\",\"timestamp\":\"2021-09-01T00:00:00.000Z\",\"userId\":\"userId\",\"userProperties\":{\"rating\":3,\"review_body\":\"OK for the price. <p>Ampersand: &;</p>. It works but the material feels flimsy.\"}}").
								SetColumnField("rudder_event", "json"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("review_body", "OK for the price. <p>Ampersand: &;</p>. It works but the material feels flimsy."),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) rules (anonymousId) being an object",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId": { "anonymousId": "anon-1234567890abcdef" },"userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								RemoveDataFields("anonymous_id").
								RemoveColumnFields("anonymous_id"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								RemoveDataFields("anonymous_id").
								RemoveColumnFields("anonymous_id"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) rules (anonymousId) being an array",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId": [ "anon-1234567890abcdef" ],"userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								RemoveDataFields("anonymous_id").
								RemoveColumnFields("anonymous_id"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								RemoveDataFields("anonymous_id").
								RemoveColumnFields("anonymous_id"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) jsonPaths (legacy destOpts for properties)",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"location": {"city":"Palo Alto","state":"California","country":"USA","coordinates":{"latitude":37.4419,"longitude":-122.143,"geo":{"altitude":30.5,"accuracy":5,"details":{"altitudeUnits":"meters","accuracyUnits":"meters"}}}},"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination: getDestination("POSTGRES", map[string]any{
					"jsonPaths": "location",
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     trackDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("location", "{\"city\":\"Palo Alto\",\"coordinates\":{\"geo\":{\"accuracy\":5,\"altitude\":30.5,\"details\":{\"accuracyUnits\":\"meters\",\"altitudeUnits\":\"meters\"}},\"latitude\":37.4419,\"longitude\":-122.143},\"country\":\"USA\",\"state\":\"California\"}").
								SetColumnField("location", "json"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) jsonPaths (legacy destOpts for user properties)",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"location": {"city":"Palo Alto","state":"California","country":"USA","coordinates":{"latitude":37.4419,"longitude":-122.143,"geo":{"altitude":30.5,"accuracy":5,"details":{"altitudeUnits":"meters","accuracyUnits":"meters"}}}}, "rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination: getDestination("POSTGRES", map[string]any{
					"jsonPaths": "location",
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     trackDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("location", "{\"city\":\"Palo Alto\",\"coordinates\":{\"geo\":{\"accuracy\":5,\"altitude\":30.5,\"details\":{\"accuracyUnits\":\"meters\",\"altitudeUnits\":\"meters\"}},\"latitude\":37.4419,\"longitude\":-122.143},\"country\":\"USA\",\"state\":\"California\"}").
								SetColumnField("location", "json"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) jsonPaths (destOpts)",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311", "location": {"city":"Palo Alto","state":"California","country":"USA","coordinates":{"latitude":37.4419,"longitude":-122.143,"geo":{"altitude":30.5,"accuracy":5,"details":{"altitudeUnits":"meters","accuracyUnits":"meters"}}}}},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination: getDestination("POSTGRES", map[string]any{
					"jsonPaths": "track.properties.location",
				}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     trackDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("location", "{\"city\":\"Palo Alto\",\"coordinates\":{\"geo\":{\"accuracy\":5,\"altitude\":30.5,\"details\":{\"accuracyUnits\":\"meters\",\"altitudeUnits\":\"meters\"}},\"latitude\":37.4419,\"longitude\":-122.143},\"country\":\"USA\",\"state\":\"California\"}").
								SetColumnField("location", "json"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) jsonPaths (intrOpts)",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311", "location": {"city":"Palo Alto","state":"California","country":"USA","coordinates":{"latitude":37.4419,"longitude":-122.143,"geo":{"altitude":30.5,"accuracy":5,"details":{"altitudeUnits":"meters","accuracyUnits":"meters"}}}}},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"},"integrations":{"POSTGRES":{"options":{"jsonPaths":["track.properties.location"]}}}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     trackDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("location", "{\"city\":\"Palo Alto\",\"coordinates\":{\"geo\":{\"accuracy\":5,\"altitude\":30.5,\"details\":{\"accuracyUnits\":\"meters\",\"altitudeUnits\":\"meters\"}},\"latitude\":37.4419,\"longitude\":-122.143},\"country\":\"USA\",\"state\":\"California\"}").
								SetColumnField("location", "json"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) jsonPaths (DATA_WAREHOUSE)",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311", "location": {"city":"Palo Alto","state":"California","country":"USA","coordinates":{"latitude":37.4419,"longitude":-122.143,"geo":{"altitude":30.5,"accuracy":5,"details":{"altitudeUnits":"meters","accuracyUnits":"meters"}}}}},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"},"integrations":{"DATA_WAREHOUSE":{"options":{"jsonPaths":["track.properties.location"]}}}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     trackDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("location", "{\"city\":\"Palo Alto\",\"coordinates\":{\"geo\":{\"accuracy\":5,\"altitude\":30.5,\"details\":{\"accuracyUnits\":\"meters\",\"altitudeUnits\":\"meters\"}},\"latitude\":37.4419,\"longitude\":-122.143},\"country\":\"USA\",\"state\":\"California\"}").
								SetColumnField("location", "json"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) jsonPaths (intrOpts with higher path)",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311", "location": {"city":"Palo Alto","state":"California","country":"USA","coordinates":{"latitude":37.4419,"longitude":-122.143,"geo":{"altitude":30.5,"accuracy":5,"details":{"altitudeUnits":"meters","accuracyUnits":"meters"}}}}},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"},"integrations":{"DATA_WAREHOUSE":{"options":{"jsonPaths":["track.properties.location"]}},"POSTGRES":{"options":{"jsonPaths":["track.properties.location.coordinates"]}}}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     trackDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("location", "{\"city\":\"Palo Alto\",\"coordinates\":{\"geo\":{\"accuracy\":5,\"altitude\":30.5,\"details\":{\"accuracyUnits\":\"meters\",\"altitudeUnits\":\"meters\"}},\"latitude\":37.4419,\"longitude\":-122.143},\"country\":\"USA\",\"state\":\"California\"}").
								SetColumnField("location", "json"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (POSTGRES) jsonPaths (DATA_WAREHOUSE with higher path)",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311", "location": {"city":"Palo Alto","state":"California","country":"USA","coordinates":{"latitude":37.4419,"longitude":-122.143,"geo":{"altitude":30.5,"accuracy":5,"details":{"altitudeUnits":"meters","accuracyUnits":"meters"}}}}},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"},"integrations":{"DATA_WAREHOUSE":{"options":{"jsonPaths":["track.properties.location.coordinates"]}},"POSTGRES":{"options":{"jsonPaths":["track.properties.location"]}}}}`,
				metadata:     getTrackMetadata("POSTGRES", "webhook"),
				destination:  getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output:     trackDefaultOutput(),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("location", "{\"city\":\"Palo Alto\",\"coordinates\":{\"geo\":{\"accuracy\":5,\"altitude\":30.5,\"details\":{\"accuracyUnits\":\"meters\",\"altitudeUnits\":\"meters\"}},\"latitude\":37.4419,\"longitude\":-122.143},\"country\":\"USA\",\"state\":\"California\"}").
								SetColumnField("location", "json"),
							Metadata:   getTrackMetadata("POSTGRES", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name: "track (BQ) merge event",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("BQ", "webhook"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								SetDataField("context_destination_type", "BQ").
								SetColumnField("loaded_at", "datetime"),
							Metadata:   getTrackMetadata("BQ", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("context_destination_type", "BQ").
								SetColumnField("loaded_at", "datetime"),
							Metadata:   getTrackMetadata("BQ", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output:     trackMergeDefaultOutput(),
							Metadata:   getTrackMetadata("BQ", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name: "track (Snowflake) with empty mergePropTwo",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("SNOWFLAKE", "webhook"),
				destination:  getDestination("SNOWFLAKE", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								SetDataField("context_destination_type", "SNOWFLAKE").
								BuildForSnowflake().
								RemoveDataFields("USER_ID").
								RemoveColumnFields("USER_ID"),
							Metadata:   getTrackMetadata("SNOWFLAKE", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("context_destination_type", "SNOWFLAKE").
								BuildForSnowflake().
								RemoveColumnFields("USER_ID").
								RemoveDataFields("USER_ID"),
							Metadata:   getTrackMetadata("SNOWFLAKE", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackMergeDefaultOutput().
								SetMetadata("mergePropTwo", "").
								BuildForSnowflake().
								RemoveDataFields("MERGE_PROPERTY_2_TYPE").RemoveColumnFields("MERGE_PROPERTY_2_TYPE").
								RemoveDataFields("MERGE_PROPERTY_2_VALUE").RemoveColumnFields("MERGE_PROPERTY_2_VALUE"),
							Metadata:   getTrackMetadata("SNOWFLAKE", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name: "track (Snowflake) with no mergePropTwo",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("SNOWFLAKE", "webhook"),
				destination:  getDestination("SNOWFLAKE", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								SetDataField("context_destination_type", "SNOWFLAKE").
								BuildForSnowflake().
								RemoveDataFields("USER_ID").
								RemoveColumnFields("USER_ID"),
							Metadata:   getTrackMetadata("SNOWFLAKE", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("context_destination_type", "SNOWFLAKE").
								BuildForSnowflake().
								RemoveColumnFields("USER_ID").
								RemoveDataFields("USER_ID"),
							Metadata:   getTrackMetadata("SNOWFLAKE", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackMergeDefaultOutput().
								RemoveMetadata("mergePropTwo").
								BuildForSnowflake().
								RemoveDataFields("MERGE_PROPERTY_2_TYPE").RemoveColumnFields("MERGE_PROPERTY_2_TYPE").
								RemoveDataFields("MERGE_PROPERTY_2_VALUE").RemoveColumnFields("MERGE_PROPERTY_2_VALUE"),
							Metadata:   getTrackMetadata("SNOWFLAKE", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name:         "track (Snowflake)",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getTrackMetadata("SNOWFLAKE", "webhook"),
				destination:  getDestination("SNOWFLAKE", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: trackDefaultOutput().
								SetDataField("context_destination_type", "SNOWFLAKE").
								BuildForSnowflake(),
							Metadata:   getTrackMetadata("SNOWFLAKE", "webhook"),
							StatusCode: http.StatusOK,
						},
						{
							Output: trackEventDefaultOutput().
								SetDataField("context_destination_type", "SNOWFLAKE").
								BuildForSnowflake(),
							Metadata:   getTrackMetadata("SNOWFLAKE", "webhook"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},

			{
				name: "merge (Postgres)",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload:     `{"type":"merge"}`,
				metadata:         getMetadata("merge", "POSTGRES"),
				destination:      getDestination("POSTGRES", map[string]any{}),
				expectedResponse: types.Response{},
			},
			{
				name: "merge (BQ)",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"merge","messageId":"messageId","receivedAt":"2021-09-01T00:00:00.000Z","mergeProperties":[{"type":"email","value":"alex@example.com"},{"type":"mobile","value":"+1-202-555-0146"}]}`,
				metadata:     getMetadata("merge", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: map[string]any{
								"data": map[string]any{
									"merge_property_1_type":  "email",
									"merge_property_1_value": "alex@example.com",
									"merge_property_2_type":  "mobile",
									"merge_property_2_value": "+1-202-555-0146",
								},
								"metadata": map[string]any{
									"table":        "rudder_identity_merge_rules",
									"columns":      map[string]any{"merge_property_1_type": "string", "merge_property_1_value": "string", "merge_property_2_type": "string", "merge_property_2_value": "string"},
									"isMergeRule":  true,
									"receivedAt":   "2021-09-01T00:00:00.000Z",
									"mergePropOne": "alex@example.com",
									"mergePropTwo": "+1-202-555-0146",
								},
								"userId": "",
							},
							Metadata:   getMetadata("merge", "BQ"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name: "merge (BQ) not enableIDResolution",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": false,
				},
				eventPayload:     `{"type":"merge"}`,
				metadata:         getMetadata("merge", "BQ"),
				destination:      getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{},
			},
			{
				name: "merge (BQ) missing mergeProperties",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"merge"}`,
				metadata:     getMetadata("merge", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					FailedEvents: []types.TransformerResponse{
						{
							Error:      response.ErrMergePropertiesMissing.Error(),
							StatusCode: response.ErrMergePropertiesMissing.StatusCode(),
							Metadata:   getMetadata("merge", "BQ"),
						},
					},
				},
			},
			{
				name: "merge (BQ) invalid mergeProperties",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"merge", "mergeProperties": "invalid"}`,
				metadata:     getMetadata("merge", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					FailedEvents: []types.TransformerResponse{
						{
							Error:      response.ErrMergePropertiesNotArray.Error(),
							StatusCode: response.ErrMergePropertiesNotArray.StatusCode(),
							Metadata:   getMetadata("merge", "BQ"),
						},
					},
				},
			},
			{
				name: "merge (BQ) empty mergeProperties",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"merge", "mergeProperties": []}`,
				metadata:     getMetadata("merge", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					FailedEvents: []types.TransformerResponse{
						{
							Error:      response.ErrMergePropertiesNotSufficient.Error(),
							StatusCode: response.ErrMergePropertiesNotSufficient.StatusCode(),
							Metadata:   getMetadata("merge", "BQ"),
						},
					},
				},
			},
			{
				name: "merge (BQ) single mergeProperties",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"merge","mergeProperties":[{"type":"email","value":"alex@example.com"}]}`,
				metadata:     getMetadata("merge", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					FailedEvents: []types.TransformerResponse{
						{
							Error:      response.ErrMergePropertiesNotSufficient.Error(),
							StatusCode: response.ErrMergePropertiesNotSufficient.StatusCode(),
							Metadata:   getMetadata("merge", "BQ"),
						},
					},
				},
			},
			{
				name: "merge (BQ) invalid merge property one",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"merge","mergeProperties":["invalid",{"type":"email","value":"alex@example.com"}]}`,
				metadata:     getMetadata("merge", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					FailedEvents: []types.TransformerResponse{
						{
							Error:      response.ErrMergePropertyOneInvalid.Error(),
							StatusCode: response.ErrMergePropertyOneInvalid.StatusCode(),
							Metadata:   getMetadata("merge", "BQ"),
						},
					},
				},
			},
			{
				name: "merge (BQ) invalid merge property two",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"merge","mergeProperties":[{"type":"email","value":"alex@example.com"},"invalid"]}`,
				metadata:     getMetadata("merge", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					FailedEvents: []types.TransformerResponse{
						{
							Error:      response.ErrMergePropertyTwoInvalid.Error(),
							StatusCode: response.ErrMergePropertyTwoInvalid.StatusCode(),
							Metadata:   getMetadata("merge", "BQ"),
						},
					},
				},
			},
			{
				name: "merge (BQ) missing mergeProperty",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"merge","mergeProperties":[{"type1":"email","value1":"alex@example.com"},{"type1":"mobile","value1":"+1-202-555-0146"}]}`,
				metadata:     getMetadata("merge", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					FailedEvents: []types.TransformerResponse{
						{
							Error:      response.ErrMergePropertyEmpty.Error(),
							StatusCode: response.ErrMergePropertyEmpty.StatusCode(),
							Metadata:   getMetadata("merge", "BQ"),
						},
					},
				},
			},
			{
				name: "merge (SNOWFLAKE)",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"merge","messageId":"messageId","receivedAt":"2021-09-01T00:00:00.000Z","mergeProperties":[{"type":"email","value":"alex@example.com"},{"type":"mobile","value":"+1-202-555-0146"}]}`,
				metadata:     getMetadata("merge", "SNOWFLAKE"),
				destination:  getDestination("SNOWFLAKE", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: map[string]any{
								"data": map[string]any{
									"MERGE_PROPERTY_1_TYPE":  "email",
									"MERGE_PROPERTY_1_VALUE": "alex@example.com",
									"MERGE_PROPERTY_2_TYPE":  "mobile",
									"MERGE_PROPERTY_2_VALUE": "+1-202-555-0146",
								},
								"metadata": map[string]any{
									"table":        "RUDDER_IDENTITY_MERGE_RULES",
									"columns":      map[string]any{"MERGE_PROPERTY_1_TYPE": "string", "MERGE_PROPERTY_1_VALUE": "string", "MERGE_PROPERTY_2_TYPE": "string", "MERGE_PROPERTY_2_VALUE": "string"},
									"isMergeRule":  true,
									"receivedAt":   "2021-09-01T00:00:00.000Z",
									"mergePropOne": "alex@example.com",
									"mergePropTwo": "+1-202-555-0146",
								},
								"userId": "",
							},
							Metadata:   getMetadata("merge", "SNOWFLAKE"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name: "alias (BQ)",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"alias","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","previousId":"previousId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("alias", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: aliasDefaultOutput().
								SetDataField("context_destination_type", "BQ").
								SetColumnField("loaded_at", "datetime"),
							Metadata:   getMetadata("alias", "BQ"),
							StatusCode: http.StatusOK,
						},
						{
							Output:     aliasMergeDefaultOutput(),
							Metadata:   getMetadata("alias", "BQ"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name: "alias (BQ) no userId and previousId",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"alias","messageId":"messageId","anonymousId":"anonymousId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("alias", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: aliasDefaultOutput().
								SetDataField("context_destination_type", "BQ").
								SetColumnField("loaded_at", "datetime").
								RemoveDataFields("user_id", "previous_id").
								RemoveColumnFields("user_id", "previous_id"),
							Metadata:   getMetadata("alias", "BQ"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name: "alias (BQ) empty userId and previousId",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"alias","messageId":"messageId","anonymousId":"anonymousId","userId":"","previousId":"","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","traits":{"title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("alias", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: aliasDefaultOutput().
								SetDataField("context_destination_type", "BQ").
								SetColumnField("loaded_at", "datetime").
								RemoveDataFields("user_id", "previous_id").
								RemoveColumnFields("user_id", "previous_id"),
							Metadata:   getMetadata("alias", "BQ"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name: "page (BQ)",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"page","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("page", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: pageDefaultOutput().
								SetDataField("context_destination_type", "BQ").
								SetColumnField("loaded_at", "datetime"),
							Metadata:   getMetadata("page", "BQ"),
							StatusCode: http.StatusOK,
						},
						{
							Output:     pageMergeDefaultOutput(),
							Metadata:   getMetadata("page", "BQ"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
			{
				name: "page (BQ) no anonymousID",
				configOverride: map[string]any{
					"Warehouse.enableIDResolution": true,
				},
				eventPayload: `{"type":"page","messageId":"messageId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","request_ip":"5.6.7.8","properties":{"name":"Home","title":"Home | RudderStack","url":"https://www.rudderstack.com"},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     getMetadata("page", "BQ"),
				destination:  getDestination("BQ", map[string]any{}),
				expectedResponse: types.Response{
					Events: []types.TransformerResponse{
						{
							Output: pageDefaultOutput().
								SetDataField("context_destination_type", "BQ").
								SetColumnField("loaded_at", "datetime").
								RemoveDataFields("anonymous_id").
								RemoveColumnFields("anonymous_id"),
							Metadata:   getMetadata("page", "BQ"),
							StatusCode: http.StatusOK,
						},
						{
							Output: map[string]any{
								"data": map[string]any{
									"merge_property_1_type":  "user_id",
									"merge_property_1_value": "userId",
								},
								"metadata": map[string]any{
									"table":        "rudder_identity_merge_rules",
									"columns":      map[string]any{"merge_property_1_type": "string", "merge_property_1_value": "string"},
									"isMergeRule":  true,
									"receivedAt":   "2021-09-01T00:00:00.000Z",
									"mergePropOne": "userId",
								},
								"userId": "",
							},
							Metadata:   getMetadata("page", "BQ"),
							StatusCode: http.StatusOK,
						},
					},
				},
			},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				c := setupConfig(transformerResource, tc.configOverride)

				processorTransformer := destination_transformer.New(c, logger.NOP, stats.Default)
				warehouseTransformer := warehouse.New(c, logger.NOP, stats.NOP)

				var singularEvent types.SingularEventT
				err := jsonrs.Unmarshal([]byte(tc.eventPayload), &singularEvent)
				require.NoError(t, err)

				ctx := context.Background()
				events := []types.TransformerEvent{{Message: singularEvent, Metadata: tc.metadata, Destination: tc.destination}}

				legacyResponse := processorTransformer.Transform(ctx, events)
				embeddedResponse := warehouseTransformer.Transform(ctx, events)
				testhelper.ValidateExpectedEvents(t, tc.expectedResponse, embeddedResponse, legacyResponse)
			})
		}
	})

	t.Run("Mandatory fields", func(t *testing.T) {
		now := time.Date(2023, 1, 1, 1, 0, 0, 0, time.UTC)
		uid := uuid.NewString()

		testCases := []struct {
			name           string
			eventPayload   string
			metadata       types.Metadata
			destination    backendconfig.DestinationT
			verifyResponse func(t *testing.T, resp types.TransformerResponse)
		}{
			{
				name:         "messageId and receivedAt both are present",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     types.Metadata{EventType: "track", DestinationType: "POSTGRES", MessageID: "messageId", ReceivedAt: "2021-09-01T00:00:00.000Z"},
				destination:  getDestination("POSTGRES", map[string]any{}),
				verifyResponse: func(t *testing.T, resp types.TransformerResponse) {
					require.Equal(t, "messageId", misc.MapLookup(resp.Output, "data", "id"))
					require.Equal(t, "2021-09-01T00:00:00.000Z", misc.MapLookup(resp.Output, "data", "received_at"))
					require.Equal(t, "messageId", resp.Metadata.MessageID)
					require.Equal(t, "2021-09-01T00:00:00.000Z", resp.Metadata.ReceivedAt)
					require.Equal(t, "2021-09-01T00:00:00.000Z", misc.MapLookup(resp.Output, "metadata", "receivedAt"))
				},
			},
			{
				name:         "messageId and receivedAt both are not present",
				eventPayload: `{"type":"track","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     types.Metadata{EventType: "track", DestinationType: "POSTGRES"},
				destination:  getDestination("POSTGRES", map[string]any{}),
				verifyResponse: func(t *testing.T, resp types.TransformerResponse) {
					require.Equal(t, "auto-"+uid, misc.MapLookup(resp.Output, "data", "id"))
					require.Equal(t, now.Format(misc.RFC3339Milli), misc.MapLookup(resp.Output, "data", "received_at"))
					require.Empty(t, resp.Metadata.MessageID)
					require.Empty(t, resp.Metadata.ReceivedAt)
					require.Equal(t, now.Format(misc.RFC3339Milli), misc.MapLookup(resp.Output, "metadata", "receivedAt"))
				},
			},
			{
				name:         "messageId and receivedAt both are present and empty",
				eventPayload: `{"type":"track","messageId":"","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     types.Metadata{EventType: "track", DestinationType: "POSTGRES"},
				destination:  getDestination("POSTGRES", map[string]any{}),
				verifyResponse: func(t *testing.T, resp types.TransformerResponse) {
					require.Equal(t, "auto-"+uid, misc.MapLookup(resp.Output, "data", "id"))
					require.Equal(t, now.Format(misc.RFC3339Milli), misc.MapLookup(resp.Output, "data", "received_at"))
					require.Empty(t, resp.Metadata.MessageID)
					require.Empty(t, resp.Metadata.ReceivedAt)
					require.Equal(t, now.Format(misc.RFC3339Milli), misc.MapLookup(resp.Output, "metadata", "receivedAt"))
				},
			},
			{
				name:         "messageId and receivedAt both are present and null",
				eventPayload: `{"type":"track","messageId":null,"anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":null,"originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     types.Metadata{EventType: "track", DestinationType: "POSTGRES"},
				destination:  getDestination("POSTGRES", map[string]any{}),
				verifyResponse: func(t *testing.T, resp types.TransformerResponse) {
					require.Equal(t, "auto-"+uid, misc.MapLookup(resp.Output, "data", "id"))
					require.Equal(t, now.Format(misc.RFC3339Milli), misc.MapLookup(resp.Output, "data", "received_at"))
					require.Empty(t, resp.Metadata.MessageID)
					require.Empty(t, resp.Metadata.ReceivedAt)
					require.Equal(t, now.Format(misc.RFC3339Milli), misc.MapLookup(resp.Output, "metadata", "receivedAt"))
				},
			},
			{
				name:         "messageId different in event and metadata",
				eventPayload: `{"type":"track","messageId":"messageId-event","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2021-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     types.Metadata{EventType: "track", DestinationType: "POSTGRES", MessageID: "messageId-metadata", ReceivedAt: "2021-09-01T00:00:00.000Z"},
				destination:  getDestination("POSTGRES", map[string]any{}),
				verifyResponse: func(t *testing.T, resp types.TransformerResponse) {
					require.Equal(t, "messageId-event", misc.MapLookup(resp.Output, "data", "id"))
					require.Equal(t, "2021-09-01T00:00:00.000Z", misc.MapLookup(resp.Output, "data", "received_at"))
					require.Equal(t, "messageId-metadata", resp.Metadata.MessageID)
					require.Equal(t, "2021-09-01T00:00:00.000Z", resp.Metadata.ReceivedAt)
					require.Equal(t, "2021-09-01T00:00:00.000Z", misc.MapLookup(resp.Output, "metadata", "receivedAt"))
				},
			},
			{
				name:         "receivedAt different in event and metadata",
				eventPayload: `{"type":"track","messageId":"messageId","anonymousId":"anonymousId","userId":"userId","sentAt":"2021-09-01T00:00:00.000Z","timestamp":"2021-09-01T00:00:00.000Z","receivedAt":"2022-09-01T00:00:00.000Z","originalTimestamp":"2021-09-01T00:00:00.000Z","channel":"web","event":"event","request_ip":"5.6.7.8","properties":{"review_id":"86ac1cd43","product_id":"9578257311"},"userProperties":{"rating":3.0,"review_body":"OK for the price. It works but the material feels flimsy."},"context":{"traits":{"name":"Richard Hendricks","email":"rhedricks@example.com","logins":2},"ip":"1.2.3.4"}}`,
				metadata:     types.Metadata{EventType: "track", DestinationType: "POSTGRES", MessageID: "messageId", ReceivedAt: "2023-09-01T00:00:00.000Z"},
				destination:  getDestination("POSTGRES", map[string]any{}),
				verifyResponse: func(t *testing.T, resp types.TransformerResponse) {
					require.Equal(t, "messageId", misc.MapLookup(resp.Output, "data", "id"))
					require.Equal(t, "2022-09-01T00:00:00.000Z", misc.MapLookup(resp.Output, "data", "received_at"))
					require.Equal(t, "messageId", resp.Metadata.MessageID)
					require.Equal(t, "2023-09-01T00:00:00.000Z", resp.Metadata.ReceivedAt)
					require.Equal(t, "2022-09-01T00:00:00.000Z", misc.MapLookup(resp.Output, "metadata", "receivedAt"))
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				c := setupConfig(transformerResource, map[string]any{})

				processorTransformer := destination_transformer.New(c, logger.NOP, stats.Default)
				warehouseTransformer := warehouse.New(c, logger.NOP, stats.NOP,
					warehouse.WithNow(func() time.Time {
						return now
					}),
					warehouse.WithUUIDGenerator(func() string {
						return uid
					}),
				)

				var singularEvent types.SingularEventT
				err := jsonrs.Unmarshal([]byte(tc.eventPayload), &singularEvent)
				require.NoError(t, err)

				ctx := context.Background()
				events := []types.TransformerEvent{{Message: singularEvent, Metadata: tc.metadata, Destination: tc.destination}}

				legacyResponse := processorTransformer.Transform(ctx, events)
				embeddedResponse := warehouseTransformer.Transform(ctx, events)

				nonEmptyFields := []string{"data.id", "data.received_at", "metadata.receivedAt"}

				require.Equal(t, len(embeddedResponse.Events), len(legacyResponse.Events))
				require.Nil(t, embeddedResponse.FailedEvents)
				require.Nil(t, legacyResponse.FailedEvents)
				for i := range embeddedResponse.Events {
					for _, field := range nonEmptyFields {
						require.NotEmpty(t, misc.MapLookup(embeddedResponse.Events[i].Output, strings.Split(field, ".")...))
						require.NotEmpty(t, misc.MapLookup(legacyResponse.Events[i].Output, strings.Split(field, ".")...))
					}
					tc.verifyResponse(t, embeddedResponse.Events[i])
				}
			})
		}
	})

	t.Run("Tracking Plan", func(t *testing.T) {
		message := map[string]any{
			"context": map[string]any{
				"trackingPlanId":      "tp_2jap9a9T5SOjEx45gsy76jiTu5q",
				"trackingPlanVersion": 8,
				"violationErrors": []types.ValidationError{
					{
						Type:    "Required-Missing",
						Message: "must have required property 'name'",
						Meta: map[string]string{
							"instancePath": "/traits",
							"schemaPath":   "#/properties/traits/required",
						},
						Property: "name",
					},
					{
						Type:    "Datatype-Mismatch",
						Message: "must be string",
						Meta: map[string]string{
							"instancePath": "/traits/email",
							"schemaPath":   "#/properties/traits/properties/email/type",
						},
						Property: "",
					},
				},
			},
			"messageId":         "messageId",
			"originalTimestamp": "2021-09-01T00:00:00.000Z",
			"receivedAt":        "2021-09-01T00:00:00.000Z",
			"sentAt":            "2021-09-01T00:00:00.000Z",
			"timestamp":         "2021-09-01T00:00:00.000Z",
			"traits": map[string]any{
				"email": float64(12),
			},
			"type": "track",
		}
		for destination := range whutils.PseudoWarehouseDestinationMap {
			t.Run(destination, func(t *testing.T) {
				c := setupConfig(transformerResource, map[string]any{})

				processorTransformer := destination_transformer.New(c, logger.NOP, stats.Default)
				warehouseTransformer := warehouse.New(c, logger.NOP, stats.NOP)

				ctx := context.Background()
				events := []types.TransformerEvent{{
					Message:     message,
					Metadata:    getMetadata("track", destination),
					Destination: getDestination(destination, map[string]any{}),
				}}
				legacyResponse := processorTransformer.Transform(ctx, events)
				embeddedResponse := warehouseTransformer.Transform(ctx, events)
				testhelper.ValidateEvents(t, embeddedResponse, legacyResponse)
			})
		}
	})

	t.Run("Enrichment", func(t *testing.T) {
		t.Run("geo (from processor)", func(t *testing.T) {
			message := map[string]any{
				"context": map[string]any{
					"geo": enricher.Geolocation{
						IP:       "192.168.1.42",
						City:     "San Francisco",
						Country:  "US",
						Region:   "CA",
						Postal:   "94107",
						Location: "37.7749,-122.4194",
						Timezone: "America/Los_Angeles",
					},
				},
				"messageId":         "messageId",
				"originalTimestamp": "2021-09-01T00:00:00.000Z",
				"receivedAt":        "2021-09-01T00:00:00.000Z",
				"sentAt":            "2021-09-01T00:00:00.000Z",
				"timestamp":         "2021-09-01T00:00:00.000Z",
				"type":              "track",
			}
			for destination := range whutils.PseudoWarehouseDestinationMap {
				t.Run(destination, func(t *testing.T) {
					c := setupConfig(transformerResource, map[string]any{})

					processorTransformer := destination_transformer.New(c, logger.NOP, stats.Default)
					warehouseTransformer := warehouse.New(c, logger.NOP, stats.NOP)

					ctx := context.Background()
					events := []types.TransformerEvent{{
						Message:     message,
						Metadata:    getMetadata("track", destination),
						Destination: getDestination(destination, map[string]any{}),
					}}
					legacyResponse := processorTransformer.Transform(ctx, events)
					embeddedResponse := warehouseTransformer.Transform(ctx, events)
					testhelper.ValidateEvents(t, embeddedResponse, legacyResponse)
				})
			}
		})
		t.Run("geo (from UT)", func(t *testing.T) {
			message := map[string]any{
				"context": map[string]any{
					"geo": "{\"city\":\"Madison\",\"country\":\"US\",\"ip\":\"104.176.44.185\",\"loc\":\"43.0371,-89.3932\",\"postal\":\"53713\",\"region\":\"Wisconsin\",\"timezone\":\"America/Chicago\"}",
				},
				"messageId":         "messageId",
				"originalTimestamp": "2021-09-01T00:00:00.000Z",
				"receivedAt":        "2021-09-01T00:00:00.000Z",
				"sentAt":            "2021-09-01T00:00:00.000Z",
				"timestamp":         "2021-09-01T00:00:00.000Z",
				"type":              "track",
			}
			for destination := range whutils.PseudoWarehouseDestinationMap {
				t.Run(destination, func(t *testing.T) {
					c := setupConfig(transformerResource, map[string]any{})

					processorTransformer := destination_transformer.New(c, logger.NOP, stats.Default)
					warehouseTransformer := warehouse.New(c, logger.NOP, stats.NOP)

					ctx := context.Background()
					events := []types.TransformerEvent{{
						Message:     message,
						Metadata:    getMetadata("track", destination),
						Destination: getDestination(destination, map[string]any{}),
					}}
					legacyResponse := processorTransformer.Transform(ctx, events)
					embeddedResponse := warehouseTransformer.Transform(ctx, events)
					testhelper.ValidateEvents(t, embeddedResponse, legacyResponse)
				})
			}
		})
		t.Run("geo (from UT) for safe key", func(t *testing.T) {
			message := map[string]any{
				"context": map[string]any{
					"geo	": "{\"city\":\"Madison\",\"country\":\"US\",\"ip\":\"104.176.44.185\",\"loc\":\"43.0371,-89.3932\",\"postal\":\"53713\",\"region\":\"Wisconsin\",\"timezone\":\"America/Chicago\"}",
				},
				"messageId":         "messageId",
				"originalTimestamp": "2021-09-01T00:00:00.000Z",
				"receivedAt":        "2021-09-01T00:00:00.000Z",
				"sentAt":            "2021-09-01T00:00:00.000Z",
				"timestamp":         "2021-09-01T00:00:00.000Z",
				"type":              "track",
			}
			for destination := range whutils.PseudoWarehouseDestinationMap {
				t.Run(destination, func(t *testing.T) {
					c := setupConfig(transformerResource, map[string]any{})

					processorTransformer := destination_transformer.New(c, logger.NOP, stats.Default)
					warehouseTransformer := warehouse.New(c, logger.NOP, stats.NOP)

					ctx := context.Background()
					events := []types.TransformerEvent{{
						Message:     message,
						Metadata:    getMetadata("track", destination),
						Destination: getDestination(destination, map[string]any{}),
					}}
					legacyResponse := processorTransformer.Transform(ctx, events)
					embeddedResponse := warehouseTransformer.Transform(ctx, events)
					testhelper.ValidateEvents(t, embeddedResponse, legacyResponse)
				})
			}
		})
		t.Run("bot (from processor)", func(t *testing.T) {
			type botDetails struct {
				Name             string `json:"name,omitempty"`
				URL              string `json:"url,omitempty"`
				IsInvalidBrowser bool   `json:"isInvalidBrowser,omitempty"`
			}

			message := map[string]any{
				"context": map[string]any{
					"isBot": true,
					"bot": botDetails{
						Name:             "ExampleBot",
						URL:              "https://example.com/bot",
						IsInvalidBrowser: true,
					},
					"botPtr": &botDetails{
						Name:             "ExampleBot",
						URL:              "https://example.com/bot",
						IsInvalidBrowser: true,
					},
				},
				"messageId":         "messageId",
				"originalTimestamp": "2021-09-01T00:00:00.000Z",
				"receivedAt":        "2021-09-01T00:00:00.000Z",
				"sentAt":            "2021-09-01T00:00:00.000Z",
				"timestamp":         "2021-09-01T00:00:00.000Z",
				"type":              "track",
			}
			for destination := range whutils.PseudoWarehouseDestinationMap {
				t.Run(destination, func(t *testing.T) {
					c := setupConfig(transformerResource, map[string]any{})

					processorTransformer := destination_transformer.New(c, logger.NOP, stats.Default)
					warehouseTransformer := warehouse.New(c, logger.NOP, stats.NOP)

					ctx := context.Background()
					events := []types.TransformerEvent{{
						Message:     message,
						Metadata:    getMetadata("track", destination),
						Destination: getDestination(destination, map[string]any{}),
					}}
					legacyResponse := processorTransformer.Transform(ctx, events)
					embeddedResponse := warehouseTransformer.Transform(ctx, events)
					testhelper.ValidateEvents(t, embeddedResponse, legacyResponse)
				})
			}
		})
		t.Run("pointers", func(t *testing.T) {
			message := map[string]any{
				"context": map[string]any{
					"boolVal": true,
					"boolPtr": lo.ToPtr(true),

					"intVal":   int(1),
					"intPtr":   lo.ToPtr(int(1)),
					"int8Val":  int8(2),
					"int8Ptr":  lo.ToPtr(int8(2)),
					"int16Val": int16(3),
					"int16Ptr": lo.ToPtr(int16(3)),
					"int32Val": int32(4),
					"int32Ptr": lo.ToPtr(int32(4)),
					"int64Val": int64(5),
					"int64Ptr": lo.ToPtr(int64(5)),

					"uintVal":    uint(6),
					"uintPtr":    lo.ToPtr(uint(6)),
					"uint8Val":   uint8(7),
					"uint8Ptr":   lo.ToPtr(uint8(7)),
					"uint16Val":  uint16(8),
					"uint16Ptr":  lo.ToPtr(uint16(8)),
					"uint32Val":  uint32(9),
					"uint32Ptr":  lo.ToPtr(uint32(9)),
					"uint64Val":  uint64(10),
					"uint64Ptr":  lo.ToPtr(uint64(10)),
					"uintptrVal": uintptr(12345),
					"uintptrPtr": lo.ToPtr(uintptr(12345)),

					"float32Val": float32(1.23),
					"float32Ptr": lo.ToPtr(float32(1.23)),
					"float64Val": float64(4.56),
					"float64Ptr": lo.ToPtr(float64(4.56)),

					"stringVal": "Hello",
					"stringPtr": lo.ToPtr("Hello"),

					"sliceVal": []any{1, 2, 3},
					"slicePtr": &[]any{1, 2, 3},

					"mapVal": map[string]any{"a": 1},
					"mapPtr": &map[string]any{"a": 1},
				},
				"messageId":         "messageId",
				"originalTimestamp": "2021-09-01T00:00:00.000Z",
				"receivedAt":        "2021-09-01T00:00:00.000Z",
				"sentAt":            "2021-09-01T00:00:00.000Z",
				"timestamp":         "2021-09-01T00:00:00.000Z",
				"type":              "track",
			}
			for destination := range whutils.PseudoWarehouseDestinationMap {
				t.Run(destination, func(t *testing.T) {
					c := setupConfig(transformerResource, map[string]any{})

					processorTransformer := destination_transformer.New(c, logger.NOP, stats.Default)
					warehouseTransformer := warehouse.New(c, logger.NOP, stats.NOP)

					ctx := context.Background()
					events := []types.TransformerEvent{{
						Message:     message,
						Metadata:    getMetadata("track", destination),
						Destination: getDestination(destination, map[string]any{}),
					}}
					legacyResponse := processorTransformer.Transform(ctx, events)
					embeddedResponse := warehouseTransformer.Transform(ctx, events)
					testhelper.ValidateEvents(t, embeddedResponse, legacyResponse)
				})
			}
		})
	})

	t.Run("Multiple fields for the same key", func(t *testing.T) {
		message := map[string]any{
			"context": map[string]any{
				"abC":  "1",
				"ab_C": "2",
				"ab_c": "3",
			},
			"messageId":         "messageId",
			"originalTimestamp": "2021-09-01T00:00:00.000Z",
			"receivedAt":        "2021-09-01T00:00:00.000Z",
			"sentAt":            "2021-09-01T00:00:00.000Z",
			"timestamp":         "2021-09-01T00:00:00.000Z",
			"type":              "identify",
		}
		c := setupConfig(transformerResource, map[string]any{})

		ctx := context.Background()
		events := []types.TransformerEvent{{
			Message:     message,
			Metadata:    getMetadata("identify", "POSTGRES"),
			Destination: getDestination("POSTGRES", map[string]any{}),
		}}

		t.Run("Reverse", func(t *testing.T) {
			processorTransformer := destination_transformer.New(c, logger.NOP, stats.Default)
			warehouseTransformer := warehouse.New(c, logger.NOP, stats.NOP, warehouse.WithSorter(func(i []string) []string {
				sort.Strings(i)
				slices.Reverse(i)
				return i
			}))

			legacyResponse := processorTransformer.Transform(ctx, events)
			embeddedResponse := warehouseTransformer.Transform(ctx, events)

			require.Equal(t, len(embeddedResponse.Events), len(legacyResponse.Events))
			require.Nil(t, legacyResponse.FailedEvents)
			require.Nil(t, embeddedResponse.FailedEvents)
			for i := range legacyResponse.Events {
				require.Equal(t, "3", misc.MapLookup(legacyResponse.Events[i].Output, "data", "context_ab_c"))
				require.Equal(t, "1", misc.MapLookup(embeddedResponse.Events[i].Output, "data", "context_ab_c"))
				delete(legacyResponse.Events[i].Output["data"].(map[string]any), "context_ab_c")
				delete(embeddedResponse.Events[i].Output["data"].(map[string]any), "context_ab_c")
				require.EqualValues(t, embeddedResponse.Events[i], legacyResponse.Events[i])
			}
		})
		t.Run("Sorted", func(t *testing.T) {
			processorTransformer := destination_transformer.New(c, logger.NOP, stats.Default)
			warehouseTransformer := warehouse.New(c, logger.NOP, stats.NOP)

			legacyResponse := processorTransformer.Transform(ctx, events)
			embeddedResponse := warehouseTransformer.Transform(ctx, events)
			testhelper.ValidateEvents(t, embeddedResponse, legacyResponse)
		})
	})
}
