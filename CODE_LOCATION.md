# 代码位置说明

本文档说明各个功能模块的代码位置。

## 充值流程相关代码

### 1. EVM 链扫描器实现

**文件**: `internal/wallet/scanner/evm_scanner.go`

#### 定期扫描新区块（默认 12 秒间隔）

```119:138:internal/wallet/scanner/evm_scanner.go
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
```

#### 解析区块中的所有交易

```200:223:internal/wallet/scanner/evm_scanner.go
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
```

#### 识别主币转账（ETH/BNB 等）

```239:260:internal/wallet/scanner/evm_scanner.go
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
```

#### 识别 ERC20 代币转账

```262:280:internal/wallet/scanner/evm_scanner.go
	// 处理 ERC20 代币转账
	if tx.TokenTransfer != nil {
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
```

**注意**: ERC20 代币转账解析需要从交易日志中解析 Transfer 事件，当前代码中 `tx.TokenTransfer` 需要从 RPC 客户端实现中获取。实际实现需要解析交易回执中的 logs。

#### 区块信息持久化

```300:303:internal/wallet/scanner/evm_scanner.go
// saveBlockInfoWithHash 保存区块信息（使用已知的哈希值）
func (s *EVMScanner) saveBlockInfoWithHash(ctx context.Context, height uint64, hash, parentHash string) error {
	return s.store.SaveBlock(ctx, s.chain, height, hash, parentHash)
}
```

保存区块信息的调用位置：

```168:178:internal/wallet/scanner/evm_scanner.go
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
```

#### 重组检测集成

```179:188:internal/wallet/scanner/evm_scanner.go
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
```

重组检测器实现：

**文件**: `internal/wallet/scanner/reorg_detector.go`

```29:118:internal/wallet/scanner/reorg_detector.go
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
```

## 充值处理流程代码

### 充值事件处理

**文件**: `internal/wallet/service/manager.go`

```133:200:internal/wallet/service/manager.go
// HandleDepositEvent 处理充值事件
func (m *Manager) HandleDepositEvent(ctx context.Context, event domain.DepositEvent) error {
	// 查找账户
	account, err := m.accounts.FindAccountByAddress(ctx, event.ToAddress, event.AssetSymbol)
	if err != nil {
		// 如果找不到账户，记录但不处理（可能是非关注地址）
		return nil
	}

	// 检查是否已存在该充值记录
	existing, err := m.deposits.GetDeposit(ctx, event.TxHash)
	if err == nil {
		// 已存在，更新确认数
		existing.Confirmations = event.Confirmations
		if existing.Confirmations >= existing.RequiredConfirmations && existing.Status == domain.DepositPending {
			existing.Status = domain.DepositConfirmed
			if err := m.deposits.UpdateDeposit(ctx, existing); err != nil {
				return err
			}
			// 如果确认数足够，执行入账
			return m.creditDeposit(ctx, account.UserID, existing)
		}
		return m.deposits.UpdateDeposit(ctx, existing)
	}

	// 创建新的充值记录
	requiredConfirmations := m.getRequiredConfirmations(event.Chain)
	record := domain.DepositRecord{
		TxHash:                event.TxHash,
		Chain:                 event.Chain,
		AssetSymbol:           event.AssetSymbol,
		Amount:                event.Amount,
		FromAddress:           event.FromAddress,
		ToAddress:             event.ToAddress,
		BlockHeight:           event.BlockHeight,
		Confirmations:         event.Confirmations,
		RequiredConfirmations: requiredConfirmations,
		Status:                domain.DepositPending,
		ObservedAt:            event.ObservedAt,
		Metadata:              event.Metadata,
	}

	// 保存充值记录
	if err := m.deposits.SaveDeposit(ctx, account.UserID, record); err != nil {
		return err
	}

	// 如果确认数足够，立即入账
	if record.Confirmations >= record.RequiredConfirmations {
		record.Status = domain.DepositConfirmed
		if err := m.deposits.UpdateDeposit(ctx, record); err != nil {
			return err
		}
		return m.creditDeposit(ctx, account.UserID, record)
	}

	return nil
}
```

### 自动入账逻辑

```162:169:internal/wallet/service/manager.go
func (m *Manager) creditDeposit(ctx context.Context, userID string, record domain.DepositRecord) error {
	// 检查是否已经入账
	if record.Status == domain.DepositCredited {
		return nil
	}

	// 增加余额
	if err := m.balances.Credit(ctx, userID, record.AssetSymbol, record.Amount); err != nil {
		return err
	}

	// 更新状态
	record.Status = domain.DepositCredited
	record.CreditedAt = time.Now()
	return m.deposits.UpdateDeposit(ctx, record)
}
```

## 区块重组处理代码

### 重组处理逻辑

**文件**: `internal/wallet/deposit/reorg.go`

