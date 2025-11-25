# 快速启动指南

## 📋 前置准备

### 1. 确保已安装 Go
```bash
go version
# 应该显示 go 1.24.2 或更高版本
```

### 2. 确保 PostgreSQL 已启动并创建数据库
```bash
# 连接到 PostgreSQL
psql -h localhost -U postgres

# 创建数据库
CREATE DATABASE wallet;

# 退出
\q
```

### 3. 执行建表语句
```bash
psql -h localhost -U postgres -d wallet -f database_schema.sql
```

## 🚀 启动服务

### 方式一：直接运行（推荐）

```bash
# 1. 安装依赖
go mod download

# 2. 启动服务
cd cmd/server
go run main.go
```

### 方式二：编译后运行

```bash
# 1. 编译
go build -o wallet-server ./cmd/server

# 2. 运行
./wallet-server
```

服务启动后，你会看到类似输出：
```
Starting server on 0.0.0.0:8080
Server started successfully on http://0.0.0.0:8080
API documentation:
  Health: GET http://0.0.0.0:8080/health
  Assets: POST http://0.0.0.0:8080/api/v1/assets
  ...
```

## 🧪 测试接口

### 方式一：使用测试脚本（推荐）

```bash
# 确保脚本有执行权限
chmod +x test_api.sh

# 运行测试脚本（需要安装 jq 用于格式化 JSON）
# macOS: brew install jq
# Ubuntu: sudo apt-get install jq
./test_api.sh
```

### 方式二：使用 curl 手动测试

#### 1. 健康检查
```bash
curl http://localhost:8080/health
```

#### 2. 注册资产
```bash
curl -X POST http://localhost:8080/api/v1/assets \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "ETH",
    "chain": "evm",
    "decimals": 18
  }'
```

#### 3. 创建账户
```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "asset_symbol": "ETH"
  }'
```

#### 4. 查询余额
```bash
curl http://localhost:8080/api/v1/balances/user123/ETH
```

#### 5. 查询账户信息
```bash
curl http://localhost:8080/api/v1/accounts/user123/ETH
```

#### 6. 列出所有资产
```bash
curl http://localhost:8080/api/v1/assets
```

#### 7. 创建提现请求
```bash
curl -X POST http://localhost:8080/api/v1/withdrawals \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "asset_symbol": "ETH",
    "to_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "amount": "1000000000000000000",
    "chain": "evm"
  }'
```

### 方式三：使用 Postman 或 HTTPie

#### 使用 HTTPie（更友好的命令行工具）
```bash
# 安装 HTTPie
# macOS: brew install httpie
# Ubuntu: sudo apt-get install httpie

# 测试接口
http GET localhost:8080/health
http POST localhost:8080/api/v1/accounts user_id=user123 asset_symbol=ETH
http GET localhost:8080/api/v1/balances/user123/ETH
```

## 📝 完整测试流程示例

```bash
# 1. 启动服务（在另一个终端）
cd cmd/server && go run main.go

# 2. 健康检查
curl http://localhost:8080/health

# 3. 注册 ETH 资产
curl -X POST http://localhost:8080/api/v1/assets \
  -H "Content-Type: application/json" \
  -d '{"symbol":"ETH","chain":"evm","decimals":18}'

# 4. 创建用户账户
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"user_id":"alice","asset_symbol":"ETH"}'

# 5. 查询账户信息（会返回生成的地址）
curl http://localhost:8080/api/v1/accounts/alice/ETH

# 6. 查询余额（初始为 0）
curl http://localhost:8080/api/v1/balances/alice/ETH

# 7. 列出所有资产
curl http://localhost:8080/api/v1/assets

# 8. 列出用户所有余额
curl http://localhost:8080/api/v1/balances/alice
```

## ⚙️ 配置说明

### 环境变量（可选）

如果需要修改默认配置，可以设置环境变量：

```bash
# 服务器配置
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

# 数据库配置（已设置默认值，通常不需要修改）
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=123456
export DB_NAME=wallet
export DB_SSLMODE=disable

# RPC 配置
export RPC_ETHEREUM=https://eth.llamarpc.com

# 钱包配置
export WALLET_MASTER_SEED=your-32-byte-seed
```

## 🔍 常见问题

### 1. 端口被占用
```bash
# 修改端口
export SERVER_PORT=8081
```

### 2. 数据库连接失败
- 检查 PostgreSQL 是否运行：`pg_isready -h localhost -p 5432`
- 检查数据库是否存在：`psql -h localhost -U postgres -l`
- 检查用户名密码是否正确

### 3. 表不存在
确保已执行建表语句：
```bash
psql -h localhost -U postgres -d wallet -f database_schema.sql
```

### 4. 依赖安装失败
```bash
# 清理并重新下载
go clean -modcache
go mod download
```

## 📚 更多信息

- 详细 API 文档：查看 [API.md](./API.md)
- 项目说明：查看 [README.md](./README.md)

