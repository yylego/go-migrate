package migrationparam_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/go-migrate/migrationparam"
)

func TestSetDebugMode(t *testing.T) {
	debugMode := migrationparam.DebugMode()
	t.Cleanup(func() { migrationparam.SetDebugMode(debugMode) })

	migrationparam.SetDebugMode(true)
	require.True(t, migrationparam.DebugMode())
	migrationparam.SetDebugMode(false)
	require.False(t, migrationparam.DebugMode())
}
