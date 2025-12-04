# 充值流程完整实现总结

## ✅ 已实现功能

根据 ROADMAP Phase 2.2.1，已完成充值流程的完整实现：

### 1. 扫描器实现 ✅

**文件**: `internal/wallet/scanner/evm_scanner.go`

**核心功能**:
- ✅ EVM 链扫描器实现
  - 定期扫描新区块（默认 12 秒间隔）
  - 解析区块中的所有交易
  - 识别主币转账（ETH/BNB 等）
  - 识别 ERC20 代币转账（TODO: 需要完善日志解析）
- ✅ 区块信息持久化
  - 保存区块高度、哈希、父哈希
  - 用于重组检测
- ✅ 重组检测集成
  - 每个新区块都会检测重组
  - 自动触发重组处理

**使用示例**:
```go
scanner := scanner.NewEVMScanner(
    domain.ChainEVM,
    rpcClient,
    store,
    12, // 确认数
    12, // 重组深度
)

scanner.Subscribe(ctx, depositHandler, reorgHandler)
```

### 2. 充值处理流程 ✅

**文件**: `internal/wallet/service/manager.go` (HandleDepositEvent)

**核心功能**:
- ✅ 充值事件监听和处理
  - 接收扫描器发送的充值事件
  - 查找对应的用户账户
- ✅ 确认数管理
  - 状态流转：pending → confirmed → credited
  - 根据链类型设置不同的确认数要求
  - 自动检查确认数并更新状态
- ✅ 自动入账逻辑
  - 确认数足够后自动增加用户余额
  - 更新充值记录状态为 credited
  - 记录入账时间
- ✅ 充值记录持久化
  - 保存所有充值记录到存储层
  - 支持查询和更新
- ✅ 重复充值检测
  - 检查交易哈希是否已存在
  - 避免重复处理

**确认数配置**:
- Bitcoin: 6 个确认
- EVM (Ethereum/Polygon/BSC): 12 个确认
- Solana: 32 个确认

### 3. 区块重组处理 ✅

**文件**: 
- `internal/wallet/scanner/reorg_detector.go` - 重组检测算法
- `internal/wallet/deposit/reorg.go` - 重组处理逻辑

**核心功能**:
- ✅ 重组检测算法
  - 基于 parentHash 比较
  - 检测新区块的父哈希是否与数据库中的上一个区块匹配
  - 如果不匹配，向上查找分叉点
- ✅ 回滚已入账的充值记录
  - 查找受影响区块范围内的所有充值
  - 将已入账的充值状态改为 failed
  - 调用 Manager.RollbackDeposit 回滚余额
- ✅ 重新处理新区块链上的交易
  - 删除重组区块数据
  - 保存新链上的区块
  - 扫描器会继续扫描新链上的交易

**重组处理流程**:
```
检测到 parentHash 不匹配
    ↓
向上查找分叉点（最多查找 100 个区块）
    ↓
查找 fromHeight 到 toHeight 范围内的充值记录
    ↓
回滚已入账的充值（状态改为 failed）
    ↓
回滚用户余额（Debit）
    ↓
删除数据库中的重组区块
    ↓
保存新链上的区块
    ↓
触发重组处理器（可选）
```

### 4. 布隆过滤器优化 ✅

**文件**: `internal/wallet/deposit/bloom_filter.go`

**功能**:
- ✅ 快速过滤非关注地址的交易
- ✅ 减少不必要的 RPC 调用和处理
- ✅ 支持自定义容量和误报率

**使用方式**:
```go
bloom := deposit.NewSimpleBloomFilter(10000, 0.01) // 10000个地址，1%误报率
bloom.Add([]byte(address))
if bloom.Test([]byte(address)) {
    // 处理交易
}
```

### 5. 确认数管理器 ✅

**文件**: `internal/wallet/deposit/confirmation_manager.go`

**功能**:
- ✅ 按链类型配置确认数
- ✅ 自动检查确认数
- ✅ 确认数足够后自动入账

## 集成到应用

### 在 app.go 中启动

```go
func (a *App) StartDepositListeners(ctx context.Context) error {
    for chainType, rpcClient := range a.RPC {
        scanner := scanner.NewEVMScanner(
            chainType,
            rpcClient,
            a.Store,
            12, // 确认数
            12, // 重组深度
        )
        
        a.Manager.RegisterScanner(scanner)
        
        depositHandler := func(ctx context.Context, event domain.DepositEvent) error {
            return a.Manager.HandleDepositEvent(ctx, event)
        }
        
        reorgHandler := func(ctx context.Context, chain domain.ChainType, fromHeight, toHeight uint64) error {
            reorgHandler := deposit.NewDefaultReorgHandlerWithConfig(
                a.Store,
                a.Manager,
                nil,
                deposit.DefaultReorgDepths,
            )
            return reorgHandler.OnReorg(ctx, chain, fromHeight, toHeight)
        }
        
        scanner.Subscribe(ctx, depositHandler, reorgHandler)
    }
    return nil
}
```

## 完整流程示例

