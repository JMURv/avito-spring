package main

import (
	"context"
	"github.com/JMURv/avito-spring/internal/auth"
	"github.com/JMURv/avito-spring/internal/config"
	"github.com/JMURv/avito-spring/internal/ctrl"
	"github.com/JMURv/avito-spring/internal/hdl/grpc"
	"github.com/JMURv/avito-spring/internal/hdl/http"
	"github.com/JMURv/avito-spring/internal/observability/metrics/prometheus"
	"github.com/JMURv/avito-spring/internal/repo/db"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

const configPath = "configs/config.yaml"

func mustRegisterLogger(mode string) {
	switch mode {
	case "prod":
		zap.ReplaceGlobals(zap.Must(zap.NewProduction()))
	case "dev":
		zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Panic("panic occurred", zap.Any("error", err))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf := config.MustLoad(configPath)
	mustRegisterLogger(conf.Mode)

	au := auth.New(conf)
	repo := db.New(conf)
	svc := ctrl.New(repo, au)
	hdl := http.New(svc, au)
	ghdl := grpc.New(conf.ServiceName, svc)

	go prometheus.New(conf.Prometheus.Port).Start(ctx)
	go hdl.Start(conf.Server.Port)
	go ghdl.Start(conf.Server.GRPCPort)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-c

	zap.L().Info("Shutting down gracefully...")
	if err := hdl.Close(ctx); err != nil {
		zap.L().Warn("Error closing handler", zap.Error(err))
	}

	if err := ghdl.Close(ctx); err != nil {
		zap.L().Warn("Error closing handler", zap.Error(err))
	}

	if err := repo.Close(); err != nil {
		zap.L().Warn("Error closing repository", zap.Error(err))
	}
}
