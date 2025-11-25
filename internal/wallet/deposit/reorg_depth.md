# 重组深度（reorgDepth）判断说明

## 什么是重组深度？

重组深度（Reorg Depth）是指当区块链发生重组时，需要回滚多少个区块内的交易。这个参数的选择直接影响系统的安全性和性能。

## 不同链的重组深度

### 1. **Bitcoin（比特币）**

- **推荐值**: 6 个区块
- **依据**: 
  - Bitcoin 网络约定俗成使用 6 个确认（约 1 小时）
  - 历史上超过 6 个区块的重组极其罕见
  - 6 个区块后基本可以认为交易不可逆

### 2. **Ethereum 及 EVM 兼容链**

- **推荐值**: 12-15 个区块
- **依据**:
  - Ethereum 主网：12 个确认（约 2.5 分钟）
  - Polygon/BSC/Arbitrum 等：通常 12-15 个确认
  - EVM 链出块快，但重组风险相对较高
  - 考虑到 Layer2 和分叉链，建议使用 12 个区块

### 3. **Solana**

- **推荐值**: 32 个区块
- **依据**:
  - Solana 出块极快（约 400ms），32 个区块约 13 秒
  - Solana 使用 PoH（历史证明），重组深度需要更大
  - 实际生产中可能需要 50-100 个确认，但重组深度通常设为 32

### 4. **其他链**

- **Litecoin**: 6 个区块（类似 Bitcoin）
- **Bitcoin Cash**: 10 个区块（更保守）
- **Avalanche**: 1-2 个区块（最终性快）
- **Cosmos 生态**: 通常 1-2 个区块

## 判断依据

### 1. **确认数（Confirmations）**

重组深度通常应该 **≥ 确认数**，因为：
- 如果重组深度 < 确认数，可能回滚已经确认但未入账的交易
- 如果重组深度 = 确认数，刚好覆盖已确认的交易
- 如果重组深度 > 确认数，更安全但可能回滚更多未确认的交易

**推荐**: `reorgDepth = RequiredConfirmations + 安全余量（2-4 个区块）`

### 2. **链的历史重组记录**

- 查看链的历史数据，统计实际发生的最大重组深度
- 例如：Bitcoin 历史上最大重组约 3-4 个区块，设置 6 个是安全的

### 3. **出块时间**

- 出块快的链（如 Solana、Polygon）需要更大的重组深度
- 出块慢的链（如 Bitcoin）重组深度可以较小

### 4. **共识机制**

- **PoW（工作量证明）**: 重组风险较高，需要更大的深度
- **PoS（权益证明）**: 最终性更快，重组深度可以较小
- **BFT（拜占庭容错）**: 几乎无重组，深度可以很小（1-2）

## 代码实现

### 方式 1: 使用默认配置（推荐）

```go
// 使用预定义的默认配置
config := deposit.DefaultReorgDepths
reorgHandler := deposit.NewDefaultReorgHandlerWithConfig(
    store, manager, ledger, config,
)
```

### 方式 2: 自定义配置

```go
// 自定义各链的重组深度
config := deposit.ReorgDepthConfig{
    domain.ChainBitcoin: 6,
    domain.ChainEVM:    15,  // 更保守
    domain.ChainSolana: 50,  // 更保守
}
reorgHandler := deposit.NewDefaultReorgHandlerWithConfig(
    store, manager, ledger, config,
)
```

### 方式 3: 实现自定义 Provider

```go
type CustomDepthProvider struct {
    // 可以从数据库或配置文件读取
}

func (p *CustomDepthProvider) GetReorgDepth(chain domain.ChainType) uint64 {
    // 根据链类型、网络状态等动态计算
    switch chain {
    case domain.ChainEVM:
        // 可以根据网络拥堵情况动态调整
        return 12
    case domain.ChainBitcoin:
        return 6
    default:
        return 6
    }
}

provider := &CustomDepthProvider{}
reorgHandler := deposit.NewDefaultReorgHandler(
    store, manager, ledger, provider, 0,
)
```

### 方式 4: 固定深度（简单场景）

```go
// 所有链使用相同的重组深度
reorgHandler := deposit.NewDefaultReorgHandler(
    store, manager, ledger, nil, 12, // 固定 12 个区块
)
```

## 最佳实践

### 1. **生产环境建议**

```go
// 保守配置（推荐）
config := deposit.ReorgDepthConfig{
    domain.ChainBitcoin: 6,
    domain.ChainEVM:    15,  // 比默认值更保守
    domain.ChainSolana:  50,  // 比默认值更保守
}
```

### 2. **测试环境**

```go
// 可以使用较小的值加快测试
config := deposit.ReorgDepthConfig{
    domain.ChainBitcoin: 3,
    domain.ChainEVM:    6,
    domain.ChainSolana:  10,
}
```

### 3. **动态调整**

对于高价值交易，可以：
- 增加重组深度（更安全）
- 增加确认数要求
- 使用多重签名验证

### 4. **监控告警**

- 监控重组发生频率
- 记录实际重组深度
- 如果发现重组深度超过配置值，及时告警并调整

## 常见问题

### Q: 重组深度设置太大有什么影响？

**A**: 
- 优点：更安全，覆盖更多潜在重组
- 缺点：回滚范围大，可能影响更多交易，性能开销增加

### Q: 重组深度设置太小有什么风险？

**A**: 
- 如果实际重组深度 > 配置值，可能漏掉需要回滚的交易
- 导致用户余额不一致，造成资金损失

### Q: 如何确定合适的重组深度？

**A**: 
1. 参考链的官方文档和社区最佳实践
2. 查看历史重组数据
3. 从保守值开始，根据实际运行情况调整
4. 考虑业务风险承受能力

## 总结

重组深度的判断需要综合考虑：
- ✅ 链的类型和特性
- ✅ 历史重组记录
- ✅ 确认数要求
- ✅ 业务风险承受能力
- ✅ 性能考虑

**推荐做法**: 使用代码中提供的 `DefaultReorgDepths` 作为起点，根据实际运行情况调整。

