import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
  Layout as AntLayout,
  Menu,
  Avatar,
  Dropdown,
  Space,
} from 'antd'
import {
  DashboardOutlined,
  WalletOutlined,
  SwapOutlined,
  UserOutlined,
  DollarOutlined,
  SafetyOutlined,
  SettingOutlined,
  LogoutOutlined,
  DatabaseOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '../stores/authStore'
import type { MenuProps } from 'antd'
import './Layout.css'

const { Header, Sider, Content } = AntLayout

const Layout = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { username, logout } = useAuthStore()
  const [collapsed, setCollapsed] = useState(false)

  const menuItems: MenuProps['items'] = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: '统计概览',
    },
    {
      key: '/deposits',
      icon: <WalletOutlined />,
      label: '充值管理',
    },
    {
      key: '/withdrawals',
      icon: <SwapOutlined />,
      label: '提现管理',
    },
    {
      key: '/accounts',
      icon: <UserOutlined />,
      label: '账号管理',
    },
    {
      key: '/assets',
      icon: <DollarOutlined />,
      label: '资产管理',
    },
    {
      key: '/risk',
      icon: <SafetyOutlined />,
      label: '风控管理',
    },
    {
      key: '/address-pool',
      icon: <DatabaseOutlined />,
      label: '地址池管理',
    },
    {
      key: '/settings',
      icon: <SettingOutlined />,
      label: '系统设置',
    },
  ]

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true,
    },
  ]

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key)
  }

  const handleUserMenuClick = ({ key }: { key: string }) => {
    if (key === 'logout') {
      logout()
      navigate('/login')
    }
  }

  return (
    <AntLayout className="admin-layout">
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        width={200}
        theme="dark"
      >
        <div className="logo">
          {!collapsed ? '钱包管理系统' : '钱包'}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>
      <AntLayout>
        <Header className="admin-header">
          <div className="header-right">
            <Space size="large">
              <Dropdown
                menu={{ items: userMenuItems, onClick: handleUserMenuClick }}
                placement="bottomRight"
              >
                <Space className="user-info" style={{ cursor: 'pointer' }}>
                  <Avatar icon={<UserOutlined />} />
                  <span>{username}</span>
                </Space>
              </Dropdown>
            </Space>
          </div>
        </Header>
        <Content className="admin-content">
          <Outlet />
        </Content>
      </AntLayout>
    </AntLayout>
  )
}

export default Layout