```98:181:internal/wallet/deposit/reorg.go
// OnReorg 处理区块重组
// 逻辑：
// 1. 查找 fromHeight 到 toHeight 范围内的所有充值记录
// 2. 将已入账的记录状态设置为 DepositFailed（回滚状态）
// 3. 回滚用户余额
// 4. 回滚资金流水
// fromHeight: 重组开始的区块高度（分叉点的下一个高度）
// toHeight: 重组结束的区块高度（当前最新区块）
func (h *DefaultReorgHandler) OnReorg(ctx context.Context, chain domain.ChainType, fromHeight, toHeight uint64) error {
	// 步骤1: 查找受影响区块范围内的所有充值记录（从 fromHeight 到 toHeight）
	records, err := h.repos.FindDepositsByBlockRange(ctx, chain, fromHeight, toHeight)
	if err != nil {
		return fmt.Errorf("failed to find deposits in reorg range: %w", err)
	}

	// 统计回滚信息
	rollbackStats := make(map[string]*rollbackInfo) // userID:asset -> info

	// 遍历所有受影响的充值记录
	for _, record := range records {
		// 只回滚已经入账的记录
		if record.Status != domain.DepositCredited {
			continue
		}

		// 查找对应的用户账户
		account, err := h.repos.FindAccountByAddress(ctx, record.ToAddress, record.AssetSymbol)
		if err != nil {
			// 如果找不到账户，记录错误但继续处理其他记录
			continue
		}

		// 聚合回滚信息（按用户+资产分组）
		key := account.UserID + ":" + record.AssetSymbol
		info, ok := rollbackStats[key]
		if !ok {
			info = &rollbackInfo{
				userID:      account.UserID,
				asset:       record.AssetSymbol,
				totalAmount: big.NewInt(0),
				records:     make([]domain.DepositRecord, 0),
			}
			rollbackStats[key] = info
		}
		info.totalAmount = new(big.Int).Add(info.totalAmount, record.Amount)
		info.records = append(info.records, record)

		// 更新充值记录状态为失败
		record.Status = domain.DepositFailed
		if err := h.repos.UpdateDeposit(ctx, record); err != nil {
			return fmt.Errorf("failed to update deposit status: %w", err)
		}

		// 调用 Ledger 回滚流水
		if h.ledger != nil {
			if err := h.ledger.Rollback(ctx, record.TxHash); err != nil {
				// Ledger 回滚失败不影响主流程，但记录错误
				// 实际生产环境应该记录告警
			}
		}
	}

	// 批量回滚余额（按用户+资产聚合，避免多次数据库操作）
	// 如果提供了 manager，使用 manager 的单笔回滚方法；否则直接批量 Debit
	if h.manager != nil {
		// 使用 manager 的单笔回滚（manager 内部会处理余额扣减）
		for _, info := range rollbackStats {
			for _, record := range info.records {
				if err := h.manager.RollbackDeposit(ctx, record); err != nil {
					return fmt.Errorf("failed to rollback deposit %s: %w", record.TxHash, err)
				}
			}
		}
	} else {
		// 直接批量回滚余额
		for _, info := range rollbackStats {
			if err := h.repos.Debit(ctx, info.userID, info.asset, info.totalAmount); err != nil {
				return fmt.Errorf("failed to debit balance for user %s asset %s: %w", info.userID, info.asset, err)
			}
		}
	}

	return nil
}
```

## 相关文件清单

### 扫描器相关
- `internal/wallet/scanner/evm_scanner.go` - EVM 扫描器主实现
- `internal/wallet/scanner/scanner.go` - 扫描器接口定义
- `internal/wallet/scanner/reorg_detector.go` - 重组检测器

### 充值处理相关
- `internal/wallet/service/manager.go` - 充值事件处理（HandleDepositEvent, creditDeposit）
- `internal/wallet/deposit/processor.go` - 充值处理器
- `internal/wallet/deposit/confirmation_manager.go` - 确认数管理
- `internal/wallet/deposit/reorg.go` - 重组处理逻辑

### RPC 客户端相关
- `internal/wallet/rpc/evm.go` - EVM RPC 客户端实现
- `internal/wallet/rpc/client.go` - RPC 客户端接口
- `internal/wallet/rpc/transaction.go` - 交易数据结构

### 应用集成
- `internal/app/app.go` - 扫描器启动集成（StartDepositListeners）

## 代码流程图

```
启动服务
  ↓
app.StartDepositListeners()
  ↓
创建 EVMScanner
  ↓
scanner.Subscribe() → 启动 scanLoop()
  ↓
scanLoop() → 每 12 秒调用 scanNewBlocks()
  ↓
scanNewBlocks() → 扫描新区块
  ↓
scanBlock() → 获取区块交易
  ↓
processTransaction() → 处理每笔交易
  ↓
识别主币/ERC20 转账 → 生成 DepositEvent
  ↓
depositHandler() → Manager.HandleDepositEvent()
  ↓
创建/更新充值记录 → 检查确认数
  ↓
确认数足够 → creditDeposit() → 增加余额
```

## 关键代码位置总结

| 功能 | 文件 | 函数/方法 |
|------|------|-----------|
| 定期扫描新区块 | `evm_scanner.go` | `scanLoop()` (119-138行) |
| 解析区块交易 | `evm_scanner.go` | `scanBlock()` (200-223行) |
| 识别主币转账 | `evm_scanner.go` | `processTransaction()` (239-260行) |
| 识别ERC20转账 | `evm_scanner.go` | `processTransaction()` (262-280行) |
| 区块信息持久化 | `evm_scanner.go` | `saveBlockInfoWithHash()` (300-303行) |
| 重组检测 | `reorg_detector.go` | `CheckBlock()` (29-118行) |
| 充值事件处理 | `manager.go` | `HandleDepositEvent()` (133-200行) |
| 自动入账 | `manager.go` | `creditDeposit()` (162-169行) |
| 重组处理 | `reorg.go` | `OnReorg()` (98-181行) |

