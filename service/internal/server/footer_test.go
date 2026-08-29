package server

import (
	"context"
	"testing"

	"github.com/jamesread/starapp/service/internal/config"
	"github.com/jamesread/starapp/service/internal/cvar"
	"github.com/jamesread/starapp/service/internal/store"
	"github.com/stretchr/testify/require"
)

func TestInitShellFooterRedactsVersion(t *testing.T) {
	ctx := context.Background()
	st := store.OpenMemory()
	cfg := &config.Config{ShowFooter: true}
	s := New(cfg, st, nil, nil)
	require.NoError(t, EnsureDefaultCvars(ctx, st, "StarApp", true))

	init := s.initShell(ctx)
	require.True(t, init.ShowFooter)
	require.True(t, init.ShowVersionNumber)
	require.NotEmpty(t, init.CurrentVersion)
	require.False(t, init.ShowNewVersions)

	require.NoError(t, st.UpdateCvar(ctx, cvar.KeyShowVersionNumber, 0, ""))

	init = s.initShell(ctx)
	require.False(t, init.ShowVersionNumber)
	require.Empty(t, init.CurrentVersion)
	require.Empty(t, init.AvailableVersion)
	require.False(t, init.ShowNewVersions)
}

func TestInitShellFooterHidden(t *testing.T) {
	ctx := context.Background()
	st := store.OpenMemory()
	cfg := &config.Config{ShowFooter: false}
	s := New(cfg, st, nil, nil)
	require.NoError(t, EnsureDefaultCvars(ctx, st, "StarApp", false))

	init := s.initShell(ctx)
	require.False(t, init.ShowFooter)
}
