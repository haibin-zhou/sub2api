package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionPeriodUsageRepository_SingleGroupedQuery(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := newSubscriptionPeriodUsageRepositoryWithSQL(db)
	s1 := time.Date(2026, 10, 22, 8, 0, 0, 0, time.UTC)
	e1 := time.Date(2026, 11, 21, 8, 0, 0, 0, time.UTC)
	s2 := time.Date(2026, 11, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	e2 := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)

	// 一次查询：unnest 出每个订阅的 [start, end) 窗口，再按 (subscription_id, created_at) 半开区间 JOIN 聚合。
	mock.ExpectQuery(`(?s)FROM unnest\(\$1::bigint\[\], \$2::timestamptz\[\], \$3::timestamptz\[\]\) AS w\(subscription_id, period_start, period_end\).*JOIN usage_logs ul\s+ON ul\.subscription_id = w\.subscription_id\s+AND ul\.created_at >= w\.period_start\s+AND ul\.created_at < w\.period_end\s+WHERE ul\.user_id = \$4\s+GROUP BY w\.subscription_id`).
		WithArgs(
			pq.Array([]int64{2, 3}),
			pq.Array([]string{"2026-10-22T08:00:00Z", "2026-10-31T16:00:00Z"}),
			pq.Array([]string{"2026-11-21T08:00:00Z", "2026-11-21T12:00:00Z"}),
			int64(42),
		).
		WillReturnRows(sqlmock.NewRows([]string{"subscription_id", "request_count", "input_tokens", "output_tokens", "cache_creation_tokens", "cache_read_tokens", "total_cost"}).
			AddRow(int64(2), int64(1753), int64(10), int64(20), int64(30), int64(40), 180.51))

	out, err := repo.SumSubscriptionPeriodUsage(context.Background(), 42, []service.SubscriptionPeriodWindow{
		{SubscriptionID: 2, Start: s1, End: e1},
		{SubscriptionID: 3, Start: s2, End: e2},
	})
	require.NoError(t, err)
	require.Equal(t, service.SubscriptionPeriodUsageTotals{RequestCount: 1753, InputTokens: 10, OutputTokens: 20, CacheCreationTokens: 30, CacheReadTokens: 40, TotalCostUSD: 180.51}, out[2])
	_, ok := out[3]
	require.False(t, ok, "subscriptions without logs are simply absent")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionPeriodUsageRepository_EmptyWindowsNoQuery(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := newSubscriptionPeriodUsageRepositoryWithSQL(db)
	out, err := repo.SumSubscriptionPeriodUsage(context.Background(), 1, nil)
	require.NoError(t, err)
	require.Empty(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}
