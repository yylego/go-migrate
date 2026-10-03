package checkmigration

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// adaptPostgresProbe scopes the metadata fix to the probe session.
// adaptPostgresProbe 仅对检查会话适配元数据查询，不修改调用方的配置。
func adaptPostgresProbe(db *gorm.DB) {
	switch dialect := db.Dialector.(type) {
	case *postgres.Dialector:
		db.Dialector = postgresProbeDialect{Dialector: *dialect}
	case postgres.Dialector:
		db.Dialector = postgresProbeDialect{Dialector: dialect}
	}
}

type postgresProbeDialect struct {
	postgres.Dialector
}

func (d postgresProbeDialect) Migrator(db *gorm.DB) gorm.Migrator {
	return postgresProbeMigration{
		Migrator: d.Dialector.Migrator(db),
		db:       db,
		dialect:  d.Dialector,
	}
}

type postgresProbeMigration struct {
	gorm.Migrator
	db      *gorm.DB
	dialect postgres.Dialector
}

func (m postgresProbeMigration) ColumnTypes(value any) ([]gorm.ColumnType, error) {
	// AlterColumn calls ColumnTypes on its DryRun session. PostgreSQL GetRows
	// needs database access; DDL stays on the existing DryRun session.
	// AlterColumn 会在 DryRun 会话上再读列信息；GetRows 需要真实查询。
	// 只为 ColumnTypes 关闭 DryRun，DDL 仍使用原来的预演会话。
	readDB := m.db.Session(&gorm.Session{})
	readDB.DryRun = false
	readDB.Dialector = m.dialect
	return m.dialect.Migrator(readDB).ColumnTypes(value)
}
