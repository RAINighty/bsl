import { useEffect, useState } from 'react';
import { Card, Col, Row, Statistic, Table, Tag, Spin } from 'antd';
import { AimOutlined, PlayCircleOutlined, BulbOutlined } from '@ant-design/icons';
import { api } from '../api';

export default function Dashboard() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.dashboard().then(setData).finally(() => setLoading(false));
    const timer = setInterval(() => api.dashboard().then(setData), 10000);
    return () => clearInterval(timer);
  }, []);

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '40px auto' }} />;

  const columns = [
    { title: '房间号', dataIndex: 'room_id', key: 'room_id' },
    { title: '直播间', dataIndex: 'name', key: 'name' },
    {
      title: '路灯数',
      dataIndex: 'streetlights',
      key: 'streetlights',
      render: (v) => <Tag color="orange">{v}</Tag>,
    },
  ];

  return (
    <>
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card><Statistic title="监听中直播间" value={data?.listening_rooms ?? 0} prefix={<AimOutlined />} /></Card>
        </Col>
        <Col span={8}>
          <Card><Statistic title="当前直播中" value={data?.live_rooms ?? 0} prefix={<PlayCircleOutlined />} valueStyle={{ color: '#3f8600' }} /></Card>
        </Col>
        <Col span={8}>
          <Card><Statistic title="今日路灯" value={data?.today_streetlights ?? 0} prefix={<BulbOutlined />} valueStyle={{ color: '#cf1322' }} /></Card>
        </Col>
      </Row>
      <Card title="正在直播">
        <Table
          dataSource={data?.live_room_list ?? []}
          columns={columns}
          rowKey="room_id"
          pagination={false}
          size="small"
        />
      </Card>
    </>
  );
}
