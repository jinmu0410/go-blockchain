package scanner

import (
	"context"
	"fmt"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
	"github.com/jinmu/go-blockchain/internal/wallet/store"
)

// ReorgDetector 重组检测器，实现基于 parentHash 的检测逻辑
type ReorgDetector struct {
	chain        domain.ChainType
	repos        store.BlockRepository
	fetcher      BlockFetcher
	reorgHandler ReorgHandler
}

// NewReorgDetector 创建重组检测器
func NewReorgDetector(chain domain.ChainType, repos store.BlockRepository, fetcher BlockFetcher, reorgHandler ReorgHandler) *ReorgDetector {
	return &ReorgDetector{
		chain:        chain,
		repos:        repos,
		fetcher:      fetcher,
		reorgHandler: reorgHandler,
	}
}

// SetReorgHandler 设置重组处理器
func (d *ReorgDetector) SetReorgHandler(handler ReorgHandler) {
	d.reorgHandler = handler
}

// SetBlockFetcher 设置区块获取器
func (d *ReorgDetector) SetBlockFetcher(fetcher BlockFetcher) {
	d.fetcher = fetcher
}

// CheckBlock 检查新区块是否发生重组
// 逻辑：
// 1. 从数据库获取上一个区块（height = newBlock.Height - 1）
// 2. 比较：newBlock.ParentHash == dbBlock.Hash
// 3. 如果一致，没有重组，正常保存
// 4. 如果不一致，发生重组，查找分叉点并回滚
// 返回: (isReorg, fromHeight, toHeight, error)
// - isReorg: 是否发生重组
// - fromHeight: 重组开始的区块高度（如果发生重组）
// - toHeight: 重组结束的区块高度（当前区块高度）
func (d *ReorgDetector) CheckBlock(ctx context.Context, newBlock BlockInfo) (bool, uint64, uint64, error) {
	// 步骤1: 获取数据库中存储的上一个区块（当前最新区块，height = newBlock.Height - 1）
	latestBlock, err := d.repos.GetLatestBlock(ctx, d.chain)
	if err != nil {
		// 如果数据库中没有区块记录，说明是第一次扫描，直接保存
		if err == domain.ErrBlockNotFound {
			if err := d.repos.SaveBlock(ctx, d.chain, newBlock.Height, newBlock.Hash, newBlock.ParentHash); err != nil {
				return false, 0, 0, fmt.Errorf("failed to save first block: %w", err)
			}
			return false, 0, newBlock.Height, nil
		}
		return false, 0, 0, fmt.Errorf("failed to get latest block: %w", err)
	}

	// 步骤2: 检查新区块的 parentHash 是否等于数据库中存储的上一个区块的 hash
	if newBlock.ParentHash == latestBlock.Hash {
		// 没有重组，正常保存新区块
		if err := d.repos.SaveBlock(ctx, d.chain, newBlock.Height, newBlock.Hash, newBlock.ParentHash); err != nil {
			return false, 0, 0, fmt.Errorf("failed to save block: %w", err)
		}
		return false, 0, newBlock.Height, nil
	}

	// parentHash 不匹配，说明发生了重组
	// 需要从链上往上查找，找到最开始发生重组的区块高度
	// 逻辑：从链上获取区块 N-1，和数据库中的区块 N-1 比较 hash
	// 如果不一致，继续往上查找，直到找到 hash 一致的区块（分叉点）
	fromHeight, err := d.findForkPoint(ctx, newBlock, latestBlock)
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to find fork point: %w", err)
	}

	// 删除数据库中从 fromHeight 开始的所有区块
	if err := d.repos.DeleteBlocksFromHeight(ctx, d.chain, fromHeight); err != nil {
		return false, 0, 0, fmt.Errorf("failed to delete blocks: %w", err)
	}

	// 保存新链上的区块（从分叉点到当前区块）
	currentHeight := fromHeight
	currentHash := ""

	// 从分叉点开始，重新获取并保存新链上的区块
	for currentHeight <= newBlock.Height {
		var block BlockInfo
		if currentHeight == newBlock.Height {
			// 使用传入的新区块
			block = newBlock
		} else {
			// 从 RPC 获取中间区块
			block, err = d.fetcher.GetBlockByHeight(ctx, currentHeight)
			if err != nil {
				return false, 0, 0, fmt.Errorf("failed to get block at height %d: %w", currentHeight, err)
			}
		}

		// 验证区块的 parentHash（除了分叉点的第一个区块）
		if currentHeight > fromHeight {
			if block.ParentHash != currentHash {
				return false, 0, 0, fmt.Errorf("block chain mismatch at height %d", currentHeight)
			}
		}

		// 保存区块
		if err := d.repos.SaveBlock(ctx, d.chain, block.Height, block.Hash, block.ParentHash); err != nil {
			return false, 0, 0, fmt.Errorf("failed to save block at height %d: %w", currentHeight, err)
		}

		currentHash = block.Hash
		currentHeight++
	}

	// 触发重组处理
	if d.reorgHandler != nil {
		if err := d.reorgHandler(ctx, d.chain, fromHeight, newBlock.Height); err != nil {
			return false, 0, 0, fmt.Errorf("failed to handle reorg: %w", err)
		}
	}

	return true, fromHeight, newBlock.Height, nil
}

