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
	"github.com/yylego/neatjson/neatjsons"
	"github.com/yylego/tint"
	"github.com/yylego/zaplog"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SQLCapture implements the GORM log interface to collect the SQL it sees
// Records each SQL statement into a slice during DryRun mode
//
// SQLCapture 实现 GORM 日志接口，收集经过的 SQL
// 在 DryRun 模式下把每条 SQL 记进切片
type SQLCapture struct {
	SQLs       []string // Collection of captured SQL statements // 捕获的 SQL 语句集合
	debugMode  bool     // Enable debug output // 调试模式
	probeFault error    // First failed metadata read // 首个元数据查询错误
}

func (c *SQLCapture) LogMode(mode logger.LogLevel) logger.Interface {
	if c.debugMode {
		zaplog.SUG.Debugln("mode", int(mode))
	}
	return c
}

func (c *SQLCapture) Info(_ context.Context, msg string, data ...any) {
	if c.debugMode {
		zaplog.SUG.Infoln("info", fmt.Sprintf(msg, data...))
	}
}

func (c *SQLCapture) Warn(_ context.Context, msg string, data ...any) {
	if c.debugMode {
		zaplog.SUG.Warnln("warn", fmt.Sprintf(msg, data...))
	}
}

func (c *SQLCapture) Error(_ context.Context, msg string, data ...any) {
	if c.debugMode {
		zaplog.SUG.Errorln("error", fmt.Sprintf(msg, data...))
	}
}

func (c *SQLCapture) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	sqx, _ := fc()
	if err != nil && c.probeFault == nil {
		c.probeFault = fmt.Errorf("schema metadata read failed (%s): %w", sqx, err)
	}
	if c.debugMode {
		zaplog.SUG.Debugln("SQL>>>", tint.GREEN.Sprint(sqx), "<<<END")
	}
	c.SQLs = append(c.SQLs, sqx)
}

// GetMigrateOps returns DDL operations with probe errors preserved.
// Failed probes return no incomplete operations. DryRun can read database metadata.
// GetMigrateOps 返回 DDL 操作及探测错误，失败时不返回不完整结果。
// DryRun 仍可能读取数据库；错误返回不代表驱动已支持所有探测操作。
func GetMigrateOps(db *gorm.DB, objects []any) (MigrationOps, error) {
	if db == nil {
		return nil, fmt.Errorf("migration probe: db is nil")
	}
	if db.Error != nil {
		return nil, fmt.Errorf("migration probe: %w", db.Error)
	}
	// Create SQLCapture to collect SQL statements
	// 创建 SQLCapture 用于 SQL 捕获
	capture := &SQLCapture{
		SQLs:      make([]string, 0),
		debugMode: migrationparam.GetDebugMode(),
	}

	// Use DryRun mode with SQLCapture to capture SQL without execution
	// 使用 DryRun 模式和 SQLCapture 来捕获 SQL 而不执行
	session := &gorm.Session{
		DryRun: true,
		Logger: capture,
	}

	// Known noise: GORM Migrator emits each DryRun SQL via fmt.Println on stdout
	// See gorm/migrator/migrator.go printSQLLogger.Trace; fmt.Println is hardcoded
	//
	// 已知噪音：GORM Migrator 在 DryRun 模式下会用 fmt.Println 把每条 SQL 打到 stdout
	// 见 gorm/migrator/migrator.go printSQLLogger.Trace，调用点是硬编码的
	probe := db.Session(session)
	adaptPostgresProbe(probe)
	err := probe.AutoMigrate(objects...)
	// APIs such as HasTable can hide metadata read errors from AutoMigrate.
	// HasTable 等布尔接口可能吞掉查询错误，不能把探测失败误报为缺表。
	if capture.probeFault != nil {
		return nil, fmt.Errorf("migration probe: %w", capture.probeFault)
	}
	if err != nil {
		return nil, fmt.Errorf("migration probe: %w", err)
	}

	// Show captured SQL statements in debug mode
	// 显示捕获的 SQL 语句用于调试
	if capture.debugMode {
		zaplog.SUG.Debugln("execute:", tint.BLUE.Sprint(neatjsons.S(capture.SQLs)))
	}

	results := make([]*MigrationOp, 0, len(capture.SQLs))
	for _, forwardSQL := range capture.SQLs {
		// Keep the DDL that AutoMigrate emits; skip the HasTable-style probing SELECT
		// 只留 AutoMigrate 产出的 DDL；跳过 HasTable 之类的探测 SELECT
		if migrationOp, match := NewMigrationOp(forwardSQL); match {
			results = append(results, migrationOp)
		}
	}
	return results, nil
}

// CheckMigrate returns SQL statements and a probe error without asserting success.
// CheckMigrate 返回 SQL 及探测错误，由调用方决定失败后的处理方式。
func CheckMigrate(db *gorm.DB, objects []any) ([]string, error) {
	steps, err := GetMigrateOps(db, objects)
	if err != nil {
		return nil, err
	}
	zaplog.LOG.Debug("missing", zap.Int("size", len(steps)))
	sqs := steps.GetForwardSQLs()
	if len(sqs) > 0 {
		debugMigrationSqs(sqs)
	}
	zaplog.SUG.Debugln("success")
	return sqs, nil
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
