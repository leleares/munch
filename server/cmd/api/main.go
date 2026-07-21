package main

import (
	"log"

	"munch/server/internal/config"
	"munch/server/internal/database"
	"munch/server/internal/handler"
	"munch/server/internal/router"
	"munch/server/internal/storage"
)

func main() {
	cfg := config.Load()

	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("[db] 连接失败: %v", err)
	}

	st := storage.New(cfg)
	h := handler.New(db, cfg, st)
	r := router.Setup(db, cfg, h)

	addr := ":" + cfg.Port
	log.Printf("[munch] listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("[munch] 启动失败: %v", err)
	}
}
