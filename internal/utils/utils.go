// Package utils handles the results of golang-migrate commands
//
// utils: 处理 golang-migrate 命令的执行结果
package utils

import (
	"errors"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/yylego/tint"
	"github.com/yylego/zaplog"
)

// HandleCause handles the result of a migration command: logs the expected outcomes, and panics on unexpected failures
// Nothing to run and no version yet are fine; os.ErrNotExist gets a warning, since it has two sources:
// a single step (Steps) has no script left in its direction; the scripts path holds no matching script
// Uses distinct colors to indicate failures and success
//
// HandleCause 处理迁移命令的执行结果：预期内的情况打日志，意外失败直接 panic
// 没有要执行的、还没有版本都属正常；os.ErrNotExist 给警告，它有两种来源：
// 单步迁移（Steps）在这个方向上已经没有脚本；脚本目录里找不到对应的脚本
// 使用颜色编码输出来区分不同的错误类型和成功状态
func HandleCause(cause error) {
	if cause != nil {
		if errors.Is(cause, migrate.ErrNoChange) {
			zaplog.SUG.Debugln(tint.BLUE.Sprint("NO MIGRATION FILES TO RUN"))
		} else if errors.Is(cause, migrate.ErrNilVersion) {
			zaplog.SUG.Debugln(tint.BLUE.Sprint("NO VERSION IN VERSION-TABLE(schema_migrations)"))
		} else if errors.Is(cause, os.ErrNotExist) {
			zaplog.SUG.Warnln(tint.YELLOW.Sprint("MIGRATION FILE NOT FOUND: NO MORE STEPS IN THIS DIRECTION, OR THE SCRIPTS PATH IS WRONG"))
		} else {
			zaplog.SUG.Panicln(tint.RED.Sprint("MIGRATION FAILED:"), cause)
		}
		return
	}
	zaplog.SUG.Debugln(tint.GREEN.Sprint("MIGRATION SUCCESS"))
}
