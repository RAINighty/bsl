package db

import "context"

func RunMigrations() error {
	sql := `
	CREATE TABLE IF NOT EXISTS rooms (
		room_id BIGINT PRIMARY KEY,
		uid BIGINT NOT NULL,
		name VARCHAR(255) NOT NULL DEFAULT '',
		is_listening BOOLEAN NOT NULL DEFAULT true,
		stats_enabled BOOLEAN NOT NULL DEFAULT true,
		last_live_start TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS danmaku_records (
		id BIGSERIAL PRIMARY KEY,
		room_id BIGINT NOT NULL REFERENCES rooms(room_id),
		username VARCHAR(255) NOT NULL DEFAULT '',
		uid BIGINT NOT NULL DEFAULT 0,
		content TEXT NOT NULL DEFAULT '',
		is_streetlight BOOLEAN NOT NULL DEFAULT false,
		streetlight_note TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_danmaku_room_time ON danmaku_records(room_id, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_danmaku_streetlight ON danmaku_records(is_streetlight) WHERE is_streetlight = true;

	CREATE TABLE IF NOT EXISTS danmaku_stats (
		id BIGSERIAL PRIMARY KEY,
		room_id BIGINT NOT NULL REFERENCES rooms(room_id),
		minute TIMESTAMP NOT NULL,
		count INT NOT NULL DEFAULT 0,
		viewer_count INT NOT NULL DEFAULT 0,
		UNIQUE(room_id, minute)
	);

	CREATE TABLE IF NOT EXISTS gift_records (
		id BIGSERIAL PRIMARY KEY,
		room_id BIGINT NOT NULL REFERENCES rooms(room_id),
		username VARCHAR(255) NOT NULL DEFAULT '',
		gift_name VARCHAR(255) NOT NULL DEFAULT '',
		count INT NOT NULL DEFAULT 0,
		price DECIMAL(10,2) NOT NULL DEFAULT 0,
		paid_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_gift_room_time ON gift_records(room_id, paid_at DESC);

	CREATE TABLE IF NOT EXISTS super_chats (
		id BIGSERIAL PRIMARY KEY,
		room_id BIGINT NOT NULL REFERENCES rooms(room_id),
		username VARCHAR(255) NOT NULL DEFAULT '',
		message TEXT NOT NULL DEFAULT '',
		price DECIMAL(10,2) NOT NULL DEFAULT 0,
		paid_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_sc_room_time ON super_chats(room_id, paid_at DESC);

	CREATE TABLE IF NOT EXISTS guard_buys (
		id BIGSERIAL PRIMARY KEY,
		room_id BIGINT NOT NULL REFERENCES rooms(room_id),
		username VARCHAR(255) NOT NULL DEFAULT '',
		guard_level VARCHAR(50) NOT NULL DEFAULT '',
		count INT NOT NULL DEFAULT 0,
		paid_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_guard_room_time ON guard_buys(room_id, paid_at DESC);

	CREATE TABLE IF NOT EXISTS blacklist (
		uid BIGINT PRIMARY KEY,
		reason TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS qq_groups (
		group_id BIGINT PRIMARY KEY,
		name VARCHAR(255) NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS room_subscriptions (
		room_id BIGINT NOT NULL REFERENCES rooms(room_id),
		group_id BIGINT NOT NULL REFERENCES qq_groups(group_id),
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		PRIMARY KEY(room_id, group_id)
	);
	`
	_, err := Pool.Exec(context.Background(), sql)
	return err
}
