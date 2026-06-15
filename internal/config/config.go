package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	RedisAddr     string
	RedisPassword string

	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration
	JWTAdminRefreshTTL time.Duration

	CookieDomain string

	SePayWebhookSecret string

	FromEmail string

	TrackingAPIBaseURL      string
	TrackingAPIKey          string
	TrackingPollInterval    time.Duration
	TrackingWebhookSecret   string

	UploadDir string

	CORSOrigins []string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	accessMin, _ := strconv.Atoi(getEnv("JWT_ACCESS_TTL_MIN", "15"))
	refreshDay, _ := strconv.Atoi(getEnv("JWT_REFRESH_TTL_DAY", "7"))
	adminRefreshDay, _ := strconv.Atoi(getEnv("JWT_ADMIN_REFRESH_TTL_DAY", "14"))
	pollSec, _ := strconv.Atoi(getEnv("TRACKING_POLL_INTERVAL_SEC", "300"))

	return &Config{
		AppPort: getEnv("APP_PORT", "8089"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "dosulogi"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "test1234"),

		RedisAddr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "dev-access-secret-change-in-production-min-32-chars"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "dev-refresh-secret-change-in-production-min-32-chars"),
		JWTAccessTTL:     time.Duration(accessMin) * time.Minute,
		JWTRefreshTTL:    time.Duration(refreshDay) * 24 * time.Hour,
		JWTAdminRefreshTTL: time.Duration(adminRefreshDay) * 24 * time.Hour,

		CookieDomain: getEnv("COOKIE_DOMAIN", ""),

		SePayWebhookSecret: getEnv("SEPAY_WEBHOOK_SECRET", "dev-sepay-secret"),

		FromEmail: getEnv("FROM_EMAIL", "no-reply@dosulogi.com"),

		TrackingAPIBaseURL:    getEnv("TRACKING_API_BASE_URL", "https://api.trackingprovider.com"),
		TrackingAPIKey:        getEnv("TRACKING_API_KEY", ""),
		TrackingPollInterval:  time.Duration(pollSec) * time.Second,
		TrackingWebhookSecret: getEnv("TRACKING_WEBHOOK_SECRET", "dev-tracking-secret"),

		UploadDir: getEnv("UPLOAD_DIR", "./uploads"),

		CORSOrigins: splitCSV(getEnv("CORS_ORIGINS", "http://localhost:5173,http://logi.dosutech.site")),
	}, nil
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (c *Config) DSN() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword + "@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName + "?sslmode=disable"
}

func (c *Config) AdminDSN() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword + "@" + c.DBHost + ":" + c.DBPort + "/postgres?sslmode=disable"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
