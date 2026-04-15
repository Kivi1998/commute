package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Amap     AmapConfig
	Doubao   DoubaoConfig
}

type AppConfig struct {
	Env     string
	Port    string
	Version string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type AmapConfig struct {
	WebServiceKey string
	BaseURL       string
	TimeoutMs     int
}

type DoubaoConfig struct {
	APIKey    string
	BaseURL   string
	Model     string
	TimeoutMs int
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

func (a AmapConfig) Timeout() time.Duration {
	return time.Duration(a.TimeoutMs) * time.Millisecond
}

func (d DoubaoConfig) Timeout() time.Duration {
	return time.Duration(d.TimeoutMs) * time.Millisecond
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Env:     getEnv("APP_ENV", "development"),
			Port:    getEnv("APP_PORT", "8080"),
			Version: getEnv("APP_VERSION", "1.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "127.0.0.1"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "commute"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Amap: AmapConfig{
			WebServiceKey: getEnv("AMAP_WS_KEY", ""),
			BaseURL:       getEnv("AMAP_WS_BASE", "https://restapi.amap.com"),
			TimeoutMs:     getEnvInt("AMAP_TIMEOUT_MS", 10000),
		},
		Doubao: DoubaoConfig{
			APIKey:    getEnv("DOUBAO_API_KEY", ""),
			BaseURL:   getEnv("DOUBAO_BASE", "https://ark.cn-beijing.volces.com"),
			Model:     getEnv("DOUBAO_MODEL", ""),
			TimeoutMs: getEnvInt("DOUBAO_TIMEOUT_MS", 60000),
		},
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
