// Package newmigrate: builds a golang-migrate instance from scripts / embed.FS against a target database
//
// newmigrate: 从脚本 / embed.FS 针对目标数据库构建 golang-migrate 实例
package newmigrate

import (
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/yylego/erero"
	"github.com/yylego/must"
	"github.com/yylego/rese"
)

func init() {
	must.Full(&file.File{}) // Enable file source registration // 启用文件源注册
}

// ScriptsAndDBSourceParam contains configuration using file-based migration with database connection string
// Supports migration from file system with database connection URL
// Used in simple setup scenarios when database connection string is available
//
// ScriptsAndDBSourceParam 包含基于文件的迁移配置和数据库连接字符串
// 支持从本地文件系统进行迁移，使用数据库连接 URL
// 用于可获得数据库连接字符串的简单设置场景
type ScriptsAndDBSourceParam struct {
	ScriptsInRoot string // Path to migration scripts DIR // 迁移脚本 DIR 路径
	ConnectSource string // Database connection string // 数据库连接字符串
}

// NewWithScriptsAndDBSource creates migration instance using file system scripts and database connection string
// Generic type param T enforces database.Driver compliance and triggers registration
// Supports multiple database types through golang-migrate
// Returns configured migration instance prepared to execute
//
// Supported database drivers:
//   - sqlite3.Sqlite from "github.com/golang-migrate/migrate/v4/database/sqlite3"
//   - mysql.Mysql from "github.com/golang-migrate/migrate/v4/database/mysql"
//   - postgres.Postgres from "github.com/golang-migrate/migrate/v4/database/postgres"
//
// Usage examples:
//   - migration, err := NewWithScriptsAndDBSource[*sqlite3.Sqlite](param)
//   - migration, err := NewWithScriptsAndDBSource[*mysql.Mysql](param)
//   - migration, err := NewWithScriptsAndDBSource[*postgres.Postgres](param)
//
// NewWithScriptsAndDBSource 使用文件系统脚本和数据库连接字符串创建迁移实例
// 泛型类型参数 T 强制数据库驱动接口兼容性并触发驱动注册
// 通过 golang-migrate 驱动系统支持多种数据库类型
// 返回已配置的迁移实例，可用于执行
func NewWithScriptsAndDBSource[T database.Driver](param *ScriptsAndDBSourceParam) (*migrate.Migrate, error) {
	sourceURL := "file://" + param.ScriptsInRoot
	migration, err := migrate.New(
		sourceURL,
		param.ConnectSource,
	)
	if err != nil {
		return nil, erero.Wro(err)
	}
	return migration, nil
}

// ScriptsAndDatabaseParam configures file-based migration with a database.Driver instance
// Provides database connection management in advanced configuration scenarios
// Enables custom database setup and connection management
//
// ScriptsAndDatabaseParam 包含基于文件的迁移配置和数据库驱动实例
// 为高级配置场景提供直接的数据库驱动控制
// 支持自定义数据库设置和连接管理
type ScriptsAndDatabaseParam struct {
	ScriptsInRoot    string          // Path to migration scripts DIR // 迁移脚本 DIR 路径
	DatabaseName     string          // Database name ID // 数据库名称标识
	DatabaseInstance database.Driver // Database migration instance // 数据库迁移驱动实例
}

// NewWithScriptsAndDatabase creates migrations using file system scripts and a database.Driver instance
// Provides database connection management with file-based migration scripts
// Returns configured migration instance prepared to execute
//
// NewWithScriptsAndDatabase 使用文件系统脚本和数据库驱动实例创建迁移实例
// 通过基于文件的迁移脚本提供直接的数据库驱动控制
// 返回已配置的迁移实例，可运行
func NewWithScriptsAndDatabase(param *ScriptsAndDatabaseParam) (*migrate.Migrate, error) {
	sourceURL := "file://" + param.ScriptsInRoot
	migration, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		param.DatabaseName,
		param.DatabaseInstance,
	)
	if err != nil {
		return nil, erero.Wro(err)
	}
	return migration, nil
}

// EmbedFsAndDatabaseParam configures a migration whose scripts live in an embed.FS
// The scripts ship inside the built program, so no separate files are needed at runtime
//
// EmbedFsAndDatabaseParam 配置迁移脚本放在 embed.FS 里的迁移
// 脚本随二进制一起打包，运行时无需额外的脚本文件
type EmbedFsAndDatabaseParam struct {
	MigrationsFS     *embed.FS       // Embedded file system with migrations // 包含迁移的嵌入文件系统
	EmbedPath        string          // Path within embedded FS // 嵌入 FS 中的路径
	DatabaseName     string          // Database name ID // 数据库名称标识
	DatabaseInstance database.Driver // Database migration instance // 数据库迁移驱动实例
}

// NewWithEmbedFsAndDatabase builds a migration whose scripts come from an embed.FS
//
// NewWithEmbedFsAndDatabase 用 embed.FS 里的脚本构建迁移实例
func NewWithEmbedFsAndDatabase(param *EmbedFsAndDatabaseParam) (*migrate.Migrate, error) {
	const sourceName = "iofs"
	// Reference: https://github.com/golang-migrate/migrate/blob/278833935c12dda022b1355f33a897d895501c45/source/iofs/example_test.go#L22
	// 详情参考: https://github.com/golang-migrate/migrate/blob/278833935c12dda022b1355f33a897d895501c45/source/iofs/example_test.go#L22
	migration, err := migrate.NewWithInstance(
		sourceName, // Fixed iofs type // 固定的 iofs 类型
		rese.V1(iofs.New(param.MigrationsFS, param.EmbedPath)), // Initialize iofs source // 初始化 iofs 源
		param.DatabaseName,
		param.DatabaseInstance,
	)
	if err != nil {
		return nil, erero.Wro(err)
	}
	return migration, nil
}
