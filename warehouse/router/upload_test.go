package router

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/filemanager"
	"github.com/rudderlabs/rudder-go-kit/filemanager/mock_filemanager"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/memstats"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/postgres"

	backendconfig "github.com/rudderlabs/rudder-server/backend-config"
	"github.com/rudderlabs/rudder-server/services/alerta"
	migrator "github.com/rudderlabs/rudder-server/services/sql-migrator"
	"github.com/rudderlabs/rudder-server/warehouse/integrations/manager"
	sqlmiddleware "github.com/rudderlabs/rudder-server/warehouse/integrations/middleware/sqlquerywrapper"
	"github.com/rudderlabs/rudder-server/warehouse/integrations/redshift"
	"github.com/rudderlabs/rudder-server/warehouse/integrations/snowflake"
	"github.com/rudderlabs/rudder-server/warehouse/internal/model"
	"github.com/rudderlabs/rudder-server/warehouse/internal/repo"
	mockupload "github.com/rudderlabs/rudder-server/warehouse/router/mocks"
	"github.com/rudderlabs/rudder-server/warehouse/schema"
	warehouseutils "github.com/rudderlabs/rudder-server/warehouse/utils"
)

type mockSchemaRepo struct{}

func (m *mockSchemaRepo) GetForNamespace(context.Context, string, string) (model.WHSchema, error) {
	return model.WHSchema{}, nil
}

func (m *mockSchemaRepo) Insert(context.Context, *model.WHSchema) error {
	return nil
}

type mockFetchSchemaRepo struct{}

func (m *mockFetchSchemaRepo) FetchSchema(context.Context) (model.Schema, error) {
	return model.Schema{}, nil
}

func TestExtractUploadErrorsByState(t *testing.T) {
	input := []struct {
		InitialErrorState []byte
		CurrentErrorState string
		CurrentError      error
		ErrorCount        int
	}{
		{
			InitialErrorState: []byte(`{}`),
			CurrentErrorState: InternalProcessingFailed,
			CurrentError:      errors.New("account locked"),
			ErrorCount:        1,
		},
		{
			InitialErrorState: []byte(`{"internal_processing_failed": {"errors": ["account locked"], "attempt": 1}}`),
			CurrentErrorState: InternalProcessingFailed,
			CurrentError:      errors.New("account locked again"),
			ErrorCount:        2,
		},
		{
			InitialErrorState: []byte(`{"internal_processing_failed": {"errors": ["account locked", "account locked again"], "attempt": 2}}`),
			CurrentErrorState: model.TableUploadExportingFailed,
			CurrentError:      errors.New("failed to load data because failed in earlier job"),
			ErrorCount:        1,
		},
	}

	for _, ip := range input {

		uploadErrors, err := extractAndUpdateUploadErrorsByState(ip.InitialErrorState, ip.CurrentErrorState, ip.CurrentError)
		if err != nil {
			t.Errorf("extracting upload errors by state should have passed: %v", err)
		}

		stateErrors := uploadErrors[ip.CurrentErrorState]
		// Below switch clause mirrors how we are
		// adding data in generic interface.

		var errorLength int
		switch stateErrors["errors"].(type) {
		case []string:
			errorLength = len(stateErrors["errors"].([]string))
		case []interface{}:
			errorLength = len(stateErrors["errors"].([]interface{}))
		}

		if errorLength != ip.ErrorCount {
			t.Errorf("expected error to be addded to list of state errors")
		}

		if stateErrors["attempt"].(int) != ip.ErrorCount {
			t.Errorf("expected attempts to be: %d, got: %d", ip.ErrorCount, stateErrors["attempt"].(int))
		}
	}
}

