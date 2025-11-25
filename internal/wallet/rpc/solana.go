package rpc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// SolanaClient Solana 链 RPC 客户端实现（占位符）
type SolanaClient struct {
	chain domain.ChainType
	// TODO: 实现 Solana RPC 客户端
}

// NewSolanaClient 创建 Solana 链客户端
func NewSolanaClient(endpoint string) (*SolanaClient, error) {
	// TODO: 实现 Solana RPC 连接
	return &SolanaClient{
		chain: domain.ChainSolana,
	}, fmt.Errorf("solana client not implemented yet")
}

// Chain 返回链类型
func (c *SolanaClient) Chain() domain.ChainType {
	return c.chain
}

// GetBlockByHeight 根据高度获取区块
func (c *SolanaClient) GetBlockByHeight(ctx context.Context, height uint64) (*BlockInfo, error) {
	// TODO: 实现 Solana 区块查询
	return nil, fmt.Errorf("not implemented")
}

// GetBlockByHash 根据哈希获取区块
func (c *SolanaClient) GetBlockByHash(ctx context.Context, hash string) (*BlockInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetLatestBlock 获取最新区块
func (c *SolanaClient) GetLatestBlock(ctx context.Context) (*BlockInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetTransaction 获取交易信息
func (c *SolanaClient) GetTransaction(ctx context.Context, txHash string) (*TransactionInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetBalance 获取地址余额
func (c *SolanaClient) GetBalance(ctx context.Context, address string) (*BalanceInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

// EstimateGas 估算 Gas（Solana 使用不同的费用模型）
func (c *SolanaClient) EstimateGas(ctx context.Context, from, to string, value *big.Int, data []byte) (*GasEstimation, error) {
	// TODO: 实现 Solana 费用估算
	return nil, fmt.Errorf("not implemented")
}

// GetFeeHistory 获取手续费历史
func (c *SolanaClient) GetFeeHistory(ctx context.Context, blockCount uint64, newestBlock string) (*FeeHistory, error) {
	return nil, fmt.Errorf("not implemented")
}

// SendRawTransaction 发送原始交易
func (c *SolanaClient) SendRawTransaction(ctx context.Context, rawTx []byte) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// GetTransactionReceipt 获取交易回执
func (c *SolanaClient) GetTransactionReceipt(ctx context.Context, txHash string) (*TransactionInfo, error) {
	return nil, fmt.Errorf("not implemented")
}
