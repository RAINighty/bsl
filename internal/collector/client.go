package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const bilibiliWS = "wss://broadcastlv.chat.bilibili.com/sub"

type BiliClient struct {
	RoomID    int64
	Token     string
	conn      *websocket.Conn
	mu        sync.Mutex
	running   bool
	stopped   chan struct{}
	OnMessage func(op uint32, body []byte)
}

func NewBiliClient(roomID int64, token string) *BiliClient {
	return &BiliClient{RoomID: roomID, Token: token, stopped: make(chan struct{})}
}

func (c *BiliClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(bilibiliWS, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	c.conn = conn
	if err := c.sendAuth(); err != nil {
		conn.Close()
		return err
	}
	log.Printf("[collector:%d] connected", c.RoomID)
	return nil
}

func (c *BiliClient) sendAuth() error {
	auth := map[string]interface{}{"uid": 0, "roomid": c.RoomID, "protover": 3, "platform": "web", "type": 2}
	if c.Token != "" {
		auth["key"] = c.Token
	}
	body, err := json.Marshal(auth)
	if err != nil {
		return err
	}
	pw := NewPacketWriter(c.conn.UnderlyingConn())
	return pw.Write(OpAuth, body)
}

func (c *BiliClient) Start() {
	c.running = true
	go c.readLoop()
	go c.heartbeatLoop()
}

func (c *BiliClient) readLoop() {
	for c.running {
		pkt, err := ParsePacket(c.conn.UnderlyingConn())
		if err != nil {
			if c.running {
				log.Printf("[collector:%d] read error: %v", c.RoomID, err)
			}
			return
		}
		if c.OnMessage != nil {
			c.OnMessage(pkt.Op, pkt.Body)
		}
	}
}

func (c *BiliClient) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for c.running {
		select {
		case <-ticker.C:
			c.mu.Lock()
			if c.conn != nil {
				pw := NewPacketWriter(c.conn.UnderlyingConn())
				pw.Write(OpHeartbeat, []byte("[object Object]"))
			}
			c.mu.Unlock()
		case <-c.stopped:
			return
		}
	}
}

func (c *BiliClient) Stop() {
	c.running = false
	close(c.stopped)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
	}
	log.Printf("[collector:%d] stopped", c.RoomID)
}
