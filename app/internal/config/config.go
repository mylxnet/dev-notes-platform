package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Config 保存服务启动所需的全部配置。
type Config struct {
	Env               string
	HTTPAddr          string
	DataDir           string
	DBPath            string
	SessionSecret     string
	AdminInitUsername string
	AdminInitPassword string
	CookieSecure      bool
	LogLevel          string
	Version           string
}

// Load 从环境变量读取配置并做启动前校验。
func Load(version string) (*Config, error) {
	cfg := &Config{
		Env:               env("APP_ENV", "production"),
		HTTPAddr:          env("HTTP_ADDR", ":8080"),
		DataDir:           env("DATA_DIR", "/app/data"),
		SessionSecret:     os.Getenv("SESSION_SECRET"),
		AdminInitUsername: os.Getenv("ADMIN_INIT_USERNAME"),
		AdminInitPassword: os.Getenv("ADMIN_INIT_PASSWORD"),
		CookieSecure:      parseBool("COOKIE_SECURE", true),
		LogLevel:          env("LOG_LEVEL", "info"),
		Version:           version,
	}
	cfg.DBPath = env("DB_PATH", cfg.DataDir+"/notes.db")

	if len(cfg.SessionSecret) < 32 {
		return nil, errors.New("SESSION_SECRET 至少需要 32 个字符")
	}
	if (cfg.AdminInitUsername == "") != (cfg.AdminInitPassword == "") {
		return nil, errors.New("ADMIN_INIT_USERNAME 与 ADMIN_INIT_PASSWORD 必须同时设置")
	}
	if cfg.AdminInitPassword != "" && len(cfg.AdminInitPassword) < 12 {
		return nil, errors.New("ADMIN_INIT_PASSWORD 至少需要 12 个字符")
	}
	if cfg.HTTPAddr == "" {
		return nil, errors.New("HTTP_ADDR 不能为空")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parseBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

// String 用于启动日志，不包含任何密钥。
func (c *Config) String() string {
	return fmt.Sprintf("env=%s addr=%s db=%s version=%s", c.Env, c.HTTPAddr, c.DBPath, c.Version)
}
