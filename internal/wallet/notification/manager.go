package notification

import (
	"context"
	"fmt"
	"sync"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// Manager 通知管理器
type Manager struct {
	notifier   Notifier
	configStore ConfigStore
	mu         sync.RWMutex
}

// NewManager 创建通知管理器
func NewManager(notifier Notifier, configStore ConfigStore) *Manager {
	return &Manager{
		notifier:    notifier,
		configStore: configStore,
	}
}

// NotifyDeposit 通知充值事件
func (m *Manager) NotifyDeposit(ctx context.Context, record domain.DepositRecord) error {
	// 获取用户的通知配置
	config, err := m.configStore.GetConfig(ctx, record.UserID)
	if err != nil {
		return fmt.Errorf("failed to get notification config: %w", err)
	}

	// 如果未启用通知，直接返回
	if !config.Enabled {
		return nil
	}

	// 检查事件类型是否在监听列表中
	eventType := string(record.Status)
	if !contains(config.Events, eventType) && !contains(config.Events, "all") {
		return nil
	}

	// 异步发送通知，避免阻塞主流程
	go func() {
		for _, url := range config.WebhookURLs {
			if err := m.notifier.NotifyDeposit(context.Background(), url, record); err != nil {
				// 记录错误但不中断其他通知
				fmt.Printf("Failed to send webhook notification to %s: %v\n", url, err)
			}
		}
	}()

	return nil
}

// contains 检查字符串切片是否包含指定值
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// SetConfig 设置用户的通知配置
func (m *Manager) SetConfig(ctx context.Context, userID string, config *NotificationConfig) error {
	return m.configStore.SaveConfig(ctx, userID, config)
}

// GetConfig 获取用户的通知配置
func (m *Manager) GetConfig(ctx context.Context, userID string) (*NotificationConfig, error) {
	return m.configStore.GetConfig(ctx, userID)
}

