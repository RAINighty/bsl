package db

import (
	"context"
	"time"

	"bsl/internal/model"
)

// ─── Room queries ────────────────────────────────────────────────

func ListRooms() ([]model.Room, error) {
	rows, err := Pool.Query(context.Background(),
		`SELECT room_id, uid, name, is_listening, stats_enabled, last_live_start FROM rooms ORDER BY room_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Room
	for rows.Next() {
		var r model.Room
		if err := rows.Scan(&r.RoomID, &r.UID, &r.Name, &r.IsListening, &r.StatsEnabled, &r.LastLiveStart); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func UpsertRoom(r model.Room) error {
	_, err := Pool.Exec(context.Background(),
		`INSERT INTO rooms (room_id, uid, name, is_listening, stats_enabled, last_live_start)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (room_id) DO UPDATE SET uid=$2, name=$3, is_listening=$4, stats_enabled=$5, last_live_start=$6`,
		r.RoomID, r.UID, r.Name, r.IsListening, r.StatsEnabled, r.LastLiveStart)
	return err
}

func DeleteRoom(roomID int64) error {
	_, err := Pool.Exec(context.Background(), `DELETE FROM rooms WHERE room_id=$1`, roomID)
	return err
}

func SetRoomListening(roomID int64, listening bool) error {
	_, err := Pool.Exec(context.Background(),
		`UPDATE rooms SET is_listening=$2 WHERE room_id=$1`, roomID, listening)
	return err
}

func SetRoomStatsEnabled(roomID int64, enabled bool) error {
	_, err := Pool.Exec(context.Background(),
		`UPDATE rooms SET stats_enabled=$2 WHERE room_id=$1`, roomID, enabled)
	return err
}

func CountListeningRooms() (int, error) {
	var n int
	err := Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM rooms WHERE is_listening=true`).Scan(&n)
	return n, err
}