func TestColumnCountStat(t *testing.T) {
	var (
		workspaceID     = "test-workspaceID"
		destinationID   = "test-destinationID"
		destinationName = "test-destinationName"
		sourceID        = "test-sourceID"
		sourceName      = "test-sourceName"
		tableName       = "test-table"
	)

	inputs := []struct {
		name             string
		columnCountLimit int
		destinationType  string
		statExpected     bool
	}{
		{
			name:            "Datalakes destination",
			destinationType: warehouseutils.S3Datalake,
		},
		{
			name:            "Unknown destination",
			destinationType: "unknown-destination",
		},
		{
			name:             "Greater than threshold",
			destinationType:  warehouseutils.RS,
			columnCountLimit: 1,
			statExpected:     true,
		},
		{
			name:             "Lesser than threshold",
			destinationType:  warehouseutils.RS,
			columnCountLimit: 10,
			statExpected:     true,
		},
	}

	for _, tc := range inputs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			statsStore, err := memstats.New()
			require.NoError(t, err)

			conf := config.New()
			conf.Set(fmt.Sprintf("Warehouse.%s.columnCountLimit", strings.ToLower(warehouseutils.WHDestNameMap[tc.destinationType])), tc.columnCountLimit)

			pool, err := dockertest.NewPool("")
			require.NoError(t, err)

			pgResource, err := postgres.Setup(pool, t)
			require.NoError(t, err)
			db := sqlmiddleware.New(pgResource.DB)
			err = (&migrator.Migrator{
				Handle:          pgResource.DB,
				MigrationsTable: "wh_schema_migrations",
			}).Migrate("warehouse")
			require.NoError(t, err)

			uploadJobFactory := &UploadJobFactory{
				logger:       logger.NOP,
				statsFactory: statsStore,
				conf:         conf,
				db:           db,
			}
			whManager, err := manager.New(warehouseutils.POSTGRES, conf, logger.NOP, statsStore)
			require.NoError(t, err)
			ctx := context.Background()
			warehouse := model.Warehouse{
				Type: tc.destinationType,
				Destination: backendconfig.DestinationT{
					ID:   destinationID,
					Name: destinationName,
				},
				Source: backendconfig.SourceT{
					ID:   sourceID,
					Name: sourceName,
				},
			}
			j := uploadJobFactory.NewUploadJob(ctx, &model.UploadJob{
				Upload: model.Upload{
					WorkspaceID:   workspaceID,
					DestinationID: destinationID,
					SourceID:      sourceID,
				},
				Warehouse: warehouse,
			}, whManager)
			j.schemaHandle, err = schema.New(ctx, warehouse, conf, logger.NOP, statsStore, &mockFetchSchemaRepo{}, &mockSchemaRepo{}, nil)
			require.NoError(t, err)
			err = j.schemaHandle.UpdateTableSchema(ctx, tableName, model.TableSchema{
				"test-column-1": "string",
				"test-column-2": "string",
				"test-column-3": "string",
			})
			require.NoError(t, err)

			tags := j.buildTags()
			tags["tableName"] = warehouseutils.TableNameForStats(tableName)

			j.columnCountStat(tableName)

			m1 := statsStore.Get("warehouse_load_table_column_count", tags)
			m2 := statsStore.Get("warehouse_load_table_column_limit", tags)

			if tc.statExpected {
				columnsCount, err := j.schemaHandle.GetColumnsCount(ctx, tableName)
				require.NoError(t, err)
				require.EqualValues(t, m1.LastValue(), columnsCount)
				require.EqualValues(t, m2.LastValue(), tc.columnCountLimit)
			} else {
				require.Nil(t, m1)
				require.Nil(t, m2)
			}
		})
	}
}

type mockAlertSender struct {
	mockError error
}

func (m *mockAlertSender) SendAlert(context.Context, string, alerta.SendAlertOpts) error {
	return m.mockError
}

