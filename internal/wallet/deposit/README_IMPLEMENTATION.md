# 充值流程完整实现说明

## 已实现功能

### 1. EVM 链扫描器 ✅

**文件**: `internal/wallet/scanner/evm_scanner.go`

**功能**:
- ✅ 监听新区块（定期扫描）
- ✅ 解析区块中的交易（主币转账和 ERC20 代币转账）
- ✅ 生成充值事件
- ✅ 区块信息持久化
- ✅ 重组检测集成

**使用方式**:
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

**功能**:
- ✅ 充值事件监听和处理
- ✅ 确认数管理（pending → confirmed → credited）
- ✅ 自动入账逻辑
- ✅ 充值记录持久化
- ✅ 重复充值检测

**流程**:
1. 扫描器发现充值事件
2. 查找对应的用户账户
3. 创建充值记录（状态：pending）
4. 检查确认数
5. 确认数足够后自动入账（状态：credited）
6. 更新用户余额

### 3. 区块重组处理 ✅

**文件**: 
- `internal/wallet/scanner/reorg_detector.go` - 重组检测
- `internal/wallet/deposit/reorg.go` - 重组处理

**功能**:
- ✅ 基于 parentHash 的重组检测算法
- ✅ 查找分叉点
- ✅ 回滚已入账的充值记录
- ✅ 回滚用户余额
- ✅ 删除重组区块数据

**重组检测流程**:
1. 比较新区块的 parentHash 与数据库中的上一个区块 hash
2. 如果不匹配，向上查找分叉点
3. 找到分叉点后，回滚从分叉点到当前高度的所有充值
4. 删除数据库中的重组区块
5. 保存新链上的区块

### 4. 布隆过滤器 ✅

**文件**: `internal/wallet/deposit/bloom_filter.go`

**功能**:
- ✅ 快速过滤非关注地址的交易
- ✅ 减少不必要的处理
- ✅ 支持自定义容量和误报率

**使用方式**:
```go
bloom := deposit.NewSimpleBloomFilter(10000, 0.01) // 10000个地址，1%误报率
bloom.Add([]byte(address))
if bloom.Test([]byte(address)) {
    // 处理交易
}
```

### 5. 确认数管理 ✅

**文件**: `internal/wallet/deposit/confirmation_manager.go`

**功能**:
- ✅ 按链类型配置确认数
- ✅ 自动检查确认数
- ✅ 确认数足够后自动入账

**默认确认数**:
- Bitcoin: 6 个确认
- EVM: 12 个确认
- Solana: 32 个确认

## 集成说明

### 在 App 中启动扫描器

```go
// 在 app.go 的 StartDepositListeners 中
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
        
        // 启动扫描
        scanner.Subscribe(ctx, depositHandler, reorgHandler)
    }
    return nil
}
```

### 充值处理流程

```
扫描器发现交易
    ↓
检查是否关注地址（布隆过滤器）
    ↓
创建充值记录（pending）
    ↓
等待确认数
    ↓
确认数足够 → 更新状态（confirmed）
    ↓
自动入账 → 更新余额
    ↓
更新状态（credited）
```

### 重组处理流程

```
检测到 parentHash 不匹配
    ↓
向上查找分叉点
    ↓
查找受影响区块的充值记录
    ↓
回滚已入账的充值
    ↓
回滚用户余额
    ↓
删除重组区块数据
    ↓
保存新链区块
```

## API 使用

### 查询充值记录

```bash
# 获取充值详情
GET /api/v1/deposits/:tx_hash

# 列出充值记录（需要实现）
GET /api/v1/deposits?user_id=xxx&asset=ETH&status=credited
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
    confirmations, // 确认数要求
    reorgDepth,    // 重组检测深度
)

// 可选配置
scanner.SetScanInterval(12 * time.Second) // 扫描间隔
scanner.SetStartHeight(1000000)           // 起始扫描高度
```

### 确认数配置

在 `Manager.getRequiredConfirmations()` 中配置各链的确认数要求。

## 注意事项

1. **RPC 节点**: 确保 RPC 节点稳定可用
2. **扫描间隔**: 根据链的出块时间调整
3. **确认数**: 根据业务需求调整确认数要求
4. **重组深度**: 设置合理的重组检测深度
5. **错误处理**: 扫描错误不应中断整个流程
6. **性能优化**: 使用布隆过滤器减少不必要的处理

## 待完善功能

- [ ] 实现 Bitcoin 和 Solana 扫描器
- [ ] 完善 ERC20 代币转账解析
- [ ] 实现充值记录查询 API
- [ ] 添加充值统计和监控
- [ ] 实现充值通知（Webhook）

