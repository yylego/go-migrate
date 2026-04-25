package checkmigration

import (
	"slices"
	"strings"

	"github.com/yylego/go-migrate/migrationkinds"
	"github.com/yylego/tern"
)

// MigrationOp represents one captured database migration operation
//
// MigrationOp 表示一条捕获到的数据库迁移操作
type MigrationOp struct {
	ForwardSQL string              // SQL statement of the forward migration // 正向迁移的 SQL 语句
	ActionKind migrationkinds.Kind // Operation kind value // 操作类型值
}

// NewMigrationOp wraps a SQL statement as MigrationOp when it matches one of the known kinds
// Returns the wrapped operation and a success flag
//
// NewMigrationOp 当 SQL 匹配某种已知类型时，将其包装成 MigrationOp
// 返回包装后的操作和匹配标志
func NewMigrationOp(forwardSQL string) (*MigrationOp, bool) {
	// Walk through each registered kind, skipping the default Unknown
	// 遍历每种已注册的 kind，跳过默认的 Unknown
	for _, kind := range migrationkinds.GetEnums().ListValid() {
		elem := migrationkinds.GetEnums().MustGet(kind)
		if strings.Contains(forwardSQL, elem.Meta().ForwardSubstr) {
			return &MigrationOp{
				ForwardSQL: forwardSQL,
				ActionKind: elem.Code(),
			}, true
		}
	}
	// Return Unknown as expected fallback when nothing matches
	// GORM DryRun also captures schema-probing SQL (SELECT, PRAGMA, etc.)
	// Such non-DDL statements receive Unknown as the chosen tag
	//
	// 匹配不到时按预期返回 Unknown 兜底
	// GORM DryRun 也会捕获探测结构的 SQL（SELECT、PRAGMA 等）
	// 这些非 DDL 语句被刻意归类为 Unknown
	return &MigrationOp{
		ForwardSQL: forwardSQL,
		ActionKind: migrationkinds.Unknown,
	}, false
}

// GetForwardSQL returns the forward migration SQL statement
//
// GetForwardSQL 返回正向迁移 SQL 语句
func (op *MigrationOp) GetForwardSQL() string {
	return op.ForwardSQL
}

// GetActionKind returns the action kind value
//
// GetActionKind 返回操作类型值
func (op *MigrationOp) GetActionKind() migrationkinds.Kind {
	return op.ActionKind
}

// GetActionEnum returns the enum item bound to the action kind
//
// GetActionEnum 返回与操作类型绑定的枚举项
func (op *MigrationOp) GetActionEnum() *migrationkinds.Enum {
	return migrationkinds.GetEnums().MustGet(op.ActionKind)
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

// GetForwardSQLs extracts forward SQL statements from each migration operation
// Returns slice of SQL strings in execution sequence
//
// GetForwardSQLs 从每个迁移操作中提取正向 SQL 语句
// 按执行顺序返回 SQL 字符串切片
func (ops MigrationOps) GetForwardSQLs() []string {
	var sqs = make([]string, 0, len(ops))
	for _, op := range ops {
		sqs = append(sqs, op.ForwardSQL)
	}
	return sqs
}

// GetForwardScript generates complete forward migration script with semicolons
// Combines each forward SQL statement into executable script format
//
// GetForwardScript 生成带分号的完整正向迁移脚本
// 将每个正向 SQL 语句组合成可执行的脚本格式
func (ops MigrationOps) GetForwardScript() string {
	var sqs = make([]string, 0, len(ops))
	for _, op := range ops {
		sqs = append(sqs, op.GetForwardSQL()+";")
	}
	res := strings.Join(sqs, "\n\n")
	if len(res) > 0 {
		res += "\n"
	}
	return res
}
