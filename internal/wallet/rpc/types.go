package rpc

import (
	"math/big"
)

// BlockInfo 区块信息（统一格式）
type BlockInfo struct {
	Height     uint64
	Hash       string
	ParentHash string
	Timestamp  uint64
	TxCount    uint64
}

// TransactionInfo 交易信息（统一格式）
type TransactionInfo struct {
	Hash        string
	From        string
	To          string
	Value       *big.Int
	GasPrice    *big.Int
	GasLimit    uint64
	Nonce       uint64
	BlockHeight uint64
	BlockHash   string
	Status      uint64 // 0: failed, 1: success
}

// GasEstimation Gas 估算结果
type GasEstimation struct {
	BaseFeePerGas    *big.Int
	PriorityFeePerGas *big.Int
	MaxFeePerGas     *big.Int
	GasLimit         uint64
	BlockNumber      uint64
}

// FeeHistory 手续费历史（用于 EIP-1559）
type FeeHistory struct {
	OldestBlock   *big.Int
	BaseFeePerGas []*big.Int
	GasUsedRatio  []float64
	Reward        [][]*big.Int
}

// BalanceInfo 余额信息
type BalanceInfo struct {
	Address string
	Balance *big.Int
	Nonce   uint64
}

