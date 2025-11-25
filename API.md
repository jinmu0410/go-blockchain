# Wallet API 文档

## 启动服务

```bash
cd cmd/server
go run main.go
```

服务默认运行在 `http://localhost:8080`

## API 端点

### 健康检查

```bash
GET /health
```

响应：
```json
{
  "status": "ok",
  "service": "wallet-api"
}
```

### 资产管理

#### 注册资产

```bash
POST /api/v1/assets
Content-Type: application/json

{
  "symbol": "ETH",
  "chain": "evm",
  "decimals": 18,
  "token_addr": ""
}
```

#### 获取资产信息

```bash
GET /api/v1/assets/:symbol
```

#### 列出所有资产

```bash
GET /api/v1/assets
```

### 账户管理

#### 创建账户

```bash
POST /api/v1/accounts
Content-Type: application/json

{
  "user_id": "user123",
  "asset_symbol": "ETH"
}
```

响应：
```json
{
  "user_id": "user123",
  "asset_symbol": "ETH",
  "address": "0x...",
  "chain": "evm",
  "created_at": "2024-01-01T00:00:00Z"
}
```

#### 获取账户信息

```bash
GET /api/v1/accounts/:user_id/:asset_symbol
```

### 余额查询

#### 获取余额

```bash
GET /api/v1/balances/:user_id/:asset_symbol
```

响应：
```json
{
  "user_id": "user123",
  "asset": "ETH",
  "balance": "1000000000000000000",
  "frozen": "0",
  "pending": "0"
}
```

#### 列出用户所有余额

```bash
GET /api/v1/balances/:user_id
```

### 充值查询

#### 获取充值记录

```bash
GET /api/v1/deposits/:tx_hash
```

#### 列出充值记录

```bash
GET /api/v1/deposits?user_id=user123&asset=ETH&status=credited
```

### 提现管理

#### 创建提现请求

```bash
POST /api/v1/withdrawals
Content-Type: application/json

{
  "user_id": "user123",
  "asset_symbol": "ETH",
  "to_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "amount": "1000000000000000000",
  "chain": "evm"
}
```

响应：
```json
{
  "id": "withdrawal_id",
  "user_id": "user123",
  "asset_symbol": "ETH",
  "to_address": "0x...",
  "amount": "1000000000000000000",
  "status": "completed",
  "tx_hash": "0x...",
  "created_at": "2024-01-01T00:00:00Z"
}
```

#### 获取提现记录

```bash
GET /api/v1/withdrawals/:id
```

#### 列出提现记录

```bash
GET /api/v1/withdrawals?user_id=user123&status=completed
```

## 使用示例

### 使用 curl

```bash
# 1. 创建账户
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "asset_symbol": "ETH"
  }'

# 2. 查询余额
curl http://localhost:8080/api/v1/balances/user123/ETH

# 3. 创建提现
curl -X POST http://localhost:8080/api/v1/withdrawals \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "asset_symbol": "ETH",
    "to_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "amount": "1000000000000000000"
  }'
```

### 使用 Postman

1. 导入环境变量：`BASE_URL = http://localhost:8080`
2. 创建请求集合，包含所有 API 端点
3. 使用环境变量设置 URL

## 错误响应格式

```json
{
  "error": "错误描述信息"
}
```

常见错误码：
- `400`: 请求参数错误
- `404`: 资源不存在
- `500`: 服务器内部错误

