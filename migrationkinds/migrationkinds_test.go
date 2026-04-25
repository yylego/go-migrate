package migrationkinds_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/go-migrate/migrationkinds"
)

// TestKindConstants verifies the raw string of one Kind constant
//
// TestKindConstants 验证一个 Kind 常量的底层字符串
func TestKindConstants(t *testing.T) {
	require.Equal(t, migrationkinds.Kind("CREATE_TABLE"), migrationkinds.CreateTable)
}

// TestEnumsList ensures the List sequence has at least the Unknown kind first
//
// TestEnumsList 确保 List 次序中第一项是 Unknown
func TestEnumsList(t *testing.T) {
	got := migrationkinds.GetEnums().List()
	require.NotEmpty(t, got)
	require.Equal(t, migrationkinds.Unknown, got[0])
}

// TestEnumsListValid checks the valid list skips the default Unknown kind
//
// TestEnumsListValid 检查有效列表跳过默认的 Unknown 类型
func TestEnumsListValid(t *testing.T) {
	got := migrationkinds.GetEnums().ListValid()
	require.NotContains(t, got, migrationkinds.Unknown)
	require.Len(t, got, len(migrationkinds.GetEnums().List())-1)
}

// TestEnumsMustGet ensures MustGet round-trips each registered kind
//
// TestEnumsMustGet 确保 MustGet 对每个已注册的 kind 能往返查到
func TestEnumsMustGet(t *testing.T) {
	for _, kind := range migrationkinds.GetEnums().List() {
		got := migrationkinds.GetEnums().MustGet(kind)
		require.NotNil(t, got)
		require.Equal(t, kind, got.Code())
		require.NotNil(t, got.Meta())
	}
}

// TestEnumsDefault verifies Unknown is configured as the default fallback
//
// TestEnumsDefault 验证 Unknown 被配置为默认兜底
func TestEnumsDefault(t *testing.T) {
	require.Equal(t, migrationkinds.Unknown, migrationkinds.GetEnums().GetDefaultCode())
}
