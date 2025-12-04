package notification

import (
	"context"
	"sync"
)

// NotificationConfig 通知配置
type NotificationConfig struct {
	WebhookURLs []string `json:"webhook_urls"` // Webhook URL列表
	Events      []string `json:"events"`       // 监听的事件类型：credited, confirmed, failed
	Enabled     bool     `json:"enabled"`      // 是否启用
}

// ConfigStore 通知配置存储接口
type ConfigStore interface {
	GetConfig(ctx context.Context, userID string) (*NotificationConfig, error)
	SaveConfig(ctx context.Context, userID string, config *NotificationConfig) error
}

// InMemoryConfigStore 内存配置存储（用于开发测试）
type InMemoryConfigStore struct {
	mu      sync.RWMutex
	configs map[string]*NotificationConfig
}

// NewInMemoryConfigStore 创建内存配置存储
func NewInMemoryConfigStore() *InMemoryConfigStore {
	return &InMemoryConfigStore{
		configs: make(map[string]*NotificationConfig),
	}
}

// GetConfig 获取用户的通知配置
func (s *InMemoryConfigStore) GetConfig(ctx context.Context, userID string) (*NotificationConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	config, ok := s.configs[userID]
	if !ok {
		// 返回默认配置（禁用）
		return &NotificationConfig{
			Enabled: false,
			Events:  []string{},
			WebhookURLs: []string{},
		}, nil
	}
	
	return config, nil
}

// SaveConfig 保存用户的通知配置
func (s *InMemoryConfigStore) SaveConfig(ctx context.Context, userID string, config *NotificationConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.configs[userID] = config
	return nil
}

