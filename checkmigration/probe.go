package checkmigration

import "gorm.io/gorm"

// probeDialect keeps SQL output scoped to the capture's debug flag.
// probeDialect 将 SQL 输出限定在捕获器的调试开关内。
type probeDialect struct {
	gorm.Dialector
	capture *SQLCapture
}

func (d probeDialect) Migrator(db *gorm.DB) gorm.Migrator {
	// AutoMigrate wraps its execution session in GORM's stdout SQL log.
	// Restore the capture at each migration dispatch; retain DryRun and the dialect.
	// AutoMigrate 为执行会话包上直接打印到 stdout 的日志器。
	// 每次分派迁移操作时恢复捕获器，保留 DryRun 和原有方言行为。
	session := db.Session(&gorm.Session{Logger: d.capture})
	return d.Dialector.Migrator(session)
}
