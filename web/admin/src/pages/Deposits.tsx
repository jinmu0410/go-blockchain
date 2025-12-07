import { useEffect, useState } from 'react'
import {
  Table,
  Card,
  Input,
  Button,
  Space,
  Tag,
  Modal,
  message,
  Select,
  DatePicker,
} from 'antd'
import { SearchOutlined, CheckCircleOutlined } from '@ant-design/icons'
import api from '../utils/api'
import dayjs from 'dayjs'

interface Deposit {
  tx_hash: string
  user_id: string
  chain: string
  asset_symbol: string
  amount: string
  from_address: string
  to_address: string
  block_height: number
  confirmations: number
  required_confirmations: number
  status: string
  observed_at: string
  credited_at?: string
}

const Deposits = () => {
  const [loading, setLoading] = useState(false)
  const [deposits, setDeposits] = useState<Deposit[]>([])
  const [searchParams, setSearchParams] = useState({
    user_id: '',
    asset: '',
    status: '',
  })

  useEffect(() => {
    fetchDeposits()
  }, [])

  const fetchDeposits = async () => {
    setLoading(true)
    try {
      const params: any = {}
      if (searchParams.user_id) params.user_id = searchParams.user_id
      if (searchParams.asset) params.asset = searchParams.asset
      if (searchParams.status) params.status = searchParams.status

      const response = await api.get('/v1/transactions/deposits', { params })
      setDeposits(response.data.deposits || [])
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取充值记录失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCredit = (txHash: string) => {
    Modal.confirm({
      title: '确认入账',
      content: `确定要手动确认交易 ${txHash} 的充值吗？`,
      onOk: async () => {
        try {
          await api.post(`/v1/transactions/deposits/${txHash}/credit`)
          message.success('充值已确认')
          fetchDeposits()
        } catch (error: any) {
          message.error(error.response?.data?.error || '确认失败')
        }
      },
    })
  }

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      pending: { color: 'orange', text: '待确认' },
      confirmed: { color: 'blue', text: '已确认' },
      credited: { color: 'green', text: '已入账' },
      failed: { color: 'red', text: '失败' },
    }
    const config = statusMap[status] || { color: 'default', text: status }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  const columns = [
    {
      title: '交易哈希',
      dataIndex: 'tx_hash',
      key: 'tx_hash',
      ellipsis: true,
      render: (text: string) => (
        <span style={{ fontFamily: 'monospace' }}>{text.slice(0, 16)}...</span>
      ),
    },
    {
      title: '用户ID',
      dataIndex: 'user_id',
      key: 'user_id',
    },
    {
      title: '链',
      dataIndex: 'chain',
      key: 'chain',
    },
    {
      title: '资产',
      dataIndex: 'asset_symbol',
      key: 'asset_symbol',
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      render: (amount: string) => amount,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => getStatusTag(status),
    },
    {
      title: '确认数',
      key: 'confirmations',
      render: (record: Deposit) => `${record.confirmations}/${record.required_confirmations}`,
    },
    {
      title: '观察时间',
      dataIndex: 'observed_at',
      key: 'observed_at',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Deposit) => (
        <Space>
          {record.status !== 'credited' && (
            <Button
              type="link"
              icon={<CheckCircleOutlined />}
              onClick={() => handleCredit(record.tx_hash)}
            >
              确认入账
            </Button>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>充值管理</h1>
      
      <Card>
        <Space style={{ marginBottom: 16 }} wrap>
          <Input
            placeholder="用户ID"
            value={searchParams.user_id}
            onChange={(e) =>
              setSearchParams({ ...searchParams, user_id: e.target.value })
            }
            style={{ width: 200 }}
          />
          <Input
            placeholder="资产符号"
            value={searchParams.asset}
            onChange={(e) =>
              setSearchParams({ ...searchParams, asset: e.target.value })
            }
            style={{ width: 200 }}
          />
          <Select
            placeholder="状态"
            value={searchParams.status}
            onChange={(value) =>
              setSearchParams({ ...searchParams, status: value })
            }
            style={{ width: 150 }}
            allowClear
          >
            <Select.Option value="pending">待确认</Select.Option>
            <Select.Option value="confirmed">已确认</Select.Option>
            <Select.Option value="credited">已入账</Select.Option>
            <Select.Option value="failed">失败</Select.Option>
          </Select>
          <Button
            type="primary"
            icon={<SearchOutlined />}
            onClick={fetchDeposits}
          >
            搜索
          </Button>
        </Space>

        <Table
          columns={columns}
          dataSource={deposits}
          loading={loading}
          rowKey="tx_hash"
          scroll={{ x: 1200 }}
        />
      </Card>
    </div>
  )
}

export default Deposits

