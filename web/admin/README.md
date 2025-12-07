# 区块链钱包管理系统 - 前端

现代化的管理后台前端应用，使用 React + TypeScript + Vite + Ant Design 构建。

## 🚀 快速开始

### 安装依赖

```bash
cd web/admin
npm install
```

### 开发模式

```bash
npm run dev
```

前端开发服务器将在 `http://localhost:3000` 启动，并自动代理后端API请求到 `http://localhost:8081`。

### 构建生产版本

```bash
npm run build
```

构建产物将输出到 `dist` 目录。

### 预览生产构建

```bash
npm run preview
```

## 📁 项目结构

```
web/admin/
├── src/
│   ├── components/      # 公共组件
│   │   ├── Layout.tsx   # 主布局组件
│   │   └── PrivateRoute.tsx  # 路由守卫
│   ├── pages/           # 页面组件
│   │   ├── Login.tsx    # 登录页
│   │   ├── Dashboard.tsx    # 统计概览
│   │   ├── Deposits.tsx     # 充值管理
│   │   ├── Withdrawals.tsx  # 提现管理
│   │   ├── Accounts.tsx     # 账号管理
│   │   ├── Assets.tsx       # 资产管理
│   │   ├── Risk.tsx         # 风控管理
│   │   └── Settings.tsx     # 系统设置
│   ├── stores/          # 状态管理
│   │   └── authStore.ts # 认证状态
│   ├── utils/           # 工具函数
│   │   └── api.ts       # API客户端
│   ├── App.tsx          # 根组件
│   └── main.tsx         # 入口文件
├── index.html
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## 🔐 登录信息

默认账号：
- 用户名：`admin`
- 密码：`admin123`

⚠️ **生产环境必须修改默认密码！**

## 🎨 功能模块

### 1. 统计概览
- 资产种类统计
- 充值数据统计
- 状态和链分布图表
- 资产统计表格

### 2. 充值管理
- 充值记录查询（支持用户ID、资产、状态筛选）
- 手动确认充值入账
- 充值状态和确认数显示

### 3. 提现管理
- 提现记录查询
- 审批/拒绝提现请求
- 风控评分显示

### 4. 账号管理
- 用户账户搜索
- 余额查询（可用/冻结/待处理）
- 余额调整（增加/减少/冻结/解冻）

### 5. 资产管理
- 资产列表查看
- 添加新资产

### 6. 风控管理
- 白名单管理
- 黑名单管理
- 风控配置

### 7. 系统设置
- 基本设置
- 功能开关
- 安全配置

## 🔧 配置

### API代理配置

在 `vite.config.ts` 中配置了API代理：

```typescript
server: {
  proxy: {
    '/admin/api': {
      target: 'http://localhost:8081',
      changeOrigin: true,
    },
  },
}
```

### 环境变量

可以创建 `.env` 文件配置环境变量：

```env
VITE_API_BASE_URL=http://localhost:8081
```

## 📦 技术栈

- **React 18** - UI框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **Ant Design 5** - UI组件库
- **React Router 6** - 路由管理
- **Zustand** - 状态管理
- **Axios** - HTTP客户端
- **Day.js** - 日期处理

## 🛠️ 开发指南

### 添加新页面

1. 在 `src/pages/` 创建新页面组件
2. 在 `src/App.tsx` 中添加路由
3. 在 `src/components/Layout.tsx` 中添加菜单项

### API调用

使用 `src/utils/api.ts` 中的 `api` 实例：

```typescript
import api from '@/utils/api'

// GET请求
const response = await api.get('/v1/statistics')

// POST请求
await api.post('/v1/assets', data)
```

### 状态管理

使用 Zustand 进行状态管理：

```typescript
import { useAuthStore } from '@/stores/authStore'

const token = useAuthStore((state) => state.token)
const login = useAuthStore((state) => state.login)
```

## 🚢 部署

### 构建并部署到后端

构建后的文件可以部署到后端的 `web/admin/dist` 目录，后端会自动提供静态文件服务。

```bash
npm run build
cp -r dist/* ../web/admin/
```

### 独立部署

也可以将前端独立部署到Nginx等Web服务器：

```nginx
server {
    listen 80;
    server_name admin.example.com;
    root /path/to/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /admin/api {
        proxy_pass http://backend:8081;
    }
}
```

## 📝 License

MIT
