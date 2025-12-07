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
} from 'antd'
import {
  SearchOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'
import api from '../utils/api'
import dayjs from 'dayjs'

interface Withdrawal {
  id: string
  user_id: string
  asset_symbol: string
  chain: string
  to_address: string
  amount: string
  fee: string
  status: string
  risk_score: number
  risk_remarks: string
  created_at: string
  updated_at: string
}

const Withdrawals = () => {
  const [loading, setLoading] = useState(false)
  const [withdrawals, setWithdrawals] = useState<Withdrawal[]>([])
  const [searchParams, setSearchParams] = useState({
    status: '',
  })

  useEffect(() => {
    fetchWithdrawals()
  }, [])

  const fetchWithdrawals = async () => {
    setLoading(true)
    try {
      const params: any = {}
      if (searchParams.status) params.status = searchParams.status

      const response = await api.get('/v1/transactions/withdrawals', { params })
      setWithdrawals(response.data.withdrawals || [])
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取提现记录失败')
    } finally {
      setLoading(false)
    }
  }

  const handleApprove = (id: string) => {
    Modal.confirm({
      title: '审批通过',
      content: `确定要审批通过提现请求 ${id} 吗？`,
      onOk: async () => {
        try {
          await api.post(`/v1/transactions/withdrawals/${id}/approve`)
          message.success('提现已审批通过')
          fetchWithdrawals()
        } catch (error: any) {
          message.error(error.response?.data?.error || '审批失败')
        }
      },
    })
  }

  const handleReject = (id: string) => {
    Modal.confirm({
      title: '拒绝提现',
      content: '请输入拒绝原因',
      onOk: async () => {
        try {
          await api.post(`/v1/transactions/withdrawals/${id}/reject`, {
            reason: '管理员拒绝',
          })
          message.success('提现已拒绝')
          fetchWithdrawals()
        } catch (error: any) {
          message.error(error.response?.data?.error || '拒绝失败')
        }
      },
    })
  }

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      requested: { color: 'orange', text: '已申请' },
      under_review: { color: 'blue', text: '审核中' },
      approved: { color: 'green', text: '已审批' },
      rejected: { color: 'red', text: '已拒绝' },
      signed: { color: 'cyan', text: '已签名' },
      broadcast: { color: 'purple', text: '已广播' },
      completed: { color: 'green', text: '已完成' },
      failed: { color: 'red', text: '失败' },
    }
    const config = statusMap[status] || { color: 'default', text: status }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
    },
    {
      title: '用户ID',
      dataIndex: 'user_id',
      key: 'user_id',
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
    },
    {
      title: '手续费',
      dataIndex: 'fee',
      key: 'fee',
    },
    {
      title: '接收地址',
      dataIndex: 'to_address',
      key: 'to_address',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => getStatusTag(status),
    },
    {
      title: '风控评分',
      dataIndex: 'risk_score',
      key: 'risk_score',
      render: (score: number) => (
        <Tag color={score > 70 ? 'red' : score > 40 ? 'orange' : 'green'}>
          {score.toFixed(2)}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Withdrawal) => (
        <Space>
          {record.status === 'under_review' && (
            <>
              <Button
                type="link"
                icon={<CheckCircleOutlined />}
                onClick={() => handleApprove(record.id)}
              >
                通过
              </Button>
              <Button
                type="link"
                danger
                icon={<CloseCircleOutlined />}
                onClick={() => handleReject(record.id)}
              >
                拒绝
              </Button>
            </>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>提现管理</h1>
      
      <Card>
        <Space style={{ marginBottom: 16 }} wrap>
          <Select
            placeholder="状态"
            value={searchParams.status}
            onChange={(value) =>
              setSearchParams({ ...searchParams, status: value })
            }
            style={{ width: 200 }}
            allowClear
          >
            <Select.Option value="under_review">审核中</Select.Option>
            <Select.Option value="approved">已审批</Select.Option>
            <Select.Option value="rejected">已拒绝</Select.Option>
            <Select.Option value="completed">已完成</Select.Option>
          </Select>
          <Button
            type="primary"
            icon={<SearchOutlined />}
            onClick={fetchWithdrawals}
          >
            搜索
          </Button>
        </Space>

        <Table
          columns={columns}
          dataSource={withdrawals}
          loading={loading}
          rowKey="id"
          scroll={{ x: 1400 }}
        />
      </Card>
    </div>
  )
}

export default Withdrawals

