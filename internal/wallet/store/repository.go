package store

import (
	"context"
	"math/big"
	"time"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// AssetRepository 资产配置存储
type AssetRepository interface {
	SaveAsset(ctx context.Context, asset domain.Asset) error
	GetAsset(ctx context.Context, symbol string) (domain.Asset, error)
	ListAssets(ctx context.Context) ([]domain.Asset, error)
}

// AccountRepository 用户钱包账户存储
type AccountRepository interface {
	SaveAccount(ctx context.Context, account domain.WalletAccount) error
	GetAccount(ctx context.Context, userID, asset string) (domain.WalletAccount, error)
	FindAccountByAddress(ctx context.Context, address, asset string) (domain.WalletAccount, error)
}

// BalanceRepository 用于管理用户余额
type BalanceRepository interface {
	Credit(ctx context.Context, userID, asset string, amount *big.Int) error
	Debit(ctx context.Context, userID, asset string, amount *big.Int) error
	Freeze(ctx context.Context, userID, asset string, amount *big.Int) error
	Unfreeze(ctx context.Context, userID, asset string, amount *big.Int) error
	GetBalance(ctx context.Context, userID, asset string) (domain.Balance, error)
}

// DepositRepository 保存充值记录
type DepositRepository interface {
	SaveDeposit(ctx context.Context, userID string, record domain.DepositRecord) error
	GetDeposit(ctx context.Context, txHash string) (domain.DepositRecord, error)
	UpdateDeposit(ctx context.Context, record domain.DepositRecord) error
	// FindDepositsByBlockRange 查找指定区块范围内的充值记录（用于重组回滚）
	FindDepositsByBlockRange(ctx context.Context, chain domain.ChainType, fromHeight, toHeight uint64) ([]domain.DepositRecord, error)
}

// WithdrawalRepository 保存提现请求
type WithdrawalRepository interface {
	SaveWithdrawal(ctx context.Context, req domain.WithdrawalRequest) error
	UpdateWithdrawal(ctx context.Context, req domain.WithdrawalRequest) error
	GetWithdrawal(ctx context.Context, id string) (domain.WithdrawalRequest, error)
}

// BlockRepository 区块信息存储（用于重组检测）
type BlockRepository interface {
	// SaveBlock 保存区块信息
	SaveBlock(ctx context.Context, chain domain.ChainType, height uint64, hash string, parentHash string) error
	// GetBlock 获取指定高度的区块信息
	GetBlock(ctx context.Context, chain domain.ChainType, height uint64) (BlockInfo, error)
	// GetLatestBlock 获取最新区块
	GetLatestBlock(ctx context.Context, chain domain.ChainType) (BlockInfo, error)
	// DeleteBlocksFromHeight 删除从指定高度开始的所有区块（重组时使用）
	DeleteBlocksFromHeight(ctx context.Context, chain domain.ChainType, fromHeight uint64) error
}

// BlockInfo 区块信息
type BlockInfo struct {
	Chain      domain.ChainType
	Height     uint64
	Hash       string
	ParentHash string
	CreatedAt  time.Time
}

// RepositoryProvider 聚合全部仓储接口
type RepositoryProvider interface {
	AssetRepository
	AccountRepository
	BalanceRepository
	DepositRepository
	WithdrawalRepository
	BlockRepository
}
