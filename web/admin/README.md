# 后台管理系统使用指南

## 启动方式

### 1. 启动后端服务

```bash
cd cmd/server
go run main.go
```

服务启动后，管理后台可通过以下地址访问：
- 前端界面: http://localhost:8081/admin
- API 接口: http://localhost:8081/admin/api/v1

### 2. 访问管理后台

在浏览器中打开：`http://localhost:8081/admin`

## 登录信息

**默认账号：**
- 用户名: `admin`
- 密码: `admin123`

⚠️ **生产环境必须修改默认密码！**

## 功能模块

### 1. 统计概览
- 总用户数
- 资产种类
- 总充值数
- 总提现数

### 2. 充值管理
- 查看所有充值记录
- 手动确认充值入账
- 查看充值状态

### 3. 提现管理
- 查看所有提现记录
- 审批提现请求
- 拒绝提现请求
- 查看风控评分

### 4. 账号管理
- 搜索用户账户
- 查看账户余额
- 调整余额（增加/减少/冻结/解冻）

### 5. 资产管理
- 查看所有资产
- 添加新资产

### 6. 风控管理
- 管理白名单
- 管理黑名单
- 配置风控参数

## API 认证

所有管理后台 API 都需要 JWT Token 认证。

### 获取 Token

```bash
curl -X POST http://localhost:8081/admin/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

响应：
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-02T12:00:00Z"
}
```

### 使用 Token

```bash
curl http://localhost:8081/admin/api/v1/statistics \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 修改默认密码

### 方式一：修改代码

编辑 `internal/api/handlers/auth.go`：

```go
// 修改登录验证逻辑
if req.Username == "admin" && req.Password == "your-new-password" {
    // ...
}
```

### 方式二：使用环境变量（推荐）

1. 在 `internal/config/config.go` 添加配置
2. 从环境变量读取用户名密码
3. 支持多用户管理

## 安全建议

1. **修改默认密码**：生产环境必须修改
2. **使用 HTTPS**：生产环境启用 HTTPS
3. **JWT Secret**：修改 `jwtSecret` 变量
4. **Token 过期时间**：根据需要调整过期时间
5. **IP 白名单**：限制管理后台访问 IP
6. **审计日志**：记录所有管理操作

## 开发扩展

### 添加新功能模块

1. 在 `internal/api/handlers/admin.go` 添加处理函数
2. 在 `internal/api/router.go` 添加路由
3. 在前端 `web/admin/index.html` 添加界面

### 集成数据库用户管理

```go
type User struct {
    Username string
    Password string // 应该存储哈希值
    Role     string
}

func (h *AuthHandler) Login(c *gin.Context) {
    // 从数据库查询用户
    // 验证密码哈希
    // 生成 token
}
```

## 故障排查

### 无法访问管理后台

1. 检查服务是否启动：`curl http://localhost:8081/health`
2. 检查端口是否正确
3. 查看浏览器控制台错误

### 登录失败

1. 确认用户名密码正确
2. 检查后端日志
3. 确认 JWT Secret 配置正确

### API 返回 401

1. 检查 Token 是否过期
2. 确认请求头包含 Authorization
3. 重新登录获取新 Token

