package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type subscriptionUsageGetter interface {
	GetForAPIKey(ctx context.Context, apiKey *service.APIKey) (*service.SubscriptionUsage, error)
	GetForGroup(ctx context.Context, groupID int64) (*service.SubscriptionUsage, error)
}

type activeSubscriptionLister interface {
	ListActiveUserSubscriptions(ctx context.Context, userID int64) ([]service.UserSubscription, error)
}

// SubscriptionUsageHandler 处理 GET /v1/subscription-usage（API Key）与
// GET /api/v1/subscriptions/usage（用户 JWT）。
type SubscriptionUsageHandler struct {
	svc  subscriptionUsageGetter
	subs activeSubscriptionLister
}

func NewSubscriptionUsageHandler(svc *service.SubscriptionUsageService, subs *service.SubscriptionService) *SubscriptionUsageHandler {
	return &SubscriptionUsageHandler{svc: svc, subs: subs}
}

// userSubscriptionUsageConcurrency 限制单次请求内并发查询的分组数。
const userSubscriptionUsageConcurrency = 4

// SubscriptionUsageItemError 是单个分组用量查询失败时的固定错误（不含原始错误文本或账号标识）。
type SubscriptionUsageItemError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// SubscriptionUsageItem 是 GET /api/v1/subscriptions/usage 中每个有效订阅分组的一项。
type SubscriptionUsageItem struct {
	GroupID int64                       `json:"group_id"`
	Usage   *service.SubscriptionUsage  `json:"usage"`
	Error   *SubscriptionUsageItemError `json:"error,omitempty"`
}

// ListForUser 返回当前用户每个有效订阅分组的官方订阅用量（逻辑同 /v1/subscription-usage）。
// GET /api/v1/subscriptions/usage — 分组只取自用户自己的有效订阅，不接受客户端传入。
func (h *SubscriptionUsageHandler) ListForUser(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	subs, err := h.subs.ListActiveUserSubscriptions(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	groupIDs := make([]int64, 0, len(subs))
	seen := make(map[int64]struct{}, len(subs))
	for i := range subs {
		sub := &subs[i]
		if !sub.IsActive() || sub.GroupID <= 0 {
			continue
		}
		if _, dup := seen[sub.GroupID]; dup {
			continue
		}
		seen[sub.GroupID] = struct{}{}
		groupIDs = append(groupIDs, sub.GroupID)
	}

	items := make([]SubscriptionUsageItem, len(groupIDs))
	var g errgroup.Group
	g.SetLimit(userSubscriptionUsageConcurrency)
	ctx := c.Request.Context()
	for i, groupID := range groupIDs {
		g.Go(func() error {
			item := SubscriptionUsageItem{GroupID: groupID}
			usage, err := h.svc.GetForGroup(ctx, groupID)
			if err != nil {
				item.Error = subscriptionUsageItemError(groupID, err)
			} else {
				item.Usage = usage
			}
			items[i] = item
			return nil
		})
	}
	_ = g.Wait()
	response.Success(c, items)
}

func subscriptionUsageItemError(groupID int64, err error) *SubscriptionUsageItemError {
	switch {
	case errors.Is(err, service.ErrSubscriptionUsageNotSubscriptionGroup):
		return &SubscriptionUsageItemError{Type: "not_subscription_group", Message: "This group is not backed by a Claude subscription"}
	case errors.Is(err, service.ErrSubscriptionUsageNoAvailableAccount):
		return &SubscriptionUsageItemError{Type: "no_available_account", Message: "No subscription account is currently available in this group"}
	case errors.Is(err, service.ErrSubscriptionUsageUpstream):
		slog.Warn("user subscription usage upstream error", "group_id", groupID, "error", err)
		return &SubscriptionUsageItemError{Type: "upstream_usage_error", Message: "Failed to query subscription usage upstream"}
	default:
		slog.Error("user subscription usage internal error", "group_id", groupID, "error", err)
		return &SubscriptionUsageItemError{Type: "api_error", Message: "Internal error"}
	}
}

func (h *SubscriptionUsageHandler) Get(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		subscriptionUsageError(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	out, err := h.svc.GetForAPIKey(c.Request.Context(), apiKey)
	if err == nil {
		c.JSON(http.StatusOK, out)
		return
	}
	switch {
	case errors.Is(err, service.ErrSubscriptionUsageNoGroup):
		subscriptionUsageError(c, http.StatusNotFound, "no_group", "API key is not bound to a group")
	case errors.Is(err, service.ErrSubscriptionUsageNotSubscriptionGroup):
		subscriptionUsageError(c, http.StatusNotFound, "not_subscription_group", "This group is not backed by a Claude subscription")
	case errors.Is(err, service.ErrSubscriptionUsageNoAvailableAccount):
		subscriptionUsageError(c, http.StatusServiceUnavailable, "no_available_account", "No subscription account is currently available in this group")
	case errors.Is(err, service.ErrSubscriptionUsageUpstream):
		slog.Warn("subscription usage upstream error", "error", err)
		subscriptionUsageError(c, http.StatusBadGateway, "upstream_usage_error", "Failed to query subscription usage upstream")
	default:
		slog.Error("subscription usage internal error", "error", err)
		subscriptionUsageError(c, http.StatusInternalServerError, "api_error", "Internal error")
	}
}

func subscriptionUsageError(c *gin.Context, status int, errType, message string) {
	c.JSON(status, gin.H{"type": "error", "error": gin.H{"type": errType, "message": message}})
}
