package scanner

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
	"github.com/jinmu/go-blockchain/internal/wallet/rpc"
	"github.com/jinmu/go-blockchain/internal/wallet/store"
)

// SolanaScanner Solana 链扫描器实现
type SolanaScanner struct {
	chain          domain.ChainType
	rpcClient      rpc.Client
	store          store.RepositoryProvider
	depositHandler DepositHandler
	reorgHandler   ReorgHandler
	blockFetcher   BlockFetcher
	reorgDetector  *ReorgDetector

	// 扫描状态
	currentHeight uint64
	scanning      bool
	mu            sync.RWMutex

	// 配置
	scanInterval  time.Duration // 扫描间隔（Solana 约 400ms 一个区块）
	confirmations uint64        // 所需确认数
	reorgDepth    uint64        // 重组检测深度
	startHeight   uint64        // 起始扫描高度（0 表示从最新开始）
}

// NewSolanaScanner 创建 Solana 链扫描器
func NewSolanaScanner(
	chain domain.ChainType,
	rpcClient rpc.Client,
	store store.RepositoryProvider,
	confirmations uint64,
	reorgDepth uint64,
) *SolanaScanner {
	scanner := &SolanaScanner{
		chain:         chain,
		rpcClient:     rpcClient,
		store:         store,
		scanInterval:  1 * time.Second, // Solana 约 400ms 一个区块，设置为1秒扫描一次
		confirmations: confirmations,
		reorgDepth:    reorgDepth,
	}

	// 初始化重组检测器
	scanner.blockFetcher = scanner.getBlockFetcher()

	return scanner
}

// getBlockFetcher 获取区块获取器
func (s *SolanaScanner) getBlockFetcher() BlockFetcher {
	return NewRPCBlockFetcher(s.rpcClient)
}

// Chain 返回链类型
func (s *SolanaScanner) Chain() domain.ChainType {
	return s.chain
}

// SetBlockFetcher 设置区块获取器
func (s *SolanaScanner) SetBlockFetcher(fetcher BlockFetcher) {
	s.blockFetcher = fetcher
	if s.reorgDetector != nil {
		s.reorgDetector.SetBlockFetcher(fetcher)
	}
}

// Subscribe 订阅新区块和充值事件
func (s *SolanaScanner) Subscribe(ctx context.Context, handler DepositHandler, reorgHandler ReorgHandler) error {
	s.mu.Lock()
	if s.scanning {
		s.mu.Unlock()
		return fmt.Errorf("scanner already running")
	}
	s.scanning = true
	s.depositHandler = handler
	s.reorgHandler = reorgHandler

	// 初始化重组检测器
	if s.reorgDetector == nil {
		s.reorgDetector = NewReorgDetector(s.chain, s.store, s.blockFetcher, reorgHandler)
	} else {
		s.reorgDetector.SetReorgHandler(reorgHandler)
	}
	s.mu.Unlock()

	// 获取当前区块高度
	currentHeight, err := s.rpcClient.GetLatestBlockHeight(ctx)
	if err != nil {
		s.mu.Lock()
		s.scanning = false
		s.mu.Unlock()
		return fmt.Errorf("failed to get latest block height: %w", err)
	}

	// 如果设置了起始高度，使用起始高度；否则从当前高度开始
	if s.startHeight > 0 && s.startHeight < currentHeight {
		s.currentHeight = s.startHeight
	} else {
		s.currentHeight = currentHeight
	}

	// 启动扫描循环
	go s.scanLoop(ctx)

	return nil
}

// scanLoop 扫描循环
func (s *SolanaScanner) scanLoop(ctx context.Context) {
	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.mu.Lock()
			s.scanning = false
			s.mu.Unlock()
			return
		case <-ticker.C:
			if err := s.scanNewBlocks(ctx); err != nil {
				// 记录错误，但继续扫描
				fmt.Printf("Solana scanner error: %v\n", err)
			}
		}
	}
}

// scanNewBlocks 扫描新区块
func (s *SolanaScanner) scanNewBlocks(ctx context.Context) error {
	// 获取最新区块高度
	latestHeight, err := s.rpcClient.GetLatestBlockHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block height: %w", err)
	}

	s.mu.RLock()
	currentHeight := s.currentHeight
	s.mu.RUnlock()

	// 如果当前高度已经是最新，跳过
	if currentHeight >= latestHeight {
		return nil
	}

	// 扫描新区块（从 currentHeight+1 到 latestHeight）
	// Solana 出块很快，限制一次最多扫描100个区块
	maxBlocks := uint64(100)
	endHeight := currentHeight + maxBlocks
	if endHeight > latestHeight {
		endHeight = latestHeight
	}

	for height := currentHeight + 1; height <= endHeight; height++ {
		if err := s.scanBlock(ctx, height); err != nil {
			return fmt.Errorf("failed to scan block %d: %w", height, err)
		}

		// 更新当前高度
		s.mu.Lock()
		s.currentHeight = height
		s.mu.Unlock()

		// 获取区块信息用于保存和重组检测
		block, err := s.rpcClient.GetBlockByHeight(ctx, height)
		if err != nil {
			return fmt.Errorf("failed to get block info: %w", err)
		}

		// 保存区块信息（用于重组检测）
		if err := s.saveBlockInfoWithHash(ctx, height, block.Hash, block.ParentHash); err != nil {
			// 记录错误但不中断扫描
			fmt.Printf("Failed to save block info: %v\n", err)
		}

		// 检测重组（使用重组检测器）
		if s.reorgDetector != nil {
			blockInfo := BlockInfo{
				Height:     height,
				Hash:       block.Hash,
				ParentHash: block.ParentHash,
			}
			isReorg, fromHeight, toHeight, err := s.reorgDetector.CheckBlock(ctx, blockInfo)
			if err != nil {
				fmt.Printf("Reorg detection error: %v\n", err)
			} else if isReorg {
				// 重组已由检测器处理，这里只需要记录
				fmt.Printf("Reorg detected: from %d to %d\n", fromHeight, toHeight)
			}
		}
	}

	return nil
}

