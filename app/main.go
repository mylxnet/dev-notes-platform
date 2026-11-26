package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dev-notes-platform/internal/api"
	"dev-notes-platform/internal/auth"
	"dev-notes-platform/internal/config"
	"dev-notes-platform/internal/store"
)

var version = "1.4.9"

func main() {
	if err := run(); err != nil {
		slog.Error("服务启动失败", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(version)
	if err != nil {
		return err
	}
	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	if err := os.MkdirAll(cfg.DataDir, 0o750); err != nil {
		return err
	}

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartup()

	st, err := store.Open(startupCtx, cfg.DBPath)
	if err != nil {
		return err
	}
	defer st.Close()

	if cfg.AdminInitUsername != "" {
		if err := st.SeedAdmin(startupCtx, cfg.AdminInitUsername, cfg.AdminInitPassword); err != nil {
			return err
		}
	} else {
		count, err := st.AdminCount(startupCtx)
		if err != nil {
			return err
		}
		if count == 0 {
			slog.Warn("尚未配置管理员账号，后台接口不可用", "hint", "设置 ADMIN_INIT_USERNAME 与 ADMIN_INIT_PASSWORD 后重启")
		}
	}

	sessions := auth.NewManager(cfg.SessionSecret, auth.CookieSecureMode(cfg.CookieSecureMode))
	router := api.New(st, version, sessions, cfg.DataDir+"/media")
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("服务已启动", "addr", cfg.HTTPAddr, "version", version, "env", cfg.Env)
		errCh <- server.ListenAndServe()
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case sig := <-stopCh:
		slog.Info("收到退出信号", "signal", sig.String())
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	return server.Shutdown(shutdownCtx)
}

func newLogger(level string) *slog.Logger {
	var parsed slog.Level
	switch level {
	case "debug":
		parsed = slog.LevelDebug
	case "warn":
		parsed = slog.LevelWarn
	case "error":
		parsed = slog.LevelError
	default:
		parsed = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parsed}))
}
