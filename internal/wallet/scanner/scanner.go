package scanner

import (
	"context"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
	"github.com/jinmu/go-blockchain/internal/wallet/rpc"
)

// DepositHandler 处理扫描到的充值
type DepositHandler func(ctx context.Context, event domain.DepositEvent) error

// ReorgHandler 处理重组事件
type ReorgHandler func(ctx context.Context, chain domain.ChainType, fromHeight, toHeight uint64) error

// BlockInfo 区块信息（从 RPC 获取）
type BlockInfo struct {
	Height     uint64
	Hash       string
	ParentHash string
}

// BlockFetcher 从 RPC 获取区块信息的接口
type BlockFetcher interface {
	// GetBlockByHeight 根据高度获取区块信息
	GetBlockByHeight(ctx context.Context, height uint64) (BlockInfo, error)
}

// RPCBlockFetcher 使用统一的 RPC 客户端实现 BlockFetcher
type RPCBlockFetcher struct {
	rpcClient rpc.Client
}

// NewRPCBlockFetcher 创建基于 RPC 客户端的 BlockFetcher
func NewRPCBlockFetcher(rpcClient rpc.Client) *RPCBlockFetcher {
	return &RPCBlockFetcher{rpcClient: rpcClient}
}

// GetBlockByHeight 实现 BlockFetcher 接口
func (f *RPCBlockFetcher) GetBlockByHeight(ctx context.Context, height uint64) (BlockInfo, error) {
	block, err := f.rpcClient.GetBlockByHeight(ctx, height)
	if err != nil {
		return BlockInfo{}, err
	}
	return BlockInfo{
		Height:     block.Height,
		Hash:       block.Hash,
		ParentHash: block.ParentHash,
	}, nil
}

// ScannerStatus 扫描器状态
type ScannerStatus struct {
	Chain         domain.ChainType `json:"chain"`
	IsRunning     bool             `json:"is_running"`
	CurrentHeight uint64           `json:"current_height"`
	LatestHeight  uint64           `json:"latest_height"`
	IsHealthy     bool             `json:"is_healthy"`
	ErrorMessage  string           `json:"error_message,omitempty"`
}

// Scanner 区块链扫描器接口
type Scanner interface {
	Chain() domain.ChainType
	// Subscribe 订阅新区块和充值事件
	Subscribe(ctx context.Context, handler DepositHandler, reorgHandler ReorgHandler) error
	// SetBlockFetcher 设置区块获取器（用于重组检测）
	SetBlockFetcher(fetcher BlockFetcher)
	// GetStatus 获取扫描器状态（用于健康检查）
	GetStatus(ctx context.Context) ScannerStatus
}
