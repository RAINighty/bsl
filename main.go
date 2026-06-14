package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bsl/internal/api"
	"bsl/internal/bot"
	"bsl/internal/collector"
	"bsl/internal/config"
	"bsl/internal/db"
	"bsl/internal/model"

	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("BSL - Bilibili StreetLight starting...")

	// Load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Init database
	if err := db.Init(cfg.Database.URL, cfg.Database.MaxConnections); err != nil {
		log.Fatalf("db init: %v", err)
	}
	defer db.Close()
	log.Println("database connected")

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("migrations applied")

	// OneBot WS server
	botServer := bot.NewServer(cfg.OneBot.WSPath)
	notifier := bot.NewNotifier(botServer)

	// Collector callbacks bridge
	callbacks := collector.Callbacks{
		OnDanmaku: func(r model.DanmakuRecord) {
			if err := db.InsertDanmaku(r); err != nil {
				log.Printf("insert danmaku: %v", err)
			}
		},
		OnGift: func(r model.GiftRecord) {
			if err := db.InsertGift(r); err != nil {
				log.Printf("insert gift: %v", err)
			}
		},
		OnSuperChat: func(r model.SuperChat) {
			if err := db.InsertSC(r); err != nil {
				log.Printf("insert sc: %v", err)
			}
		},
		OnGuard: func(r model.GuardBuy) {
			if err := db.InsertGuard(r); err != nil {
				log.Printf("insert guard: %v", err)
			}
		},
		OnStreetlight: func(event model.StreetlightEvent) {
			notifier.SendStreetlight(event)
		},
		OnStats: func(stat model.DanmakuStat) {
			db.UpsertDanmakuStat(stat)
		},
		IsBlacklisted: func(uid int64) bool {
			yes, _ := db.IsBlacklisted(uid)
			return yes
		},
	}

	// Collector manager
	mgr := collector.NewManager(callbacks, cfg.Bilibili.Cookie)

	// Load existing rooms and start collection
	rooms, err := db.ListRooms()
	if err != nil {
		log.Printf("list rooms: %v", err)
	}
	for _, room := range rooms {
		if room.IsListening {
			mgr.AddRoom(room)
		}
	}

	// Bot command processor
	_ = bot.NewCommandProc(botServer, bot.CommandCallbacks{
		EnsureGroup: func(groupID int64, name string) error {
			return db.EnsureQQGroup(groupID, name)
		},
		SubscribeRoom: func(roomID, groupID int64) error {
			return db.SubscribeRoom(roomID, groupID)
		},
		UnsubscribeRoom: func(roomID, groupID int64) error {
			return db.UnsubscribeRoom(roomID, groupID)
		},
		ListSubscriptions: func(groupID int64) ([]int64, error) {
			return db.GetSubscribedRooms(groupID)
		},
		ListBlacklist: func() ([]model.BlacklistEntry, error) {
			return db.ListBlacklist()
		},
		AddBlacklist: func(uid int64, reason string) error {
			return db.AddBlacklist(uid, reason)
		},
		RemoveBlacklist: func(uid int64) error {
			return db.RemoveBlacklist(uid)
		},
	})

	// API Server
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	api.SetupRoutes(r)
	r.GET(cfg.OneBot.WSPath, func(c *gin.Context) {
		botServer.HandleWS(c.Writer, c.Request)
	})

	// Static files for web panel
	r.Static("/", "./web/dist")

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("shutting down...")
		cancel()
	}()

	// Start API
	go func() {
		addr := cfg.Server.Host + ":" + itoa(cfg.Server.Port)
		log.Printf("API server listening on %s", addr)
		if err := r.Run(addr); err != nil {
			log.Printf("api server: %v", err)
		}
	}()

	// Wait for signal
	<-ctx.Done()
	log.Println("BSL stopped.")
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
