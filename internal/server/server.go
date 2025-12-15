package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"
	"time"

	"todo/internal/gen/todo/v1"
	"todo/internal/storage"
	todorepo "todo/internal/todo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Addr        string
	DatabaseURL string
}

func Run(ctx context.Context, cfg Config) error {
	db, err := openDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := storage.ApplyMigrations(ctx, db); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	repo := todorepo.NewRepository(db)
	grpcSrv := grpc.NewServer()

	service := NewTodoService(repo)
	todo.RegisterTodoServiceServer(grpcSrv, service)

	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcSrv, healthSrv)
	reflection.Register(grpcSrv)

	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.Addr, err)
	}

	log.Printf("gRPC server listening on %s", cfg.Addr)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		stopServer(shutdownCtx, grpcSrv)
	}()

	if err := grpcSrv.Serve(lis); err != nil {
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

func openDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(4)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

func WaitForSignal(parent context.Context) context.Context {
	ctx, cancel := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		cancel()
	}()
	return ctx
}
