package store

import (
	"context"
	"errors"

	iamsqlite "github.com/jamesread/armature-iam/store/sqlite"

	"github.com/jamesread/starapp/service/internal/config"
)

var errCvarNotFound = errors.New("cvar not found")

// Memory is an in-memory Store for tests that bypass migration checks.
type Memory struct {
	*iamsqlite.SQLite
	cvars      map[string]CvarRow
	userPrefs  map[int]UserPreferencesRow
	webhooks   *memoryWebhook
	family     *memoryFamily
	chores     *memoryChore
	starCharts *memoryStarChart
}

// OpenMemory returns a new in-memory store.
func OpenMemory() *Memory {
	iam, err := iamsqlite.OpenMemory()
	if err != nil {
		panic(err)
	}
	if err := seedDomainRBAC(iam.DB()); err != nil {
		panic(err)
	}
	return &Memory{SQLite: iam, cvars: map[string]CvarRow{}, userPrefs: map[int]UserPreferencesRow{}}
}

func (m *Memory) Close() error {
	if m.SQLite != nil {
		return m.SQLite.Close()
	}
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
