package migrationstate_test

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/yylego/go-migrate/migrationparam"
	"github.com/yylego/go-migrate/migrationstate"
	"github.com/yylego/go-migrate/newmigrate"
	"github.com/yylego/must"
	"github.com/yylego/rese"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// debugFlag toggles migration debug mode in the test run via the -debug-mode CLI flag
//
// debugFlag 通过 -debug-mode 命令行标志切换测试运行时的迁移调试模式
var debugFlag = flag.Bool("debug-mode", false, "enable migration debug mode in tests")

// TestMain wires the -debug-mode flag into migrationparam before each test runs
//
// TestMain 在测试运行前把 -debug-mode 标志接入 migrationparam
func TestMain(m *testing.M) {
	flag.Parse()
	migrationparam.SetDebugMode(*debugFlag)
	os.Exit(m.Run())
}

// Student is a GORM struct used to drive schema diff detection
//
// Student 是用于驱动结构差异检测的简单 GORM 模型
type Student struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"size:100"`
	Mailbox string `gorm:"size:200;uniqueIndex"`
}

// TableName binds this struct to the users table
//
// TableName 把该结构绑定到 users 表
func (*Student) TableName() string { return "users" }

func TestGetStatus(t *testing.T) {
	scriptsPath := t.TempDir()
	writeScript(t, filepath.Join(scriptsPath, "00001_create_students.up.sql"),
		"CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT);")
	writeScript(t, filepath.Join(scriptsPath, "00001_create_students.down.sql"),
		"DROP TABLE users;")

	param := newSqliteParam(scriptsPath)
	mig, cleanup := param.Migration()
	defer cleanup()
	db, _ := param.DB()

	// Before the migration runs: one pending script, and the models do not match the database
	// 迁移前：有一个待执行脚本，模型和数据库有差异
	status, err := migrationstate.GetStatus(db, mig, scriptsPath, []any{&Student{}})
	require.NoError(t, err)
	require.False(t, status.HasMigrated)
	require.Equal(t, []uint{1}, status.PendingVersions)
	require.False(t, status.IsUpToDate)
	require.True(t, status.SchemaChecked)
	require.Positive(t, status.SchemaDiffCount)

	// Once the migration runs: up to date
	// 迁移后：已是最新
	require.NoError(t, mig.Up())
	status, err = migrationstate.GetStatus(db, mig, scriptsPath, []any{&Student{}})
	require.NoError(t, err)
	require.Equal(t, uint(1), status.DatabaseVersion)
	require.False(t, status.IsUnfinished)
	require.True(t, status.IsUpToDate)

	// Without models, the schema is not compared
	// 不传模型就不比对结构
	status, err = migrationstate.GetStatus(db, mig, scriptsPath, nil)
	require.NoError(t, err)
	require.False(t, status.SchemaChecked)
	require.Zero(t, status.SchemaDiffCount)
}

func TestGetStatus_Unfinished(t *testing.T) {
	// A script with broken SQL fails in the middle and leaves the migration unfinished: no pending script, yet not up to date
	// SQL 写错的脚本跑到一半失败，数据库留在未完成状态：没有待执行脚本，但也不算已是最新
	scriptsPath := t.TempDir()
	writeScript(t, filepath.Join(scriptsPath, "00001_create_students.up.sql"),
		"CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT);")
	writeScript(t, filepath.Join(scriptsPath, "00001_create_students.down.sql"),
		"DROP TABLE users;")
	writeScript(t, filepath.Join(scriptsPath, "00002_add_age.up.sql"),
		"ALTER TABLE users ADD COLUMN;")
	writeScript(t, filepath.Join(scriptsPath, "00002_add_age.down.sql"),
		"ALTER TABLE users DROP COLUMN age;")

	param := newSqliteParam(scriptsPath)
	mig, cleanup := param.Migration()
	defer cleanup()
	db, _ := param.DB()

	err := mig.Up()
	require.Error(t, err)
	require.NotErrorIs(t, err, migrate.ErrNoChange)

	status, err := migrationstate.GetStatus(db, mig, scriptsPath, []any{&Student{}})
	require.NoError(t, err)
	require.Equal(t, uint(2), status.DatabaseVersion)
	require.True(t, status.IsUnfinished)
	require.Zero(t, status.PendingCount)
	require.False(t, status.IsUpToDate)
}

// TestShowStatus_NoiseDemo runs the status command end-to-end against an in-memory SQLite database
// The test prints raw output to expose noise from nested functions
//
// TestShowStatus_NoiseDemo 针对内存 SQLite 数据库端到端运行 status 命令
// 测试打印原始输出，使调用链中的任何噪音都能被看到
func TestShowStatus_NoiseDemo(t *testing.T) {
	scriptsPath := t.TempDir()
	writeScript(t, filepath.Join(scriptsPath, "00001_create_students.up.sql"),
		"CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT);")
	writeScript(t, filepath.Join(scriptsPath, "00001_create_students.down.sql"),
		"DROP TABLE users;")

	dsn := fmt.Sprintf("file:db-%s?mode=memory&cache=shared", uuid.New().String())

	param := migrationparam.NewMigrationParam(
		func() *gorm.DB {
			db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{
				Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
					SlowThreshold:             200 * time.Millisecond,
					LogLevel:                  logger.Silent,
					IgnoreRecordNotFoundError: true,
					Colorful:                  true,
				}),
			}))
			must.Done(rese.P1(db.DB()).Ping())
			return db
		},
		func(db *gorm.DB) *migrate.Migrate {
			conn := rese.P1(db.DB())
			migrationDB := rese.V1(sqlite3.WithInstance(conn, &sqlite3.Config{}))
			return rese.P1(newmigrate.NewWithScriptsAndDatabase(
				&newmigrate.ScriptsAndDatabaseParam{
					ScriptsInRoot:    scriptsPath,
					DatabaseName:     "sqlite3",
					DatabaseInstance: migrationDB,
				},
			))
		},
	)

	mig, cleanup := param.Migration()
	defer cleanup()

	db, cleanup2 := param.DB()
	defer cleanup2()

	status, e := migrationstate.GetStatus(db, mig, scriptsPath, []any{&Student{}})
	require.NoError(t, e)

	fmt.Println()
	fmt.Println("==== END OF NOISE ====")
	fmt.Println("==== STATUS RESULT BELOW ====")
	migrationstate.ShowStatus(status)
}

func TestShowStatus(t *testing.T) {
	t.Run("unfinished", func(t *testing.T) {
		// No pending script, but the last migration did not finish: the output does not claim up to date
		// 没有待执行脚本，但上次迁移没有跑完：输出里不说已是最新
		output := captureStatusOutput(t, &migrationstate.Status{
			DatabaseVersion:     2,
			IsUnfinished:        true,
			HasMigrated:         true,
			LatestScriptVersion: 2,
			ScriptCount:         2,
		})
		require.Equal(t, []string{
			"=== Migration Status ===",
			"Database Version: 2 (DIRTY)",
			"Scripts Latest: 2 (2 scripts)",
			"Pending Migrations: 0 (not up to date - the last migration did not finish)",
			"Schema Differences: (not checked - no models given)",
		}, output)
	})

	t.Run("up-to-date", func(t *testing.T) {
		output := captureStatusOutput(t, &migrationstate.Status{
			DatabaseVersion:     2,
			HasMigrated:         true,
			LatestScriptVersion: 2,
			ScriptCount:         2,
			IsUpToDate:          true,
			SchemaChecked:       true,
		})
		require.Equal(t, []string{
			"=== Migration Status ===",
			"Database Version: 2",
			"Scripts Latest: 2 (2 scripts)",
			"Pending Migrations: 0 (up to date)",
			"Schema Differences: 0 (Models match database)",
		}, output)
	})

	t.Run("pending-with-diff", func(t *testing.T) {
		output := captureStatusOutput(t, &migrationstate.Status{
			LatestScriptVersion: 1,
			ScriptCount:         1,
			PendingCount:        1,
			PendingVersions:     []uint{1},
			SchemaChecked:       true,
			SchemaDiffCount:     1,
			SchemaDiffSQLs:      []string{"CREATE TABLE `users` (`id` integer)"},
		})
		require.Equal(t, []string{
			"=== Migration Status ===",
			"Database Version: (none - no migration records)",
			"Scripts Latest: 1 (1 scripts)",
			"Pending Migrations: 1",
			"  Versions: [1]",
			"Schema Differences: 1",
			"  (GORM detected schema differences; inspect migration scripts before applying changes)",
			"-> 1 -> CREATE TABLE `users` (`id` integer)",
		}, output)
	})
}

// captureStatusOutput runs ShowStatus and returns the lines it prints, with the colors and the separating lines taken out
//
// captureStatusOutput 运行 ShowStatus 并返回打印的各行，去掉颜色和分隔线
func captureStatusOutput(t *testing.T, status *migrationstate.Status) []string {
	t.Helper()
	pipeIn, pipeOut, err := os.Pipe()
	require.NoError(t, err)
	stdout := os.Stdout
	os.Stdout = pipeOut
	migrationstate.ShowStatus(status)
	os.Stdout = stdout
	require.NoError(t, pipeOut.Close())
	content, err := io.ReadAll(pipeIn)
	require.NoError(t, err)
	require.NoError(t, pipeIn.Close())

	plain := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(string(content), "")
	var lines []string
	for line := range strings.SplitSeq(strings.TrimSuffix(plain, "\n"), "\n") {
		if strings.Trim(line, "-") != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func TestNewStatusCmd(t *testing.T) {
	cmd := migrationstate.NewStatusCmd(&migrationstate.Config{
		Param:       newSqliteParam(t.TempDir()),
		ScriptsPath: t.TempDir(),
		Objects:     []any{&Student{}},
	})
	require.Equal(t, "status", cmd.Name())
	require.Error(t, cmd.Args(cmd, []string{"unexpected"}))
	require.NoError(t, cmd.Args(cmd, nil))
}

// newSqliteParam builds a MigrationParam on a fresh in-memory SQLite database
//
// newSqliteParam 在一个新的内存 SQLite 数据库上构建 MigrationParam
func newSqliteParam(scriptsPath string) *migrationparam.MigrationParam {
	dsn := fmt.Sprintf("file:db-%s?mode=memory&cache=shared", uuid.New().String())
	return migrationparam.NewMigrationParam(
		func() *gorm.DB {
			return rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent),
			}))
		},
		func(db *gorm.DB) *migrate.Migrate {
			migrationDB := rese.V1(sqlite3.WithInstance(rese.P1(db.DB()), &sqlite3.Config{}))
			return rese.P1(newmigrate.NewWithScriptsAndDatabase(&newmigrate.ScriptsAndDatabaseParam{
				ScriptsInRoot:    scriptsPath,
				DatabaseName:     "sqlite3",
				DatabaseInstance: migrationDB,
			}))
		},
	)
}

// writeScript writes content to the given path, failing the test on errors
//
// writeScript 把内容写入指定路径，出错时让测试失败
func writeScript(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}
