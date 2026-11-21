// Package migrationparam: holds the database and migration, each built on first access
// Releases both via cleanup once the operations complete
//
// migrationparam: 持有数据库和迁移实例，均在首次访问时才创建
// 操作完成后通过 cleanup 一起释放
package migrationparam

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/yylego/must"
	"github.com/yylego/rese"
	"gorm.io/gorm"
)

// MigrationParam holds the database and migration behind on-demand build funcs
// It creates each on first access and releases both on cleanup
//
// MigrationParam 用按需构建函数持有数据库和迁移实例
// 首次访问时创建，cleanup 时一起释放
type MigrationParam struct {
	newDB        func() *gorm.DB // Builds the database connection on demand // 按需创建数据库连接的工厂函数
	db           *gorm.DB
	newMigration func(db *gorm.DB) *migrate.Migrate // Builds the migration from a shared connection // 接受共享数据库连接的工厂函数
	migration    *migrate.Migrate
}

// NewMigrationParam creates param with database and migration construction functions
// Uses delayed initialization to create connections when needed
//
// NewMigrationParam 使用数据库和迁移工厂函数创建参数
// 使用延迟初始化在需要时创建连接
func NewMigrationParam(newDB func() *gorm.DB, newMigration func(db *gorm.DB) *migrate.Migrate) *MigrationParam {
	return &MigrationParam{
		newDB:        newDB,
		newMigration: newMigration,
	}
}

// DB returns database connection with cleanup function
// Creates connection on first access using delayed initialization
//
// DB 返回数据库连接和清理函数
// 使用延迟初始化在首次访问时创建连接
func (p *MigrationParam) DB() (*gorm.DB, func()) {
	return p.sharedDB(), p.cleanup
}

// Migration returns migration instance with cleanup function
// Creates migration on first access using delayed initialization; it shares the connection of DB
//
// Migration 返回迁移实例和清理函数
// 使用延迟初始化在首次访问时创建迁移，和 DB 共用同一个连接
func (p *MigrationParam) Migration() (*migrate.Migrate, func()) {
	if p.migration == nil {
		p.migration = p.newMigration(p.sharedDB())
	}
	return p.migration, p.cleanup
}

// sharedDB builds the connection on first access; cleanup closes it along with the migration
//
// sharedDB 首次访问时创建连接；cleanup 会把它和迁移实例一起关闭
func (p *MigrationParam) sharedDB() *gorm.DB {
	if p.db == nil {
		p.db = p.newDB()
	}
	return p.db
}

// cleanup releases resources once migration operations complete
// Closes the migration instance and database connection; failures cause panic
//
// cleanup 在迁移操作完成后释放资源
// 正确关闭迁移实例和数据库连接
// 关闭失败时抛出异常
func (p *MigrationParam) cleanup() {
	// Close migration instance
	// 关闭迁移实例
	if p.migration != nil {
		err1, err2 := p.migration.Close()
		must.Done(err1)
		must.Done(err2)
		p.migration = nil
	}

	// Close database connection
	// 关闭数据库连接
	if p.db != nil {
		must.Done(rese.P1(p.db.DB()).Close())
		p.db = nil
	}
}
