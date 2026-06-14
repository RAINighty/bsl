import { useEffect, useState } from 'react';
import { Table, Button, Modal, Input, Switch, Space, Popconfirm, message } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { api } from '../api';

export default function Rooms() {
  const [rooms, setRooms] = useState([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [newRoom, setNewRoom] = useState('');
  const navigate = useNavigate();

  const load = () => api.listRooms().then(setRooms).finally(() => setLoading(false));
  useEffect(() => { load(); }, []);

  const addRoom = async () => {
    const roomID = parseInt(newRoom, 10);
    if (!roomID) return message.error('请输入有效的房间号');
    try {
      await api.addRoom({ room_id: roomID, uid: 0, name: '' });
      message.success('添加成功');
      setModalOpen(false);
      setNewRoom('');
      load();
    } catch (e) { message.error('添加失败: ' + e.message); }
  };

  const deleteRoom = async (id) => {
    try {
      await api.deleteRoom(id);
      message.success('删除成功');
      load();
    } catch (e) { message.error('删除失败: ' + e.message); }
  };

  const toggleListening = async (id, checked) => {
    try {
      await api.setListening(id, checked);
      load();
    } catch (e) { message.error('操作失败'); }
  };

  const columns = [
    { title: '房间号', dataIndex: 'room_id', key: 'room_id' },
    { title: '直播间名称', dataIndex: 'name', key: 'name', render: (v) => v || '（待获取）' },
    {
      title: '监听',
      dataIndex: 'is_listening',
      key: 'is_listening',
      render: (v, record) => <Switch checked={v} onChange={(c) => toggleListening(record.room_id, c)} />,
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Button type="link" onClick={() => navigate(`/rooms/${record.room_id}`)}>详情</Button>
          <Popconfirm title="确定删除？" onConfirm={() => deleteRoom(record.room_id)}>
            <Button type="link" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2>房间管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>添加房间</Button>
      </div>
      <Table dataSource={rooms} columns={columns} rowKey="room_id" loading={loading} pagination={false} />
      <Modal title="添加直播间" open={modalOpen} onOk={addRoom} onCancel={() => setModalOpen(false)}>
        <Input placeholder="B站直播间房间号" value={newRoom} onChange={(e) => setNewRoom(e.target.value)} onPressEnter={addRoom} />
      </Modal>
    </>
  );
}
