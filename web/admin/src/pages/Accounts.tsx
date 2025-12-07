import { useState } from 'react'
import {
  Card,
  Input,
  Button,
  Table,
  Space,
  Modal,
  Form,
  InputNumber,
  Select,
  Input as AntInput,
  message,
} from 'antd'
import { SearchOutlined, EditOutlined } from '@ant-design/icons'
import api from '../utils/api'

const Accounts = () => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [accounts, setAccounts] = useState<any[]>([])
  const [modalVisible, setModalVisible] = useState(false)
  const [selectedAccount, setSelectedAccount] = useState<any>(null)

  const handleSearch = async (userID: string, assetSymbol: string) => {
    if (!userID || !assetSymbol) {
      message.warning('请输入用户ID和资产符号')
      return
    }

    setLoading(true)
    try {
      const [accountRes, balanceRes] = await Promise.all([
        api.get(`/v1/accounts/${userID}/${assetSymbol}`),
        api.get(`/v1/accounts/${userID}/${assetSymbol}/balance`),
      ])

      setAccounts([
        {
          ...accountRes.data,
          balance: balanceRes.data,
        },
      ])
    } catch (error: any) {
      message.error(error.response?.data?.error || '查询失败')
    } finally {
      setLoading(false)
    }
  }

  const handleAdjust = (account: any) => {
    setSelectedAccount(account)
    setModalVisible(true)
    form.resetFields()
  }

  const handleSubmit = async (values: any) => {
    try {
      await api.post('/v1/accounts/balance/adjust', {
        user_id: selectedAccount.user_id,
        asset_symbol: selectedAccount.asset_symbol,
        type: values.type,
        amount: values.amount.toString(),
        reason: values.reason,
      })
      message.success('余额调整成功')
      setModalVisible(false)
      handleSearch(selectedAccount.user_id, selectedAccount.asset_symbol)
    } catch (error: any) {
      message.error(error.response?.data?.error || '调整失败')
    }
  }

  const columns = [
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
      title: '地址',
      dataIndex: 'address',
      key: 'address',
      ellipsis: true,
    },
    {
      title: '可用余额',
      key: 'available',
      render: (_: any, record: any) =>
        record.balance?.available || '0',
    },
    {
      title: '冻结余额',
      key: 'frozen',
      render: (_: any, record: any) => record.balance?.frozen || '0',
    },
    {
      title: '待处理',
      key: 'pending',
      render: (_: any, record: any) => record.balance?.pending || '0',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: any) => (
        <Button
          type="link"
          icon={<EditOutlined />}
          onClick={() => handleAdjust(record)}
        >
          调整余额
        </Button>
      ),
    },
  ]

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>账号管理</h1>

      <Card>
        <Space style={{ marginBottom: 16 }}>
          <Input
            placeholder="用户ID"
            id="user-id-input"
            style={{ width: 200 }}
          />
          <Input
            placeholder="资产符号"
            id="asset-input"
            style={{ width: 200 }}
          />
          <Button
            type="primary"
            icon={<SearchOutlined />}
            onClick={() => {
              const userID = (
                document.getElementById('user-id-input') as HTMLInputElement
              ).value
              const assetSymbol = (
                document.getElementById('asset-input') as HTMLInputElement
              ).value
              handleSearch(userID, assetSymbol)
            }}
          >
            搜索
          </Button>
        </Space>

        <Table
          columns={columns}
          dataSource={accounts}
          loading={loading}
          rowKey={(record) => `${record.user_id}-${record.asset_symbol}`}
        />
      </Card>

      <Modal
        title="调整余额"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item label="操作类型" name="type" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="credit">增加</Select.Option>
              <Select.Option value="debit">减少</Select.Option>
              <Select.Option value="freeze">冻结</Select.Option>
              <Select.Option value="unfreeze">解冻</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            label="金额"
            name="amount"
            rules={[{ required: true }]}
          >
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label="原因" name="reason">
            <AntInput.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Accounts

