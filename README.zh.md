<!-- TEMPLATE (ZH) BEGIN: BADGES -->

[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/go-migrate/release.yml?branch=main&label=BUILD)](https://github.com/yylego/go-migrate/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/go-migrate)](https://pkg.go.dev/github.com/yylego/go-migrate)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/go-migrate/main.svg)](https://coveralls.io/github/yylego/go-migrate?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.26%2B-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/go-migrate.svg)](https://github.com/yylego/go-migrate/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/go-migrate)](https://goreportcard.com/report/github.com/yylego/go-migrate)

<!-- TEMPLATE (ZH) CLOSE: BADGES -->

# go-migrate

数据库迁移工具包，集成 GORM 模型分析、迁移执行和状态管理功能。

---

## 生态系统

![go-migrate overview](assets/go-migrate-overview.svg)

![go-migrate workflow](assets/go-migrate-workflow.svg)

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->

## 英文文档

[ENGLISH README](README.md)

<!-- TEMPLATE (ZH) CLOSE: LANGUAGE NAVIGATION -->

## 核心特性

- **智能结构分析**：自动对比 GORM 模型与现有数据库结构
- **多数据库支持**：通过 golang-migrate 支持 MySQL、PostgreSQL、SQLite
- **全面 CLI 支持**：直观的 Cobra 命令覆盖所有迁移操作
- **状态检查功能**：检查数据库版本、待处理迁移和结构差异

## 核心包

| 包名             | 用途                                  |
| ---------------- | ------------------------------------- |
| `checkmigration` | 对比 GORM 模型与数据库，捕获 SQL 差异 |
| `newmigrate`     | 创建 golang-migrate 实例              |
| `migrationparam` | 迁移连接管理和调试模式控制            |
| `cobramigration` | Cobra CLI 命令 (inc/dec/batch)        |
| `migrationstate` | 检查迁移状态                          |

## 结构探测的错误处理

`checkmigration.CheckMigrate(db, objects)` 返回 `([]string, error)`，`GetMigrateOps` 返回 `(MigrationOps, error)`。探测失败不返回不完整结果，由调用方处理错误。升级原先的单返回值调用时，需要接收新增的 `error`；需要失败即 panic 的调用方可显式使用 `must.Done(err)`：

```go
sqls, err := checkmigration.CheckMigrate(db, objects)
must.Done(err)
```

这些接口通过 GORM DryRun 捕获 SQL，仍会读取数据库结构。PostgreSQL 的列信息查询使用独立的非 DryRun 会话，DDL 继续预演，不改变调用方配置。查询失败会返回错误，不会被当成“表不存在”。探测范围取决于 GORM AutoMigrate，不能代替完整的 schema diff，探测失败也不代表结构一致。

## 安装

```bash
go get github.com/yylego/go-migrate
```

## 快速开始

### 1. 定义 GORM 模型

```go
type Account struct {
    ID   uint   `gorm:"primarykey"`
    Name string `gorm:"size:100"`
    Age  int
}
```

### 2. 配置 CLI 工具

```go
package main

import (
    "github.com/yylego/go-migrate/cobramigration"
    "github.com/yylego/go-migrate/migrationparam"
    "github.com/yylego/go-migrate/migrationstate"
    "github.com/yylego/go-migrate/newmigrate"
    "github.com/golang-migrate/migrate/v4"
    mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
    "github.com/spf13/cobra"
    "github.com/yylego/must"
    "github.com/yylego/rese"
    "gorm.io/gorm"
)

func main() {
    scriptsPath := "./scripts"

    // MigrationParam 延迟初始化和统一资源管理
    param := migrationparam.NewMigrationParam(
        func() *gorm.DB {
            return setupDatabase() // 你的 GORM 配置
        },
        func(db *gorm.DB) *migrate.Migrate {
            conn := rese.P1(db.DB())
            migrationDB := rese.V1(mysqlmigrate.WithInstance(conn, &mysqlmigrate.Config{}))
            return rese.P1(newmigrate.NewWithScriptsAndDatabase(&newmigrate.ScriptsAndDatabaseParam{
                ScriptsInRoot:    scriptsPath,
                DatabaseName:     "mysql",
                DatabaseInstance: migrationDB,
            }))
        },
    )

    objects := []any{
        &Account{},
        &Product{},
        &Cart{},
    }

    rootCmd := &cobra.Command{Use: "app"}
    rootCmd.AddCommand(cobramigration.NewMigrateCmd(param))
    rootCmd.AddCommand(migrationstate.NewStatusCmd(&migrationstate.Config{
        Param:       param,
        ScriptsPath: scriptsPath,
        Objects:     objects,
    }))

    must.Done(rootCmd.Execute())
}
```

### 3. 常用工作流

```bash
# 步骤 1: 检查当前状态
go run main.go status

# 步骤 2: 编写迁移脚本（手动编写或借助 AI）
# 创建: scripts/000001_xxx.up.sql 和 scripts/000001_xxx.down.sql

# 步骤 3: 执行迁移
go run main.go migrate inc    # 单步执行
go run main.go migrate batch  # 执行所有待处理
```

## 由 AI 驱动的脚本编写

该项目早期版本包含内置的脚本生成（`newscripts`）和预览（`previewmigrate`）功能包，现已移除。具体改动原因：

- **AI 写脚本更准确**：Claude Code 等 AI 工具可以直接阅读 GORM 模型变更，准确生成迁移 SQL 语句。
- **AI 总是绕过工具**：即使在文档中反复强调要走工具生成脚本，AI 阅读项目时一看到 migrations 目录和已有脚本样例，就会自行推断格式直接手写新脚本，绕过工具。手写结果反而稳定准确，而工具产物时好时坏；与其混用造成不一致，不如直接放弃脚本生成功能，只保留校验部分。
- **脚本创建不依赖数据库**：原工具需要运行中的数据库，通过 GORM DryRun 模式捕获结构差异。AI 只需阅读代码即可生成脚本。
- **预览功能有局限**：MySQL 不支持 DDL 事务回滚，导致预览功能在最常用的数据库上不可靠。
- **出错有 AI 协助**：原先设计预览是为了避免迁移中途出错后需要人工逐步排查；如今出错后只需交给 AI，它能机智引导快速回滚到出错前的状态。考虑到迁移通常先在本地或测试环境演练、失败本就少见，加上 AI 的处理能力又足够强，预览这一环节就显得多余。

当前架构聚焦于 AI 无法替代的能力：**结构验证、迁移执行和状态管理**。脚本编写交由 AI 或手动完成。

## CLI 命令清单

| 命令            | 描述                                 |
| --------------- | ------------------------------------ |
| `status`        | 显示数据库版本、待处理迁移、结构差异 |
| `migrate`       | 显示当前迁移版本                     |
| `migrate inc`   | 执行下一次迁移                       |
| `migrate dec`   | 回滚一次迁移                         |
| `migrate batch` | 执行所有待处理迁移                   |

## 数据库支持

通过 golang-migrate 驱动支持 MySQL、PostgreSQL、SQLite：

```go
// MySQL
import mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
migrationDB := rese.V1(mysqlmigrate.WithInstance(conn, &mysqlmigrate.Config{}))

// PostgreSQL
import postgresmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
migrationDB := rese.V1(postgresmigrate.WithInstance(conn, &postgresmigrate.Config{}))

// SQLite
import sqlite3migrate "github.com/golang-migrate/migrate/v4/database/sqlite3"
migrationDB := rese.V1(sqlite3migrate.WithInstance(conn, &sqlite3migrate.Config{}))
```

## 高级配置

### 调试模式

启用调试模式以查看详细的 SQL 捕获和迁移分析输出：

```go
import "github.com/yylego/go-migrate/migrationparam"

// 启用迁移操作的调试模式
migrationparam.SetDebugMode(true)

// 检查当前调试模式状态
if migrationparam.GetDebugMode() {
    // 调试日志已启用
}
```

### 嵌入式迁移

```go
//go:embed migrations
var migrationsFS embed.FS

migration := rese.V1(newmigrate.NewWithEmbedFsAndDatabase(&newmigrate.EmbedFsAndDatabaseParam{
    MigrationsFS:     &migrationsFS,
    EmbedPath:        "migrations",
    DatabaseName:     "mysql",
    DatabaseInstance: migrationDB,
}))
```

## 示例

参见 [internal/demos](internal/demos) 中的完整工作示例：

- [demo1x](internal/demos/demo1x)：MySQL 集成与 Makefile 命令
- [demo2x](internal/demos/demo2x)：PostgreSQL 集成与 Makefile 命令

```bash
cd internal/demos/demo1x
make STATUS       # 检查状态
make MIGRATE-INC  # 单步执行
make MIGRATE-ALL  # 执行全部
```

<!-- TEMPLATE (ZH) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 许可证类型

MIT 许可证 - 详见 [LICENSE](LICENSE)。

---

## 💬 联系与反馈

非常欢迎贡献代码！报告 BUG、建议功能、贡献代码：

- 🐛 **问题报告？** 在 GitHub 上提交问题并附上重现步骤
- 💡 **新颖思路？** 创建 issue 讨论
- 📖 **文档疑惑？** 报告问题，帮助我们完善文档
- 🚀 **需要功能？** 分享使用场景，帮助理解需求
- ⚡ **性能瓶颈？** 报告慢操作，协助解决性能问题
- 🔧 **配置困扰？** 询问复杂设置的相关问题
- 📢 **关注进展？** 关注仓库以获取新版本和功能
- 🌟 **成功案例？** 分享这个包如何改善工作流程
- 💬 **反馈意见？** 欢迎提出建议和意见

---

## 🔧 代码贡献

新代码贡献，请遵循此流程：

1. **Fork**：在 GitHub 上 Fork 仓库（使用网页界面）
2. **克隆**：克隆 Fork 的项目（`git clone https://github.com/yourname/repo-name.git`）
3. **导航**：进入克隆的项目（`cd repo-name`）
4. **分支**：创建功能分支（`git checkout -b feature/xxx`）
5. **编码**：实现您的更改并编写全面的测试
6. **测试**：（Golang 项目）确保测试通过（`go test ./...`）并遵循 Go 代码风格约定
7. **文档**：面向用户的更改需要更新文档
8. **暂存**：暂存更改（`git add .`）
9. **提交**：提交更改（`git commit -m "Add feature xxx"`）确保向后兼容的代码
10. **推送**：推送到分支（`git push origin feature/xxx`）
11. **PR**：在 GitHub 上打开 Merge Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Merge Request 和报告问题来贡献此项目。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**祝你用这个包编程愉快！** 🎉🎉🎉

<!-- TEMPLATE (ZH) CLOSE: STANDARD PROJECT FOOTER -->

---

<!-- TEMPLATE (ZH) BEGIN: GITHUB STARS -->

## GitHub 标星点赞

[![Stargazers](https://starchart.cc/yylego/go-migrate.svg?variant=adaptive)](https://starchart.cc/yylego/go-migrate)

<!-- TEMPLATE (ZH) CLOSE: GITHUB STARS -->
