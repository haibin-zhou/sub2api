package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type subscriptionUsageGetter interface {
	GetForAPIKey(ctx context.Context, apiKey *service.APIKey) (*service.SubscriptionUsage, error)
}

// SubscriptionUsageHandler 处理 GET /v1/subscription-usage。
type SubscriptionUsageHandler struct {
	svc subscriptionUsageGetter
}

func NewSubscriptionUsageHandler(svc *service.SubscriptionUsageService) *SubscriptionUsageHandler {
	return &SubscriptionUsageHandler{svc: svc}
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
