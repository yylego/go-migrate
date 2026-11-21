package migrationparam

var debugModeOpen = false

// SetDebugMode sets the project-wide debug mode used in migration operations
// SetDebugMode 设置项目级别的迁移操作调试模式
func SetDebugMode(enable bool) {
	debugModeOpen = enable
}

// DebugMode returns the project-wide debug mode used in migration operations
// DebugMode 返回项目级别的迁移操作调试模式
func DebugMode() bool {
	return debugModeOpen
}
