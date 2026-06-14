import { useEffect, useState, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { Tabs, Table, Tag, Spin, Card, Statistic } from 'antd';
import { BulbOutlined } from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import { api } from '../api';

export default function RoomDetail() {
  const { id } = useParams();
  const roomID = parseInt(id, 10);
  const [room, setRoom] = useState(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('danmaku');

  useEffect(() => {
    api.roomDetail(roomID).then((d) => { setRoom(d.room); setLoading(false); });
  }, [roomID]);

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '40px auto' }} />;
  if (!room) return <div>房间未找到</div>;

  return (
    <>
      <div style={{ marginBottom: 16 }}>
        <h2>{room.name || `直播间 ${room.room_id}`}</h2>
        <Statistic title="今日路灯" value={room.today_streetlights ?? 0} prefix={<BulbOutlined />} valueStyle={{ color: '#cf1322' }} />
      </div>
      <Tabs activeKey={activeTab} onChange={setActiveTab} items={[
        { key: 'danmaku', label: '弹幕', children: <DataTable fetcher={(l,o) => api.danmakuList(roomID, l, o)} columns={danmakuColumns} /> },
        { key: 'streetlights', label: '路灯', children: <DataTable fetcher={(l,o) => api.streetlightList(roomID, l, o)} columns={streetlightColumns} /> },
        { key: 'gifts', label: '礼物', children: <DataTable fetcher={(l,o) => api.giftList(roomID, l, o)} columns={giftColumns} /> },
        { key: 'sc', label: 'SC', children: <DataTable fetcher={(l,o) => api.scList(roomID, l, o)} columns={scColumns} /> },
        { key: 'guards', label: '舰长', children: <DataTable fetcher={(l,o) => api.guardList(roomID, l, o)} columns={guardColumns} /> },
        { key: 'stats', label: '统计图表', children: <StatsChart roomID={roomID} /> },
      ]} />
    </>
  );
}

function DataTable({ fetcher, columns }) {
  const [data, setData] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const pageSize = 50;

  const load = useCallback((p) => {
    setLoading(true);
    const offset = (p - 1) * pageSize;
    fetcher(pageSize, offset).then((d) => {
      setData(d.items || []);
      setTotal(d.total || 0);
      setLoading(false);
    });
  }, [fetcher]);

  useEffect(() => { load(page); }, [page, load]);

  return (
    <Table
      dataSource={data}
      columns={columns}
      rowKey="id"
      loading={loading}
      size="small"
      pagination={{ current: page, total, pageSize, onChange: setPage, showTotal: (t) => `共 ${t} 条` }}
    />
  );
}

const danmakuColumns = [
  { title: '时间', dataIndex: 'created_at', key: 'created_at', render: (v) => new Date(v).toLocaleString(), width: 180 },
  { title: '用户', dataIndex: 'username', key: 'username', width: 120 },
  { title: '内容', dataIndex: 'content', key: 'content', ellipsis: true },
  {
    title: '路灯', dataIndex: 'is_streetlight', key: 'is_streetlight', width: 80,
    render: (v) => v ? <Tag color="orange">路灯</Tag> : null,
  },
];

const streetlightColumns = [
  { title: '时间', dataIndex: 'created_at', key: 'created_at', render: (v) => new Date(v).toLocaleString(), width: 180 },
  { title: '用户', dataIndex: 'username', key: 'username', width: 120 },
  { title: '路灯内容', dataIndex: 'streetlight_note', key: 'streetlight_note', ellipsis: true },
];

const giftColumns = [
  { title: '时间', dataIndex: 'paid_at', key: 'paid_at', render: (v) => new Date(v).toLocaleString(), width: 180 },
  { title: '用户', dataIndex: 'username', key: 'username', width: 120 },
  { title: '礼物', dataIndex: 'gift_name', key: 'gift_name' },
  { title: '数量', dataIndex: 'count', key: 'count' },
  { title: '价格', dataIndex: 'price', key: 'price', render: (v) => `¥${v}` },
];

const scColumns = [
  { title: '时间', dataIndex: 'paid_at', key: 'paid_at', render: (v) => new Date(v).toLocaleString(), width: 180 },
  { title: '用户', dataIndex: 'username', key: 'username', width: 120 },
  { title: '内容', dataIndex: 'message', key: 'message', ellipsis: true },
  { title: '金额', dataIndex: 'price', key: 'price', render: (v) => `¥${v}`, width: 80 },
];

const guardColumns = [
  { title: '时间', dataIndex: 'paid_at', key: 'paid_at', render: (v) => new Date(v).toLocaleString(), width: 180 },
  { title: '用户', dataIndex: 'username', key: 'username', width: 120 },
  { title: '等级', dataIndex: 'guard_level', key: 'guard_level' },
  { title: '数量', dataIndex: 'count', key: 'count' },
];

function StatsChart({ roomID }) {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const end = new Date().toISOString();
    const start = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString();
    api.danmakuStats(roomID, start, end).then(setData).finally(() => setLoading(false));
  }, [roomID]);

  if (loading) return <Spin />;

  const option = {
    tooltip: { trigger: 'axis' },
    legend: { data: ['弹幕数', '观众数'] },
    xAxis: { type: 'time', name: '时间' },
    yAxis: [
      { type: 'value', name: '弹幕数' },
      { type: 'value', name: '观众数' },
    ],
    series: [
      {
        name: '弹幕数', type: 'line', smooth: true,
        data: data.map((d) => [d.minute, d.count]),
      },
      {
        name: '观众数', type: 'line', smooth: true, yAxisIndex: 1,
        data: data.map((d) => [d.minute, d.viewer_count]),
      },
    ],
  };

  return (
    <Card>
      <ReactECharts option={option} style={{ height: 400 }} />
    </Card>
  );
}
