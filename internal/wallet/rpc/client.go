package rpc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// Client 统一的 RPC 客户端接口
type Client interface {
	// Chain 返回链类型
	Chain() domain.ChainType

	// GetBlockByHeight 根据高度获取区块信息
	GetBlockByHeight(ctx context.Context, height uint64) (*BlockInfo, error)

	// GetBlockByHash 根据哈希获取区块信息
	GetBlockByHash(ctx context.Context, hash string) (*BlockInfo, error)

	// GetLatestBlock 获取最新区块
	GetLatestBlock(ctx context.Context) (*BlockInfo, error)

	// GetTransaction 获取交易信息
	GetTransaction(ctx context.Context, txHash string) (*TransactionInfo, error)

	// GetBalance 获取地址余额
	GetBalance(ctx context.Context, address string) (*BalanceInfo, error)

	// EstimateGas 估算 Gas
	EstimateGas(ctx context.Context, from, to string, value *big.Int, data []byte) (*GasEstimation, error)

	// GetFeeHistory 获取手续费历史（EIP-1559）
	GetFeeHistory(ctx context.Context, blockCount uint64, newestBlock string) (*FeeHistory, error)

	// SendRawTransaction 发送原始交易
	SendRawTransaction(ctx context.Context, rawTx []byte) (string, error)

	// GetTransactionReceipt 获取交易回执
	GetTransactionReceipt(ctx context.Context, txHash string) (*TransactionInfo, error)
}

// ClientFactory RPC 客户端工厂
type ClientFactory interface {
	CreateClient(chain domain.ChainType, endpoint string) (Client, error)
}

// NewClient 创建 RPC 客户端（根据链类型）
func NewClient(chain domain.ChainType, endpoint string) (Client, error) {
	switch chain {
	case domain.ChainEVM:
		return NewEVMClient(endpoint)
	case domain.ChainBitcoin:
		return NewBitcoinClient(endpoint)
	case domain.ChainSolana:
		return NewSolanaClient(endpoint)
	default:
		return nil, fmt.Errorf("unsupported chain: %s", chain)
	}
}
