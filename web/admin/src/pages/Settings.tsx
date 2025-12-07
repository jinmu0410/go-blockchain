import { Card, Form, Input, Button, Switch, message, Divider } from 'antd'
import { SaveOutlined } from '@ant-design/icons'
import { useState } from 'react'

const Settings = () => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (values: any) => {
    setLoading(true)
    try {
      // TODO: 实现设置保存API
      await new Promise((resolve) => setTimeout(resolve, 1000))
      message.success('设置保存成功')
    } catch (error) {
      message.error('保存失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>系统设置</h1>

      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            system_name: '区块链钱包管理系统',
            notification_enabled: true,
            auto_credit_enabled: true,
          }}
        >
          <h2>基本设置</h2>
          <Form.Item label="系统名称" name="system_name">
            <Input />
          </Form.Item>

          <Divider />

          <h2>功能设置</h2>
          <Form.Item
            label="启用通知"
            name="notification_enabled"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>
          <Form.Item
            label="自动入账"
            name="auto_credit_enabled"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>

          <Divider />

          <h2>安全设置</h2>
          <Form.Item label="JWT密钥">
            <Input.Password placeholder="修改JWT密钥" />
          </Form.Item>
          <Form.Item label="会话超时（分钟）">
            <Input type="number" defaultValue={1440} />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              icon={<SaveOutlined />}
              loading={loading}
            >
              保存设置
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}

export default Settings

