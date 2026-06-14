package collector

import (
	"encoding/json"
	"strings"
)

type MessageType int

const (
	MsgDanmaku  MessageType = iota
	MsgGift
	MsgSuperChat
	MsgGuard
	MsgViewerCount
	MsgOther
)

type ParsedMessage struct {
	Type        MessageType
	Raw         []byte
	Username    string
	UID         int64
	Content     string
	GiftName    string
	GiftCount   int
	GiftPrice   float64
	SCMessage   string
	SCPrice     float64
	GuardLevel  string
	GuardCount  int
	ViewerCount int
}

// SplitMessages handles multiple JSON objects concatenated in one body.
func SplitMessages(body []byte) [][]byte {
	var out [][]byte
	dec := json.NewDecoder(strings.NewReader(string(body)))
	for {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			break
		}
		out = append(out, raw)
	}
	return out
}

func ParseMessage(raw []byte) ParsedMessage {
	m := ParsedMessage{Raw: raw}
	var base struct {
		Cmd  string          `json:"cmd"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &base); err != nil {
		return m
	}
	switch base.Cmd {
	case "DANMU_MSG":
		return parseDanmaku(raw)
	case "SEND_GIFT":
		return parseGift(raw)
	case "SUPER_CHAT_MESSAGE":
		return parseSC(raw)
	case "GUARD_BUY":
		return parseGuard(raw)
	case "WATCHED_CHANGE":
		return parseViewer(raw)
	}
	m.Type = MsgOther
	return m
}

func parseDanmaku(raw []byte) ParsedMessage {
	m := ParsedMessage{Type: MsgDanmaku, Raw: raw}
	var msg struct {
		Info []interface{} `json:"info"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil || len(msg.Info) < 4 {
		return m
	}
	if info2, ok := msg.Info[2].([]interface{}); ok && len(info2) > 1 {
		if uid, ok := info2[0].(float64); ok {
			m.UID = int64(uid)
		}
		if name, ok := info2[1].(string); ok {
			m.Username = name
		}
	}
	if content, ok := msg.Info[1].(string); ok {
		m.Content = content
	}
	return m
}

func parseGift(raw []byte) ParsedMessage {
	m := ParsedMessage{Type: MsgGift, Raw: raw}
	var msg struct {
		Data struct {
			Uname    string  `json:"uname"`
			UID      int64   `json:"uid"`
			GiftName string  `json:"giftName"`
			Num      int     `json:"num"`
			Price    float64 `json:"price"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return m
	}
	m.Username = msg.Data.Uname
	m.UID = msg.Data.UID
	m.GiftName = msg.Data.GiftName
	m.GiftCount = msg.Data.Num
	m.GiftPrice = msg.Data.Price
	return m
}

func parseSC(raw []byte) ParsedMessage {
	m := ParsedMessage{Type: MsgSuperChat, Raw: raw}
	var msg struct {
		Data struct {
			UserName string  `json:"user_info>uname"`
			UID      int64   `json:"uid"`
			Message  string  `json:"message"`
			Price    float64 `json:"price"`
		} `json:"data"`
	}
	json.Unmarshal(raw, &msg)
	m.Username = msg.Data.UserName
	m.UID = msg.Data.UID
	m.SCMessage = msg.Data.Message
	m.SCPrice = msg.Data.Price
	return m
}

func parseGuard(raw []byte) ParsedMessage {
	m := ParsedMessage{Type: MsgGuard, Raw: raw}
	var msg struct {
		Data struct {
			Username   string `json:"username"`
			UID        int64  `json:"uid"`
			GuardLevel int    `json:"guard_level"`
			Num        int    `json:"num"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return m
	}
	m.Username = msg.Data.Username
	m.UID = msg.Data.UID
	m.GuardCount = msg.Data.Num
	switch msg.Data.GuardLevel {
	case 1:
		m.GuardLevel = "总督"
	case 2:
		m.GuardLevel = "提督"
	case 3:
		m.GuardLevel = "舰长"
	default:
		m.GuardLevel = "未知"
	}
	return m
}

func parseViewer(raw []byte) ParsedMessage {
	m := ParsedMessage{Type: MsgViewerCount, Raw: raw}
	var msg struct {
		Data struct {
			Number int `json:"number"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return m
	}
	m.ViewerCount = msg.Data.Number
	return m
}
