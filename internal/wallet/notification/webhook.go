package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// WebhookNotifier Webhook通知器
type WebhookNotifier struct {
	client *http.Client
}

// NewWebhookNotifier 创建Webhook通知器
func NewWebhookNotifier() *WebhookNotifier {
	return &WebhookNotifier{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DepositNotification 充值通知消息
type DepositNotification struct {
	Event      string    `json:"event"`       // 事件类型：deposit.credited, deposit.confirmed, deposit.failed
	TxHash     string    `json:"tx_hash"`     // 交易哈希
	UserID     string    `json:"user_id"`     // 用户ID
	Chain      string    `json:"chain"`       // 链类型
	Asset      string    `json:"asset"`        // 资产符号
	Amount     string    `json:"amount"`      // 金额（字符串格式）
	FromAddr   string    `json:"from_address"` // 发送地址
	ToAddr     string    `json:"to_address"`  // 接收地址
	Status     string    `json:"status"`      // 状态
	BlockHeight uint64   `json:"block_height"` // 区块高度
	Confirmations uint64 `json:"confirmations"` // 确认数
	ObservedAt  string   `json:"observed_at"`  // 观察时间
	CreditedAt  string   `json:"credited_at,omitempty"` // 入账时间（如果已入账）
	Timestamp   string   `json:"timestamp"`    // 通知时间戳
}

// NotifyDeposit 发送充值通知
func (n *WebhookNotifier) NotifyDeposit(ctx context.Context, url string, record domain.DepositRecord) error {
	notification := DepositNotification{
		Event:        "deposit." + string(record.Status),
		TxHash:       record.TxHash,
		UserID:       record.UserID,
		Chain:        string(record.Chain),
		Asset:        record.AssetSymbol,
		Amount:       record.Amount.String(),
		FromAddr:     record.FromAddress,
		ToAddr:       record.ToAddress,
		Status:       string(record.Status),
		BlockHeight:  record.BlockHeight,
		Confirmations: record.Confirmations,
		ObservedAt:   record.ObservedAt.Format(time.RFC3339),
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	if !record.CreditedAt.IsZero() {
		notification.CreditedAt = record.CreditedAt.Format(time.RFC3339)
	}

	// 序列化为JSON
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "go-blockchain-wallet/1.0")

	// 发送请求
	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}

// Notifier 通知器接口
type Notifier interface {
	NotifyDeposit(ctx context.Context, url string, record domain.DepositRecord) error
}