func TestUploadJobT_UpdateTableSchema(t *testing.T) {
	t.Parallel()

	var (
		testNamespace       = "test_namespace"
		testTable           = "test_table"
		testColumn          = "test_column"
		testColumnType      = "text"
		testDestinationID   = "test_destination_id"
		testDestinationType = "test_destination_type"
	)

	t.Run("alter column", func(t *testing.T) {
		t.Parallel()

		t.Run("basic", func(t *testing.T) {
			t.Parallel()

			testCases := []struct {
				name           string
				createView     bool
				mockAlertError error
				wantError      error
			}{
				{
					name: "success",
				},
				{
					name:       "with view attached to table",
					createView: true,
				},
				{
					name:           "with alert error",
					createView:     true,
					mockAlertError: errors.New("alert error"),
					wantError:      errors.New("alert error"),
				},
				{
					name:           "skipping columns",
					createView:     true,
					mockAlertError: errors.New("alert error"),
					wantError:      errors.New("alert error"),
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					pool, err := dockertest.NewPool("")
					require.NoError(t, err)

					pgResource, err := postgres.Setup(pool, t)
					require.NoError(t, err)

					t.Log("db:", pgResource.DBDsn)

					rs := redshift.New(config.New(), logger.NOP, stats.NOP)
					rs.DB = sqlmiddleware.New(pgResource.DB)
					rs.Namespace = testNamespace

					ujf := &UploadJobFactory{
						conf:         config.New(),
						logger:       logger.NOP,
						statsFactory: stats.NOP,
						db:           sqlmiddleware.New(pgResource.DB),
					}

					job := ujf.NewUploadJob(context.Background(), &model.UploadJob{
						Upload: model.Upload{
							DestinationID:   testDestinationID,
							DestinationType: testDestinationType,
						},
						Warehouse: model.Warehouse{
							Type: testDestinationType,
						},
					}, rs)
					job.alertSender = &mockAlertSender{
						mockError: tc.mockAlertError,
					}

					_, err = rs.DB.Exec(
						fmt.Sprintf("CREATE SCHEMA %s;",
							testNamespace,
						),
					)
					require.NoError(t, err)

					_, err = rs.DB.Exec(
						fmt.Sprintf("CREATE TABLE %q.%q (%s VARCHAR(512));",
							testNamespace,
							testTable,
							testColumn,
						),
					)
					require.NoError(t, err)

					if tc.createView {
						_, err = rs.DB.Exec(
							fmt.Sprintf("CREATE VIEW %[1]q.%[2]q AS SELECT * FROM %[1]q.%[3]q;",
								testNamespace,
								fmt.Sprintf("%s_view", testTable),
								testTable,
							),
						)
						require.NoError(t, err)
					}

					err = job.UpdateTableSchema(testTable, warehouseutils.TableSchemaDiff{
						AlteredColumnMap: model.TableSchema{
							testColumn: testColumnType,
						},
					})
					if tc.wantError != nil {
						require.ErrorContains(t, err, tc.wantError.Error())
					} else {
						require.NoError(t, err)
					}
				})
			}
		})

		t.Run("process all columns", func(t *testing.T) {
			t.Parallel()

			pool, err := dockertest.NewPool("")
			require.NoError(t, err)

			pgResource, err := postgres.Setup(pool, t)
			require.NoError(t, err)

			t.Log("db:", pgResource.DBDsn)

			rs := redshift.New(config.New(), logger.NOP, stats.NOP)
			rs.DB = sqlmiddleware.New(pgResource.DB)
			rs.Namespace = testNamespace

			ujf := &UploadJobFactory{
				conf:         config.New(),
				logger:       logger.NOP,
				statsFactory: stats.NOP,
				db:           sqlmiddleware.New(pgResource.DB),
			}

			job := ujf.NewUploadJob(context.Background(), &model.UploadJob{
				Upload: model.Upload{
					DestinationID:   testDestinationID,
					DestinationType: testDestinationType,
				},
				Warehouse: model.Warehouse{
					Type: testDestinationType,
				},
			}, rs)
			job.alertSender = &mockAlertSender{}

			_, err = rs.DB.Exec(
				fmt.Sprintf("CREATE SCHEMA %s;",
					testNamespace,
				),
			)
			require.NoError(t, err)

			_, err = rs.DB.Exec(
				fmt.Sprintf("CREATE TABLE %q.%q (%s VARCHAR(512));",
					testNamespace,
					testTable,
					testColumn,
				),
			)
			require.NoError(t, err)

			for i := range [10]int{} {
				if i%3 == 0 {
					continue
				}

				_, err = rs.DB.Exec(
					fmt.Sprintf("ALTER TABLE %q.%q ADD COLUMN %s_%d VARCHAR(512);",
						testNamespace,
						testTable,
						testColumn,
						i,
					),
				)
				require.NoError(t, err)
			}

			_, err = rs.DB.Exec(
				fmt.Sprintf("CREATE VIEW %[1]q.%[2]q AS SELECT * FROM %[1]q.%[3]q;",
					testNamespace,
					fmt.Sprintf("%s_view", testTable),
					testTable,
				),
			)
			require.NoError(t, err)

			alteredColumnsMap := model.TableSchema{}
			for i := range [10]int{} {
				alteredColumnsMap[fmt.Sprintf("%s_%d", testColumn, i)] = testColumnType
			}

			err = job.UpdateTableSchema(testTable, warehouseutils.TableSchemaDiff{
				AlteredColumnMap: alteredColumnsMap,
			})
			require.Error(t, err)

			for i := range [10]int{} {
				column := fmt.Sprintf("test_column_%d", i)

				if i%3 == 0 {
					require.Contains(t, err.Error(), column)
				} else {
					require.NotContains(t, err.Error(), column)
				}
			}
		})
	})
}

