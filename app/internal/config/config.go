package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
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
	// CookieSecureMode 取 auto / always / never，默认 auto：按请求协议决定 Secure 标记。
	CookieSecureMode string
	LogLevel         string
	Version          string
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
		LogLevel:          env("LOG_LEVEL", "info"),
		Version:           version,
	}
	cfg.DBPath = env("DB_PATH", cfg.DataDir+"/notes.db")

	secureMode, err := parseCookieSecure(os.Getenv("COOKIE_SECURE"))
	if err != nil {
		return nil, err
	}
	cfg.CookieSecureMode = secureMode

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

// parseCookieSecure 把 COOKIE_SECURE 归一化为 auto / always / never。
// 默认 auto：按请求协议判断 Secure 标记，避免 http 部署下 Cookie 被浏览器丢弃。
func parseCookieSecure(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "auto":
		return "auto", nil
	case "true", "1", "yes", "always":
		return "always", nil
	case "false", "0", "no", "never":
		return "never", nil
	default:
		return "", fmt.Errorf("COOKIE_SECURE 只能是 auto / true / false，当前为 %q", raw)
	}
}

// String 用于启动日志，不包含任何密钥。
func (c *Config) String() string {
	return fmt.Sprintf("env=%s addr=%s db=%s version=%s", c.Env, c.HTTPAddr, c.DBPath, c.Version)
}
