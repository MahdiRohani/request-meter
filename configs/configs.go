package configs

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type AppConfigs struct {
	DatabaseURL  string
	RedisAddr    string
	ServerPort   string
	SyncInterval time.Duration
	InstanceName string
}

var ApplicationConfigs AppConfigs

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found. Using system environment variables.")
	}

	interval, err := time.ParseDuration(os.Getenv("SYNC_INTERVAL"))

	if err != nil || interval <= 0 {
		interval = 500 * time.Millisecond
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	ApplicationConfigs = AppConfigs{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		RedisAddr:    os.Getenv("REDIS_ADDR"),
		ServerPort:   port,
		SyncInterval: interval,
		InstanceName: os.Getenv("INSTANCE_NAME"),
	}
}
