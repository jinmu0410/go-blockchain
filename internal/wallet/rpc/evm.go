package rpc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum"
	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// EVMClient EVM 链 RPC 客户端实现
type EVMClient struct {
	client *ethclient.Client
	chain  domain.ChainType
}

// NewEVMClient 创建 EVM 链客户端
func NewEVMClient(endpoint string) (*EVMClient, error) {
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to EVM node: %w", err)
	}

	return &EVMClient{
		client: client,
		chain:  domain.ChainEVM,
	}, nil
}

// Chain 返回链类型
func (c *EVMClient) Chain() domain.ChainType {
	return c.chain
}

// GetBlockByHeight 根据高度获取区块
func (c *EVMClient) GetBlockByHeight(ctx context.Context, height uint64) (*BlockInfo, error) {
	block, err := c.client.BlockByNumber(ctx, big.NewInt(int64(height)))
	if err != nil {
		return nil, fmt.Errorf("failed to get block by height %d: %w", height, err)
	}

	return &BlockInfo{
		Height:     block.Number().Uint64(),
		Hash:       block.Hash().Hex(),
		ParentHash: block.ParentHash().Hex(),
		Timestamp:  block.Time(),
		TxCount:    uint64(block.Transactions().Len()),
	}, nil
}

// GetBlockByHash 根据哈希获取区块
func (c *EVMClient) GetBlockByHash(ctx context.Context, hash string) (*BlockInfo, error) {
	blockHash := common.HexToHash(hash)
	block, err := c.client.BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash %s: %w", hash, err)
	}

	return &BlockInfo{
		Height:     block.Number().Uint64(),
		Hash:       block.Hash().Hex(),
		ParentHash: block.ParentHash().Hex(),
		Timestamp:  block.Time(),
		TxCount:    uint64(block.Transactions().Len()),
	}, nil
}

// GetLatestBlock 获取最新区块
func (c *EVMClient) GetLatestBlock(ctx context.Context) (*BlockInfo, error) {
	block, err := c.client.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	return &BlockInfo{
		Height:     block.Number().Uint64(),
		Hash:       block.Hash().Hex(),
		ParentHash: block.ParentHash().Hex(),
		Timestamp:  block.Time(),
		TxCount:    uint64(block.Transactions().Len()),
	}, nil
}

// GetTransaction 获取交易信息
func (c *EVMClient) GetTransaction(ctx context.Context, txHash string) (*TransactionInfo, error) {
	hash := common.HexToHash(txHash)
	tx, isPending, err := c.client.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction %s: %w", txHash, err)
	}

	info := &TransactionInfo{
		Hash:     txHash,
		GasPrice: tx.GasPrice(),
		GasLimit: tx.Gas(),
		Nonce:    tx.Nonce(),
	}

	if tx.To() != nil {
		info.To = tx.To().Hex()
	}

	if tx.Value() != nil {
		info.Value = tx.Value()
	}

	// 如果交易已确认，获取区块信息
	if !isPending {
		receipt, err := c.client.TransactionReceipt(ctx, hash)
		if err == nil {
			info.BlockHeight = receipt.BlockNumber.Uint64()
			info.BlockHash = receipt.BlockHash.Hex()
			info.Status = receipt.Status
		}
	}

	return info, nil
}

// GetBalance 获取地址余额
func (c *EVMClient) GetBalance(ctx context.Context, address string) (*BalanceInfo, error) {
	addr := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance for %s: %w", address, err)
	}

	nonce, err := c.client.NonceAt(ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce for %s: %w", address, err)
	}

	return &BalanceInfo{
		Address: address,
		Balance: balance,
		Nonce:   nonce,
	}, nil
}

// EstimateGas 估算 Gas
func (c *EVMClient) EstimateGas(ctx context.Context, from, to string, value *big.Int, data []byte) (*GasEstimation, error) {
	fromAddr := common.HexToAddress(from)
	var toAddr *common.Address
	if to != "" {
		addr := common.HexToAddress(to)
		toAddr = &addr
	}

	msg := ethereum.CallMsg{
		From:  fromAddr,
		To:    toAddr,
		Value: value,
		Data:  data,
	}

	gasLimit, err := c.client.EstimateGas(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// 获取最新区块的 BaseFee
	header, err := c.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest header: %w", err)
	}

	baseFee := big.NewInt(0)
	if header.BaseFee != nil {
		baseFee = header.BaseFee
	}

	// 估算 PriorityFee（可以使用 feeHistory 获取）
	priorityFee := big.NewInt(2000000000) // 2 Gwei 默认值

	return &GasEstimation{
		BaseFeePerGas:    baseFee,
		PriorityFeePerGas: priorityFee,
		MaxFeePerGas:     new(big.Int).Add(baseFee, priorityFee),
		GasLimit:         gasLimit,
		BlockNumber:      header.Number.Uint64(),
	}, nil
}

// GetFeeHistory 获取手续费历史（EIP-1559）
func (c *EVMClient) GetFeeHistory(ctx context.Context, blockCount uint64, newestBlock string) (*FeeHistory, error) {
	var blockNumber *big.Int
	if newestBlock != "" && newestBlock != "latest" {
		blockNumber = new(big.Int)
		blockNumber.SetString(newestBlock, 0)
	}

	history, err := c.client.FeeHistory(ctx, blockCount, blockNumber, []float64{0.5})
	if err != nil {
		return nil, fmt.Errorf("failed to get fee history: %w", err)
	}

	return &FeeHistory{
		OldestBlock:   history.OldestBlock,
		BaseFeePerGas: history.BaseFee,
		GasUsedRatio:  history.GasUsedRatio,
		Reward:        history.Reward,
	}, nil
}

// SendRawTransaction 发送原始交易
func (c *EVMClient) SendRawTransaction(ctx context.Context, rawTx []byte) (string, error) {
	tx := new(types.Transaction)
	if err := tx.UnmarshalBinary(rawTx); err != nil {
		return "", fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	err := c.client.SendTransaction(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return tx.Hash().Hex(), nil
}

// GetTransactionReceipt 获取交易回执
func (c *EVMClient) GetTransactionReceipt(ctx context.Context, txHash string) (*TransactionInfo, error) {
	hash := common.HexToHash(txHash)
	receipt, err := c.client.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	tx, _, err := c.client.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	from, err := c.client.TransactionSender(ctx, tx, receipt.BlockHash, receipt.TransactionIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction sender: %w", err)
	}

	return &TransactionInfo{
		Hash:        txHash,
		From:        from.Hex(),
		To:          receipt.ContractAddress.Hex(),
		BlockHeight: receipt.BlockNumber.Uint64(),
		BlockHash:   receipt.BlockHash.Hex(),
		Status:      receipt.Status,
	}, nil
}

