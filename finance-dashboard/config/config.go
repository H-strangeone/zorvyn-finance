package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort           string
	JWTSecret         string
	JWTExpiryHours    int
	SeedAdminName     string
	SeedAdminEmail    string
	SeedAdminPassword string
}

var App *Config

func Load() {
	godotenv.Load()

	jwtExpiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		log.Fatal("JWT_EXPIRY_HOURS must be a number")
	}

	App = &Config{
		AppPort:           getEnv("APP_PORT", "8080"),
		JWTSecret:         mustGetEnv("JWT_SECRET"),
		JWTExpiryHours:    jwtExpiry,
		SeedAdminName:     getEnv("SEED_ADMIN_NAME", "Super Admin"),
		SeedAdminEmail:    strings.ToLower(strings.TrimSpace(mustGetEnv("SEED_ADMIN_EMAIL"))),
		SeedAdminPassword: mustGetEnv("SEED_ADMIN_PASSWORD"),
	}

	// Security: weak JWT secret is as dangerous as no secret
	if len(App.JWTSecret) < 16 {
		log.Fatal("FATAL: JWT_SECRET must be at least 16 characters long")
	}

	log.Println("✅ Config loaded successfully")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("FATAL: required environment variable %s is not set", key)
	}
	return val
}