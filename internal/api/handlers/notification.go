package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinmu/go-blockchain/internal/wallet/notification"
	"github.com/jinmu/go-blockchain/internal/wallet/service"
)

// NotificationHandler 通知相关处理器
type NotificationHandler struct {
	manager *service.Manager
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler(manager *service.Manager) *NotificationHandler {
	return &NotificationHandler{manager: manager}
}

// GetNotificationConfig 获取用户的通知配置
// GET /api/v1/notifications/config?user_id=xxx
func (h *NotificationHandler) GetNotificationConfig(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// TODO: 需要添加Manager方法获取通知配置
	// 暂时返回空配置
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"config": gin.H{
			"enabled":      false,
			"webhook_urls": []string{},
			"events":       []string{},
		},
	})
}

// SetNotificationConfig 设置用户的通知配置
// POST /api/v1/notifications/config
func (h *NotificationHandler) SetNotificationConfig(c *gin.Context) {
	var req struct {
		UserID      string   `json:"user_id" binding:"required"`
		WebhookURLs []string `json:"webhook_urls"`
		Events      []string `json:"events"` // credited, confirmed, failed, all
		Enabled     bool     `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := &notification.NotificationConfig{
		Enabled:     req.Enabled,
		WebhookURLs: req.WebhookURLs,
		Events:      req.Events,
	}

	// TODO: 需要添加Manager方法设置通知配置
	// 暂时返回成功
	c.JSON(http.StatusOK, gin.H{
		"message": "Notification config updated",
		"config":  config,
	})
}

