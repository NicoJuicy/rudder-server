package loadfiles_test

import (
	"context"
	"slices"
	"sync"

	"github.com/rudderlabs/rudder-server/warehouse/internal/model"
)

type mockLoadFilesRepo struct {
	id    int64
	store []model.LoadFile
	mu    sync.Mutex
}

func (m *mockLoadFilesRepo) Insert(_ context.Context, loadFiles []model.LoadFile) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, lf := range loadFiles {
		lf.ID = m.id + 1
		m.id += 1

		m.store = append(m.store, lf)
	}

	return nil
}

func (m *mockLoadFilesRepo) Delete(_ context.Context, uploadID int64, stagingFileIDs []int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	store := make([]model.LoadFile, 0)
	for _, loadFile := range m.store {
		if !slices.Contains(stagingFileIDs, loadFile.StagingFileID) {
			store = append(store, loadFile)
		}
	}
	m.store = store

	return nil
}

func (m *mockLoadFilesRepo) Get(_ context.Context, uploadID int64) ([]model.LoadFile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var loadFiles []model.LoadFile
	for _, loadFile := range m.store {
		if *loadFile.UploadID == uploadID {
			loadFiles = append(loadFiles, loadFile)
		}
	}
	return loadFiles, nil
}