### 正常充值流程

```
1. 用户向地址 0xABC... 转账 1 ETH
   ↓
2. 扫描器扫描到新区块，发现交易
   ↓
3. 检查地址是否在布隆过滤器中
   ↓
4. 查找地址对应的用户账户（user123）
   ↓
5. 创建充值记录（状态：pending，确认数：1/12）
   ↓
6. 等待更多确认...
   ↓
7. 确认数达到 12，更新状态（confirmed）
   ↓
8. 自动入账：增加 user123 的 ETH 余额
   ↓
9. 更新状态（credited），记录入账时间
```

### 重组处理流程

```
1. 扫描器发现新区块的 parentHash 不匹配
   ↓
2. 向上查找分叉点（假设在高度 1000）
   ↓
3. 查找高度 1000-1010 范围内的所有充值记录
   ↓
4. 找到 3 笔已入账的充值
   ↓
5. 回滚这 3 笔充值：
   - 状态改为 failed
   - 从用户余额中扣除金额
   ↓
6. 删除数据库中的区块 1000-1010
   ↓
7. 保存新链上的区块
   ↓
8. 扫描器继续扫描新链上的交易
```

## API 接口

### 查询充值记录

```bash
# 获取充值详情
GET /api/v1/deposits/:tx_hash

# 列出充值记录（需要实现查询接口）
GET /api/v1/deposits?user_id=user123&asset=ETH&status=credited
```

### 手动确认充值（管理员）

```bash
POST /admin/api/v1/transactions/deposits/:tx_hash/credit
```

## 配置说明

### 扫描器配置

```go
scanner := scanner.NewEVMScanner(
    chain,
    rpcClient,
    store,
    confirmations, // 确认数要求（默认 12）
    reorgDepth,    // 重组检测深度（默认 12）
)

// 可选配置
scanner.SetScanInterval(12 * time.Second) // 扫描间隔
scanner.SetStartHeight(1000000)           // 起始扫描高度（用于从指定高度开始扫描）
```

### 确认数配置

在 `Manager.getRequiredConfirmations()` 中配置：

```go
func (m *Manager) getRequiredConfirmations(chain domain.ChainType) uint64 {
    switch chain {
    case domain.ChainBitcoin:
        return 6
    case domain.ChainEVM:
        return 12
    case domain.ChainSolana:
        return 32
    default:
        return 6
    }
}
```

### 重组深度配置

在 `deposit.DefaultReorgDepths` 中配置：

```go
var DefaultReorgDepths = ReorgDepthConfig{
    domain.ChainBitcoin: 6,  // Bitcoin: 6 个区块
    domain.ChainEVM:     12, // EVM: 12 个区块
    domain.ChainSolana:  32, // Solana: 32 个区块
}
```

## 测试建议

### 1. 单元测试

```go
// 测试扫描器
func TestEVMScanner_ScanBlock(t *testing.T) {
    // 测试区块扫描和交易解析
}

// 测试充值处理
func TestManager_HandleDepositEvent(t *testing.T) {
    // 测试充值事件处理
}

// 测试重组处理
func TestReorgHandler_OnReorg(t *testing.T) {
    // 测试重组回滚
}
```

### 2. 集成测试

- 使用测试网 RPC 节点
- 发送测试交易
- 验证充值流程
- 模拟重组场景

## 注意事项

1. **RPC 节点稳定性**: 确保 RPC 节点稳定可用，建议使用多个备用节点
2. **扫描间隔**: 根据链的出块时间调整（EVM 约 12 秒，Bitcoin 约 10 分钟）
3. **确认数**: 根据业务需求调整，确认数越高越安全但入账越慢
4. **重组深度**: 设置合理的重组检测深度，避免误报
5. **错误处理**: 扫描错误不应中断整个流程，应该记录日志并继续
6. **性能优化**: 
   - 使用布隆过滤器减少不必要的处理
   - 批量处理交易
   - 异步处理充值事件
7. **数据一致性**: 所有余额操作必须使用事务，确保一致性

## 待完善功能

- [ ] 完善 ERC20 代币转账解析（解析 Transfer 事件）
- [ ] 实现 Bitcoin 和 Solana 扫描器
- [ ] 实现充值记录查询 API（按用户、资产、状态查询）
- [ ] 添加充值统计和监控指标
- [ ] 实现充值通知（Webhook）
- [ ] 优化扫描性能（批量查询、并发处理）
- [ ] 添加扫描器健康检查

## 相关文件

- `internal/wallet/scanner/evm_scanner.go` - EVM 扫描器实现
- `internal/wallet/scanner/reorg_detector.go` - 重组检测器
- `internal/wallet/deposit/processor.go` - 充值处理器
- `internal/wallet/deposit/reorg.go` - 重组处理
- `internal/wallet/deposit/bloom_filter.go` - 布隆过滤器
- `internal/wallet/deposit/confirmation_manager.go` - 确认数管理
- `internal/wallet/service/manager.go` - 业务逻辑集成
- `internal/app/app.go` - 应用启动集成

