package model

import "time"

type Room struct {
	RoomID        int64      `json:"room_id"`
	UID           int64      `json:"uid"`
	Name          string     `json:"name"`
	IsListening   bool       `json:"is_listening"`
	StatsEnabled  bool       `json:"stats_enabled"`
	LastLiveStart *time.Time `json:"last_live_start"`
}

type DanmakuRecord struct {
	ID              int64     `json:"id"`
	RoomID          int64     `json:"room_id"`
	Username        string    `json:"username"`
	UID             int64     `json:"uid"`
	Content         string    `json:"content"`
	IsStreetlight   bool      `json:"is_streetlight"`
	StreetlightNote string    `json:"streetlight_note"`
	CreatedAt       time.Time `json:"created_at"`
}

type DanmakuStat struct {
	ID          int64     `json:"id"`
	RoomID      int64     `json:"room_id"`
	Minute      time.Time `json:"minute"`
	Count       int       `json:"count"`
	ViewerCount int       `json:"viewer_count"`
}

type GiftRecord struct {
	ID       int64     `json:"id"`
	RoomID   int64     `json:"room_id"`
	Username string    `json:"username"`
	GiftName string    `json:"gift_name"`
	Count    int       `json:"count"`
	Price    float64   `json:"price"`
	PaidAt   time.Time `json:"paid_at"`
}

type SuperChat struct {
	ID       int64     `json:"id"`
	RoomID   int64     `json:"room_id"`
	Username string    `json:"username"`
	Message  string    `json:"message"`
	Price    float64   `json:"price"`
	PaidAt   time.Time `json:"paid_at"`
}

type GuardBuy struct {
	ID         int64     `json:"id"`
	RoomID     int64     `json:"room_id"`
	Username   string    `json:"username"`
	GuardLevel string    `json:"guard_level"`
	Count      int       `json:"count"`
	PaidAt     time.Time `json:"paid_at"`
}

type BlacklistEntry struct {
	UID       int64     `json:"uid"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type QQGroup struct {
	GroupID int64  `json:"group_id"`
	Name    string `json:"name"`
}

type RoomSubscription struct {
	RoomID    int64     `json:"room_id"`
	GroupID   int64     `json:"group_id"`
	CreatedAt time.Time `json:"created_at"`
}

type StreetlightEvent struct {
	RoomID    int64
	RoomName  string
	Username  string
	UID       int64
	Content   string
	Timestamp time.Time
	LiveStart time.Time
}

type DashboardData struct {
	ListeningRooms    int            `json:"listening_rooms"`
	LiveRooms         int            `json:"live_rooms"`
	TodayStreetlights int            `json:"today_streetlights"`
	LiveRoomList      []LiveRoomInfo `json:"live_room_list"`
}

type LiveRoomInfo struct {
	RoomID       int64  `json:"room_id"`
	Name         string `json:"name"`
	Streetlights int    `json:"streetlights"`
}
