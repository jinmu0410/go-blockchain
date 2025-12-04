# 风控系统使用指南

## 概述

风控系统提供了完整的提现风险评估和审批功能，包括：
- 规则引擎：可配置的多规则评估系统
- 白名单/黑名单：地址管理
- 限额管理：单笔、日累计、月累计限额
- 频率限制：提现次数和间隔控制
- 风控日志：完整的审计追踪

## 快速开始

### 1. 基本使用

```go
import (
    "github.com/jinmu/go-blockchain/internal/wallet/risk"
    "github.com/jinmu/go-blockchain/internal/wallet/store"
)

// 创建风控控制器
store := store.NewInMemoryStore()
riskController := risk.NewRiskController(store, nil) // nil 使用默认配置

// 在 Manager 中使用
manager := service.NewManager(store,
    service.WithRiskController(riskController),
)
```

### 2. 自定义配置

```go
config := &risk.Config{
    SingleMaxAmount:   big.NewInt(1000000000000000000), // 1 ETH
    DailyMaxAmount:    big.NewInt(10000000000000000000), // 10 ETH
    MonthlyMaxAmount:  big.NewInt(100000000000000000000), // 100 ETH
    MaxCountPerDay:    5,
    MinInterval:       10 * time.Minute,
    MinAccountAge:     7 * 24 * time.Hour, // 7天
    AutoApproveScore:  0.3,
    ManualReviewScore: 0.5,
    RejectScore:       0.7,
}

riskController := risk.NewRiskController(store, config)
```

## API 使用

### 白名单管理

```bash
# 添加地址到白名单
curl -X POST http://localhost:8081/api/v1/risk/whitelist \
  -H "Content-Type: application/json" \
  -d '{
    "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "chain": "evm",
    "remark": "合作伙伴地址"
  }'

# 从白名单移除
curl -X DELETE http://localhost:8081/api/v1/risk/whitelist/0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb?chain=evm

# 列出白名单
curl http://localhost:8081/api/v1/risk/whitelist?chain=evm&limit=100
```

### 黑名单管理

```bash
# 添加地址到黑名单
curl -X POST http://localhost:8081/api/v1/risk/blacklist \
  -H "Content-Type: application/json" \
  -d '{
    "address": "0x1234567890123456789012345678901234567890",
    "chain": "evm",
    "remark": "已知恶意地址"
  }'

# 从黑名单移除
curl -X DELETE http://localhost:8081/api/v1/risk/blacklist/0x1234567890123456789012345678901234567890?chain=evm

# 列出黑名单
curl http://localhost:8081/api/v1/risk/blacklist?chain=evm&limit=100
```

### 配置管理

```bash
# 获取当前配置
curl http://localhost:8081/api/v1/risk/config

# 更新配置
curl -X PUT http://localhost:8081/api/v1/risk/config \
  -H "Content-Type: application/json" \
  -d '{
    "single_max_amount": "1000000000000000000",
    "daily_max_amount": "10000000000000000000",
    "max_count_per_day": 10,
    "min_interval": 300000000000,
    "auto_approve_score": 0.3,
    "manual_review_score": 0.5,
    "reject_score": 0.7
  }'
```

### 风控日志查询

```bash
# 查询所有日志
curl http://localhost:8081/api/v1/risk/logs

# 按用户查询
curl http://localhost:8081/api/v1/risk/logs?user_id=user123

# 按提现ID查询
curl http://localhost:8081/api/v1/risk/logs?withdrawal_id=withdrawal_123

# 分页查询
curl http://localhost:8081/api/v1/risk/logs?limit=50&offset=0
```

## 风控规则说明

### 1. 金额限制规则

- **单笔限额**：超过此金额，风险评分 +0.5
- **日累计限额**：超过此金额，风险评分 +0.4
- **月累计限额**：超过此金额，风险评分 +0.3

### 2. 频率限制规则

- **每日最大次数**：超过此次数，风险评分 +0.4
- **最小间隔**：间隔过短，风险评分 +0.3

### 3. 地址规则

- **黑名单**：直接拒绝（评分 = 1.0）
- **白名单**：降低风险评分（-0.2）

### 4. 账户年龄规则

- 账户创建时间越短，风险越高
- 最高贡献 0.3 的风险评分

### 5. 地址风险评分规则

- 根据地址的历史行为评分
- 高风险地址贡献 0.5 * 地址风险评分

## 决策流程

```
风险评分 < 0.3  → 自动通过
0.3 ≤ 评分 < 0.5 → 建议人工审核
0.5 ≤ 评分 < 0.7 → 需要人工审核
评分 ≥ 0.7       → 直接拒绝
```

## 编程接口

### 添加自定义规则

```go
// 实现 Rule 接口
type CustomRule struct{}

func (r *CustomRule) Name() string {
    return "自定义规则"
}

func (r *CustomRule) Priority() int {
    return 5
}

func (r *CustomRule) Evaluate(ctx context.Context, req domain.WithdrawalRequest, data *EvaluationData) (float64, string, error) {
    // 实现评估逻辑
    if someCondition {
        return 0.2, "触发自定义规则", nil
    }
    return 0, "", nil
}

// 使用自定义规则
rules := []risk.Rule{
    risk.NewAddressRule(),
    risk.NewAmountLimitRule(...),
    &CustomRule{},
}
engine := risk.NewRuleEngine(rules)
```

## 注意事项

1. **配置更新**：更新配置会重新创建规则引擎，确保配置正确
2. **白名单/黑名单**：优先级最高，黑名单地址会直接拒绝
3. **日志记录**：所有风控决策都会记录日志，用于审计
4. **性能考虑**：规则评估是同步的，确保规则执行快速
5. **数据收集**：需要从存储层查询历史数据，确保存储层实现正确

## 扩展开发

### 实现数据库存储

```go
// 实现 AddressListRepository 和 RiskLogRepository
type PostgresAddressListStore struct {
    db *sql.DB
}

func (s *PostgresAddressListStore) AddToWhitelist(ctx context.Context, address string, chain string, remark string) error {
    // 实现数据库插入
}

// 使用数据库存储
addressList := NewPostgresAddressListStore(db)
riskLog := NewPostgresRiskLogStore(db)
riskController := risk.NewRiskControllerWithRepos(store, config, addressList, riskLog)
```

### 集成外部风控服务

```go
type ExternalRiskController struct {
    client *http.Client
    url    string
}

func (r *ExternalRiskController) EvaluateWithdrawal(ctx context.Context, req domain.WithdrawalRequest) (domain.WithdrawalDecision, error) {
    // 调用外部风控API
    // 返回决策结果
}
```

