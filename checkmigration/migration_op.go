package checkmigration

import (
	"slices"
	"strings"

	"github.com/yylego/tern"
)

// MigrationOp represents one captured database migration operation
//
// MigrationOp 表示一条捕获到的数据库迁移操作
type MigrationOp struct {
	ForwardSQL string // SQL statement of the forward migration // 正向迁移的 SQL 语句
}

// NewMigrationOp wraps a SQL statement as MigrationOp when it is a DDL migration
// Returns the wrapped operation and a match flag
//
// NewMigrationOp 当 SQL 是一条 DDL 迁移时，将其包装成 MigrationOp
// 返回包装后的操作和匹配标志
func NewMigrationOp(forwardSQL string) (*MigrationOp, bool) {
	// GORM DryRun also captures schema-probing SQL (SELECT, PRAGMA, etc.)
	// Keep the DDL that AutoMigrate emits (CREATE / ALTER / COMMENT ON / DROP SEQUENCE); skip the rest
	//
	// GORM DryRun 也会捕获探测结构的 SQL（SELECT、PRAGMA 等）
	// 只留 AutoMigrate 产出的 DDL（CREATE / ALTER / COMMENT ON / DROP SEQUENCE），其余跳过
	if !isMigrateDDL(forwardSQL) {
		return nil, false
	}
	return &MigrationOp{ForwardSQL: forwardSQL}, true
}

// isMigrateDDL reports if the SQL is one of the DDL statements AutoMigrate emits
// AutoMigrate builds tables, adds and alters columns, and creates indexes; on PostgreSQL it also
// sets column comments (COMMENT ON), creates sequences when a column turns auto-increment,
// and drops them when it stops being one
//
// isMigrateDDL 判断 SQL 是不是 AutoMigrate 会产出的那几种 DDL
// AutoMigrate 建表、加列改列、建索引；在 PostgreSQL 上还会改列注释（COMMENT ON），
// 列改成自增时建序列、取消自增时删序列；命中其一即保留
func isMigrateDDL(forwardSQL string) bool {
	head := strings.ToUpper(strings.TrimSpace(forwardSQL))
	switch {
	case strings.HasPrefix(head, "CREATE TABLE"),
		strings.HasPrefix(head, "ALTER TABLE"),
		strings.HasPrefix(head, "CREATE INDEX"),
		strings.HasPrefix(head, "CREATE UNIQUE INDEX"),
		strings.HasPrefix(head, "COMMENT ON"),
		strings.HasPrefix(head, "CREATE SEQUENCE"),
		strings.HasPrefix(head, "ALTER SEQUENCE"),
		strings.HasPrefix(head, "DROP SEQUENCE"):
		return true
	default:
		return false
	}
}

// MigrationOps represents a collection of migration operations
//
// MigrationOps 表示迁移操作集合
type MigrationOps []*MigrationOp

// SearchOp finds migration operation matching the given forward SQL statement
// Uses case-insensitive comparison to locate operations
//
// SearchOp 查找匹配给定正向 SQL 语句的迁移操作
// 使用不区分大小写的比较来定位操作
func (ops MigrationOps) SearchOp(forwardSQL string) *MigrationOp {
	posIndex := slices.IndexFunc(ops, func(op *MigrationOp) bool {
		return strings.EqualFold(op.ForwardSQL, forwardSQL)
	})
	return tern.BF(posIndex >= 0, func() *MigrationOp {
		return ops[posIndex]
	})
}

// ForwardSQLs extracts forward SQL statements from each migration operation
// Returns slice of SQL strings in execution sequence
//
// ForwardSQLs 从每个迁移操作中提取正向 SQL 语句
// 按执行顺序返回 SQL 字符串切片
func (ops MigrationOps) ForwardSQLs() []string {
	var sqs = make([]string, 0, len(ops))
	for _, op := range ops {
		sqs = append(sqs, op.ForwardSQL)
	}
	return sqs
}

// ForwardScript generates complete forward migration script with semicolons
// Combines each forward SQL statement into executable script format
//
// ForwardScript 生成带分号的完整正向迁移脚本
// 将每个正向 SQL 语句组合成可执行的脚本格式
func (ops MigrationOps) ForwardScript() string {
	var sqs = make([]string, 0, len(ops))
	for _, op := range ops {
		sqs = append(sqs, op.ForwardSQL+";")
	}
	res := strings.Join(sqs, "\n\n")
	if len(res) > 0 {
		res += "\n"
	}
	return res
}
