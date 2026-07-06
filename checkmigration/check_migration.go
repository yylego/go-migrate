// Package checkmigration: reports the schema changes that GORM AutoMigrate would run
// Captures AutoMigrate's SQL via DryRun mode and keeps the DDL statements
//
// checkmigration: 报告 GORM AutoMigrate 会执行的 schema 变更
// 用 DryRun 模式捕获 AutoMigrate 的 SQL，只留 DDL 语句
package checkmigration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yylego/go-migrate/migrationparam"
	"github.com/yylego/must"
	"github.com/yylego/neatjson/neatjsons"
	"github.com/yylego/tint"
	"github.com/yylego/zaplog"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SqlCapture implements the GORM log interface to collect the SQL it sees
// Records each SQL statement into a slice during DryRun mode
//
// SqlCapture 实现 GORM 日志接口，收集经过的 SQL
// 在 DryRun 模式下把每条 SQL 记进切片
type SqlCapture struct {
	SQLs      []string // Collection of captured SQL statements // 捕获的 SQL 语句集合
	debugMode bool     // Enable debug output // 调试模式
}

func (c *SqlCapture) LogMode(level logger.LogLevel) logger.Interface {
	if c.debugMode {
		zaplog.SUG.Debugln("mode", int(level))
	}
	return c
}

func (c *SqlCapture) Info(_ context.Context, msg string, data ...any) {
	if c.debugMode {
		zaplog.SUG.Infoln("info", fmt.Sprintf(msg, data...))
	}
}

func (c *SqlCapture) Warn(_ context.Context, msg string, data ...any) {
	if c.debugMode {
		zaplog.SUG.Warnln("warn", fmt.Sprintf(msg, data...))
	}
}

func (c *SqlCapture) Error(_ context.Context, msg string, data ...any) {
	if c.debugMode {
		zaplog.SUG.Errorln("error", fmt.Sprintf(msg, data...))
	}
}

func (c *SqlCapture) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	sqx, _ := fc()
	if c.debugMode {
		zaplog.SUG.Debugln("SQL>>>", tint.GREEN.Sprint(sqx), "<<<END")
	}
	c.SQLs = append(c.SQLs, sqx)
}

// GetMigrateOps captures the SQL AutoMigrate would run and keeps the DDL operations
// Uses DryRun mode with a custom SQL capture; debug output via SetDebugMode
//
// GetMigrateOps 捕获 AutoMigrate 会执行的 SQL，只留 DDL 操作
// 用 DryRun 模式配自定义 SQL 捕获；调试输出由 SetDebugMode 控制
func GetMigrateOps(db *gorm.DB, objects []any) MigrationOps {
	// Create SqlCapture for SQL capture
	// 创建 SqlCapture 用于 SQL 捕获
	sqlCapture := &SqlCapture{
		SQLs:      make([]string, 0),
		debugMode: migrationparam.GetDebugMode(),
	}

	// Use DryRun mode with SqlCapture to capture SQL without execution
	// 使用 DryRun 模式和 SqlCapture 来捕获 SQL 而不执行
	session := &gorm.Session{
		DryRun: true,
		Logger: sqlCapture,
	}

	// Known noise: GORM Migrator emits each DryRun SQL via fmt.Println on stdout
	// See gorm/migrator/migrator.go printSQLLogger.Trace; the call is hardcoded
	//
	// 已知噪音：GORM Migrator 在 DryRun 模式下会用 fmt.Println 把每条 SQL 打到 stdout
	// 见 gorm/migrator/migrator.go printSQLLogger.Trace，调用点是硬编码的
	must.Done(db.Session(session).AutoMigrate(objects...))

	// Display captured SQL statements for debugging
	// 显示捕获的 SQL 语句用于调试
	if sqlCapture.debugMode {
		zaplog.SUG.Debugln("execute:", tint.BLUE.Sprint(neatjsons.S(sqlCapture.SQLs)))
	}

	results := make([]*MigrationOp, 0, len(sqlCapture.SQLs))
	for _, forwardSQL := range sqlCapture.SQLs {
		// Keep the DDL that AutoMigrate emits; skip the HasTable-style probing SELECT
		// 只留 AutoMigrate 产出的 DDL；跳过 HasTable 之类的探测 SELECT
		if migrationOp, match := NewMigrationOp(forwardSQL); match {
			results = append(results, migrationOp)
		}
	}
	return results
}

// CheckMigrate returns the forward SQL statements GORM AutoMigrate would run
// Wraps GetMigrateOps and pulls out each forward SQL
//
// CheckMigrate 返回 GORM AutoMigrate 会执行的正向 SQL 语句
// 封装 GetMigrateOps，取出每条正向 SQL
func CheckMigrate(db *gorm.DB, objects []any) []string {
	steps := GetMigrateOps(db, objects)
	zaplog.LOG.Debug("missing", zap.Int("size", len(steps)))
	sqs := steps.GetForwardSQLs()
	if len(sqs) > 0 {
		debugMigrationSqs(sqs)
	}
	zaplog.SUG.Debugln("success")
	return sqs
}

func debugMigrationSqs(sqs []string) {
	zaplog.SUG.Debugln("-")
	for idx, sqx := range sqs {
		zaplog.SUG.Debug(
			"missing:",
			fmt.Sprintf("(%d/%d)", idx, len(sqs)),
			"\n",
			tint.PINK.Sprint("----------------"),
			"\n\n",
			tint.PINK.Sprint(sqx+";"),
			"\n\n",
			tint.PINK.Sprint("----------------"),
		)
	}
	zaplog.SUG.Debugln("-")
	zaplog.SUG.Debugln("-")
	zaplog.SUG.Debug(
		"scripts:",
		"\n",
		tint.CYAN.Sprint("----------------"),
		"\n\n",
		tint.CYAN.Sprint(strings.Join(sqs, ";\n\n")+";"),
		"\n\n",
		tint.CYAN.Sprint("----------------"),
	)
	zaplog.SUG.Debugln("-")
	zaplog.SUG.Debugln("-")
}
