import { useEffect, useState } from 'react';
import { Table, Button, Modal, Input, Popconfirm, Space, message } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { api } from '../api';

export default function Blacklist() {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [uid, setUid] = useState('');
  const [reason, setReason] = useState('');

  const load = () => api.listBlacklist().then(setData).finally(() => setLoading(false));
  useEffect(() => { load(); }, []);

  const add = async () => {
    const uidNum = parseInt(uid, 10);
    if (!uidNum) return message.error('请输入有效的UID');
    try {
      await api.addBlacklist(uidNum, reason);
      message.success('添加成功');
      setModalOpen(false);
      setUid('');
      setReason('');
      load();
    } catch (e) { message.error('添加失败: ' + e.message); }
  };

  const remove = async (uid) => {
    try {
      await api.removeBlacklist(uid);
      message.success('已移除');
      load();
    } catch (e) { message.error('移除失败'); }
  };

  const columns = [
    { title: 'UID', dataIndex: 'uid', key: 'uid' },
    { title: '原因', dataIndex: 'reason', key: 'reason', ellipsis: true },
    { title: '添加时间', dataIndex: 'created_at', key: 'created_at', render: (v) => new Date(v).toLocaleString() },
    {
      title: '操作', key: 'actions',
      render: (_, record) => (
        <Popconfirm title="确定移除？" onConfirm={() => remove(record.uid)}>
          <Button type="link" danger icon={<DeleteOutlined />}>移除</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2>黑名单管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>添加</Button>
      </div>
      <Table dataSource={data} columns={columns} rowKey="uid" loading={loading} pagination={false} />
      <Modal title="添加黑名单" open={modalOpen} onOk={add} onCancel={() => setModalOpen(false)}>
        <Input placeholder="B站 UID" value={uid} onChange={(e) => setUid(e.target.value)} style={{ marginBottom: 12 }} />
        <Input placeholder="原因（可选）" value={reason} onChange={(e) => setReason(e.target.value)} />
      </Modal>
    </>
  );
}
