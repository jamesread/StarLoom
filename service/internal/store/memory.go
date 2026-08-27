package store

import (
	"context"
	"errors"

	"github.com/jamesread/starapp/service/internal/config"
)

var errCvarNotFound = errors.New("cvar not found")

// Memory is an in-memory Store for tests that bypass migration checks.
type Memory struct {
	cvars    map[string]CvarRow
	webhooks *memoryWebhook
	iam      *memoryIAM
	family   *memoryFamily
}

// OpenMemory returns a new in-memory store.
func OpenMemory() *Memory {
	return &Memory{cvars: map[string]CvarRow{}}
}

func (m *Memory) Close() error {
	return nil
}

func (m *Memory) Ping(_ context.Context) error {
	return nil
}

func (m *Memory) HasMigration(_ context.Context, id string) (bool, error) {
	return id == config.RequiredMigration, nil
}

func (m *Memory) LatestMigration(_ context.Context) (string, error) {
	return config.RequiredMigration, nil
}
