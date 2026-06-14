package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"bsl/internal/model"
)

const HelpText = `BSL 路灯系统指令：
#路灯 内容  - 标记当前高能时刻
help / 帮助  - 显示此帮助
房间 [数字]  - 订阅/取消本群与指定房间的路灯通知
黑名单 list   - 查看黑名单
黑名单 add [uid] [原因] - 添加黑名单
黑名单 remove [uid] - 移除黑名单
订阅列表 - 查看本群已订阅的房间
`

type CommandProc struct {
	server *Server
	callbacks CommandCallbacks
}

type CommandCallbacks struct {
	GetGroups          func() ([]model.QQGroup, error)
	EnsureGroup        func(groupID int64, name string) error
	SubscribeRoom      func(roomID, groupID int64) error
	UnsubscribeRoom    func(roomID, groupID int64) error
	ListSubscriptions  func(groupID int64) ([]int64, error)
	ListBlacklist      func() ([]model.BlacklistEntry, error)
	AddBlacklist       func(uid int64, reason string) error
	RemoveBlacklist    func(uid int64) error
}

func NewCommandProc(srv *Server, callbacks CommandCallbacks) *CommandProc {
	cp := &CommandProc{server: srv, callbacks: callbacks}
	srv.OnMessage = cp.Handle
	return cp
}

func (c *CommandProc) Handle(from FromInfo, msg string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return
	}

	if strings.HasPrefix(msg, "#路灯") || strings.HasPrefix(msg, "#路燈") {
		c.reply(from, "路灯标记功能由B站直播间直接处理，请将 #路灯 内容发送到B站弹幕中。\nQQ群内支持以下管理指令：\n"+HelpText)
		return
	}

	switch {
	case msg == "help" || msg == "帮助":
		c.reply(from, HelpText)
	case strings.HasPrefix(msg, "房间 "):
		c.handleRoom(from, msg)
	case strings.HasPrefix(msg, "黑名单 list"):
		c.handleBlacklistList(from)
	case strings.HasPrefix(msg, "黑名单 add "):
		c.handleBlacklistAdd(from, msg)
	case strings.HasPrefix(msg, "黑名单 remove "):
		c.handleBlacklistRemove(from, msg)
	case msg == "订阅列表":
		c.handleSubList(from)
	default:
		c.reply(from, "未知指令。发送 help 查看可用指令。")
	}
}

func (c *CommandProc) reply(from FromInfo, text string) {
	if from.MessageType == "group" && from.GroupID > 0 {
		c.server.SendGroupMessage(from.GroupID, text)
	} else {
		log.Printf("[bot:private] %s: %s", from.Nickname, text)
	}
}

func (c *CommandProc) handleRoom(from FromInfo, msg string) {
	if from.GroupID == 0 {
		c.reply(from, "仅群聊可用。")
		return
	}
	parts := strings.Fields(msg)
	if len(parts) < 2 {
		c.reply(from, "用法：房间 [房间号]")
		return
	}
	roomID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		c.reply(from, "房间号格式错误。")
		return
	}
	if err := c.callbacks.EnsureGroup(from.GroupID, from.Nickname+"的群"); err != nil {
		log.Printf("[bot] ensure group: %v", err)
	}
	subs, _ := c.callbacks.ListSubscriptions(from.GroupID)
	for _, rid := range subs {
		if rid == roomID {
			if err := c.callbacks.UnsubscribeRoom(roomID, from.GroupID); err != nil {
				c.reply(from, fmt.Sprintf("取消订阅失败: %v", err))
				return
			}
			c.reply(from, fmt.Sprintf("已取消订阅房间 %d。", roomID))
			return
		}
	}
	if err := c.callbacks.SubscribeRoom(roomID, from.GroupID); err != nil {
		c.reply(from, fmt.Sprintf("订阅失败: %v", err))
		return
	}
	c.reply(from, fmt.Sprintf("已订阅房间 %d 的路灯通知。", roomID))
}

func (c *CommandProc) handleBlacklistList(from FromInfo) {
	entries, err := c.callbacks.ListBlacklist()
	if err != nil {
		c.reply(from, fmt.Sprintf("查询失败: %v", err))
		return
	}
	if len(entries) == 0 {
		c.reply(from, "黑名单为空。")
		return
	}
	var lines []string
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("UID: %d - %s", e.UID, e.Reason))
	}
	c.reply(from, "黑名单:\n"+strings.Join(lines, "\n"))
}

func (c *CommandProc) handleBlacklistAdd(from FromInfo, msg string) {
	parts := strings.Fields(msg)
	if len(parts) < 3 {
		c.reply(from, "用法：黑名单 add [uid] [原因]")
		return
	}
	uid, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		c.reply(from, "UID格式错误。")
		return
	}
	reason := ""
	if len(parts) > 3 {
		reason = strings.Join(parts[3:], " ")
	}
	if err := c.callbacks.AddBlacklist(uid, reason); err != nil {
		c.reply(from, fmt.Sprintf("添加失败: %v", err))
		return
	}
	c.reply(from, fmt.Sprintf("已添加 UID %d 到黑名单。", uid))
}

func (c *CommandProc) handleBlacklistRemove(from FromInfo, msg string) {
	parts := strings.Fields(msg)
	if len(parts) < 3 {
		c.reply(from, "用法：黑名单 remove [uid]")
		return
	}
	uid, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		c.reply(from, "UID格式错误。")
		return
	}
	if err := c.callbacks.RemoveBlacklist(uid); err != nil {
		c.reply(from, fmt.Sprintf("移除失败: %v", err))
		return
	}
	c.reply(from, fmt.Sprintf("已从黑名单移除 UID %d。", uid))
}

func (c *CommandProc) handleSubList(from FromInfo) {
	if from.GroupID == 0 {
		c.reply(from, "仅群聊可用。")
		return
	}
	rooms, err := c.callbacks.ListSubscriptions(from.GroupID)
	if err != nil {
		c.reply(from, fmt.Sprintf("查询失败: %v", err))
		return
	}
	if len(rooms) == 0 {
		c.reply(from, "本群未订阅任何房间。发送 房间 [房号] 订阅。")
		return
	}
	var lines []string
	for _, rid := range rooms {
		lines = append(lines, fmt.Sprintf("- 房间 %d", rid))
	}
	c.reply(from, "本群订阅:\n"+strings.Join(lines, "\n"))
}
