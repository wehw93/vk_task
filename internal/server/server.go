package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"vk_task/internal/config"

	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	config     *config.Config
}

func NewServer(cfg *config.Config, grpcServer *grpc.Server) *Server {
	return &Server{
		grpcServer: grpcServer,
		config:     cfg,
	}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GRPC.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()

	s.grpcServer.GracefulStop()
	return nil
}
