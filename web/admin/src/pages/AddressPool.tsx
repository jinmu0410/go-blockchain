import { useEffect, useState } from 'react'
import {
  Table,
  Card,
  Button,
  Space,
  Modal,
  Form,
  InputNumber,
  Select,
  message,
  Tag,
  Statistic,
  Row,
  Col,
  Input,
} from 'antd'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import api from '../utils/api'
import dayjs from 'dayjs'

interface AddressPoolEntry {
  id: string
  chain: string
  asset_symbol: string
  address: string
  status: 'available' | 'used'
  created_at: string
  used_at?: string
}

interface PoolStats {
  available: number
  used: number
  total: number
}

const AddressPool = () => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [addresses, setAddresses] = useState<AddressPoolEntry[]>([])
  const [stats, setStats] = useState<PoolStats>({ available: 0, used: 0, total: 0 })
  const [modalVisible, setModalVisible] = useState(false)
  const [chainFilter, setChainFilter] = useState<string>('')
  const [assetFilter, setAssetFilter] = useState<string>('')
  const [statusFilter, setStatusFilter] = useState<string>('')

  useEffect(() => {
    fetchAddresses()
    fetchStats()
  }, [chainFilter, assetFilter, statusFilter])

  const fetchAddresses = async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams()
      if (chainFilter) params.append('chain', chainFilter)
      if (assetFilter) params.append('asset', assetFilter)
      if (statusFilter) params.append('status', statusFilter)
      params.append('limit', '100')
      params.append('offset', '0')

      const response = await api.get(`/v1/address-pool?${params.toString()}`)
      setAddresses(response.data.addresses || [])
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取地址列表失败')
    } finally {
      setLoading(false)
    }
  }

  const fetchStats = async () => {
    try {
      const params = new URLSearchParams()
      if (chainFilter) params.append('chain', chainFilter)
      if (assetFilter) params.append('asset', assetFilter)

      const response = await api.get(`/v1/address-pool/stats?${params.toString()}`)
      setStats(response.data)
    } catch (error: any) {
      console.error('Failed to fetch stats:', error)
    }
  }

  const handleGenerate = () => {
    form.resetFields()
    setModalVisible(true)
  }

  const handleSubmit = async (values: any) => {
    try {
      await api.post('/v1/address-pool/generate', {
        chain: values.chain,
        asset_symbol: values.asset_symbol,
        count: values.count,
      })
      message.success('地址生成成功')
      setModalVisible(false)
      fetchAddresses()
      fetchStats()
    } catch (error: any) {
      message.error(error.response?.data?.error || '生成失败')
    }
  }

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      ellipsis: true,
    },
    {
      title: '链',
      dataIndex: 'chain',
      key: 'chain',
      width: 100,
    },
    {
      title: '资产',
      dataIndex: 'asset_symbol',
      key: 'asset_symbol',
      width: 120,
    },
    {
      title: '地址',
      dataIndex: 'address',
      key: 'address',
      ellipsis: true,
      render: (text: string) => (
        <span style={{ fontFamily: 'monospace' }}>{text}</span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={status === 'available' ? 'green' : 'default'}>
          {status === 'available' ? '可用' : '已使用'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '使用时间',
      dataIndex: 'used_at',
      key: 'used_at',
      width: 180,
      render: (text: string) => text ? dayjs(text).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
  ]

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>地址池管理</h1>

      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card>
            <Statistic
              title="可用地址"
              value={stats.available}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="已使用地址"
              value={stats.used}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic title="总地址数" value={stats.total} />
          </Card>
        </Col>
      </Row>

      <Card>
        <Space style={{ marginBottom: 16 }} wrap>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleGenerate}
          >
            批量生成地址
          </Button>
          <Button icon={<ReloadOutlined />} onClick={fetchAddresses}>
            刷新
          </Button>
          <Select
            style={{ width: 120 }}
            placeholder="筛选链"
            allowClear
            value={chainFilter}
            onChange={(value) => setChainFilter(value || '')}
          >
            <Select.Option value="evm">EVM</Select.Option>
            <Select.Option value="bitcoin">Bitcoin</Select.Option>
            <Select.Option value="solana">Solana</Select.Option>
          </Select>
          <Input
            style={{ width: 150 }}
            placeholder="筛选资产"
            allowClear
            value={assetFilter}
            onChange={(e) => setAssetFilter(e.target.value)}
          />
          <Select
            style={{ width: 120 }}
            placeholder="筛选状态"
            allowClear
            value={statusFilter}
            onChange={(value) => setStatusFilter(value || '')}
          >
            <Select.Option value="available">可用</Select.Option>
            <Select.Option value="used">已使用</Select.Option>
          </Select>
        </Space>

        <Table
          columns={columns}
          dataSource={addresses}
          loading={loading}
          rowKey="id"
          scroll={{ x: 1200 }}
        />
      </Card>

      <Modal
        title="批量生成地址"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item
            label="链"
            name="chain"
            rules={[{ required: true, message: '请选择链' }]}
          >
            <Select placeholder="选择链">
              <Select.Option value="evm">EVM</Select.Option>
              <Select.Option value="bitcoin">Bitcoin</Select.Option>
              <Select.Option value="solana">Solana</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            label="资产符号"
            name="asset_symbol"
            rules={[{ required: true, message: '请输入资产符号' }]}
          >
            <Input placeholder="如：ETH" />
          </Form.Item>
          <Form.Item
            label="生成数量"
            name="count"
            rules={[
              { required: true, message: '请输入生成数量' },
              { type: 'number', min: 1, max: 1000, message: '数量必须在1-1000之间' },
            ]}
          >
            <InputNumber
              min={1}
              max={1000}
              style={{ width: '100%' }}
              placeholder="1-1000"
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default AddressPool

