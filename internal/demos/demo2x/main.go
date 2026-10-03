package main

import (
	"log"
	"math/rand/v2"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	postgresmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/spf13/cobra"
	"github.com/yylego/go-migrate/cobramigration"
	"github.com/yylego/go-migrate/internal/demos/demo2x/internal/models"
	"github.com/yylego/go-migrate/migrationparam"
	"github.com/yylego/go-migrate/migrationstate"
	"github.com/yylego/go-migrate/newmigrate"
	"github.com/yylego/must"
	"github.com/yylego/rese"
	"github.com/yylego/runpath"
	"github.com/yylego/zaplog"
	"gorm.io/driver/postgres"
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

	// docker run -d --name=postgres -e POSTGRES_PASSWORD=123456 -p 5432:5432 postgres
	cfg := &PostgresConfig{
		Dsn: "postgres://postgres:123456@localhost:5432/yylego_migrate_demo2x?sslmode=disable&TimeZone=UTC",
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
			migrationDB := rese.V1(postgresmigrate.WithInstance(conn, &postgresmigrate.Config{}))
			return rese.P1(newmigrate.NewWithScriptsAndDatabase(
				&newmigrate.ScriptsAndDatabaseParam{
					ScriptsInRoot:    scriptsInRoot,
					DatabaseName:     "postgres",
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

type PostgresConfig struct {
	Dsn string
}

func newGormDB(cfg *PostgresConfig) *gorm.DB {
	db := rese.P1(gorm.Open(
		postgres.Open(cfg.Dsn),
		&gorm.Config{
			Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
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