// scanBlock 扫描单个区块
func (s *SolanaScanner) scanBlock(ctx context.Context, height uint64) error {
	// 获取区块中的所有交易
	transactions, err := s.rpcClient.GetBlockTransactions(ctx, height)
	if err != nil {
		return fmt.Errorf("failed to get block transactions: %w", err)
	}

	// 获取区块哈希（用于交易处理）
	block, err := s.rpcClient.GetBlockByHeight(ctx, height)
	if err != nil {
		return fmt.Errorf("failed to get block: %w", err)
	}

	// 处理每笔交易
	for _, tx := range transactions {
		if err := s.processTransaction(ctx, tx, height, block.Hash); err != nil {
			// 记录错误但继续处理其他交易
			fmt.Printf("Failed to process transaction %s: %v\n", tx.Hash, err)
		}
	}

	return nil
}

// processTransaction 处理单笔交易
func (s *SolanaScanner) processTransaction(ctx context.Context, tx rpc.Transaction, blockHeight uint64, blockHash string) error {
	// 只处理成功的主币转账（SOL）和 SPL 代币转账
	if !tx.Success {
		return nil
	}

	// 计算确认数（当前高度 - 交易所在区块高度）
	latestHeight, err := s.rpcClient.GetLatestBlockHeight(ctx)
	if err != nil {
		return err
	}
	confirmations := latestHeight - blockHeight + 1

	// 处理主币转账（SOL）
	if tx.Value != nil && tx.Value.Cmp(big.NewInt(0)) > 0 {
		event := domain.DepositEvent{
			Chain:         s.chain,
			AssetSymbol:   "SOL",
			FromAddress:   tx.From,
			ToAddress:     tx.To,
			Amount:        tx.Value,
			TxHash:        tx.Hash,
			BlockHeight:   blockHeight,
			Confirmations: confirmations,
			ObservedAt:    time.Now(),
			Metadata: map[string]string{
				"block_hash": blockHash,
			},
		}

		if s.depositHandler != nil {
			if err := s.depositHandler(ctx, event); err != nil {
				return err
			}
		}
	}

	// 处理 SPL 代币转账（类似 ERC20）
	if len(tx.TokenTransfers) > 0 {
		for _, transfer := range tx.TokenTransfers {
			event := domain.DepositEvent{
				Chain:         s.chain,
				AssetSymbol:   transfer.Symbol,
				FromAddress:   transfer.From,
				ToAddress:     transfer.To,
				Amount:        transfer.Amount,
				TxHash:        tx.Hash,
				BlockHeight:   blockHeight,
				Confirmations: confirmations,
				ObservedAt:    time.Now(),
				Metadata: map[string]string{
					"block_hash":    blockHash,
					"token_address": transfer.TokenAddress,
				},
			}

			if s.depositHandler != nil {
				if err := s.depositHandler(ctx, event); err != nil {
					// 记录错误但继续处理其他转账
					fmt.Printf("Failed to handle token transfer event: %v\n", err)
				}
			}
		}
	}

	return nil
}

// saveBlockInfoWithHash 保存区块信息（使用已知的哈希值）
func (s *SolanaScanner) saveBlockInfoWithHash(ctx context.Context, height uint64, hash, parentHash string) error {
	return s.store.SaveBlock(ctx, s.chain, height, hash, parentHash)
}

// SetScanInterval 设置扫描间隔
func (s *SolanaScanner) SetScanInterval(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scanInterval = interval
}

// SetStartHeight 设置起始扫描高度
func (s *SolanaScanner) SetStartHeight(height uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.startHeight = height
}

// GetCurrentHeight 获取当前扫描高度
func (s *SolanaScanner) GetCurrentHeight() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentHeight
}

// GetStatus 获取扫描器状态（实现Scanner接口）
func (s *SolanaScanner) GetStatus(ctx context.Context) ScannerStatus {
	s.mu.RLock()
	isRunning := s.scanning
	currentHeight := s.currentHeight
	s.mu.RUnlock()

	status := ScannerStatus{
		Chain:         s.chain,
		IsRunning:     isRunning,
		CurrentHeight: currentHeight,
		IsHealthy:     true,
	}

	// 尝试获取最新区块高度
	latestHeight, err := s.rpcClient.GetLatestBlockHeight(ctx)
	if err != nil {
		status.IsHealthy = false
		status.ErrorMessage = fmt.Sprintf("failed to get latest block height: %v", err)
		return status
	}

	status.LatestHeight = latestHeight

	// 如果扫描器正在运行，检查是否落后太多（超过1000个区块认为不健康，因为Solana出块很快）
	if isRunning {
		if latestHeight > currentHeight+1000 {
			status.IsHealthy = false
			status.ErrorMessage = fmt.Sprintf("scanner is lagging: current=%d, latest=%d", currentHeight, latestHeight)
		}
	}

	return status
}

