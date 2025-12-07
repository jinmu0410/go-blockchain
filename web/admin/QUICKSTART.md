# 快速开始指南

## 📦 安装依赖

```bash
cd web/admin
npm install
```

## 🚀 开发模式

启动前端开发服务器：

```bash
npm run dev
```

前端将在 `http://localhost:3000` 启动，并自动代理API请求到后端。

**注意**：确保后端服务已启动在 `http://localhost:8081`

## 🏗️ 构建生产版本

```bash
npm run build
```

构建完成后，将 `dist` 目录中的文件复制到后端的 `web/admin/dist` 目录：

```bash
# 如果后端已经运行，需要重启以加载新的静态文件
cp -r dist/* ../../web/admin/dist/
```

## 🔐 登录

访问 `http://localhost:3000` 或 `http://localhost:8081/admin`

默认账号：
- 用户名：`admin`
- 密码：`admin123`

## 📱 功能页面

- `/dashboard` - 统计概览
- `/deposits` - 充值管理
- `/withdrawals` - 提现管理
- `/accounts` - 账号管理
- `/assets` - 资产管理
- `/risk` - 风控管理
- `/settings` - 系统设置

## 🛠️ 开发提示

1. **API代理**：开发模式下，所有 `/admin/api` 请求会自动代理到 `http://localhost:8081`
2. **热重载**：修改代码后自动刷新页面
3. **TypeScript**：项目使用TypeScript，提供类型安全
4. **Ant Design**：使用Ant Design组件库，UI美观现代

## 🐛 常见问题

### 1. 无法连接后端API

- 确认后端服务已启动
- 检查 `vite.config.ts` 中的代理配置
- 查看浏览器控制台的网络请求

### 2. 登录后跳转失败

- 检查Token是否正常存储（查看localStorage）
- 确认路由配置正确

### 3. 构建后页面空白

- 确认构建成功（检查dist目录）
- 确认后端路由配置正确
- 查看浏览器控制台错误

