package risk

import (
	"context"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// Controller 风控接口
type Controller interface {
	EvaluateWithdrawal(ctx context.Context, req domain.WithdrawalRequest) (domain.WithdrawalDecision, error)
	
	// 地址列表管理（可选实现）
	AddToWhitelist(ctx context.Context, address string, chain string, remark string) error
	RemoveFromWhitelist(ctx context.Context, address string, chain string) error
	AddToBlacklist(ctx context.Context, address string, chain string, remark string) error
	RemoveFromBlacklist(ctx context.Context, address string, chain string) error
	
	// 配置管理（可选实现）
	GetConfig() *Config
	UpdateConfig(config *Config)
}

// NoopController 默认通过所有提现
type NoopController struct{}

// EvaluateWithdrawal 实现风控接口
func (NoopController) EvaluateWithdrawal(ctx context.Context, req domain.WithdrawalRequest) (domain.WithdrawalDecision, error) {
	return domain.WithdrawalDecision{Approved: true, Score: 0.0, Remarks: "auto-approved"}, nil
}

// 实现可选接口（空实现）
func (NoopController) AddToWhitelist(ctx context.Context, address string, chain string, remark string) error {
	return nil
}

func (NoopController) RemoveFromWhitelist(ctx context.Context, address string, chain string) error {
	return nil
}

func (NoopController) AddToBlacklist(ctx context.Context, address string, chain string, remark string) error {
	return nil
}

func (NoopController) RemoveFromBlacklist(ctx context.Context, address string, chain string) error {
	return nil
}

func (NoopController) GetConfig() *Config {
	return DefaultConfig()
}

func (NoopController) UpdateConfig(config *Config) {
	// 空实现
}
