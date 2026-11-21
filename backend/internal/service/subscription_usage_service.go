package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var (
	ErrSubscriptionUsageNoGroup              = errors.New("api key is not bound to a group")
	ErrSubscriptionUsageNotSubscriptionGroup = errors.New("group has no anthropic subscription account")
	ErrSubscriptionUsageNoAvailableAccount   = errors.New("no available subscription account in group")
	ErrSubscriptionUsageUpstream             = errors.New("upstream usage query failed")
)

// SubscriptionUsageWindow 对外暴露的单个用量窗口（不含任何账号标识）。
type SubscriptionUsageWindow struct {
	Utilization      float64    `json:"utilization"`
	ResetsAt         *time.Time `json:"resets_at"`
	RemainingSeconds int        `json:"remaining_seconds"`
}

// SubscriptionUsage 是 GET /v1/subscription-usage 的响应体。
type SubscriptionUsage struct {
	Plan           string                   `json:"plan,omitempty"`
	FiveHour       *SubscriptionUsageWindow `json:"five_hour,omitempty"`
	SevenDay       *SubscriptionUsageWindow `json:"seven_day,omitempty"`
	SevenDaySonnet *SubscriptionUsageWindow `json:"seven_day_sonnet,omitempty"`
	SevenDayFable  *SubscriptionUsageWindow `json:"seven_day_fable,omitempty"`
	UpdatedAt      *time.Time               `json:"updated_at,omitempty"`
}

type subscriptionUsageAccountLister interface {
	ListByGroup(ctx context.Context, groupID int64) ([]Account, error)
}

type subscriptionUsageFetcher interface {
	GetUsageForAccount(ctx context.Context, account *Account, force ...bool) (*UsageInfo, error)
}

// SubscriptionUsageService 为普通 API Key 返回其分组首选 Anthropic 订阅账号的官方用量。
type SubscriptionUsageService struct {
	accounts subscriptionUsageAccountLister
	usage    subscriptionUsageFetcher
}

func NewSubscriptionUsageService(accountRepo AccountRepository, usageService *AccountUsageService) *SubscriptionUsageService {
	return &SubscriptionUsageService{accounts: accountRepo, usage: usageService}
}

func (s *SubscriptionUsageService) GetForAPIKey(ctx context.Context, apiKey *APIKey) (*SubscriptionUsage, error) {
	if apiKey == nil || apiKey.GroupID == nil || *apiKey.GroupID <= 0 {
		return nil, ErrSubscriptionUsageNoGroup
	}
	return s.GetForGroup(ctx, *apiKey.GroupID)
}

// GetForGroup 返回指定分组首选 Anthropic 订阅账号的官方用量（3 分钟缓存，不强制刷新）。
func (s *SubscriptionUsageService) GetForGroup(ctx context.Context, groupID int64) (*SubscriptionUsage, error) {
	if groupID <= 0 {
		return nil, ErrSubscriptionUsageNoGroup
	}
	accounts, err := s.accounts.ListByGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("list group accounts: %w", err)
	}
	account, err := selectSubscriptionAccount(accounts, groupID)
	if err != nil {
		return nil, err
	}
	info, err := s.usage.GetUsageForAccount(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSubscriptionUsageUpstream, err)
	}
	return toSubscriptionUsage(info), nil
}

func selectSubscriptionAccount(accounts []Account, groupID int64) (*Account, error) {
	var subscription []*Account
	for i := range accounts {
		a := &accounts[i]
		if a.Platform == PlatformAnthropic && a.Type == AccountTypeOAuth {
			subscription = append(subscription, a)
		}
	}
	if len(subscription) == 0 {
		return nil, ErrSubscriptionUsageNotSubscriptionGroup
	}
	var available []*Account
	for _, a := range subscription {
		if a.IsSchedulable() {
			available = append(available, a)
		}
	}
	if len(available) == 0 {
		return nil, ErrSubscriptionUsageNoAvailableAccount
	}
	sort.SliceStable(available, func(i, j int) bool {
		gi, gj := accountGroupPriority(available[i], groupID), accountGroupPriority(available[j], groupID)
		if gi != gj {
			return gi < gj
		}
		if available[i].Priority != available[j].Priority {
			return available[i].Priority < available[j].Priority
		}
		return available[i].ID < available[j].ID
	})
	return available[0], nil
}

// accountGroupPriority 返回账号在该分组内的优先级；未加载关联时按 schema 默认值 50。
func accountGroupPriority(a *Account, groupID int64) int {
	for _, ag := range a.AccountGroups {
		if ag.GroupID == groupID {
			return ag.Priority
		}
	}
	return 50
}

func toSubscriptionUsage(info *UsageInfo) *SubscriptionUsage {
	out := &SubscriptionUsage{}
	if info == nil {
		return out
	}
	out.Plan = strings.ToLower(strings.TrimSpace(info.SubscriptionTier))
	out.FiveHour = toSubscriptionWindow(info.FiveHour)
	out.SevenDay = toSubscriptionWindow(info.SevenDay)
	out.SevenDaySonnet = toSubscriptionWindow(info.SevenDaySonnet)
	out.SevenDayFable = toSubscriptionWindow(info.SevenDayFable)
	out.UpdatedAt = info.UpdatedAt
	return out
}

func toSubscriptionWindow(p *UsageProgress) *SubscriptionUsageWindow {
	if p == nil {
		return nil
	}
	return &SubscriptionUsageWindow{Utilization: p.Utilization, ResetsAt: p.ResetsAt, RemainingSeconds: p.RemainingSeconds}
}
