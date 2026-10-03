package migrationstate_test

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

	mig, cleanup := param.GetMigration()
	defer cleanup()

	db, cleanup2 := param.GetDB()
	defer cleanup2()

	status, e := migrationstate.GetStatus(db, mig, scriptsPath, []any{&Student{}})
	require.NoError(t, e)

	fmt.Println()
	fmt.Println("==== END OF NOISE ====")
	fmt.Println("==== STATUS RESULT BELOW ====")
	migrationstate.ShowStatus(status)
}

// writeScript writes content to the given path, failing the test on errors
//
// writeScript 把内容写入指定路径，出错时让测试失败
func writeScript(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}
