# Go Blockchain Wallet

一个基于 Go 语言的交易所钱包系统，支持多链资产管理、充值、提现等功能。

## 项目结构

```
go-blockchain/
├── cmd/
│   └── server/          # 服务器入口
│       └── main.go
├── internal/
│   ├── api/             # API 层
│   │   ├── handlers/    # HTTP 处理器
│   │   ├── middleware/  # 中间件
│   │   └── router.go    # 路由配置
│   ├── app/             # 应用初始化
│   │   └── app.go
│   ├── config/          # 配置管理
│   │   └── config.go
│   ├── wallet/          # 钱包核心模块
│   │   ├── domain/      # 领域模型
│   │   ├── service/     # 业务服务
│   │   ├── store/       # 数据仓储
│   │   ├── deposit/     # 充值模块
│   │   ├── withdrawal/  # 提现模块
│   │   ├── scanner/     # 链扫描器
│   │   ├── signer/      # 签名机
│   │   ├── risk/        # 风控模块
│   │   ├── rpc/         # RPC 客户端
│   │   ├── bip/         # BIP 钱包协议
│   │   └── funding/     # 资金调度
│   └── blockchain/      # 区块链客户端（旧代码，可迁移）
├── pkg/                 # 公共包
├── test/                # 测试文件
├── go.mod
├── go.sum
├── README.md
└── API.md               # API 文档
```

## 功能特性

- ✅ 多链支持（EVM、Bitcoin、Solana）
- ✅ BIP32/BIP44 钱包分层协议
- ✅ 充值监听和自动入账
- ✅ 提现审批和签名
- ✅ 区块重组（Reorg）处理
- ✅ 风控集成
- ✅ RESTful API
- ✅ 内存存储（可扩展为数据库）

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置环境变量（可选）

```bash
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080
export RPC_ETHEREUM=https://eth.llamarpc.com
export WALLET_MASTER_SEED=your-32-byte-seed
```

### 3. 启动服务

```bash
cd cmd/server
go run main.go
```

服务将在 `http://localhost:8080` 启动

### 4. 测试 API

```bash
# 健康检查
curl http://localhost:8080/health

# 创建账户
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "asset_symbol": "ETH"}'

# 查询余额
curl http://localhost:8080/api/v1/balances/user123/ETH
```

详细 API 文档请参考 [API.md](./API.md)

## 架构设计

### 分层架构

```
┌─────────────────┐
│   API Layer     │  HTTP Handlers
├─────────────────┤
│  Service Layer  │  Business Logic
├─────────────────┤
│  Domain Layer   │  Domain Models
├─────────────────┤
│  Store Layer    │  Data Persistence
└─────────────────┘
```

### 核心模块

1. **Service Manager**: 核心业务服务，协调各个模块
2. **Deposit Processor**: 处理充值事件，支持布隆过滤器和重组回滚
3. **Withdrawal Processor**: 处理提现，支持热钱包选择、Gas 估算、批量交易
4. **Scanner**: 区块链扫描器，监听新区块和交易
5. **Signer**: 签名机，隔离私钥管理
6. **RPC Client**: 统一的 RPC 客户端接口

## 开发指南

### 添加新链支持

1. 在 `internal/wallet/rpc/` 实现 `Client` 接口
2. 在 `internal/wallet/bip/coin_type.go` 添加币种类型
3. 在 `internal/wallet/domain/types.go` 添加链类型

### 扩展存储

实现 `store.RepositoryProvider` 接口，支持：
- PostgreSQL
- MySQL
- MongoDB
- Redis

### 自定义风控

实现 `risk.Controller` 接口，集成风控系统。

## 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| SERVER_HOST | 服务器地址 | 0.0.0.0 |
| SERVER_PORT | 服务器端口 | 8080 |
| RPC_ETHEREUM | Ethereum RPC 端点 | https://eth.llamarpc.com |
| RPC_BITCOIN | Bitcoin RPC 端点 | - |
| RPC_SOLANA | Solana RPC 端点 | - |
| WALLET_MASTER_SEED | 主种子（32字节） | - |

## 安全注意事项

⚠️ **生产环境必须配置**：

1. 使用安全的密钥管理服务（HSM/KMS）
2. 主种子必须安全存储，不能硬编码
3. 所有私钥必须加密存储
4. 启用 HTTPS
5. 添加认证和授权中间件
6. 实现请求限流
7. 添加审计日志

## 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/wallet/service/...
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License
