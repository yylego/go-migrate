// Package migrationkinds defines the enum of database migration operation kinds
// Each kind keeps its forward SQL substring and reverse SQL keyword via metadata
//
// migrationkinds 定义数据库迁移操作类型的枚举
// 每种类型通过元数据持有正向 SQL 子串和反向 SQL 关键字
package migrationkinds

import (
	"github.com/yylego/enum"
)

// Kind identifies one kind of database migration operation
//
// Kind 标识一种数据库迁移操作类型
type Kind string

const (
	Unknown           Kind = "UNKNOWN"
	CreateTable       Kind = "CREATE_TABLE"
	AlterTable        Kind = "ALTER_TABLE"
	AddColumn         Kind = "ADD_COLUMN"
	AddIndex          Kind = "ADD_INDEX"
	CreateIndex       Kind = "CREATE_INDEX"
	CreateUniqueIndex Kind = "CREATE_UNIQUE_INDEX"
)

// Meta holds the SQL keyword pattern of a migration kind
// ReverseSubstr is kept on hand to support upcoming scripting needs
//
// Meta 持有迁移类型的 SQL 关键字模式
// ReverseSubstr 保留供未来脚本需求使用
type Meta struct {
	ForwardSubstr string // SQL substring matching the forward operation // 正向操作匹配 SQL 子串
	ReverseSubstr string // SQL keyword of the reverse operation // 反向操作 SQL 关键字
}

// Enum binds a Kind with its Meta
//
// Enum 把 Kind 与元数据进行关联
type Enum = enum.Enum[Kind, *Meta]

// enums indexes each migration kind with metadata
//
// enums 以 kind 索引每种迁移类型并关联元数据
var enums = enum.NewEnums(
	enum.NewEnumWithMeta(Unknown, &Meta{ForwardSubstr: "UNKNOWN", ReverseSubstr: "UNKNOWN"}),
	enum.NewEnumWithMeta(CreateTable, &Meta{ForwardSubstr: "CREATE TABLE", ReverseSubstr: "DROP TABLE"}),
	enum.NewEnumWithMeta(AlterTable, &Meta{ForwardSubstr: "ALTER TABLE", ReverseSubstr: "ALTER TABLE"}),
	enum.NewEnumWithMeta(AddColumn, &Meta{ForwardSubstr: "ADD COLUMN", ReverseSubstr: "DROP COLUMN"}),
	enum.NewEnumWithMeta(AddIndex, &Meta{ForwardSubstr: "ADD INDEX", ReverseSubstr: "DROP INDEX"}),
	enum.NewEnumWithMeta(CreateIndex, &Meta{ForwardSubstr: "CREATE INDEX", ReverseSubstr: "DROP INDEX"}),
	enum.NewEnumWithMeta(CreateUniqueIndex, &Meta{ForwardSubstr: "CREATE UNIQUE INDEX", ReverseSubstr: "DROP INDEX"}),
)

// GetEnums exposes the underlying enum collection
// Use this to fetch enum items via MustGet, ListValid, etc.
//
// GetEnums 暴露底层枚举集合
// 通过它可使用 MustGet 获取枚举项或 List 进行迭代
func GetEnums() *enum.Enums[Kind, *Meta] {
	return enums
}
