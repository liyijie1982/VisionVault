package config

import (
	"os"
	"strconv"
)

type Config struct {
	App   AppConfig
	Agent AgentConfig
	HTTP  HTTPConfig
	Auth  AuthConfig
	DB    DBConfig
	Redis RedisConfig
}

type AppConfig struct {
	Name string
	Env  string
}

type HTTPConfig struct {
	Addr string
}

type AuthConfig struct {
	AdminUsername string
	AdminPassword string
	CaptchaTTL    string
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Addr       string
	Password   string
	DB         int
	SessionTTL string
}

type AgentConfig struct {
	Version      string
	DownloadID   string
	DownloadFile string
}

func Load() Config {
	return Config{
		App: AppConfig{
			Name: getEnv("SKYBASE_APP_NAME", "SkyBase"),
			Env:  getEnv("SKYBASE_ENV", "dev"),
		},
		Agent: AgentConfig{
			Version:      getEnv("SKYBASE_AGENT_VERSION", ""),
			DownloadID:   getEnv("SKYBASE_AGENT_DOWNLOAD_ID", "latest"),
			DownloadFile: getEnv("SKYBASE_AGENT_DOWNLOAD_FILE", ""),
		},
		HTTP: HTTPConfig{
			Addr: getEnv("SKYBASE_HTTP_ADDR", ":8080"),
		},
		Auth: AuthConfig{
			AdminUsername: getEnv("SKYBASE_ADMIN_USERNAME", "admin"),
			AdminPassword: getEnv("SKYBASE_ADMIN_PASSWORD", "admin123"),
			CaptchaTTL:    getEnv("SKYBASE_CAPTCHA_TTL", "5m"),
		},
		DB: DBConfig{
			Host:     getEnv("SKYBASE_DB_HOST", "127.0.0.1"),
			Port:     getEnvAsInt("SKYBASE_DB_PORT", 3306),
			User:     getEnv("SKYBASE_DB_USER", "root"),
			Password: getEnv("SKYBASE_DB_PASSWORD", ""),
			Name:     getEnv("SKYBASE_DB_NAME", "skyvv"),
		},
		Redis: RedisConfig{
			Addr:       getEnv("SKYBASE_REDIS_ADDR", "127.0.0.1:6379"),
			Password:   getEnv("SKYBASE_REDIS_PASSWORD", ""),
			DB:         getEnvAsInt("SKYBASE_REDIS_DB", 0),
			SessionTTL: getEnv("SKYBASE_SESSION_TTL", "12h"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}
