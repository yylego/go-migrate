package migrationparam_test

import (
	"fmt"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/yylego/go-migrate/migrationparam"
	"github.com/yylego/go-migrate/newmigrate"
	"github.com/yylego/rese"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMigrationParam(t *testing.T) {
	scriptsPath := t.TempDir()
	dsn := fmt.Sprintf("file:db-%s?mode=memory&cache=shared", uuid.New().String())

	var dbCount, migrationCount int
	param := migrationparam.NewMigrationParam(
		func() *gorm.DB {
			dbCount++
			return rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent),
			}))
		},
		func(db *gorm.DB) *migrate.Migrate {
			migrationCount++
			migrationDB := rese.V1(sqlite3.WithInstance(rese.P1(db.DB()), &sqlite3.Config{}))
			return rese.P1(newmigrate.NewWithScriptsAndDatabase(&newmigrate.ScriptsAndDatabaseParam{
				ScriptsInRoot:    scriptsPath,
				DatabaseName:     "sqlite3",
				DatabaseInstance: migrationDB,
			}))
		},
	)

	// Nothing is built before the first access
	// 首次访问之前什么都不创建
	require.Zero(t, dbCount)
	require.Zero(t, migrationCount)

	// The migration and the database share one connection, each built once
	// 迁移实例和数据库共用一个连接，各只创建一次
	migration, cleanup := param.Migration()
	db, _ := param.DB()
	sameMigration, _ := param.Migration()
	require.Same(t, migration, sameMigration)
	require.Equal(t, 1, dbCount)
	require.Equal(t, 1, migrationCount)
	conn := rese.P1(db.DB())
	require.NoError(t, conn.Ping())

	// cleanup closes the connection; the next access builds a new one
	// cleanup 关闭连接；下次访问重新创建
	cleanup()
	require.Error(t, conn.Ping())
	_, cleanup = param.DB()
	require.Equal(t, 2, dbCount)
	require.Equal(t, 1, migrationCount)
	cleanup()
}