// findForkPoint 查找分叉点（重组开始的区块高度）
// 逻辑：
// 1. 从链上获取区块 N-1（newBlock.Height - 1）
// 2. 从数据库获取区块 N-1
// 3. 比较链上的 hash 和数据库的 hash
// 4. 如果不一致，继续往上查找 N-2, N-3...
// 5. 直到找到 hash 一致的区块，这就是分叉点
// 6. 返回分叉点的下一个高度（重组开始的区块高度）
func (d *ReorgDetector) findForkPoint(ctx context.Context, newBlock BlockInfo, latestBlockInfo store.BlockInfo) (uint64, error) {
	// 从新区块的父区块开始往上查找（newBlock.Height - 1）
	currentHeight := newBlock.Height - 1

	// 最多查找 maxDepth 个区块（避免无限循环）
	maxDepth := uint64(100) // 可以根据链类型配置
	searched := uint64(0)

	// 限制查找范围：不能低于数据库中最老区块的高度
	minHeight := uint64(0)
	if latestBlockInfo.Height > maxDepth {
		minHeight = latestBlockInfo.Height - maxDepth
	}

	for searched < maxDepth && currentHeight >= minHeight {
		// 步骤1: 从链上获取当前高度的区块
		chainBlock, err := d.fetcher.GetBlockByHeight(ctx, currentHeight)
		if err != nil {
			return 0, fmt.Errorf("failed to get block from chain at height %d: %w", currentHeight, err)
		}

		// 步骤2: 从数据库获取当前高度的区块
		dbBlockInfo, err := d.repos.GetBlock(ctx, d.chain, currentHeight)
		if err != nil {
			// 如果数据库中不存在该高度的区块，继续往上查找
			currentHeight--
			searched++
			continue
		}

		// 步骤3: 比较链上的 hash 和数据库的 hash
		if chainBlock.Hash == dbBlockInfo.Hash {
			// 找到了匹配的区块，这就是分叉点
			// 分叉点就是 currentHeight，重组从 currentHeight+1 开始
			return currentHeight + 1, nil
		}

		// hash 不一致，说明这个区块也发生了重组，继续往上查找
		currentHeight--
		searched++
	}

	// 如果找不到匹配点，说明重组深度很大，返回一个安全的分叉点
	// 使用 latestBlockInfo.Height + 1 作为分叉点（回滚所有新区块）
	if latestBlockInfo.Height > 0 {
		return latestBlockInfo.Height + 1, nil
	}

	return 0, fmt.Errorf("failed to find fork point after searching %d blocks", maxDepth)
}
