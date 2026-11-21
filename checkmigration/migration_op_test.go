package checkmigration_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/go-migrate/checkmigration"
)

func TestNewMigrationOp(t *testing.T) {
	t.Run("postgres-comment", func(t *testing.T) {
		op, match := checkmigration.NewMigrationOp(`COMMENT ON COLUMN "users"."name" IS '名称'`)
		require.True(t, match)
		require.Equal(t, `COMMENT ON COLUMN "users"."name" IS '名称'`, op.ForwardSQL)
	})

	t.Run("postgres-sequence", func(t *testing.T) {
		_, match := checkmigration.NewMigrationOp(`CREATE SEQUENCE IF NOT EXISTS "users_id_seq" AS bigint`)
		require.True(t, match)
		_, match = checkmigration.NewMigrationOp(`ALTER SEQUENCE "users_id_seq" OWNED BY "users"."id"`)
		require.True(t, match)
		_, match = checkmigration.NewMigrationOp(`DROP SEQUENCE IF EXISTS "users_id_seq"`)
		require.True(t, match)
	})

	t.Run("schema-probe-skipped", func(t *testing.T) {
		_, match := checkmigration.NewMigrationOp(`SELECT description FROM pg_catalog.pg_description`)
		require.False(t, match)
	})
}

func TestMigrationOps_ForwardSQLs(t *testing.T) {
	createOp, match := checkmigration.NewMigrationOp("CREATE TABLE `users` (`id` integer)")
	require.True(t, match)
	addColumnOp, match := checkmigration.NewMigrationOp("ALTER TABLE `users` ADD `age` bigint")
	require.True(t, match)

	ops := checkmigration.MigrationOps{createOp, addColumnOp}
	require.Equal(t, []string{"CREATE TABLE `users` (`id` integer)", "ALTER TABLE `users` ADD `age` bigint"}, ops.ForwardSQLs())
}

func TestMigrationOps_ForwardScript(t *testing.T) {
	createOp, match := checkmigration.NewMigrationOp("CREATE TABLE `users` (`id` integer)")
	require.True(t, match)
	addColumnOp, match := checkmigration.NewMigrationOp("ALTER TABLE `users` ADD `age` bigint")
	require.True(t, match)

	ops := checkmigration.MigrationOps{createOp, addColumnOp}
	require.Equal(t, "CREATE TABLE `users` (`id` integer);\n\nALTER TABLE `users` ADD `age` bigint;\n", ops.ForwardScript())
	require.Empty(t, checkmigration.MigrationOps{}.ForwardScript())
}

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