func TestUploadJobT_Aborted(t *testing.T) {
	t.Parallel()

	var (
		minAttempts    = 3
		minRetryWindow = 3 * time.Hour
		now            = time.Date(2021, 1, 1, 6, 0, 0, 0, time.UTC)
	)

	testCases := []struct {
		name      string
		attempts  int
		startTime time.Time
		expected  bool
	}{
		{
			name:      "empty start time",
			startTime: time.Time{},
			expected:  false,
		},
		{
			name:      "crossing max attempts but not retry window",
			attempts:  5,
			startTime: time.Date(2021, 1, 1, 5, 30, 0, 0, time.UTC),
			expected:  false,
		},
		{
			name:      "crossing max retry window but not attempts",
			attempts:  2,
			startTime: time.Date(2021, 1, 1, 2, 0, 0, 0, time.UTC),
			expected:  false,
		},
		{
			name:      "crossing max retry window but not attempts",
			attempts:  5,
			startTime: time.Date(2021, 1, 1, 2, 0, 0, 0, time.UTC),
			expected:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			job := &UploadJob{
				now: func() time.Time { return now },
				ctx: context.Background(),
			}
			job.config.minRetryAttempts = minAttempts
			job.config.retryTimeWindow = minRetryWindow

			require.Equal(t, tc.expected, job.Aborted(tc.attempts, tc.startTime))
		})
	}
}

type mockPendingTablesRepo struct {
	pendingTables []model.PendingTableUpload
	err           error
	called        int
}

func (m *mockPendingTablesRepo) PendingTableUploads(context.Context, string, string, int, time.Time, int64) ([]model.PendingTableUpload, error) {
	m.called++
	return m.pendingTables, m.err
}

func TestUploadJobT_TablesToSkip(t *testing.T) {
	t.Run("repo error", func(t *testing.T) {
		job := &UploadJob{
			upload: model.Upload{
				ID: 1,
			},
			pendingTableUploadsRepo: &mockPendingTablesRepo{
				err: errors.New("some error"),
			},
			ctx: context.Background(),
		}

		previouslyFailedTables, currentJobSucceededTables, err := job.TablesToSkip()
		require.EqualError(t, err, "pending table uploads: some error")
		require.Empty(t, previouslyFailedTables)
		require.Empty(t, currentJobSucceededTables)
	})

	t.Run("should populate only once", func(t *testing.T) {
		ptRepo := &mockPendingTablesRepo{}

		job := &UploadJob{
			upload: model.Upload{
				ID: 1,
			},
			pendingTableUploadsRepo: ptRepo,
			ctx:                     context.Background(),
		}

		for i := 0; i < 5; i++ {
			_, _, _ = job.TablesToSkip()
			require.Equal(t, 1, ptRepo.called)
		}
	})

	t.Run("skip tables", func(t *testing.T) {
		pendingTables := []model.PendingTableUpload{
			{
				UploadID:      1,
				DestinationID: "destID",
				Namespace:     "namespace",
				Status:        model.TableUploadExportingFailed,
				TableName:     "previously_failed_table_1",
				Error:         "some error",
			},
			{
				UploadID:      1,
				DestinationID: "destID",
				Namespace:     "namespace",
				Status:        model.TableUploadUpdatingSchemaFailed,
				TableName:     "previously_failed_table_2",
				Error:         "",
			},
			{
				UploadID:      1,
				DestinationID: "destID",
				Namespace:     "namespace",
				Status:        model.TableUploadExported,
				TableName:     "previously_succeeded_table_1",
				Error:         "",
			},
			{
				UploadID:      5,
				DestinationID: "destID",
				Namespace:     "namespace",
				Status:        model.TableUploadExportingFailed,
				TableName:     "current_failed_table_1",
				Error:         "some error",
			},
			{
				UploadID:      5,
				DestinationID: "destID",
				Namespace:     "namespace",
				Status:        model.TableUploadExported,
				TableName:     "current_succeeded_table_1",
				Error:         "",
			},
		}

		testCases := []struct {
			name                           string
			skipPreviouslyFailedTables     bool
			expectedPreviouslyFailedTables map[string]model.PendingTableUpload
		}{
			{
				name:                           "skip previously failed tables",
				skipPreviouslyFailedTables:     true,
				expectedPreviouslyFailedTables: map[string]model.PendingTableUpload{},
			},
			{
				name:                       "do not skip previously failed tables",
				skipPreviouslyFailedTables: false,
				expectedPreviouslyFailedTables: map[string]model.PendingTableUpload{
					"previously_failed_table_1": pendingTables[0],
				},
			},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				job := &UploadJob{
					upload: model.Upload{
						ID: 5,
					},
					pendingTableUploadsRepo: &mockPendingTablesRepo{
						pendingTables: pendingTables,
					},
					ctx: context.Background(),
				}
				job.config.skipPreviouslyFailedTables = tc.skipPreviouslyFailedTables

				previouslyFailedTables, currentJobSucceededTables, err := job.TablesToSkip()
				require.NoError(t, err)
				require.Equal(t, tc.expectedPreviouslyFailedTables, previouslyFailedTables)
				require.Equal(t, map[string]model.PendingTableUpload{
					"current_succeeded_table_1": pendingTables[4],
				}, currentJobSucceededTables)
			})
		}
	})
}

