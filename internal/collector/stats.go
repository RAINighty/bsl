package collector

import (
	"log"
	"sync"
	"time"
)

type StatsAccumulator struct {
	mu          sync.Mutex
	roomID      int64
	currentMin  time.Time
	count       int
	viewerCount int
	onFlush     func(DanmakuStat)
}

type DanmakuStat struct {
	RoomID      int64
	Minute      time.Time
	Count       int
	ViewerCount int
}

func NewStatsAccumulator(roomID int64, onFlush func(DanmakuStat)) *StatsAccumulator {
	return &StatsAccumulator{
		roomID:     roomID,
		currentMin: time.Now().Truncate(time.Minute),
		onFlush:    onFlush,
	}
}

func (s *StatsAccumulator) Add() {
	s.mu.Lock()
	defer s.mu.Unlock()
	nowMin := time.Now().Truncate(time.Minute)
	if !nowMin.Equal(s.currentMin) {
		s.flushLocked()
		s.currentMin = nowMin
		s.count = 0
	}
	s.count++
}

func (s *StatsAccumulator) SetViewers(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.viewerCount = n
}

func (s *StatsAccumulator) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flushLocked()
}

func (s *StatsAccumulator) flushLocked() {
	if s.count == 0 && s.viewerCount == 0 {
		return
	}
	if s.onFlush != nil {
		s.onFlush(DanmakuStat{
			RoomID:      s.roomID,
			Minute:      s.currentMin,
			Count:       s.count,
			ViewerCount: s.viewerCount,
		})
		log.Printf("[stats:%d] flushed: %s count=%d viewers=%d", s.roomID, s.currentMin.Format("15:04"), s.count, s.viewerCount)
	}
}
