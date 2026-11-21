//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// 运行：go test -tags integration -run TestSubscriptionPeriodUsageRepositoryIntegration ./internal/repository/
// （需要 Docker，由 integration_harness_test.go 的 TestMain 拉起 Postgres。）
func TestSubscriptionPeriodUsageRepositoryIntegration(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	client := tx.Client()
	usageRepo := newUsageLogRepositoryWithSQL(client, tx)
	repo := newSubscriptionPeriodUsageRepositoryWithSQL(tx)

	user := mustCreateUser(t, client, &service.User{Email: "period-" + uuid.NewString() + "@test.com"})
	other := mustCreateUser(t, client, &service.User{Email: "period-other-" + uuid.NewString() + "@test.com"})
	group := mustCreateGroup(t, client, &service.Group{Name: "period-g-" + uuid.NewString()})
	apiKey := mustCreateApiKey(t, client, &service.APIKey{UserID: user.ID, Key: "sk-" + uuid.NewString(), Name: "k"})
	otherKey := mustCreateApiKey(t, client, &service.APIKey{UserID: other.ID, Key: "sk-" + uuid.NewString(), Name: "k"})
	account := mustCreateAccount(t, client, &service.Account{Name: "period-acc-" + uuid.NewString()})
	sub := mustCreateSubscription(t, client, &service.UserSubscription{UserID: user.ID, GroupID: group.ID})
	sub2 := mustCreateSubscription(t, client, &service.UserSubscription{UserID: user.ID, GroupID: group.ID})

	start := time.Date(2026, 10, 22, 8, 0, 0, 0, time.UTC)
	end := start.Add(30 * 24 * time.Hour)
	mk := func(uid, keyID int64, subID *int64, at time.Time, in, out, cc, cr int, cost float64) {
		_, err := usageRepo.Create(ctx, &service.UsageLog{
			UserID: uid, APIKeyID: keyID, AccountID: account.ID, RequestID: uuid.NewString(), Model: "claude-3",
			SubscriptionID: subID, InputTokens: in, OutputTokens: out, CacheCreationTokens: cc, CacheReadTokens: cr,
			TotalCost: cost, ActualCost: cost * 0.1, CreatedAt: at,
		})
		require.NoError(t, err)
	}
	mk(user.ID, apiKey.ID, &sub.ID, start, 1, 2, 3, 4, 1.25)                    // 边界：包含
	mk(user.ID, apiKey.ID, &sub.ID, end.Add(-time.Second), 10, 20, 30, 40, 2.5) // 包含
	mk(user.ID, apiKey.ID, &sub.ID, end, 100, 100, 100, 100, 100)               // 边界：排除
	mk(user.ID, apiKey.ID, &sub.ID, start.Add(-time.Second), 100, 100, 100, 100, 100)
	mk(user.ID, apiKey.ID, nil, start.Add(time.Hour), 100, 100, 100, 100, 100)        // 非订阅
	mk(other.ID, otherKey.ID, &sub.ID, start.Add(time.Hour), 100, 100, 100, 100, 100) // 其他用户（防御）

	out, err := repo.SumSubscriptionPeriodUsage(ctx, user.ID, []service.SubscriptionPeriodWindow{
		{SubscriptionID: sub.ID, Start: start, End: end},
		{SubscriptionID: sub2.ID, Start: start, End: end},
	})
	require.NoError(t, err)
	got := out[sub.ID]
	require.Equal(t, int64(2), got.RequestCount)
	require.Equal(t, int64(11), got.InputTokens)
	require.Equal(t, int64(22), got.OutputTokens)
	require.Equal(t, int64(33), got.CacheCreationTokens)
	require.Equal(t, int64(44), got.CacheReadTokens)
	require.InDelta(t, 3.75, got.TotalCostUSD, 1e-9, "uses total_cost (official price), not actual_cost")
	_, ok := out[sub2.ID]
	require.False(t, ok)
}
