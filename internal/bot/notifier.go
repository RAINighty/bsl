package bot

import (
	"fmt"
	"log"
	"time"

	"bsl/internal/db"
	"bsl/internal/model"
)

type Notifier struct {
	server *Server
}

func NewNotifier(srv *Server) *Notifier {
	return &Notifier{server: srv}
}

func (n *Notifier) SendStreetlight(event model.StreetlightEvent) {
	subscribers, err := db.GetRoomSubscribers(event.RoomID)
	if err != nil {
		log.Printf("[notifier] get subscribers: %v", err)
		return
	}
	if len(subscribers) == 0 {
		return
	}
	liveDuration := event.Timestamp.Sub(event.LiveStart).Truncate(time.Second)
	msg := fmt.Sprintf("💡 路灯 #%s\n房间: %s (%d)\n用户: %s\n内容: %s\n直播时长: %s",
		event.Timestamp.Format("15:04:05"),
		event.RoomName, event.RoomID,
		event.Username,
		event.Content,
		formatDuration(liveDuration),
	)
	for _, groupID := range subscribers {
		if err := n.server.SendGroupMessage(groupID, msg); err != nil {
			log.Printf("[notifier] send to %d: %v", groupID, err)
		}
	}
}

func formatDuration(d time.Duration) string {
	d = d.Truncate(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	}
	return fmt.Sprintf("%dm%02ds", m, s)
}
