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

// EVMScanner EVM 链扫描器实现
type EVMScanner struct {
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
	scanInterval  time.Duration // 扫描间隔
	confirmations uint64        // 所需确认数
	reorgDepth    uint64        // 重组检测深度
	startHeight   uint64        // 起始扫描高度（0 表示从最新开始）
	maxWorkers    int           // 最大并发worker数（默认5）
	batchSize     uint64        // 批量处理大小（默认10）
}

// NewEVMScanner 创建 EVM 链扫描器
func NewEVMScanner(
	chain domain.ChainType,
	rpcClient rpc.Client,
	store store.RepositoryProvider,
	confirmations uint64,
	reorgDepth uint64,
) *EVMScanner {
	scanner := &EVMScanner{
		chain:         chain,
		rpcClient:     rpcClient,
		store:         store,
		scanInterval:  12 * time.Second, // EVM 链约 12 秒一个区块
		confirmations: confirmations,
		reorgDepth:    reorgDepth,
		maxWorkers:    5,  // 默认5个并发worker
		batchSize:     10, // 默认批量处理10个区块
	}

	// 初始化重组检测器
	scanner.blockFetcher = scanner.getBlockFetcher()

	return scanner
}

// getBlockFetcher 获取区块获取器
func (s *EVMScanner) getBlockFetcher() BlockFetcher {
	return NewRPCBlockFetcher(s.rpcClient)
}

// Chain 返回链类型
func (s *EVMScanner) Chain() domain.ChainType {
	return s.chain
}

// SetBlockFetcher 设置区块获取器
func (s *EVMScanner) SetBlockFetcher(fetcher BlockFetcher) {
	s.blockFetcher = fetcher
	if s.reorgDetector != nil {
		s.reorgDetector.SetBlockFetcher(fetcher)
	}
}

// Subscribe 订阅新区块和充值事件
func (s *EVMScanner) Subscribe(ctx context.Context, handler DepositHandler, reorgHandler ReorgHandler) error {
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
func (s *EVMScanner) scanLoop(ctx context.Context) {
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
				fmt.Printf("EVM scanner error: %v\n", err)
			}
		}
	}
}

// scanNewBlocks 扫描新区块（优化版本：支持并发和批量处理）
func (s *EVMScanner) scanNewBlocks(ctx context.Context) error {
	// 获取最新区块高度
	latestHeight, err := s.rpcClient.GetLatestBlockHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block height: %w", err)
	}

	s.mu.RLock()
	currentHeight := s.currentHeight
	batchSize := s.batchSize
	maxWorkers := s.maxWorkers
	s.mu.RUnlock()

	// 如果当前高度已经是最新，跳过
	if currentHeight >= latestHeight {
		return nil
	}

	// 计算需要扫描的区块数量
	blocksToScan := latestHeight - currentHeight

	// 如果区块数量较少，使用顺序处理
	if blocksToScan <= batchSize {
		return s.scanBlocksSequentially(ctx, currentHeight+1, latestHeight)
	}

	// 批量并发处理
	return s.scanBlocksConcurrently(ctx, currentHeight+1, latestHeight, batchSize, maxWorkers)
}

// scanBlocksSequentially 顺序扫描区块（用于少量区块）
func (s *EVMScanner) scanBlocksSequentially(ctx context.Context, startHeight, endHeight uint64) error {
	for height := startHeight; height <= endHeight; height++ {
		if err := s.scanAndProcessBlock(ctx, height); err != nil {
			return fmt.Errorf("failed to scan block %d: %w", height, err)
		}
		// 更新当前高度
		s.mu.Lock()
		s.currentHeight = height
		s.mu.Unlock()
	}
	return nil
}

// scanBlocksConcurrently 并发扫描区块（用于大量区块）
func (s *EVMScanner) scanBlocksConcurrently(ctx context.Context, startHeight, endHeight uint64, batchSize uint64, maxWorkers int) error {
	// 创建worker池
	type blockJob struct {
		height uint64
	}

	jobs := make(chan blockJob, maxWorkers*2)
	errors := make(chan error, maxWorkers)

	// 启动workers
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				if err := s.scanAndProcessBlock(ctx, job.height); err != nil {
					errors <- fmt.Errorf("failed to scan block %d: %w", job.height, err)
					return
				}
			}
		}()
	}

	// 发送任务
	go func() {
		defer close(jobs)
		for height := startHeight; height <= endHeight; height++ {
			select {
			case <-ctx.Done():
				return
			case jobs <- blockJob{height: height}:
			}
		}
	}()

	// 等待完成
	go func() {
		wg.Wait()
		close(errors)
	}()

	// 收集错误
	var lastError error
	processedCount := uint64(0)
	for err := range errors {
		if err != nil {
			lastError = err
		}
		processedCount++
	}

	// 更新当前高度（批量更新）
	if processedCount > 0 {
		s.mu.Lock()
		s.currentHeight = startHeight + processedCount - 1
		s.mu.Unlock()
	}

	return lastError
}

