package service

import (
	"context"
	"time"
)

// subscriptionPeriodLength 是订阅"本周期"的长度：自 starts_at 起每 30 天为一期（滚动）。
const subscriptionPeriodLength = 30 * 24 * time.Hour

// subscriptionPeriod 计算 now 所处的 30 天周期。
// k = floor((now - startsAt) / 30d)，now 早于 startsAt 时 k = 0；
// 返回 1 基期号 k+1、周期起点 startsAt + k·30d、终点（不含）起点 + 30d，均为 UTC。
func subscriptionPeriod(startsAt, now time.Time) (index int, start, end time.Time) {
	startsAt = startsAt.UTC()
	k := int64(0)
	if elapsed := now.Sub(startsAt); elapsed > 0 {
		k = int64(elapsed / subscriptionPeriodLength)
	}
	start = startsAt.Add(time.Duration(k) * subscriptionPeriodLength)
	return int(k) + 1, start, start.Add(subscriptionPeriodLength)
}

// SubscriptionPeriodWindow 是一个订阅的聚合区间 [Start, End)。
type SubscriptionPeriodWindow struct {
	SubscriptionID int64
	Start          time.Time
	End            time.Time
}

// SubscriptionPeriodUsageTotals 是一个订阅在区间内的用量汇总；TotalCostUSD 为按官方模型价格计算的 total_cost（倍率前）。
type SubscriptionPeriodUsageTotals struct {
	RequestCount        int64
	InputTokens         int64
	OutputTokens        int64
	CacheCreationTokens int64
	CacheReadTokens     int64
	TotalCostUSD        float64
}

// SubscriptionPeriodUsageRepository 以一次查询聚合多个订阅各自区间内的 usage_logs。
// 结果仅包含有日志的订阅；userID 作为纵深防御过滤条件。
type SubscriptionPeriodUsageRepository interface {
	SumSubscriptionPeriodUsage(ctx context.Context, userID int64, windows []SubscriptionPeriodWindow) (map[int64]SubscriptionPeriodUsageTotals, error)
}

type periodActiveSubscriptionLister interface {
	ListActiveUserSubscriptions(ctx context.Context, userID int64) ([]UserSubscription, error)
}

// SubscriptionPeriodUsage 是 GET /api/v1/subscriptions/period-usage 的单项。
type SubscriptionPeriodUsage struct {
	SubscriptionID      int64     `json:"subscription_id"`
	GroupID             int64     `json:"group_id"`
	PeriodIndex         int       `json:"period_index"`
	PeriodStart         time.Time `json:"period_start"`
	PeriodEnd           time.Time `json:"period_end"`
	RequestCount        int64     `json:"request_count"`
	InputTokens         int64     `json:"input_tokens"`
	OutputTokens        int64     `json:"output_tokens"`
	CacheCreationTokens int64     `json:"cache_creation_tokens"`
	CacheReadTokens     int64     `json:"cache_read_tokens"`
	TotalCostUSD        float64   `json:"total_cost_usd"`
}

// SubscriptionPeriodUsageService 统计用户每个有效订阅在当前 30 天周期内的累计用量。
type SubscriptionPeriodUsageService struct {
	subs periodActiveSubscriptionLister
	repo SubscriptionPeriodUsageRepository
	now  func() time.Time
}

func NewSubscriptionPeriodUsageService(subs *SubscriptionService, repo SubscriptionPeriodUsageRepository) *SubscriptionPeriodUsageService {
	return NewSubscriptionPeriodUsageServiceWith(subs, repo, time.Now)
}

// NewSubscriptionPeriodUsageServiceWith 允许注入订阅列表来源与时钟（测试用）。
func NewSubscriptionPeriodUsageServiceWith(subs periodActiveSubscriptionLister, repo SubscriptionPeriodUsageRepository, now func() time.Time) *SubscriptionPeriodUsageService {
	return &SubscriptionPeriodUsageService{subs: subs, repo: repo, now: now}
}

// ListForUser 返回 userID 的每个有效订阅（全部平台）在当前周期内的用量；订阅只取自该用户自己。
func (s *SubscriptionPeriodUsageService) ListForUser(ctx context.Context, userID int64) ([]SubscriptionPeriodUsage, error) {
	subs, err := s.subs.ListActiveUserSubscriptions(ctx, userID)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	items := make([]SubscriptionPeriodUsage, 0, len(subs))
	windows := make([]SubscriptionPeriodWindow, 0, len(subs))
	for i := range subs {
		sub := &subs[i]
		if !sub.IsActive() || sub.ID <= 0 {
			continue
		}
		idx, start, end := subscriptionPeriod(sub.StartsAt, now)
		upper := end
		if now.Before(upper) {
			upper = now
		}
		items = append(items, SubscriptionPeriodUsage{
			SubscriptionID: sub.ID,
			GroupID:        sub.GroupID,
			PeriodIndex:    idx,
			PeriodStart:    start,
			PeriodEnd:      end,
		})
		windows = append(windows, SubscriptionPeriodWindow{SubscriptionID: sub.ID, Start: start, End: upper})
	}
	if len(windows) == 0 {
		return items, nil
	}
	totals, err := s.repo.SumSubscriptionPeriodUsage(ctx, userID, windows)
	if err != nil {
		return nil, err
	}
	for i := range items {
		t, ok := totals[items[i].SubscriptionID]
		if !ok {
			continue
		}
		items[i].RequestCount = t.RequestCount
		items[i].InputTokens = t.InputTokens
		items[i].OutputTokens = t.OutputTokens
		items[i].CacheCreationTokens = t.CacheCreationTokens
		items[i].CacheReadTokens = t.CacheReadTokens
		items[i].TotalCostUSD = t.TotalCostUSD
	}
	return items, nil
}
