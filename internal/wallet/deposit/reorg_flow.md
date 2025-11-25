# 区块重组处理流程图

## 时序图

```mermaid
sequenceDiagram
    participant Scanner as 扫描器
    participant Processor as deposit.Processor
    participant ReorgHandler as DefaultReorgHandler
    participant Store as RepositoryProvider
    participant Manager as ReorgManager (可选)
    participant Ledger as Ledger

    Scanner->>Processor: HandleReorg(chain, blockHeight)
    Processor->>ReorgHandler: OnReorg(chain, blockHeight)
    
    ReorgHandler->>ReorgHandler: 计算回滚范围<br/>fromHeight = blockHeight - reorgDepth
    
    ReorgHandler->>Store: FindDepositsByBlockRange<br/>(chain, fromHeight, blockHeight)
    Store-->>ReorgHandler: []DepositRecord
    
    loop 遍历每条充值记录
        alt 状态 == DepositCredited
            ReorgHandler->>Store: FindAccountByAddress
            Store-->>ReorgHandler: WalletAccount
            
            ReorgHandler->>ReorgHandler: 聚合回滚信息<br/>(userID:asset)
            
            ReorgHandler->>Store: UpdateDeposit<br/>(Status=Failed)
            
            opt 提供 Manager
                ReorgHandler->>Manager: RollbackDeposit(record)
                Manager->>Store: Debit(userID, asset, amount)
                Manager->>Store: UpdateDeposit(Status=Failed)
            end
            
            opt 提供 Ledger
                ReorgHandler->>Ledger: Rollback(txHash)
            end
        else 其他状态
            Note over ReorgHandler: 跳过（未入账或已失败）
        end
    end
    
    alt 未提供 Manager
        ReorgHandler->>ReorgHandler: 按 userID:asset 聚合
        loop 批量回滚余额
            ReorgHandler->>Store: Debit(userID, asset, totalAmount)
        end
    end
    
    ReorgHandler-->>Processor: Success/Error
    Processor-->>Scanner: 处理完成
```

## 状态转换图

```mermaid
stateDiagram-v2
    [*] --> DepositPending: 扫描到交易
    DepositPending --> DepositConfirmed: 确认数 >= RequiredConfirmations
    DepositConfirmed --> DepositCredited: 余额入账
    
    DepositCredited --> DepositFailed: 区块重组回滚
    DepositPending --> DepositFailed: 区块重组（可选）
    DepositConfirmed --> DepositFailed: 区块重组（可选）
    
    DepositFailed --> [*]
```

## 数据流图

```mermaid
flowchart TD
    A[扫描器检测重组] --> B[计算回滚范围<br/>fromHeight ~ blockHeight]
    B --> C[查询受影响充值记录]
    C --> D{记录状态?}
    
    D -->|DepositCredited| E[查找用户账户]
    D -->|其他状态| F[跳过]
    
    E --> G[聚合回滚信息<br/>userID:asset]
    G --> H[更新记录状态=Failed]
    H --> I{提供Manager?}
    
    I -->|是| J[Manager.RollbackDeposit<br/>单笔处理]
    I -->|否| K[批量Debit余额]
    
    J --> L[Ledger.Rollback]
    K --> L
    L --> M[完成]
    F --> M
```

## 核心代码逻辑

### 1. 查找受影响记录

```go
// 计算回滚范围
fromHeight := blockHeight - reorgDepth  // 默认回滚 6 个区块
toHeight := blockHeight

// 查询该范围内的所有充值记录
records, err := repos.FindDepositsByBlockRange(ctx, chain, fromHeight, toHeight)
```

### 2. 筛选并聚合

```go
rollbackStats := make(map[string]*rollbackInfo)  // key: userID:asset

for _, record := range records {
    if record.Status != domain.DepositCredited {
        continue  // 只回滚已入账的
    }
    
    account, _ := repos.FindAccountByAddress(ctx, record.ToAddress, record.AssetSymbol)
    key := account.UserID + ":" + record.AssetSymbol
    
    // 聚合金额
    rollbackStats[key].totalAmount += record.Amount
}
```

### 3. 执行回滚

```go
// 方式1：使用 Manager（推荐，包含业务逻辑）
if manager != nil {
    for _, record := range info.records {
        manager.RollbackDeposit(ctx, record)  // 内部会 Debit + UpdateDeposit
    }
}

// 方式2：直接批量 Debit
else {
    for _, info := range rollbackStats {
        repos.Debit(ctx, info.userID, info.asset, info.totalAmount)
    }
}
```

## 关键设计点

1. **只回滚已入账记录**：`DepositCredited` 状态才需要回滚余额
2. **批量优化**：按 `userID:asset` 聚合，减少数据库操作
3. **可选组件**：Manager 和 Ledger 都是可选的，提供灵活性
4. **错误隔离**：单条记录失败不影响其他记录的处理
5. **幂等性**：重复调用应该安全（已回滚的记录状态为 Failed）

