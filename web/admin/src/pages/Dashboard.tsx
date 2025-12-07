import { useEffect, useState } from 'react'
import { Card, Row, Col, Statistic, Table, Spin } from 'antd'
import {
  DollarOutlined,
  WalletOutlined,
  SwapOutlined,
} from '@ant-design/icons'
import api from '../utils/api'

interface Statistics {
  total_assets: number
  deposits: {
    total_count: number
    total_amount: string
    by_asset: Record<string, { count: number; amount: string }>
    by_status: Record<string, number>
    by_chain: Record<string, number>
  }
  today_deposits: number
  today_deposit_amount: string
  updated_at: string
}

const Dashboard = () => {
  const [loading, setLoading] = useState(true)
  const [stats, setStats] = useState<Statistics | null>(null)

  useEffect(() => {
    fetchStatistics()
  }, [])

  const fetchStatistics = async () => {
    try {
      const response = await api.get('/v1/statistics')
      setStats(response.data)
    } catch (error) {
      console.error('Failed to fetch statistics:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>统计概览</h1>
      
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="资产种类"
              value={stats?.total_assets || 0}
              prefix={<DollarOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总充值数"
              value={stats?.deposits.total_count || 0}
              prefix={<WalletOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="今日充值"
              value={stats?.today_deposits || 0}
              prefix={<SwapOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总充值金额"
              value={stats?.deposits.total_amount || '0'}
              precision={0}
              prefix={<DollarOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={12}>
          <Card title="充值状态分布" style={{ height: '100%' }}>
            <Row gutter={[16, 16]}>
              {stats?.deposits.by_status &&
                Object.entries(stats.deposits.by_status).map(([status, count]) => (
                  <Col span={12} key={status}>
                    <Statistic
                      title={status}
                      value={count}
                    />
                  </Col>
                ))}
            </Row>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="链分布" style={{ height: '100%' }}>
            <Row gutter={[16, 16]}>
              {stats?.deposits.by_chain &&
                Object.entries(stats.deposits.by_chain).map(([chain, count]) => (
                  <Col span={12} key={chain}>
                    <Statistic
                      title={chain}
                      value={count}
                    />
                  </Col>
                ))}
            </Row>
          </Card>
        </Col>
      </Row>

      <Card title="资产统计" style={{ marginTop: 24 }}>
        <Table
          dataSource={stats?.deposits.by_asset ? Object.entries(stats.deposits.by_asset).map(([asset, data]) => ({
            key: asset,
            asset,
            count: data.count,
            amount: data.amount,
          })) : []}
          columns={[
            { title: '资产', dataIndex: 'asset', key: 'asset' },
            { title: '充值次数', dataIndex: 'count', key: 'count' },
            { title: '总金额', dataIndex: 'amount', key: 'amount' },
          ]}
          pagination={false}
        />
      </Card>
    </div>
  )
}

export default Dashboard
