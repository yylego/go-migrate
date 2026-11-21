package cobramigration_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/go-migrate/cobramigration"
	"github.com/yylego/go-migrate/migrationparam"
)

func TestNewMigrateCmd(t *testing.T) {
	// Building the command tree opens no connection, so the builders can be left out here
	// 构建命令树不会建立连接，所以这里可以不传构建函数
	cmd := cobramigration.NewMigrateCmd(migrationparam.NewMigrationParam(nil, nil))
	require.Equal(t, "migrate", cmd.Name())

	var names []string
	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}
	require.ElementsMatch(t, []string{"batch", "inc", "dec"}, names)
}