func TestUploadJob_DurationBeforeNextAttempt(t *testing.T) {
	testCases := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{
			name:     "attempt 0",
			attempt:  0,
			expected: time.Duration(0),
		},
		{
			name:     "attempt 1",
			attempt:  1,
			expected: time.Second * 60,
		},
		{
			name:     "attempt 2",
			attempt:  2,
			expected: time.Second * 120,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			job := &UploadJob{}
			job.config.minUploadBackoff = time.Second * 60
			job.config.maxUploadBackoff = time.Second * 1800
			require.Equal(t, tc.expected, job.durationBeforeNextAttempt(int64(tc.attempt)))
		})
	}
}

func TestUploadJob_CanAppend(t *testing.T) {
	testCases := []struct {
		name           string
		sourceCategory string
		sourceJobRunID string // if not empty then it's an ETL source
		originalID     string // if not empty then it's a replay
		expected       bool
	}{
		{
			name:           "not a merge category",
			sourceCategory: "event-stream",
			sourceJobRunID: "",
			originalID:     "",
			expected:       true,
		},
		{
			name:           "cloud merge category",
			sourceCategory: "cloud",
			sourceJobRunID: "",
			originalID:     "",
			expected:       false,
		},
		{
			name:           "singer-protocol merge category",
			sourceCategory: "singer-protocol",
			sourceJobRunID: "",
			originalID:     "",
			expected:       false,
		},
		{
			name:           "etl source",
			sourceCategory: "event-stream",
			sourceJobRunID: "some-job-run-id",
			originalID:     "",
			expected:       false,
		},
		{
			name:           "replay",
			sourceCategory: "event-stream",
			sourceJobRunID: "",
			originalID:     "some-original-id",
			expected:       false,
		},
		{
			name:           "replay of etl source in merge category map",
			sourceCategory: "cloud",
			sourceJobRunID: "some-job-run-id",
			originalID:     "some-original-id",
			expected:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uj := UploadJob{
				upload: model.Upload{
					SourceJobRunID: tc.sourceJobRunID,
				},
				warehouse: model.Warehouse{
					Source: backendconfig.SourceT{
						OriginalID: tc.originalID,
						SourceDefinition: backendconfig.SourceDefinitionT{
							Category: tc.sourceCategory,
						},
					},
				},
			}
			require.Equal(t, uj.CanAppend(), tc.expected)
		})
	}
}

