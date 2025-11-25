package funding

import (
	"context"
	"math/big"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
	"github.com/jinmu/go-blockchain/internal/wallet/service"
)

// MultisigController 管理多签钱包
type MultisigController interface {
	Propose(ctx context.Context, plan domain.TransferPlan) (string, error)
	Approve(ctx context.Context, proposalID string) error
	Execute(ctx context.Context, proposalID string) error
}

// TreasuryConfig 资金层级配置
type TreasuryConfig struct {
	HotWalletLimit  *big.Int
	WarmWalletLimit *big.Int
	ColdWalletLimit *big.Int
}

// RebalanceStrategy 资金调度策略
type RebalanceStrategy interface {
	Plan(ctx context.Context, asset domain.Asset, balances map[string]*big.Int, cfg TreasuryConfig) ([]domain.TransferPlan, error)
}

// Manager 负责多签归集和平衡
type Manager struct {
	service  *service.Manager
	multi    MultisigController
	strategy RebalanceStrategy
	treasury TreasuryConfig
}

// NewManager 初始化资金管理
func NewManager(service *service.Manager, multi MultisigController, strategy RebalanceStrategy, cfg TreasuryConfig) *Manager {
	return &Manager{service: service, multi: multi, strategy: strategy, treasury: cfg}
}

// Rebalance 执行资金调度
func (f *Manager) Rebalance(ctx context.Context, assetSymbol string, balances map[string]*big.Int) error {
	asset, err := f.service.GetAsset(ctx, assetSymbol)
	if err != nil {
		return err
	}
	plans, err := f.strategy.Plan(ctx, asset, balances, f.treasury)
	if err != nil {
		return err
	}
	for _, plan := range plans {
		proposalID, err := f.multi.Propose(ctx, plan)
		if err != nil {
			return err
		}
		if err := f.multi.Approve(ctx, proposalID); err != nil {
			return err
		}
		if err := f.multi.Execute(ctx, proposalID); err != nil {
			return err
		}
	}
	return nil
}
