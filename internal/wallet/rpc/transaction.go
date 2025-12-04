package rpc

import (
	"math/big"
)

// Transaction 区块交易信息（用于扫描器）
type Transaction struct {
	Hash           string
	From           string
	To             string
	Value          *big.Int
	Success        bool
	GasUsed        uint64
	TokenTransfer  *TokenTransfer  // ERC20 代币转账信息（向后兼容，保留第一个）
	TokenTransfers []*TokenTransfer // ERC20 代币转账信息列表（支持多个）
}

// TokenTransfer ERC20 代币转账信息
type TokenTransfer struct {
	TokenAddress string
	From         string
	To           string
	Amount       *big.Int
	Symbol       string
	Decimals     uint8
}

