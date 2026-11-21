package checkmigration_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/go-migrate/checkmigration"
	"github.com/yylego/go-migrate/migrationparam"
	"gorm.io/gorm"
)

func TestProbeErrors(t *testing.T) {
	ops, err := checkmigration.GetMigrateOps(nil, nil)
	require.Error(t, err)
	require.Nil(t, ops)
	// A failed session must preserve its cause, not return incomplete SQL as success.
	// 失败会话应保留错误链，不把不完整 SQL 当成成功结果。
	cause := errors.New("probe session failed")
	db := caseDB.Session(&gorm.Session{})
	require.ErrorIs(t, db.AddError(cause), cause)
	sqls, err := checkmigration.CheckMigrate(db, []any{&AccountV1{}})
	require.ErrorIs(t, err, cause)
	require.Nil(t, sqls)
	db = caseDB.Session(&gorm.Session{})
	db.Dialector = failedProbeDialect{Dialector: db.Dialector, cause: cause}
	sqls, err = checkmigration.CheckMigrate(db, []any{&AccountV1{}})
	require.ErrorIs(t, err, cause)
	require.Nil(t, sqls)
}

// TestProbeModes preserves captured DDL, the source session, and the database.
// TestProbeModes 核对两种模式的 DDL 相同，且不改变原会话或执行建表。
func TestProbeModes(t *testing.T) {
	debugMode := migrationparam.DebugMode()
	t.Cleanup(func() { migrationparam.SetDebugMode(debugMode) })
	db := caseDB.Session(&gorm.Session{})
	dialect, log := db.Dialector, db.Logger
	require.False(t, db.Migrator().HasTable(&probeOutputRecord{}))

	migrationparam.SetDebugMode(false)
	quiet, err := checkmigration.CheckMigrate(db, []any{&probeOutputRecord{}})
	require.NoError(t, err)
	require.Len(t, quiet, 1)
	require.Contains(t, quiet[0], "CREATE TABLE")
	migrationparam.SetDebugMode(true)
	debugSQL, err := checkmigration.CheckMigrate(db, []any{&probeOutputRecord{}})
	require.NoError(t, err)
	require.Equal(t, quiet, debugSQL)

	require.Same(t, dialect, db.Dialector)
	require.Equal(t, log, db.Logger)
	require.False(t, db.DryRun)
	require.False(t, db.Migrator().HasTable(&probeOutputRecord{}))
}

type probeOutputRecord struct {
	ID uint `gorm:"primaryKey"`
}

// TestProbeColumnType changes a column type on SQLite: the SQLite dialect panics in DryRun, and the probe returns it as an error
// TestProbeColumnType 在 SQLite 上改列类型：SQLite 方言在 DryRun 下会 panic，探测把它接成错误返回
func TestProbeColumnType(t *testing.T) {
	db := caseDB
	t.Cleanup(func() {
		require.NoError(t, db.Migrator().DropTable(&probeGradeV1{}))
	})
	require.NoError(t, db.AutoMigrate(&probeGradeV1{}))

	var sqls []string
	var err error
	require.NotPanics(t, func() {
		sqls, err = checkmigration.CheckMigrate(db, []any{&probeGradeV2{}})
	})
	require.Error(t, err)
	require.Nil(t, sqls)
}

type probeGradeV1 struct {
	ID    uint `gorm:"primaryKey"`
	Grade string
}

func (*probeGradeV1) TableName() string { return "probe_grades" }

type probeGradeV2 struct {
	ID    uint `gorm:"primaryKey"`
	Grade int
}

func (*probeGradeV2) TableName() string { return "probe_grades" }

type failedProbeDialect struct {
	gorm.Dialector
	cause error
}

func (d failedProbeDialect) Migrator(db *gorm.DB) gorm.Migrator {
	return failedProbeMigration{Migrator: d.Dialector.Migrator(db), cause: d.cause}
}

type failedProbeMigration struct {
	gorm.Migrator
	cause error
}

func (m failedProbeMigration) AutoMigrate(...any) error { return m.cause }
