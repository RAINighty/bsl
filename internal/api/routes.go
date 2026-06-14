package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"bsl/internal/db"
	"bsl/internal/model"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.GET("/dashboard", handleDashboard)
	api.GET("/rooms", handleListRooms)
	api.POST("/rooms", handleAddRoom)
	api.DELETE("/rooms/:id", handleDeleteRoom)
	api.PUT("/rooms/:id/listening", handleSetListening)
	api.PUT("/rooms/:id/stats", handleSetStats)
	api.GET("/rooms/:id", handleRoomDetail)
	api.GET("/rooms/:id/danmaku", handleDanmakuList)
	api.GET("/rooms/:id/streetlights", handleStreetlightList)
	api.GET("/rooms/:id/gifts", handleGiftList)
	api.GET("/rooms/:id/sc", handleSCList)
	api.GET("/rooms/:id/guards", handleGuardList)
	api.GET("/rooms/:id/stats", handleDanmakuStats)
	api.GET("/blacklist", handleBlacklistList)
	api.POST("/blacklist", handleBlacklistAdd)
	api.DELETE("/blacklist/:uid", handleBlacklistRemove)
}

func handleDashboard(c *gin.Context) {
	listening, _ := db.CountListeningRooms()
	liveRooms, _ := db.GetLiveRooms()
	todaySL, _ := db.TodayStreetlightCount()
	c.JSON(http.StatusOK, model.DashboardData{
		ListeningRooms:    listening,
		LiveRooms:         len(liveRooms),
		TodayStreetlights: todaySL,
		LiveRoomList:      liveRooms,
	})
}

func handleListRooms(c *gin.Context) {
	rooms, err := db.ListRooms()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rooms == nil {
		rooms = []model.Room{}
	}
	c.JSON(http.StatusOK, rooms)
}

func handleAddRoom(c *gin.Context) {
	var r model.Room
	if err := json.NewDecoder(c.Request.Body).Decode(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if r.RoomID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id is required"})
		return
	}
	r.IsListening = true
	r.StatsEnabled = true
	if err := db.UpsertRoom(r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, r)
}

func handleDeleteRoom(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := db.DeleteRoom(roomID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func handleSetListening(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		Listening bool `json:"listening"`
	}
	json.NewDecoder(c.Request.Body).Decode(&body)
	if err := db.SetRoomListening(roomID, body.Listening); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func handleSetStats(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		Enabled bool `json:"enabled"`
	}
	json.NewDecoder(c.Request.Body).Decode(&body)
	if err := db.SetRoomStatsEnabled(roomID, body.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func handleRoomDetail(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rooms, _ := db.ListRooms()
	var found *model.Room
	for _, r := range rooms {
		if r.RoomID == roomID {
			found = &r
			break
		}
	}
	if found == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	todaySL, _ := db.TodayStreetlightCount()
	c.JSON(http.StatusOK, gin.H{"room": found, "today_streetlights": todaySL})
}

func pagination(c *gin.Context) (limit, offset int) {
	limit = 50
	offset = 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	return
}

func handleDanmakuList(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	limit, offset := pagination(c)
	items, total, err := db.GetDanmakuList(roomID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

func handleStreetlightList(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	limit, offset := pagination(c)
	items, total, err := db.GetStreetlightList(roomID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

func handleGiftList(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	limit, offset := pagination(c)
	items, total, err := db.GetGiftList(roomID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

func handleSCList(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	limit, offset := pagination(c)
	items, total, err := db.GetSCList(roomID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

func handleGuardList(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	limit, offset := pagination(c)
	items, total, err := db.GetGuardList(roomID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

func handleDanmakuStats(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	startStr := c.DefaultQuery("start", time.Now().Add(-24*time.Hour).Format(time.RFC3339))
	endStr := c.DefaultQuery("end", time.Now().Format(time.RFC3339))
	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)
	stats, err := db.GetDanmakuStats(roomID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if stats == nil {
		stats = []model.DanmakuStat{}
	}
	c.JSON(http.StatusOK, stats)
}

func handleBlacklistList(c *gin.Context) {
	entries, err := db.ListBlacklist()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if entries == nil {
		entries = []model.BlacklistEntry{}
	}
	c.JSON(http.StatusOK, entries)
}

func handleBlacklistAdd(c *gin.Context) {
	var entry struct {
		UID    int64  `json:"uid"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.AddBlacklist(entry.UID, entry.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func handleBlacklistRemove(c *gin.Context) {
	uid, _ := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err := db.RemoveBlacklist(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
