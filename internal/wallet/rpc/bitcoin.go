package rpc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// BitcoinClient Bitcoin 链 RPC 客户端实现（占位符）
type BitcoinClient struct {
	chain domain.ChainType
	// TODO: 实现 Bitcoin RPC 客户端
}

// NewBitcoinClient 创建 Bitcoin 链客户端
func NewBitcoinClient(endpoint string) (*BitcoinClient, error) {
	// TODO: 实现 Bitcoin RPC 连接
	return &BitcoinClient{
		chain: domain.ChainBitcoin,
	}, fmt.Errorf("bitcoin client not implemented yet")
}

// Chain 返回链类型
func (c *BitcoinClient) Chain() domain.ChainType {
	return c.chain
}

// GetBlockByHeight 根据高度获取区块
func (c *BitcoinClient) GetBlockByHeight(ctx context.Context, height uint64) (*BlockInfo, error) {
	// TODO: 实现 Bitcoin 区块查询
	return nil, fmt.Errorf("not implemented")
}

// GetBlockByHash 根据哈希获取区块
func (c *BitcoinClient) GetBlockByHash(ctx context.Context, hash string) (*BlockInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetLatestBlock 获取最新区块
func (c *BitcoinClient) GetLatestBlock(ctx context.Context) (*BlockInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetTransaction 获取交易信息
func (c *BitcoinClient) GetTransaction(ctx context.Context, txHash string) (*TransactionInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetBalance 获取地址余额
func (c *BitcoinClient) GetBalance(ctx context.Context, address string) (*BalanceInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

// EstimateGas 估算 Gas（Bitcoin 不需要）
func (c *BitcoinClient) EstimateGas(ctx context.Context, from, to string, value *big.Int, data []byte) (*GasEstimation, error) {
	return nil, fmt.Errorf("bitcoin does not support gas estimation")
}

// GetFeeHistory 获取手续费历史（Bitcoin 不需要）
func (c *BitcoinClient) GetFeeHistory(ctx context.Context, blockCount uint64, newestBlock string) (*FeeHistory, error) {
	return nil, fmt.Errorf("bitcoin does not support fee history")
}

// SendRawTransaction 发送原始交易
func (c *BitcoinClient) SendRawTransaction(ctx context.Context, rawTx []byte) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// GetTransactionReceipt 获取交易回执（Bitcoin 不需要）
func (c *BitcoinClient) GetTransactionReceipt(ctx context.Context, txHash string) (*TransactionInfo, error) {
	return nil, fmt.Errorf("bitcoin does not support transaction receipt")
}