func TestUploadJob_GetLoadFilesMetadata(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name              string
		tableName         string
		limit             int64
		expectedLoadFiles int
	}{
		{
			name:              "query with upload ID",
			expectedLoadFiles: 4,
		},
		{
			name:              "query with upload ID and table name",
			tableName:         "test_table2",
			expectedLoadFiles: 3,
		},
		{
			name:              "query with upload ID, table name and limit",
			tableName:         "test_table2",
			limit:             2,
			expectedLoadFiles: 2,
		},
		{
			name:              "query with upload ID and limit",
			limit:             1,
			expectedLoadFiles: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)

			conf := config.New()

			job := &UploadJob{
				ctx:            ctx,
				db:             db,
				upload:         model.Upload{},
				stagingFileIDs: []int64{1, 2, 3},
				logger:         logger.NOP,
			}
			var stagingFileId int64
			stagingFileId, job.upload.ID = createUpload(t, ctx, db)
			loadFiles := []model.LoadFile{
				{
					UploadID:      &job.upload.ID,
					StagingFileID: stagingFileId,
					TableName:     "test_table",
				},
				{
					UploadID:  &job.upload.ID,
					TableName: "test_table2",
				},
				{
					UploadID:  &job.upload.ID,
					TableName: "test_table2",
				},
				{
					UploadID:  &job.upload.ID,
					TableName: "test_table2",
				},
			}
			err := repo.NewLoadFiles(db, conf).Insert(ctx, loadFiles)
			require.NoError(t, err)
			result, err := job.GetLoadFilesMetadata(ctx, warehouseutils.GetLoadFilesOptions{
				Table: tc.tableName,
				Limit: tc.limit,
			})
			require.NoError(t, err)
			require.Equal(t, tc.expectedLoadFiles, len(result))
		})
	}
}

func createUpload(t testing.TB, ctx context.Context, db *sqlmiddleware.DB) (int64, int64) {
	t.Helper()
	stagingFilesRepo := repo.NewStagingFiles(db, config.New())
	stagingFile := model.StagingFileWithSchema{
		StagingFile: model.StagingFile{},
	}
	var err error
	stagingFile.ID, err = stagingFilesRepo.Insert(ctx, &stagingFile)
	require.NoError(t, err)
	stagingFiles := []*model.StagingFile{&stagingFile.StagingFile}
	uploadRepo := repo.NewUploads(db)
	upload := model.Upload{}
	uploadID, err := uploadRepo.CreateWithStagingFiles(ctx, upload, stagingFiles)
	require.NoError(t, err)
	return stagingFile.ID, uploadID
}

