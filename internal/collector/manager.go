package collector

import (
	"log"
	"sync"
	"time"

	"bsl/internal/model"
)

type Callbacks struct {
	OnDanmaku    func(model.DanmakuRecord)
	OnGift       func(model.GiftRecord)
	OnSuperChat  func(model.SuperChat)
	OnGuard      func(model.GuardBuy)
	OnStreetlight func(model.StreetlightEvent)
	OnStats      func(model.DanmakuStat)
	IsBlacklisted func(uid int64) bool
}

type Manager struct {
	rooms     map[int64]*RoomSession
	callbacks Callbacks
	mu        sync.RWMutex
	cookie    string
}

type RoomSession struct {
	Client     *BiliClient
	Stats      *StatsAccumulator
	Room       model.Room
	stopCh     chan struct{}
}

func NewManager(callbacks Callbacks, cookie string) *Manager {
	return &Manager{
		rooms:     make(map[int64]*RoomSession),
		callbacks: callbacks,
		cookie:    cookie,
	}
}

func (m *Manager) AddRoom(room model.Room) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.rooms[room.RoomID]; exists {
		return
	}
	session := &RoomSession{Room: room, stopCh: make(chan struct{})}
	session.Client = NewBiliClient(room.RoomID, m.cookie)
	session.Stats = NewStatsAccumulator(room.RoomID, func(stat DanmakuStat) {
		if m.callbacks.OnStats != nil {
			m.callbacks.OnStats(model.DanmakuStat{
				RoomID:      stat.RoomID,
				Minute:      stat.Minute,
				Count:       stat.Count,
				ViewerCount: stat.ViewerCount,
			})
		}
	})
	session.Client.OnMessage = m.makeHandler(session)
	m.rooms[room.RoomID] = session
	go m.runRoom(session)
}

func (m *Manager) RemoveRoom(roomID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.rooms[roomID]; ok {
		s.Client.Stop()
		delete(m.rooms, roomID)
	}
}

func (m *Manager) UpdateRooms(rooms []model.Room) {
	m.mu.Lock()
	current := make(map[int64]bool)
	for _, r := range rooms {
		current[r.RoomID] = true
		if _, exists := m.rooms[r.RoomID]; !exists {
			session := &RoomSession{Room: r, stopCh: make(chan struct{})}
			session.Client = NewBiliClient(r.RoomID, m.cookie)
			session.Stats = NewStatsAccumulator(r.RoomID, func(stat DanmakuStat) {
				if m.callbacks.OnStats != nil {
					m.callbacks.OnStats(model.DanmakuStat{
						RoomID:      stat.RoomID,
						Minute:      stat.Minute,
						Count:       stat.Count,
						ViewerCount: stat.ViewerCount,
					})
				}
			})
			session.Client.OnMessage = m.makeHandler(session)
			m.rooms[r.RoomID] = session
			go m.runRoom(session)
		}
	}
	for id := range m.rooms {
		if !current[id] {
			m.rooms[id].Client.Stop()
			delete(m.rooms, id)
		}
	}
	m.mu.Unlock()
}

func (m *Manager) runRoom(session *RoomSession) {
	backoff := time.Second
	maxBackoff := time.Duration(60) * time.Second
	for {
		if err := session.Client.Connect(); err != nil {
			log.Printf("[manager:%d] connect error: %v (retry in %v)", session.Room.RoomID, err, backoff)
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}
		backoff = time.Second
		session.Client.Start()
		<-session.stopCh
		return
	}
}

func (m *Manager) makeHandler(session *RoomSession) func(uint32, []byte) {
	return func(op uint32, body []byte) {
		if op != OpServerMessage || len(body) == 0 {
			return
		}
		for _, raw := range SplitMessages(body) {
			msg := ParseMessage(raw)
			switch msg.Type {
			case MsgDanmaku:
				m.handleDanmaku(session, msg)
			case MsgGift:
				m.handleGift(session, msg)
			case MsgSuperChat:
				m.handleSC(session, msg)
			case MsgGuard:
				m.handleGuard(session, msg)
			case MsgViewerCount:
				session.Stats.SetViewers(msg.ViewerCount)
			}
		}
	}
}

func (m *Manager) handleDanmaku(session *RoomSession, msg ParsedMessage) {
	now := time.Now()

	if m.callbacks.IsBlacklisted != nil && m.callbacks.IsBlacklisted(msg.UID) {
		return
	}

	isSL, note := DetectStreetlight(msg.Content)

	record := model.DanmakuRecord{
		RoomID:          session.Room.RoomID,
		Username:        msg.Username,
		UID:             msg.UID,
		Content:         msg.Content,
		IsStreetlight:   isSL,
		StreetlightNote: note,
		CreatedAt:       now,
	}

	if m.callbacks.OnDanmaku != nil {
		m.callbacks.OnDanmaku(record)
	}

	session.Stats.Add()

	if isSL && m.callbacks.OnStreetlight != nil {
		m.callbacks.OnStreetlight(model.StreetlightEvent{
			RoomID:   session.Room.RoomID,
			RoomName: session.Room.Name,
			Username: msg.Username,
			UID:      msg.UID,
			Content:  note,
			Timestamp: now,
		})
	}
}

func (m *Manager) handleGift(session *RoomSession, msg ParsedMessage) {
	if m.callbacks.IsBlacklisted != nil && m.callbacks.IsBlacklisted(msg.UID) {
		return
	}
	if m.callbacks.OnGift != nil {
		m.callbacks.OnGift(model.GiftRecord{
			RoomID:   session.Room.RoomID,
			Username: msg.Username,
			GiftName: msg.GiftName,
			Count:    msg.GiftCount,
			Price:    msg.GiftPrice,
			PaidAt:   time.Now(),
		})
	}
}

func (m *Manager) handleSC(session *RoomSession, msg ParsedMessage) {
	if m.callbacks.IsBlacklisted != nil && m.callbacks.IsBlacklisted(msg.UID) {
		return
	}
	if m.callbacks.OnSuperChat != nil {
		m.callbacks.OnSuperChat(model.SuperChat{
			RoomID:   session.Room.RoomID,
			Username: msg.Username,
			Message:  msg.SCMessage,
			Price:    msg.SCPrice,
			PaidAt:   time.Now(),
		})
	}
}

func (m *Manager) handleGuard(session *RoomSession, msg ParsedMessage) {
	if m.callbacks.IsBlacklisted != nil && m.callbacks.IsBlacklisted(msg.UID) {
		return
	}
	if m.callbacks.OnGuard != nil {
		m.callbacks.OnGuard(model.GuardBuy{
			RoomID:     session.Room.RoomID,
			Username:   msg.Username,
			GuardLevel: msg.GuardLevel,
			Count:      msg.GuardCount,
			PaidAt:     time.Now(),
		})
	}
}
