// Package cobramigration: Cobra CLI commands to run golang-migrate migrations
// Covers version reporting, batch migration, and step-by-step control
//
// cobramigration: 运行 golang-migrate 迁移的 Cobra CLI 命令
// 涵盖版本查看、全量迁移、逐步迁移控制
package cobramigration

import (
	"github.com/spf13/cobra"
	"github.com/yylego/go-migrate/internal/utils"
	"github.com/yylego/go-migrate/migrationparam"
	"github.com/yylego/tint"
)

// NewMigrateCmd builds the migrate command tree with its subcommands
// Connections open just when a command runs, not while the tree is built; each run cleans them up
//
// NewMigrateCmd 构建 migrate 命令树及其子命令
// 连接在命令运行时才建立（而非构建命令树时），每次运行结束会清理
func NewMigrateCmd(param *migrationparam.MigrationParam) *cobra.Command {
	// Create root command
	var rootCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Database migration",
		Long:  "Database migration",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			migration, cleanup := param.GetMigration()
			defer cleanup()

			version, dirtyFlag, err := migration.Version()
			utils.WhistleCause(err) // panic when cause is not expected
			if dirtyFlag {
				tint.RED.ShowMessage(version, "(DIRTY)")
			} else {
				tint.GREEN.ShowMessage(version)
			}
		},
	}

	rootCmd.AddCommand(newAllCmd(param)) // Append `all` subcommand // 添加 `all` 子命令
	rootCmd.AddCommand(newIncCMD(param)) // Append `inc` subcommand // 添加 `inc` 子命令
	rootCmd.AddCommand(newDecCMD(param)) // Append `dec` subcommand // 添加 `dec` 子命令

	return rootCmd
}

// newAllCmd creates command for executing all pending migrations
// Performs complete database upgrade to latest schema version
//
// newAllCmd 创建用于执行所有待处理迁移的命令
// 将数据库升级到最新的结构版本
func newAllCmd(param *migrationparam.MigrationParam) *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "Run all migration files",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			migration, cleanup := param.GetMigration()
			defer cleanup()

			// Perform complete database upgrade
			// 执行完整的数据库升级
			utils.WhistleCause(migration.Up())
		},
	}
}

// newDecCMD creates command for rolling back one migration step
// Safely reverts database schema by one version
//
// newDecCMD 创建用于回滚一个迁移步骤的命令
// 安全地将数据库结构回退一个版本
func newDecCMD(param *migrationparam.MigrationParam) *cobra.Command {
	return &cobra.Command{
		Use:   "dec",
		Short: "Rollback one step (-1)",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			migration, cleanup := param.GetMigration()
			defer cleanup()

			// Rollback database by one migration step
			// 将数据库回滚一个迁移步骤
			utils.WhistleCause(migration.Steps(-1))
		},
	}
}

// newIncCMD creates command for executing next migration step
// Advances database schema by one version forward
//
// newIncCMD 创建用于执行下一个迁移步骤的命令
// 将数据库结构向前推进一个版本
func newIncCMD(param *migrationparam.MigrationParam) *cobra.Command {
	return &cobra.Command{
		Use:   "inc",
		Short: "Run next step (+1)",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			migration, cleanup := param.GetMigration()
			defer cleanup()

			// Execute next migration step forward
			// 向前执行下一个迁移步骤
			utils.WhistleCause(migration.Steps(+1))
		},
	}
}
