package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"vk_task/internal/config"
	"vk_task/internal/service"
	"vk_task/pkg/subpub"
	"vk_task/proto"

	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	config     *config.Config
	bus        subpub.SubPub
}

func NewServer(cfg *config.Config) *Server {
	bus := subpub.NewSubPub()
	service := service.NewPubSubService(bus)

	grpcServer := grpc.NewServer()
	proto.RegisterPubSubServer(grpcServer, service)

	return &Server{
		grpcServer: grpcServer,
		config:     cfg,
		bus:        bus,
	}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GRPC.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
		}
	}()

	fmt.Printf("Server started on port %d\n", s.config.GRPC.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.GRPC.Timeout)
	defer cancel()

	s.grpcServer.GracefulStop()
	s.bus.Close(ctx)

	return nil
}
