package domain

import (
	"math/big"
	"time"
)

// ChainType 标识支持的链类型
type ChainType string

const (
	ChainEVM     ChainType = "evm"
	ChainSolana  ChainType = "solana"
	ChainBitcoin ChainType = "bitcoin"
)

// Asset 表示可交易/存取款资产
type Asset struct {
	Symbol    string
	Chain     ChainType
	Decimals  uint8
	TokenAddr string
	Tags      map[string]string
}

// WalletAccount 表示用户在特定资产上的钱包账户
type WalletAccount struct {
	UserID      string
	AssetSymbol string
	Address     string
	Chain       ChainType
	CreatedAt   time.Time
	Metadata    map[string]string
}

// Balance 聚合余额信息
type Balance struct {
	Available *big.Int
	Frozen    *big.Int
	Pending   *big.Int
}

// DepositStatus 记录充值处理状态
type DepositStatus string

const (
	DepositPending   DepositStatus = "pending"
	DepositConfirmed DepositStatus = "confirmed"
	DepositCredited  DepositStatus = "credited"
	DepositFailed    DepositStatus = "failed"
)

// DepositRecord 表示一笔充值
type DepositRecord struct {
	TxHash                string
	Chain                 ChainType
	AssetSymbol           string
	Amount                *big.Int
	FromAddress           string
	ToAddress             string
	BlockHeight           uint64 // 交易所在区块高度
	Confirmations         uint64
	RequiredConfirmations uint64
	Status                DepositStatus
	ObservedAt            time.Time
	CreditedAt            time.Time
	Metadata              map[string]string
}

// DepositEvent 扫描器产生的充值事件
type DepositEvent struct {
	Chain         ChainType
	AssetSymbol   string
	FromAddress   string
	ToAddress     string
	Amount        *big.Int
	TxHash        string
	BlockHeight   uint64
	Confirmations uint64
	ObservedAt    time.Time
	Metadata      map[string]string
}

// WithdrawalStatus 表示提现状态
type WithdrawalStatus string

const (
	WithdrawalRequested   WithdrawalStatus = "requested"
	WithdrawalUnderReview WithdrawalStatus = "under_review"
	WithdrawalRejected    WithdrawalStatus = "rejected"
	WithdrawalApproved    WithdrawalStatus = "approved"
	WithdrawalSigned      WithdrawalStatus = "signed"
	WithdrawalBroadcast   WithdrawalStatus = "broadcast"
	WithdrawalCompleted   WithdrawalStatus = "completed"
	WithdrawalFailed      WithdrawalStatus = "failed"
)

// WithdrawalRequest 表示用户提现请求
type WithdrawalRequest struct {
	ID          string
	UserID      string
	AssetSymbol string
	Chain       ChainType
	ToAddress   string
	Amount      *big.Int
	Fee         *big.Int
	Status      WithdrawalStatus
	RiskScore   float64
	RiskRemarks string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Metadata    map[string]string
	RawTx       []byte
	TxHash      string
}

// TransferPlan 用于热冷钱包资金调度
type TransferPlan struct {
	ID          string
	AssetSymbol string
	FromAddress string
	ToAddress   string
	Amount      *big.Int
	CreatedAt   time.Time
	Status      string
	Metadata    map[string]string
}

// WithdrawalDecision 风控返回的决策
type WithdrawalDecision struct {
	Approved bool
	Score    float64
	Remarks  string
	Metadata map[string]string
}
