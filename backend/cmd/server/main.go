package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	appconfig "github.com/time/card/backend/internal/config"
	"github.com/time/card/backend/internal/database"
	"github.com/time/card/backend/internal/router"
)

func main() {
	defaultConfig := filepath.Join("config", "config.yaml")
	configPath := flag.String("config", envOrDefault("CARD_CONFIG_PATH", defaultConfig), "path to config yaml")
	flag.Parse()

	absConfig, err := filepath.Abs(*configPath)
	if err != nil {
		log.Fatalf("resolve config path: %v", err)
	}

	cfg, err := appconfig.Load(absConfig)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.NewMySQL(&cfg.MySQL)
	if err != nil {
		log.Fatalf("init mysql: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	rdb, err := database.NewRedis(&cfg.Redis)
	if err != nil {
		log.Fatalf("init redis: %v", err)
	}

	engine := router.New(cfg, db, rdb)
	log.Printf("server starting on %s (config: %s)", cfg.Addr(), absConfig)
	if err := engine.Run(cfg.Addr()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
