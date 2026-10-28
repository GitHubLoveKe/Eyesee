package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"coordinator/config"
	zc "coordinator"

	// 模块自注册（init 顺序无关，由 enabled_modules 控制实际加载）
	_ "coordinator/module/logs"
	_ "coordinator/module/metrics"
	_ "coordinator/module/network"
	_ "coordinator/module/traces"
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	logLevel := flag.String("log-level", "info", "日志级别 (debug/info/warn/error)")
	flag.Parse()

	var level slog.Level
	switch *logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("加载配置文件失败", "path", *configPath, "error", err)
		os.Exit(1)
	}

	logger.Info("配置加载完成",
		"zone_id", cfg.ZoneID,
		"http_addr", cfg.Server.HTTPAddr,
		"gossip_port", cfg.Gossip.BindPort,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("收到退出信号, 正在关闭...", "signal", sig)
		cancel()
	}()

	coord := zc.New(cfg, logger)
	if err := coord.Run(ctx); err != nil {
		logger.Error("Coordinator 运行错误", "error", err)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Coordinator 已退出")
}
