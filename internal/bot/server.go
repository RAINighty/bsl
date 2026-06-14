package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

type Server struct {
	Path       string
	OnMessage  func(FromInfo, string)
	clients    map[*websocket.Conn]bool
	mu         sync.RWMutex
}

type FromInfo struct {
	UserID     int64  `json:"user_id"`
	GroupID    int64  `json:"group_id"`
	Nickname   string `json:"nickname"`
	MessageType string `json:"message_type"`
}

func NewServer(path string) *Server {
	return &Server{Path: path, clients: make(map[*websocket.Conn]bool)}
}

func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[bot] upgrade: %v", err)
		return
	}
	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()
	log.Printf("[bot] client connected (%d total)", len(s.clients))
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		conn.Close()
	}()
	s.readLoop(conn)
}

func (s *Server) readLoop(conn *websocket.Conn) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var envelope struct {
			PostType    string          `json:"post_type"`
			MessageType string          `json:"message_type"`
			RawMessage  string          `json:"raw_message"`
			Sender      OneBotSender    `json:"sender"`
			GroupID     int64           `json:"group_id"`
			UserID      int64           `json:"user_id"`
		}
		if err := json.Unmarshal(msg, &envelope); err != nil {
			continue
		}
		if envelope.PostType != "message" {
			continue
		}
		if s.OnMessage == nil {
			continue
		}
		from := FromInfo{
			UserID:      envelope.UserID,
			GroupID:     envelope.GroupID,
			Nickname:    envelope.Sender.Nickname,
			MessageType: envelope.MessageType,
		}
		go s.OnMessage(from, envelope.RawMessage)
	}
}

type OneBotSender struct {
	Nickname string `json:"nickname"`
}

type GroupMsg struct {
	Action string     `json:"action"`
	Params GroupParams `json:"params"`
}

type GroupParams struct {
	GroupID int64  `json:"group_id"`
	Message string `json:"message"`
}

func (s *Server) SendGroupMessage(groupID int64, msg string) error {
	payload := GroupMsg{
		Action: "send_group_msg",
		Params: GroupParams{GroupID: groupID, Message: msg},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.clients) == 0 {
		return fmt.Errorf("no onebot client connected")
	}
	var lastErr error
	for conn := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
