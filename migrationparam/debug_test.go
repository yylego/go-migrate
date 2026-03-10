package migrationparam_test

import (
	"testing"

	"github.com/yylego/go-migrate/migrationparam"
)

func TestGetDebugMode(t *testing.T) {
	t.Log(migrationparam.GetDebugMode())
}
