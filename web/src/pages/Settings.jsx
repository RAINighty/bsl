import { Card, Switch, Form, Input, Typography, Space, theme } from 'antd';
import { useState } from 'react';

const { Title, Text } = Typography;

export default function Settings() {
  const [darkMode, setDarkMode] = useState(false);
  const { token: themeToken } = theme.useToken();

  return (
    <>
      <Title level={3}>系统设置</Title>
      <Space direction="vertical" size="large" style={{ width: '100%', maxWidth: 600 }}>
        <Card title="外观">
          <Form.Item label="暗色模式">
            <Switch checked={darkMode} onChange={setDarkMode} />
            <Text type="secondary" style={{ marginLeft: 8 }}>（主题切换将在后续版本完善）</Text>
          </Form.Item>
        </Card>

        <Card title="B站配置">
          <Text type="secondary">B站 Cookie 配置通过 config.yaml 的 bilibili.cookie 字段设置。</Text>
          <br />
          <Text type="secondary">设置 Cookie 后可获取完整用户名信息。</Text>
        </Card>

        <Card title="OneBot 配置">
          <Form.Item label="WebSocket 路径">
            <Input disabled value="/onebot" />
          </Form.Item>
          <Text type="secondary">NapCat 反向 WebSocket 地址应配置为: ws://your-host:8080/onebot</Text>
        </Card>

        <Card title="系统信息">
          <Text>BSL (Bilibili StreetLight) v1.0.0</Text>
          <br />
          <Text type="secondary">Go + React 构建，PostgreSQL 数据存储</Text>
        </Card>
      </Space>
    </>
  );
}
