package deposit

import (
	"context"
	"time"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
	"github.com/jinmu/go-blockchain/internal/wallet/store"
)

// ConfirmationManager 确认数管理器
type ConfirmationManager struct {
	store        store.RepositoryProvider
	confirmations map[domain.ChainType]uint64 // 各链所需确认数
	checkInterval time.Duration                // 检查间隔
}

// NewConfirmationManager 创建确认数管理器
func NewConfirmationManager(store store.RepositoryProvider) *ConfirmationManager {
	return &ConfirmationManager{
		store:        store,
		confirmations: make(map[domain.ChainType]uint64),
		checkInterval: 30 * time.Second,
	}
}

// SetConfirmations 设置链的确认数要求
func (m *ConfirmationManager) SetConfirmations(chain domain.ChainType, confirmations uint64) {
	m.confirmations[chain] = confirmations
}

// GetConfirmations 获取链的确认数要求
func (m *ConfirmationManager) GetConfirmations(chain domain.ChainType) uint64 {
	if conf, ok := m.confirmations[chain]; ok {
		return conf
	}
	// 默认确认数
	switch chain {
	case domain.ChainBitcoin:
		return 6
	case domain.ChainEVM:
		return 12
	case domain.ChainSolana:
		return 32
	default:
		return 6
	}
}

// CheckAndCredit 检查确认数并自动入账
func (m *ConfirmationManager) CheckAndCredit(ctx context.Context, record domain.DepositRecord, currentHeight uint64, creditFunc func(ctx context.Context, record domain.DepositRecord) error) error {
	// 计算当前确认数
	confirmations := currentHeight - record.BlockHeight + 1
	requiredConfirmations := m.GetConfirmations(record.Chain)

	// 更新确认数
	record.Confirmations = confirmations
	record.RequiredConfirmations = requiredConfirmations

	// 如果确认数足够且还未入账，执行入账
	if confirmations >= requiredConfirmations && record.Status == domain.DepositPending {
		record.Status = domain.DepositConfirmed
		if err := m.store.UpdateDeposit(ctx, record); err != nil {
			return err
		}

		// 执行入账
		if err := creditFunc(ctx, record); err != nil {
			return err
		}

		record.Status = domain.DepositCredited
		record.CreditedAt = time.Now()
		return m.store.UpdateDeposit(ctx, record)
	}

	// 如果确认数不足，更新确认数
	if confirmations < requiredConfirmations {
		record.Status = domain.DepositPending
		return m.store.UpdateDeposit(ctx, record)
	}

	return nil
}

// StartConfirmationChecker 启动确认数检查器（定期检查待确认的充值）
func (m *ConfirmationManager) StartConfirmationChecker(ctx context.Context, getCurrentHeight func(chain domain.ChainType) (uint64, error), creditFunc func(ctx context.Context, record domain.DepositRecord) error) {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// TODO: 查询所有待确认的充值记录并检查确认数
			// 这里需要添加查询接口
		}
	}
}

