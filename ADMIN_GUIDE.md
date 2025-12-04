# 后台管理系统使用指南

## 🚀 启动服务

### 1. 启动后端服务

```bash
cd cmd/server
go run main.go
```

服务启动后，你会看到：
```
Starting server on 0.0.0.0:8081
Server started successfully on http://0.0.0.0:8081
```

### 2. 访问管理后台

在浏览器中打开：**http://localhost:8081/admin**

## 🔐 登录

**默认账号：**
- 用户名：`admin`
- 密码：`admin123`

⚠️ **生产环境必须修改默认密码！**

## 📋 功能模块

### 1. 统计概览
- 总用户数
- 资产种类
- 总充值数
- 总提现数

### 2. 充值管理
- 查看所有充值记录
- 手动确认充值入账
- 查看充值状态和确认数

### 3. 提现管理
- 查看所有提现记录
- 审批提现请求（通过/拒绝）
- 查看风控评分和状态

### 4. 账号管理
- 搜索用户账户
- 查看账户余额（可用/冻结/待处理）
- **余额调整**（增加/减少/冻结/解冻）

### 5. 资产管理
- 查看所有资产配置
- 添加新资产

### 6. 风控管理
- **白名单管理**：添加/移除地址
- **黑名单管理**：添加/移除地址
- **风控配置**：调整自动通过、人工审核、拒绝阈值

## 🔧 API 接口

### 登录接口

```bash
POST /admin/api/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

响应：
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-02T12:00:00Z"
}
```

### 管理接口（需要 Token）

所有管理接口都需要在请求头中携带 Token：

```bash
Authorization: Bearer YOUR_TOKEN
```

**示例：**
```bash
# 获取统计信息
curl http://localhost:8081/admin/api/v1/statistics \
  -H "Authorization: Bearer YOUR_TOKEN"

# 手动确认充值
curl -X POST http://localhost:8081/admin/api/v1/transactions/deposits/TX_HASH/credit \
  -H "Authorization: Bearer YOUR_TOKEN"

# 审批提现
curl -X POST http://localhost:8081/admin/api/v1/transactions/withdrawals/WITHDRAWAL_ID/approve \
  -H "Authorization: Bearer YOUR_TOKEN"

# 调整余额
curl -X POST http://localhost:8081/admin/api/v1/accounts/balance/adjust \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "asset_symbol": "ETH",
    "type": "credit",
    "amount": "1000000000000000000",
    "reason": "测试充值"
  }'
```

## 🔒 安全配置

### 修改默认密码

编辑 `internal/api/handlers/auth.go`：

```go
// 修改登录验证逻辑
if req.Username == "admin" && req.Password == "your-new-password" {
    // ...
}
```

### 修改 JWT Secret

编辑 `internal/api/handlers/auth.go`：

```go
var jwtSecret = []byte("your-secret-key-change-in-production")
```

### 生产环境建议

1. ✅ 修改默认密码
2. ✅ 使用强 JWT Secret
3. ✅ 启用 HTTPS
4. ✅ 限制管理后台访问 IP
5. ✅ 添加操作审计日志
6. ✅ 实现多用户权限管理

## 🐛 故障排查

### 无法访问管理后台

1. 检查服务是否启动：`curl http://localhost:8081/health`
2. 检查端口是否正确（默认 8081）
3. 查看浏览器控制台错误

### 登录失败

1. 确认用户名密码正确（admin/admin123）
2. 检查后端日志
3. 确认 JWT Secret 配置正确

### API 返回 401

1. 检查 Token 是否过期（24小时）
2. 确认请求头包含 `Authorization: Bearer TOKEN`
3. 重新登录获取新 Token

### 静态文件无法加载

1. 确认 `web/admin/index.html` 文件存在
2. 检查文件权限
3. 查看服务器日志

## 📝 使用示例

### 完整操作流程

1. **登录管理后台**
   - 访问 http://localhost:8081/admin
   - 输入 admin/admin123

2. **查看统计信息**
   - 点击"统计概览"查看系统数据

3. **管理充值**
   - 点击"充值管理"查看充值记录
   - 点击"确认入账"手动确认充值

4. **管理提现**
   - 点击"提现管理"查看提现记录
   - 点击"通过"或"拒绝"审批提现

5. **管理账号**
   - 点击"账号管理"
   - 输入用户ID和资产符号搜索
   - 调整余额（增加/减少/冻结/解冻）

6. **配置风控**
   - 点击"风控管理"
   - 添加白名单/黑名单地址
   - 调整风控阈值

## 🎯 下一步

- [ ] 实现多用户权限管理
- [ ] 添加操作日志和审计
- [ ] 实现数据导出功能
- [ ] 添加图表和可视化
- [ ] 实现实时通知

