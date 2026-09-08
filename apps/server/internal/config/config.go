package config

import (
	"os"
	"strings"
)

type Config struct {
	MongoURI        string
	DBName          string
	JWTSecret       string
	ServerPort      string
	AllowedOrigins  []string
	CronSecret      string
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string
}

func Load() *Config {
	return &Config{
		MongoURI:        mustGetEnv("MONGO_URI"),
		DBName:          getEnv("DB_NAME", "reliva"),
		JWTSecret:       mustGetEnv("JWT_SECRET"),
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		AllowedOrigins:  strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
		CronSecret:      getEnv("CRON_SECRET", ""),
		VAPIDPublicKey:  getEnv("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey: getEnv("VAPID_PRIVATE_KEY", ""),
		VAPIDSubject:    getEnv("VAPID_SUBJECT", ""),
	}
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required environment variable not set: " + key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
