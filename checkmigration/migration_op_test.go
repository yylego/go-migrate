package checkmigration_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_extractTableNameFromCreateTable(t *testing.T) {
	t.Run("case-1", func(t *testing.T) {
		statement := "CREATE TABLE `users` (`id` integer PRIMARY KEY AUTOINCREMENT,`name` text,`code` text,CONSTRAINT `uni_users_code` UNIQUE (`code`))"
		tableName := extractTableNameFromCreateTable(statement)
		t.Log(tableName)
		require.Equal(t, "users", tableName)
	})
	t.Run("case-2", func(t *testing.T) {
		statement := "CREATE TABLE `infos` (`code` varchar(255) COMMENT '编码(通用缩写)', `name` varchar(255) COMMENT '名称(英文名称)', PRIMARY KEY (`code`))"
		tableName := extractTableNameFromCreateTable(statement)
		t.Log(tableName)
		require.Equal(t, "infos", tableName)
	})
}

func TestExtractAddedColumn(t *testing.T) {
	t.Run("case-1", func(t *testing.T) {
		table, column := extractAddedColumn("ALTER TABLE `users` ADD `age` bigint")
		require.Equal(t, "users", table)
		require.Equal(t, "age", column)
	})

	t.Run("case-2", func(t *testing.T) {
		table, column := extractAddedColumn("ALTER TABLE `users` ADD `from` varchar(255)")
		require.Equal(t, "users", table)
		require.Equal(t, "from", column)
	})

	t.Run("case-3", func(t *testing.T) {
		table, column := extractAddedColumn("ALTER TABLE `users` ADD `student_no` varchar(255)")
		require.Equal(t, "users", table)
		require.Equal(t, "student_no", column)
	})

	t.Run("case-4", func(t *testing.T) {
		table, column := extractAddedColumn("ALTER TABLE `users` ADD `rank` integer")
		require.Equal(t, "users", table)
		require.Equal(t, "rank", column)
	})
}

func Test_extractIndexAndTableFromCreateIndex(t *testing.T) {
	t.Run("case-1", func(t *testing.T) {
		index, table := extractIndexAndTableFromCreateIndex("CREATE UNIQUE INDEX `idx_users_student_no` ON `users`(`student_no`)")
		require.Equal(t, "idx_users_student_no", index)
		require.Equal(t, "users", table)
	})

	t.Run("case-2", func(t *testing.T) {
		index, table := extractIndexAndTableFromCreateIndex("CREATE INDEX `idx_users_rank` ON `users`(`rank`)")
		require.Equal(t, "idx_users_rank", index)
		require.Equal(t, "users", table)
	})
}

func extractTableNameFromCreateTable(statement string) string {
	re := regexp.MustCompile(`(?i)^CREATE TABLE\s+` + "`" + `(\w+)` + "`" + `\s*\(.*\)$`)
	match := re.FindStringSubmatch(statement)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}

func extractAddedColumn(statement string) (table string, column string) {
	re := regexp.MustCompile(`(?i)^ALTER TABLE\s+` + "`" + `(\w+)` + "`" + `\s+ADD(?: COLUMN)?\s+` + "`" + `(\w+)` + "`" + `\s+\w+`)
	match := re.FindStringSubmatch(statement)
	if len(match) >= 3 {
		return match[1], match[2]
	}
	return "", ""
}

func extractIndexAndTableFromCreateIndex(statement string) (index string, table string) {
	re := regexp.MustCompile(`(?i)^CREATE(?: UNIQUE)? INDEX\s+` + "`" + `(\w+)` + "`" + `\s+ON\s+` + "`" + `(\w+)` + "`")
	match := re.FindStringSubmatch(statement)
	if len(match) >= 3 {
		return match[1], match[2]
	}
	return "", ""
}
