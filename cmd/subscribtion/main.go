package main

import (
	"log/slog"
	"os"

	"vk_task/internal/config"
	"vk_task/internal/lib/logger"
	"vk_task/internal/server"
	"vk_task/internal/service"
	"vk_task/pkg/subpub"
	"vk_task/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.MustLoad()
	log := logger.SetupLogger(cfg.Env)
	bus := subpub.NewSubPub()
	grpcService := service.NewPubSubService(bus)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(logger.LoggingInterceptor(log)),
	)
	proto.RegisterPubSubServer(grpcServer, grpcService)
	reflection.Register(grpcServer)
	srv := server.NewServer(cfg, grpcServer)

	log.Info("starting server", slog.String("env", cfg.Env), slog.Int("port", cfg.GRPC.Port))
	if err := srv.Run(); err != nil {
		log.Error("failed to run server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