func TestCleanupObjectStorageFiles(t *testing.T) {
	stagingFiles := lo.RepeatBy(4, func(i int) *model.StagingFile {
		return &model.StagingFile{
			ID:       int64(i + 1),
			Location: fmt.Sprintf("test-location-%d", i+1),
		}
	})
	loadFiles := lo.RepeatBy(4, func(i int) model.LoadFile {
		return model.LoadFile{
			Location: fmt.Sprintf("test-load-location-%d", i+1),
		}
	})
	createUploadJob := func(destConfig map[string]interface{}, destName string, mockFileManager filemanager.FileManager, mockLoadFilesRepo loadFilesRepo, stagingFiles []*model.StagingFile) *UploadJob {
		job := &UploadJob{
			ctx: context.Background(),
			upload: model.Upload{
				WorkspaceID: "test-workspace",
				ID:          1,
			},
			warehouse: model.Warehouse{
				Destination: backendconfig.DestinationT{
					Config: destConfig,
					DestinationDefinition: backendconfig.DestinationDefinitionT{
						Name: destName,
					},
				},
			},
			conf: config.New(),
			fileManagerFactory: func(settings *filemanager.Settings) (filemanager.FileManager, error) {
				return mockFileManager, nil
			},
			loadFilesRepo:  mockLoadFilesRepo,
			stagingFiles:   stagingFiles,
			stagingFileIDs: lo.Map(stagingFiles, func(f *model.StagingFile, _ int) int64 { return f.ID }),
			statsFactory:   stats.NOP,
			now:            time.Now,
			logger:         logger.NOP,
		}
		job.stats.objectsDeleted = stats.NOP.NewStat("objects_deleted_count", stats.GaugeType)
		job.stats.objectsDeletionTime = stats.NOP.NewStat("objects_deletion_time", stats.GaugeType)
		return job
	}
	setupMockFileLocationCalls := func(mockFileManager *mock_filemanager.MockFileManager, stagingFiles []*model.StagingFile, loadFiles []model.LoadFile) {
		for _, file := range stagingFiles {
			mockFileManager.EXPECT().GetDownloadKeyFromFileLocation(file.Location).Return(file.Location).Times(1)
		}
		for _, file := range loadFiles {
			mockFileManager.EXPECT().GetDownloadKeyFromFileLocation(file.Location).Return(file.Location).Times(1)
		}
	}
	createDestConfig := func(cleanupEnabled bool, bucketProvider string) map[string]interface{} {
		return map[string]interface{}{
			model.CleanupObjectStorageFilesSetting.String(): cleanupEnabled,
			"bucketProvider": bucketProvider,
		}
	}
	setupMocks := func() (*gomock.Controller, *mock_filemanager.MockFileManager, *mockupload.MockloadFilesRepo) {
		ctrl := gomock.NewController(t)
		return ctrl, mock_filemanager.NewMockFileManager(ctrl), mockupload.NewMockloadFilesRepo(ctrl)
	}

	t.Run("cleanup disabled", func(t *testing.T) {
		ctrl, mockFileManager, mockLoadFilesRepo := setupMocks()
		defer ctrl.Finish()

		destConfig := createDestConfig(false, warehouseutils.S3)
		job := createUploadJob(destConfig, warehouseutils.SNOWFLAKE, mockFileManager, mockLoadFilesRepo, stagingFiles)

		err := job.cleanupObjectStorageFiles()
		require.NoError(t, err)
	})

	t.Run("cleanup enabled, successful deletion for Non-Datalakes", func(t *testing.T) {
		ctrl, mockFileManager, mockLoadFilesRepo := setupMocks()
		defer ctrl.Finish()

		setupMockFileLocationCalls(mockFileManager, stagingFiles, loadFiles)
		expectedKeys := append(
			lo.Map(stagingFiles, func(f *model.StagingFile, _ int) string { return f.Location }),
			lo.Map(loadFiles, func(f model.LoadFile, _ int) string { return f.Location })...,
		)
		mockFileManager.EXPECT().Delete(gomock.Any(), expectedKeys).Return(nil).Times(1)
		mockLoadFilesRepo.EXPECT().Get(context.Background(), int64(1)).Return(loadFiles, nil).Times(1)

		destConfig := createDestConfig(true, warehouseutils.S3)
		job := createUploadJob(destConfig, warehouseutils.SNOWFLAKE, mockFileManager, mockLoadFilesRepo, stagingFiles)

		err := job.cleanupObjectStorageFiles()
		require.NoError(t, err)
	})

	t.Run("cleanup enabled, successful deletion for Datalakes", func(t *testing.T) {
		ctrl, mockFileManager, mockLoadFilesRepo := setupMocks()
		defer ctrl.Finish()

		setupMockFileLocationCalls(mockFileManager, stagingFiles, nil)
		expectedKeys := lo.Map(stagingFiles, func(f *model.StagingFile, _ int) string { return f.Location })
		mockFileManager.EXPECT().Delete(gomock.Any(), expectedKeys).Return(nil).Times(1)

		destConfig := createDestConfig(true, warehouseutils.S3)
		job := createUploadJob(destConfig, warehouseutils.S3Datalake, mockFileManager, mockLoadFilesRepo, stagingFiles)

		err := job.cleanupObjectStorageFiles()
		require.NoError(t, err)
	})

	t.Run("cleanup enabled, deletion error", func(t *testing.T) {
		ctrl, mockFileManager, mockLoadFilesRepo := setupMocks()
		defer ctrl.Finish()

		setupMockFileLocationCalls(mockFileManager, stagingFiles, loadFiles)
		expectedKeys := append(
			lo.Map(stagingFiles, func(f *model.StagingFile, _ int) string { return f.Location }),
			lo.Map(loadFiles, func(f model.LoadFile, _ int) string { return f.Location })...,
		)
		mockFileManager.EXPECT().Delete(gomock.Any(), expectedKeys).Return(errors.New("delete error")).Times(1)
		mockLoadFilesRepo.EXPECT().Get(context.Background(), int64(1)).Return(loadFiles, nil).Times(1)

		destConfig := createDestConfig(true, warehouseutils.S3)
		job := createUploadJob(destConfig, warehouseutils.SNOWFLAKE, mockFileManager, mockLoadFilesRepo, stagingFiles)

		err := job.cleanupObjectStorageFiles()
		require.EqualError(t, err, "deleting files from object storage: delete error")
	})

	t.Run("GCS cleanup enabled, chunked deletion", func(t *testing.T) {
		ctrl, mockFileManager, mockLoadFilesRepo := setupMocks()
		defer ctrl.Finish()

		setupMockFileLocationCalls(mockFileManager, stagingFiles, loadFiles)
		expectedKeys := append(
			lo.Map(stagingFiles, func(f *model.StagingFile, _ int) string { return f.Location }),
			lo.Map(loadFiles, func(f model.LoadFile, _ int) string { return f.Location })...,
		)
		for _, chunk := range lo.Chunk(expectedKeys, 4) {
			mockFileManager.EXPECT().Delete(gomock.Any(), chunk).Return(nil).Times(1)
		}
		mockLoadFilesRepo.EXPECT().Get(context.Background(), int64(1)).Return(loadFiles, nil).Times(1)

		destConfig := createDestConfig(true, warehouseutils.GCS)
		job := createUploadJob(destConfig, warehouseutils.BQ, mockFileManager, mockLoadFilesRepo, stagingFiles)
		job.config.maxConcurrentObjDeleteRequests = func(workspaceID string) int { return 2 }
		job.config.objDeleteBatchSize = func(workspaceID string) int { return 4 }

		err := job.cleanupObjectStorageFiles()
		require.NoError(t, err)
	})
}