// scanAndProcessBlock 扫描并处理单个区块（提取公共逻辑）
func (s *EVMScanner) scanAndProcessBlock(ctx context.Context, height uint64) error {
	// 扫描区块交易
	if err := s.scanBlock(ctx, height); err != nil {
		return err
	}

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
			// 清除 ERC20 缓存，确保数据一致性
			s.clearERC20CacheIfNeeded()
		}
	}

	return nil
}

// scanBlock 扫描单个区块
func (s *EVMScanner) scanBlock(ctx context.Context, height uint64) error {
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
func (s *EVMScanner) processTransaction(ctx context.Context, tx rpc.Transaction, blockHeight uint64, blockHash string) error {
	// 只处理成功的主币转账和 ERC20 转账
	if !tx.Success {
		return nil
	}

	// 计算确认数（当前高度 - 交易所在区块高度）
	latestHeight, err := s.rpcClient.GetLatestBlockHeight(ctx)
	if err != nil {
		return err
	}
	confirmations := latestHeight - blockHeight + 1

	// 处理主币转账（ETH, BNB 等）
	if tx.Value != nil && tx.Value.Cmp(big.NewInt(0)) > 0 {
		event := domain.DepositEvent{
			Chain:         s.chain,
			AssetSymbol:   s.getNativeAssetSymbol(),
			FromAddress:   tx.From,
			ToAddress:     tx.To,
			Amount:        tx.Value,
			TxHash:        tx.Hash,
			BlockHeight:   blockHeight,
			Confirmations: confirmations,
			ObservedAt:    time.Now(),
			Metadata: map[string]string{
				"block_hash": blockHash,
				"gas_used":   fmt.Sprintf("%d", tx.GasUsed),
			},
		}

		if s.depositHandler != nil {
			return s.depositHandler(ctx, event)
		}
	}

	// 处理 ERC20 代币转账（支持多个 Transfer 事件）
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
					"gas_used":      fmt.Sprintf("%d", tx.GasUsed),
				},
			}

			if s.depositHandler != nil {
				if err := s.depositHandler(ctx, event); err != nil {
					// 记录错误但继续处理其他转账
					fmt.Printf("Failed to handle token transfer event: %v\n", err)
				}
			}
		}
	} else if tx.TokenTransfer != nil {
		// 向后兼容：如果没有 TokenTransfers，使用 TokenTransfer
		event := domain.DepositEvent{
			Chain:         s.chain,
			AssetSymbol:   tx.TokenTransfer.Symbol,
			FromAddress:   tx.TokenTransfer.From,
			ToAddress:     tx.TokenTransfer.To,
			Amount:        tx.TokenTransfer.Amount,
			TxHash:        tx.Hash,
			BlockHeight:   blockHeight,
			Confirmations: confirmations,
			ObservedAt:    time.Now(),
			Metadata: map[string]string{
				"block_hash":    blockHash,
				"token_address": tx.TokenTransfer.TokenAddress,
				"gas_used":      fmt.Sprintf("%d", tx.GasUsed),
			},
		}

		if s.depositHandler != nil {
			return s.depositHandler(ctx, event)
		}
	}

	return nil
}

// getNativeAssetSymbol 获取主币符号
func (s *EVMScanner) getNativeAssetSymbol() string {
	switch s.chain {
	case domain.ChainEVM:
		// 可以根据 RPC 端点判断是哪个链
		return "ETH" // 默认 ETH，实际应该从配置获取
	default:
		return "ETH"
	}
}

// saveBlockInfoWithHash 保存区块信息（使用已知的哈希值）
func (s *EVMScanner) saveBlockInfoWithHash(ctx context.Context, height uint64, hash, parentHash string) error {
	return s.store.SaveBlock(ctx, s.chain, height, hash, parentHash)
}

// SetScanInterval 设置扫描间隔
func (s *EVMScanner) SetScanInterval(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scanInterval = interval
}

// SetStartHeight 设置起始扫描高度
func (s *EVMScanner) SetStartHeight(height uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.startHeight = height
}

// SetMaxWorkers 设置最大并发worker数
func (s *EVMScanner) SetMaxWorkers(workers int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if workers > 0 {
		s.maxWorkers = workers
	}
}

// SetBatchSize 设置批量处理大小
func (s *EVMScanner) SetBatchSize(size uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if size > 0 {
		s.batchSize = size
	}
}

// GetCurrentHeight 获取当前扫描高度
func (s *EVMScanner) GetCurrentHeight() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentHeight
}

// clearERC20CacheIfNeeded 如果需要，清除 ERC20 缓存
// 在区块重组发生时调用，确保缓存数据的一致性
func (s *EVMScanner) clearERC20CacheIfNeeded() {
	// 类型断言，检查是否是 EVMClient
	if evmClient, ok := s.rpcClient.(*rpc.EVMClient); ok {
		evmClient.ClearERC20Cache()
	}
}

// GetStatus 获取扫描器状态（实现Scanner接口）
func (s *EVMScanner) GetStatus(ctx context.Context) ScannerStatus {
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

	// 如果扫描器正在运行，检查是否落后太多（超过100个区块认为不健康）
	if isRunning {
		if latestHeight > currentHeight+100 {
			status.IsHealthy = false
			status.ErrorMessage = fmt.Sprintf("scanner is lagging: current=%d, latest=%d", currentHeight, latestHeight)
		}
	}

	return status
}