func GetLiveRooms() ([]model.LiveRoomInfo, error) {
	rows, err := Pool.Query(context.Background(),
		`SELECT r.room_id, r.name, COUNT(d.id) AS sl
		 FROM rooms r LEFT JOIN danmaku_records d
		   ON r.room_id = d.room_id AND d.is_streetlight = true
		 WHERE r.last_live_start > NOW() - INTERVAL '1 hour'
		 GROUP BY r.room_id, r.name ORDER BY sl DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.LiveRoomInfo
	for rows.Next() {
		var info model.LiveRoomInfo
		if err := rows.Scan(&info.RoomID, &info.Name, &info.Streetlights); err != nil {
			return nil, err
		}
		out = append(out, info)
	}
	return out, nil
}

// ─── Danmaku queries ─────────────────────────────────────────────

func InsertDanmaku(r model.DanmakuRecord) error {
	_, err := Pool.Exec(context.Background(),
		`INSERT INTO danmaku_records (room_id, username, uid, content, is_streetlight, streetlight_note, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		r.RoomID, r.Username, r.UID, r.Content, r.IsStreetlight, r.StreetlightNote, r.CreatedAt)
	return err
}

func GetDanmakuList(roomID int64, limit, offset int) ([]model.DanmakuRecord, int, error) {
	var total int
	if err := Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM danmaku_records WHERE room_id=$1`, roomID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := Pool.Query(context.Background(),
		`SELECT id, room_id, username, uid, content, is_streetlight, streetlight_note, created_at
		 FROM danmaku_records WHERE room_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		roomID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []model.DanmakuRecord
	for rows.Next() {
		var d model.DanmakuRecord
		if err := rows.Scan(&d.ID, &d.RoomID, &d.Username, &d.UID,
			&d.Content, &d.IsStreetlight, &d.StreetlightNote, &d.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, d)
	}
	return out, total, nil
}

func GetStreetlightList(roomID int64, limit, offset int) ([]model.DanmakuRecord, int, error) {
	var total int
	if err := Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM danmaku_records WHERE room_id=$1 AND is_streetlight=true`, roomID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := Pool.Query(context.Background(),
		`SELECT id, room_id, username, uid, content, is_streetlight, streetlight_note, created_at
		 FROM danmaku_records WHERE room_id=$1 AND is_streetlight=true ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		roomID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []model.DanmakuRecord
	for rows.Next() {
		var d model.DanmakuRecord
		if err := rows.Scan(&d.ID, &d.RoomID, &d.Username, &d.UID,
			&d.Content, &d.IsStreetlight, &d.StreetlightNote, &d.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, d)
	}
	return out, total, nil
}

func TodayStreetlightCount() (int, error) {
	var n int
	err := Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM danmaku_records
		 WHERE is_streetlight=true AND created_at >= CURRENT_DATE`).Scan(&n)
	return n, err
}

func GetStreetlightForNotif(roomID int64, lastID int64) ([]model.StreetlightEvent, error) {
	rows, err := Pool.Query(context.Background(),
		`SELECT d.room_id, r.name, d.username, d.uid, d.streetlight_note, d.created_at, r.last_live_start
		 FROM danmaku_records d JOIN rooms r ON d.room_id = r.room_id
		 WHERE d.is_streetlight=true AND d.id > $1 AND ($2=0 OR d.room_id=$2)
		 ORDER BY d.id`, lastID, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.StreetlightEvent
	for rows.Next() {
		var e model.StreetlightEvent
		if err := rows.Scan(&e.RoomID, &e.RoomName, &e.Username, &e.UID,
			&e.Content, &e.Timestamp, &e.LiveStart); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

// ─── Danmaku stats ───────────────────────────────────────────────

func UpsertDanmakuStat(stat model.DanmakuStat) error {
	_, err := Pool.Exec(context.Background(),
		`INSERT INTO danmaku_stats (room_id, minute, count, viewer_count)
		 VALUES ($1,$2,$3,$4)
		 ON CONFLICT (room_id, minute) DO UPDATE SET count=$3, viewer_count=$4`,
		stat.RoomID, stat.Minute, stat.Count, stat.ViewerCount)
	return err
}

func GetDanmakuStats(roomID int64, start, end time.Time) ([]model.DanmakuStat, error) {
	rows, err := Pool.Query(context.Background(),
		`SELECT room_id, minute, count, viewer_count FROM danmaku_stats
		 WHERE room_id=$1 AND minute BETWEEN $2 AND $3 ORDER BY minute`,
		roomID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.DanmakuStat
	for rows.Next() {
		var s model.DanmakuStat
		if err := rows.Scan(&s.RoomID, &s.Minute, &s.Count, &s.ViewerCount); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

// ─── Gift / SC / Guard queries ───────────────────────────────────

func InsertGift(r model.GiftRecord) error {
	_, err := Pool.Exec(context.Background(),
		`INSERT INTO gift_records (room_id, username, gift_name, count, price, paid_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		r.RoomID, r.Username, r.GiftName, r.Count, r.Price, r.PaidAt)
	return err
}

func GetGiftList(roomID int64, limit, offset int) ([]model.GiftRecord, int, error) {
	var total int
	if err := Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM gift_records WHERE room_id=$1`, roomID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := Pool.Query(context.Background(),
		`SELECT id, room_id, username, gift_name, count, price, paid_at
		 FROM gift_records WHERE room_id=$1 ORDER BY paid_at DESC LIMIT $2 OFFSET $3`,
		roomID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []model.GiftRecord
	for rows.Next() {
		var g model.GiftRecord
		if err := rows.Scan(&g.ID, &g.RoomID, &g.Username, &g.GiftName, &g.Count, &g.Price, &g.PaidAt); err != nil {
			return nil, 0, err
		}
		out = append(out, g)
	}
	return out, total, nil
}

func InsertSC(sc model.SuperChat) error {
	_, err := Pool.Exec(context.Background(),
		`INSERT INTO super_chats (room_id, username, message, price, paid_at) VALUES ($1,$2,$3,$4,$5)`,
		sc.RoomID, sc.Username, sc.Message, sc.Price, sc.PaidAt)
	return err
}

func GetSCList(roomID int64, limit, offset int) ([]model.SuperChat, int, error) {
	var total int
	if err := Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM super_chats WHERE room_id=$1`, roomID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := Pool.Query(context.Background(),
		`SELECT id, room_id, username, message, price, paid_at
		 FROM super_chats WHERE room_id=$1 ORDER BY paid_at DESC LIMIT $2 OFFSET $3`,
		roomID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []model.SuperChat
	for rows.Next() {
		var s model.SuperChat
		if err := rows.Scan(&s.ID, &s.RoomID, &s.Username, &s.Message, &s.Price, &s.PaidAt); err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	return out, total, nil
}

func InsertGuard(g model.GuardBuy) error {
	_, err := Pool.Exec(context.Background(),
		`INSERT INTO guard_buys (room_id, username, guard_level, count, paid_at) VALUES ($1,$2,$3,$4,$5)`,
		g.RoomID, g.Username, g.GuardLevel, g.Count, g.PaidAt)
	return err
}

func GetGuardList(roomID int64, limit, offset int) ([]model.GuardBuy, int, error) {
	var total int
	if err := Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM guard_buys WHERE room_id=$1`, roomID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := Pool.Query(context.Background(),
		`SELECT id, room_id, username, guard_level, count, paid_at
		 FROM guard_buys WHERE room_id=$1 ORDER BY paid_at DESC LIMIT $2 OFFSET $3`,
		roomID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []model.GuardBuy
	for rows.Next() {
		var g model.GuardBuy
		if err := rows.Scan(&g.ID, &g.RoomID, &g.Username, &g.GuardLevel, &g.Count, &g.PaidAt); err != nil {
			return nil, 0, err
		}
		out = append(out, g)
	}
	return out, total, nil
}

// ─── Blacklist queries ───────────────────────────────────────────

func IsBlacklisted(uid int64) (bool, error) {
	var exists bool
	err := Pool.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM blacklist WHERE uid=$1)`, uid).Scan(&exists)
	return exists, err
}

func ListBlacklist() ([]model.BlacklistEntry, error) {
	rows, err := Pool.Query(context.Background(),
		`SELECT uid, reason, created_at FROM blacklist ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.BlacklistEntry
	for rows.Next() {
		var b model.BlacklistEntry
		if err := rows.Scan(&b.UID, &b.Reason, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

func AddBlacklist(uid int64, reason string) error {
	_, err := Pool.Exec(context.Background(),
		`INSERT INTO blacklist (uid, reason) VALUES ($1,$2) ON CONFLICT (uid) DO UPDATE SET reason=$2`,
		uid, reason)
	return err
}

func RemoveBlacklist(uid int64) error {
	_, err := Pool.Exec(context.Background(), `DELETE FROM blacklist WHERE uid=$1`, uid)
	return err
}

// ─── QQ group / subscription queries ────────────────────────────

func EnsureQQGroup(groupID int64, name string) error {
	_, err := Pool.Exec(context.Background(),
		`INSERT INTO qq_groups (group_id, name) VALUES ($1,$2) ON CONFLICT (group_id) DO UPDATE SET name=$2`,
		groupID, name)
	return err
}

func ListQQGroups() ([]model.QQGroup, error) {
	rows, err := Pool.Query(context.Background(),
		`SELECT group_id, name FROM qq_groups ORDER BY group_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.QQGroup
	for rows.Next() {
		var g model.QQGroup
		if err := rows.Scan(&g.GroupID, &g.Name); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, nil
}

func SubscribeRoom(roomID, groupID int64) error {
	_, err := Pool.Exec(context.Background(),
		`INSERT INTO room_subscriptions (room_id, group_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		roomID, groupID)
	return err
}

func UnsubscribeRoom(roomID, groupID int64) error {
	_, err := Pool.Exec(context.Background(),
		`DELETE FROM room_subscriptions WHERE room_id=$1 AND group_id=$2`,
		roomID, groupID)
	return err
}

func GetRoomSubscribers(roomID int64) ([]int64, error) {
	rows, err := Pool.Query(context.Background(),
		`SELECT group_id FROM room_subscriptions WHERE room_id=$1`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var gid int64
		if err := rows.Scan(&gid); err != nil {
			return nil, err
		}
		out = append(out, gid)
	}
	return out, nil
}

func GetSubscribedRooms(groupID int64) ([]int64, error) {
	rows, err := Pool.Query(context.Background(),
		`SELECT room_id FROM room_subscriptions WHERE group_id=$1`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var rid int64
		if err := rows.Scan(&rid); err != nil {
			return nil, err
		}
		out = append(out, rid)
	}
	return out, nil
}
