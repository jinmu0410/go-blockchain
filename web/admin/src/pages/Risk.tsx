import { useEffect, useState } from 'react'
import {
  Tabs,
  Card,
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  InputNumber,
  message,
  Tag,
  Alert,
} from 'antd'
import {
  PlusOutlined,
  DeleteOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import api from '../utils/api'

const Risk = () => {
  const [whitelist, setWhitelist] = useState<any[]>([])
  const [blacklist, setBlacklist] = useState<any[]>([])
  const [config, setConfig] = useState<any>({})
  const [loading, setLoading] = useState(false)
  const [addModalVisible, setAddModalVisible] = useState(false)
  const [configModalVisible, setConfigModalVisible] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [form] = Form.useForm()
  const [configForm] = Form.useForm()
  const [activeTab, setActiveTab] = useState('whitelist')

  useEffect(() => {
    fetchWhitelist()
    fetchBlacklist()
    fetchConfig()
  }, [])

  const fetchWhitelist = async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await api.get('/v1/risk/whitelist')
      // 后端返回格式：{chain, count, entries}
      setWhitelist(response.data.entries || [])
    } catch (error: any) {
      console.error('Failed to fetch whitelist:', error)
      const errorMsg = error.response?.data?.error || '获取白名单失败'
      setError(errorMsg)
      if (error.response?.status === 501) {
        message.warning('风控功能未启用，当前使用内存存储模式')
      } else {
        message.error(errorMsg)
      }
    } finally {
      setLoading(false)
    }
  }

  const fetchBlacklist = async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await api.get('/v1/risk/blacklist')
      // 后端返回格式：{chain, count, entries}
      setBlacklist(response.data.entries || [])
    } catch (error: any) {
      console.error('Failed to fetch blacklist:', error)
      const errorMsg = error.response?.data?.error || '获取黑名单失败'
      setError(errorMsg)
      if (error.response?.status === 501) {
        message.warning('风控功能未启用，当前使用内存存储模式')
      } else {
        message.error(errorMsg)
      }
    } finally {
      setLoading(false)
    }
  }

  const fetchConfig = async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await api.get('/v1/risk/config')
      setConfig(response.data)
      configForm.setFieldsValue(response.data)
    } catch (error: any) {
      console.error('Failed to fetch config:', error)
      const errorMsg = error.response?.data?.error || '获取配置失败'
      setError(errorMsg)
      if (error.response?.status === 501) {
        message.warning('风控功能未启用，当前使用内存存储模式')
      } else {
        message.error(errorMsg)
      }
    } finally {
      setLoading(false)
    }
  }

  const handleAdd = async (type: 'whitelist' | 'blacklist') => {
    try {
      const values = await form.validateFields()
      await api.post(`/v1/risk/${type}`, {
        address: values.address,
        chain: values.chain,
        remark: values.remark || '',
      })
      message.success('添加成功')
      setAddModalVisible(false)
      form.resetFields()
      if (type === 'whitelist') {
        fetchWhitelist()
      } else {
        fetchBlacklist()
      }
    } catch (error: any) {
      const errorMsg = error.response?.data?.error || '添加失败'
      if (error.response?.status === 501) {
        message.warning('风控功能未启用，当前使用内存存储模式')
      } else {
        message.error(errorMsg)
      }
    }
  }

  const handleRemove = async (type: 'whitelist' | 'blacklist', record: any) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要从${type === 'whitelist' ? '白名单' : '黑名单'}中删除 ${record.address} 吗？`,
      onOk: async () => {
        try {
          // 删除时需要传递chain参数
          await api.delete(`/v1/risk/${type}/${encodeURIComponent(record.address)}`, {
            params: { chain: record.chain || 'evm' }
          })
          message.success('删除成功')
          if (type === 'whitelist') {
            fetchWhitelist()
          } else {
            fetchBlacklist()
          }
        } catch (error: any) {
          message.error(error.response?.data?.error || '删除失败')
        }
      },
    })
  }

  const handleConfigUpdate = async () => {
    try {
      const values = await configForm.validateFields()
      await api.put('/v1/risk/config', values)
      message.success('配置更新成功')
      setConfigModalVisible(false)
      fetchConfig()
    } catch (error: any) {
      message.error(error.response?.data?.error || '更新失败')
    }
  }

  const whitelistColumns = [
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
      title: '链',
      dataIndex: 'chain',
      key: 'chain',
    },
    {
      title: '备注',
      dataIndex: 'remark',
      key: 'remark',
      ellipsis: true,
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: any) => (
        <Button
          danger
          icon={<DeleteOutlined />}
          onClick={() => handleRemove('whitelist', record)}
        >
          删除
        </Button>
      ),
    },
  ]

  const blacklistColumns = [
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
      title: '链',
      dataIndex: 'chain',
      key: 'chain',
    },
    {
      title: '备注',
      dataIndex: 'remark',
      key: 'remark',
      ellipsis: true,
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: any) => (
        <Button
          danger
          icon={<DeleteOutlined />}
          onClick={() => handleRemove('blacklist', record)}
        >
          删除
        </Button>
      ),
    },
  ]

  const tabItems = [
    {
      key: 'whitelist',
      label: '白名单',
      children: (
        <Card>
          <Space style={{ marginBottom: 16 }}>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setActiveTab('whitelist')
                setAddModalVisible(true)
              }}
            >
              添加白名单
            </Button>
          </Space>
          <Table
            columns={whitelistColumns}
            dataSource={whitelist}
            loading={loading}
            rowKey={(record) => `${record.address}-${record.chain}`}
          />
        </Card>
      ),
    },
    {
      key: 'blacklist',
      label: '黑名单',
      children: (
        <Card>
          <Space style={{ marginBottom: 16 }}>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setActiveTab('blacklist')
                setAddModalVisible(true)
              }}
            >
              添加黑名单
            </Button>
          </Space>
          <Table
            columns={blacklistColumns}
            dataSource={blacklist}
            loading={loading}
            rowKey={(record) => `${record.address}-${record.chain}`}
          />
        </Card>
      ),
    },
    {
      key: 'config',
      label: '风控配置',
      children: (
        <Card>
          <Space style={{ marginBottom: 16 }}>
            <Button
              type="primary"
              icon={<SettingOutlined />}
              onClick={() => setConfigModalVisible(true)}
            >
              编辑配置
            </Button>
          </Space>
          <Card>
            <p>
              <strong>自动通过阈值：</strong>
              {config.auto_approve_threshold || 'N/A'}
            </p>
            <p>
              <strong>人工审核阈值：</strong>
              {config.manual_review_threshold || 'N/A'}
            </p>
            <p>
              <strong>拒绝阈值：</strong>
              {config.reject_threshold || 'N/A'}
            </p>
          </Card>
        </Card>
      ),
    },
  ]

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>风控管理</h1>

      {error && (
        <Alert
          message="提示"
          description={error}
          type="warning"
          showIcon
          closable
          onClose={() => setError(null)}
          style={{ marginBottom: 16 }}
        />
      )}

      <Tabs items={tabItems} activeKey={activeTab} onChange={setActiveTab} />

      <Modal
        title={activeTab === 'whitelist' ? '添加白名单' : '添加黑名单'}
        open={addModalVisible}
        onCancel={() => {
          setAddModalVisible(false)
          form.resetFields()
        }}
        onOk={() => handleAdd(activeTab as 'whitelist' | 'blacklist')}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="地址"
            name="address"
            rules={[{ required: true, message: '请输入地址' }]}
          >
            <Input placeholder="0x..." />
          </Form.Item>
          <Form.Item
            label="链"
            name="chain"
            rules={[{ required: true, message: '请选择链' }]}
          >
            <Input placeholder="evm" />
          </Form.Item>
          <Form.Item
            label="备注"
            name="remark"
          >
            <Input.TextArea rows={2} placeholder="可选备注信息" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="编辑风控配置"
        open={configModalVisible}
        onCancel={() => setConfigModalVisible(false)}
        onOk={() => configForm.submit()}
      >
        <Form form={configForm} onFinish={handleConfigUpdate} layout="vertical">
          <Form.Item label="自动通过阈值" name="auto_approve_threshold">
            <InputNumber style={{ width: '100%' }} min={0} max={100} />
          </Form.Item>
          <Form.Item label="人工审核阈值" name="manual_review_threshold">
            <InputNumber style={{ width: '100%' }} min={0} max={100} />
          </Form.Item>
          <Form.Item label="拒绝阈值" name="reject_threshold">
            <InputNumber style={{ width: '100%' }} min={0} max={100} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Risk

