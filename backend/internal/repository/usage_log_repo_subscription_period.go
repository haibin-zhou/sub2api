package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

// subscriptionPeriodUsageRepository 聚合订阅在各自周期窗口内的 usage_logs（只读）。
type subscriptionPeriodUsageRepository struct {
	sql sqlExecutor
}

func NewSubscriptionPeriodUsageRepository(sqlDB *sql.DB) service.SubscriptionPeriodUsageRepository {
	return newSubscriptionPeriodUsageRepositoryWithSQL(sqlDB)
}

func newSubscriptionPeriodUsageRepositoryWithSQL(sqlq sqlExecutor) *subscriptionPeriodUsageRepository {
	return &subscriptionPeriodUsageRepository{sql: sqlq}
}

// 一次查询覆盖用户全部订阅：unnest 出每个订阅自己的 [period_start, period_end) 窗口，
// 再与 usage_logs 做等值 + 范围 JOIN。每个窗口行都能走 idx_usage_logs_sub_created
// (subscription_id, created_at) 的索引范围扫描（嵌套循环），user_id 仅作纵深防御过滤。
const subscriptionPeriodUsageQuery = `
	SELECT
		w.subscription_id,
		COUNT(*) AS request_count,
		COALESCE(SUM(ul.input_tokens), 0) AS input_tokens,
		COALESCE(SUM(ul.output_tokens), 0) AS output_tokens,
		COALESCE(SUM(ul.cache_creation_tokens), 0) AS cache_creation_tokens,
		COALESCE(SUM(ul.cache_read_tokens), 0) AS cache_read_tokens,
		COALESCE(SUM(ul.total_cost), 0) AS total_cost
	FROM unnest($1::bigint[], $2::timestamptz[], $3::timestamptz[]) AS w(subscription_id, period_start, period_end)
	JOIN usage_logs ul
		ON ul.subscription_id = w.subscription_id
		AND ul.created_at >= w.period_start
		AND ul.created_at < w.period_end
	WHERE ul.user_id = $4
	GROUP BY w.subscription_id
`

func (r *subscriptionPeriodUsageRepository) SumSubscriptionPeriodUsage(ctx context.Context, userID int64, windows []service.SubscriptionPeriodWindow) (map[int64]service.SubscriptionPeriodUsageTotals, error) {
	result := make(map[int64]service.SubscriptionPeriodUsageTotals, len(windows))
	if len(windows) == 0 {
		return result, nil
	}
	ids := make([]int64, len(windows))
	starts := make([]string, len(windows))
	ends := make([]string, len(windows))
	for i, w := range windows {
		ids[i] = w.SubscriptionID
		starts[i] = w.Start.UTC().Format(time.RFC3339Nano)
		ends[i] = w.End.UTC().Format(time.RFC3339Nano)
	}

	rows, err := r.sql.QueryContext(ctx, subscriptionPeriodUsageQuery, pq.Array(ids), pq.Array(starts), pq.Array(ends), userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var subID int64
		var t service.SubscriptionPeriodUsageTotals
		if err := rows.Scan(&subID, &t.RequestCount, &t.InputTokens, &t.OutputTokens, &t.CacheCreationTokens, &t.CacheReadTokens, &t.TotalCostUSD); err != nil {
			return nil, err
		}
		result[subID] = t
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