type mockManager struct {
	manager.Manager
	isSchemaOutdated bool
	fetchSchemaCount int
}

func (m *mockManager) Setup(ctx context.Context, warehouse model.Warehouse, uploader warehouseutils.Uploader) error {
	return nil
}

func (m *mockManager) CreateTable(ctx context.Context, tableName string, schema model.TableSchema) error {
	return fmt.Errorf("error in creating table")
}

func (m *mockManager) FetchSchema(ctx context.Context) (model.Schema, error) {
	m.fetchSchemaCount++
	// For the first call in the outdated case, we expect the table to not exist in the warehouse
	// So we return an empty schema. After that we are returning a non-empty schema to simulate the diff
	if m.isSchemaOutdated && m.fetchSchemaCount > 1 {
		return model.Schema{
			"test_table": model.TableSchema{
				"id": "int",
			},
		}, nil
	}
	return model.Schema{}, nil
}

type mockLoadFilesRepo struct {
	loadFilesRepo
}

func (m *mockLoadFilesRepo) DistinctTableName(ctx context.Context, sourceID, destinationID string, startID, endID int64) ([]string, error) {
	return []string{"test_table"}, nil
}

func TestUploadJob_SchemaResetState(t *testing.T) {
	db := setupDB(t)
	factory := UploadJobFactory{
		db:           db,
		logger:       logger.NOP,
		statsFactory: stats.NOP,
		conf:         config.New(),
	}
	manager := snowflake.New(config.New(), logger.NOP, stats.NOP)
	_, uploadId := createUpload(t, context.Background(), db)
	testCases := []struct {
		name             string
		isSchemaOutdated bool
		namespace        string // Using different namespaces to ensure that one test case doesn't affect the other
		expectedStatus   string
	}{
		{
			name:             "schema outdated",
			isSchemaOutdated: true,
			namespace:        "test_namespace1",
			expectedStatus:   model.Waiting,
		},
		{
			name:             "schema up to date",
			isSchemaOutdated: false,
			namespace:        "test_namespace2",
			expectedStatus:   model.ExportingDataFailed,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockManager := &mockManager{
				Manager:          manager,
				isSchemaOutdated: tc.isSchemaOutdated,
			}
			uploadJob := factory.NewUploadJob(context.Background(), &model.UploadJob{
				Warehouse: model.Warehouse{
					Namespace: tc.namespace,
				},
				StagingFiles: []*model.StagingFile{
					{},
				},
				Upload: model.Upload{
					ID:     uploadId,
					Status: model.ExportingData,
					UploadSchema: model.Schema{
						"test_table": model.TableSchema{
							"id": "string",
						},
					},
				},
			}, mockManager)
			uploadJob.loadFilesRepo = &mockLoadFilesRepo{
				loadFilesRepo: uploadJob.loadFilesRepo,
			}
			err := uploadJob.run()
			require.ErrorContains(t, err, "error in creating table")
			require.Equal(t, tc.expectedStatus, uploadJob.upload.Status)
		})
	}
}
