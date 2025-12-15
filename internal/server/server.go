// Package server поднимает gRPC-сервер и обрабатывает его жизненный цикл.
package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Config описывает параметры запуска gRPC-сервера.
type Config struct {
	Addr string
}

// RegisterFunc регистрирует gRPC-сервисы на сервере.
type RegisterFunc func(grpcServer *grpc.Server)

// Run создаёт и запускает gRPC-сервер с указанной регистрацией сервисов.
func Run(ctx context.Context, cfg Config, register RegisterFunc) error {
	if register == nil {
		return fmt.Errorf("register function is required")
	}

	srv := grpc.NewServer()
	register(srv)
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, healthSrv)
	reflection.Register(srv)

	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.Addr, err)
	}

	log.Printf("gRPC server listening on %s", cfg.Addr)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		stopServer(shutdownCtx, srv)
	}()

	if err := srv.Serve(lis); err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}

func stopServer(ctx context.Context, srv *grpc.Server) {
	ch := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(ch)
	}()
	select {
	case <-ctx.Done():
		srv.Stop()
	case <-ch:
	}
}

// WaitForSignal создаёт контекст, отменяемый по SIGINT/SIGTERM.
func WaitForSignal(parent context.Context) context.Context {
	ctx, cancel := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		cancel()
	}()
	return ctx
}
