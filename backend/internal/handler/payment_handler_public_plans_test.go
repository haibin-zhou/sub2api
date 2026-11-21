//go:build unit

package handler

import (
	"encoding/json"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/stretchr/testify/require"
)

func TestToPublicPlansKeepsOnlyForSaleAndPublicFields(t *testing.T) {
	orig := 60.0
	plans := []*dbent.SubscriptionPlan{
		{ID: 1, GroupID: 5, Name: "Standard", Description: "约 1.25× Pro 用量", Price: 45, OriginalPrice: &orig,
			Currency: "USD", ValidityDays: 30, ValidityUnit: "days", Features: "适合日常单项目", ForSale: true, SortOrder: 1, ProductName: "internal"},
		{ID: 2, GroupID: 6, Name: "Hidden", Price: 99, ForSale: false},
		nil,
	}

	got := toPublicPlans(plans)

	require.Len(t, got, 1)
	require.Equal(t, int64(1), got[0].ID)
	require.Equal(t, "Standard", got[0].Name)
	require.Equal(t, "约 1.25× Pro 用量", got[0].Description)
	require.Equal(t, 45.0, got[0].Price)
	require.NotNil(t, got[0].OriginalPrice)
	require.Equal(t, 60.0, *got[0].OriginalPrice)
	require.Equal(t, "USD", got[0].Currency)

	raw, err := json.Marshal(got[0])
	require.NoError(t, err)
	for _, leaked := range []string{"group_id", "group_name", "rate_multiplier", "product_name", "peak"} {
		require.NotContains(t, string(raw), leaked)
	}
}

func TestToPublicPlansEmptyInputReturnsEmptySlice(t *testing.T) {
	got := toPublicPlans(nil)
	require.NotNil(t, got)
	raw, _ := json.Marshal(got)
	require.Equal(t, "[]", string(raw))
}
