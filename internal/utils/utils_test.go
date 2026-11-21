package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/yylego/go-migrate/internal/utils"
	"github.com/yylego/go-migrate/newmigrate"
)

func TestHandleCause(t *testing.T) {
	t.Run("no-scripts", func(t *testing.T) {
		// A scripts DIR with no files gives os.ErrNotExist: a warning, no panic
		// 脚本目录是空的会得到 os.ErrNotExist：只警告，不 panic
		migration := newSqliteMigration(t, t.TempDir())
		cause := migration.Up()
		require.ErrorIs(t, cause, os.ErrNotExist)
		require.NotPanics(t, func() { utils.HandleCause(cause) })
	})

	t.Run("up-to-date", func(t *testing.T) {
		// Running Up a second time gives migrate.ErrNoChange: nothing to run, no panic
		// 第二次 Up 会得到 migrate.ErrNoChange：没有要执行的，不 panic
		scriptsPath := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(scriptsPath, "000001_create_users.up.sql"), []byte("CREATE TABLE users (id INTEGER PRIMARY KEY);"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(scriptsPath, "000001_create_users.down.sql"), []byte("DROP TABLE users;"), 0o644))

		migration := newSqliteMigration(t, scriptsPath)
		require.NotPanics(t, func() { utils.HandleCause(migration.Up()) })
		cause := migration.Up()
		require.ErrorIs(t, cause, migrate.ErrNoChange)
		require.NotPanics(t, func() { utils.HandleCause(cause) })

		// One more step at the latest version also gives os.ErrNotExist, with the scripts path being right: a warning, no panic
		// 已是最新版本时再单步前进同样得到 os.ErrNotExist，这时脚本目录并没有写错：只警告，不 panic
		cause = migration.Steps(+1)
		require.ErrorIs(t, cause, os.ErrNotExist)
		require.NotPanics(t, func() { utils.HandleCause(cause) })
	})
}

func newSqliteMigration(t *testing.T, scriptsPath string) *migrate.Migrate {
	migration, err := newmigrate.NewWithScriptsAndDBSource[*sqlite3.Sqlite](&newmigrate.ScriptsAndDBSourceParam{
		ScriptsInRoot: scriptsPath,
		ConnectSource: "sqlite3://file:/" + uuid.NewString() + "?mode=memory&cache=private",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		sourceCause, databaseCause := migration.Close()
		require.NoError(t, sourceCause)
		require.NoError(t, databaseCause)
	})
	return migration
}
