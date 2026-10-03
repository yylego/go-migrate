package main

import (
	"log"
	"math/rand/v2"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/spf13/cobra"
	"github.com/yylego/go-migrate/cobramigration"
	"github.com/yylego/go-migrate/internal/demos/demo1x/internal/models"
	"github.com/yylego/go-migrate/migrationparam"
	"github.com/yylego/go-migrate/migrationstate"
	"github.com/yylego/go-migrate/newmigrate"
	"github.com/yylego/must"
	"github.com/yylego/rese"
	"github.com/yylego/runpath"
	"github.com/yylego/zaplog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	var debugMode bool

	var rootCmd = &cobra.Command{
		Use:   "main",
		Short: "main",
		Long:  "main",
		Args:  cobra.NoArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			migrationparam.SetDebugMode(debugMode)
		},
	}
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable debug mode")

	cfg := &MySQLConfig{
		Dsn: "root:123456@tcp(localhost:3306)/yylego_migrate_demo1x?charset=utf8mb4&parseTime=true&multiStatements=true",
	}
	scriptsInRoot := runpath.PARENT.Join("scripts")

	// Migration connection with on-demand initialization and unified resource management
	// 迁移连接，支持延迟初始化和统一资源管理
	param := migrationparam.NewMigrationParam(
		func() *gorm.DB {
			return newGormDB(cfg)
		},
		func(db *gorm.DB) *migrate.Migrate {
			conn := rese.P1(db.DB())
			migrationDB := rese.V1(mysqlmigrate.WithInstance(conn, &mysqlmigrate.Config{}))
			return rese.P1(newmigrate.NewWithScriptsAndDatabase(
				&newmigrate.ScriptsAndDatabaseParam{
					ScriptsInRoot:    scriptsInRoot,
					DatabaseName:     "mysql",
					DatabaseInstance: migrationDB,
				},
			))
		},
	)

	// Random version objects to simulate different development stages
	// 随机版本对象，模拟不同开发阶段的迁移场景
	objects := []any{
		randomSample(&models.AccountV1{}, &models.AccountV2{}, &models.AccountV3{}),
		randomSample(&models.InfoV1{}, &models.InfoV2{}, &models.InfoV3{}),
	}

	rootCmd.AddCommand(cobramigration.NewMigrateCmd(param))
	rootCmd.AddCommand(migrationstate.NewStatusCmd(&migrationstate.Config{
		Param:       param,
		ScriptsPath: scriptsInRoot,
		Objects:     objects,
	}))

	must.Done(rootCmd.Execute())
}

func randomSample(objects ...interface{}) any {
	must.Have(objects)
	idx := rand.IntN(len(objects))
	return objects[idx]
}

type MySQLConfig struct {
	Dsn string
}

func newGormDB(cfg *MySQLConfig) *gorm.DB {
	db := rese.P1(gorm.Open(
		mysql.Open(cfg.Dsn),
		&gorm.Config{
			Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Info, //设置级别
				IgnoreRecordNotFoundError: true,        //有些场景就是需要查不到时才创建，因此查询不到时不算错误
				Colorful:                  true,
			}),
			TranslateError: true,
		},
	))
	conn := rese.P1(db.DB())
	conn.SetConnMaxIdleTime(60 * time.Second)
	conn.SetMaxIdleConns(500)

	zaplog.SUG.Debugln("正在检查数据库连接")
	must.Done(conn.Ping())
	zaplog.SUG.Debugln("已经检查数据库连接")
	return db
}
