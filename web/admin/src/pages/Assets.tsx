import { useEffect, useState } from 'react'
import {
  Table,
  Card,
  Button,
  Space,
  Modal,
  Form,
  Input,
  InputNumber,
  Select,
  message,
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import api from '../utils/api'

interface Asset {
  symbol: string
  chain: string
  decimals: number
  token_addr: string
}

const Assets = () => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [assets, setAssets] = useState<Asset[]>([])
  const [modalVisible, setModalVisible] = useState(false)

  useEffect(() => {
    fetchAssets()
  }, [])

  const fetchAssets = async () => {
    setLoading(true)
    try {
      const response = await api.get('/v1/assets')
      setAssets(response.data.assets || [])
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取资产列表失败')
    } finally {
      setLoading(false)
    }
  }

  const handleAdd = () => {
    form.resetFields()
    setModalVisible(true)
  }

  const handleSubmit = async (values: any) => {
    try {
      await api.post('/v1/assets', {
        symbol: values.symbol,
        chain: values.chain,
        decimals: values.decimals,
        token_addr: values.token_addr || '',
      })
      message.success('资产添加成功')
      setModalVisible(false)
      fetchAssets()
    } catch (error: any) {
      message.error(error.response?.data?.error || '添加失败')
    }
  }

  const columns = [
    {
      title: '符号',
      dataIndex: 'symbol',
      key: 'symbol',
    },
    {
      title: '链',
      dataIndex: 'chain',
      key: 'chain',
    },
    {
      title: '精度',
      dataIndex: 'decimals',
      key: 'decimals',
    },
    {
      title: '代币地址',
      dataIndex: 'token_addr',
      key: 'token_addr',
      ellipsis: true,
    },
  ]

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>资产管理</h1>

      <Card>
        <Space style={{ marginBottom: 16 }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            添加资产
          </Button>
        </Space>

        <Table
          columns={columns}
          dataSource={assets}
          loading={loading}
          rowKey="symbol"
        />
      </Card>

      <Modal
        title="添加资产"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item
            label="符号"
            name="symbol"
            rules={[{ required: true }]}
          >
            <Input placeholder="如：ETH" />
          </Form.Item>
          <Form.Item
            label="链"
            name="chain"
            rules={[{ required: true }]}
          >
            <Select>
              <Select.Option value="evm">EVM</Select.Option>
              <Select.Option value="bitcoin">Bitcoin</Select.Option>
              <Select.Option value="solana">Solana</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            label="精度"
            name="decimals"
            rules={[{ required: true }]}
          >
            <InputNumber min={0} max={18} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="代币地址" name="token_addr">
            <Input placeholder="主币留空" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Assets

