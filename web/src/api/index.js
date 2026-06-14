const API_BASE = '/api';

export async function fetchJSON(url, options = {}) {
  const res = await fetch(API_BASE + url, {
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options,
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export const api = {
  dashboard: () => fetchJSON('/dashboard'),
  listRooms: () => fetchJSON('/rooms'),
  addRoom: (data) => fetchJSON('/rooms', { method: 'POST', body: JSON.stringify(data) }),
  deleteRoom: (id) => fetchJSON(`/rooms/${id}`, { method: 'DELETE' }),
  setListening: (id, listening) => fetchJSON(`/rooms/${id}/listening`, { method: 'PUT', body: JSON.stringify({ listening }) }),
  setStatsEnabled: (id, enabled) => fetchJSON(`/rooms/${id}/stats`, { method: 'PUT', body: JSON.stringify({ enabled }) }),
  roomDetail: (id) => fetchJSON(`/rooms/${id}`),
  danmakuList: (id, limit = 50, offset = 0) => fetchJSON(`/rooms/${id}/danmaku?limit=${limit}&offset=${offset}`),
  streetlightList: (id, limit = 50, offset = 0) => fetchJSON(`/rooms/${id}/streetlights?limit=${limit}&offset=${offset}`),
  giftList: (id, limit = 50, offset = 0) => fetchJSON(`/rooms/${id}/gifts?limit=${limit}&offset=${offset}`),
  scList: (id, limit = 50, offset = 0) => fetchJSON(`/rooms/${id}/sc?limit=${limit}&offset=${offset}`),
  guardList: (id, limit = 50, offset = 0) => fetchJSON(`/rooms/${id}/guards?limit=${limit}&offset=${offset}`),
  danmakuStats: (id, start, end) => fetchJSON(`/rooms/${id}/stats?start=${start}&end=${end}`),
  listBlacklist: () => fetchJSON('/blacklist'),
  addBlacklist: (uid, reason) => fetchJSON('/blacklist', { method: 'POST', body: JSON.stringify({ uid, reason }) }),
  removeBlacklist: (uid) => fetchJSON(`/blacklist/${uid}`, { method: 'DELETE' }),
};
