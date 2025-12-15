package main

import (
	"context"
	"log"
	"os"

	"todo/internal/config"
	gen "todo/internal/gen/todo/v1"
	todogrpc "todo/internal/handler/grpc/todo"
	"todo/internal/server"
	todosvc "todo/internal/service/todo"
	"todo/internal/storage"
	todorepo "todo/internal/todo"

	"google.golang.org/grpc"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := server.WaitForSignal(context.Background())

	db, err := storage.OpenDB(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	if err := storage.ApplyMigrations(ctx, db); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	todoRepo := todorepo.NewRepository(db)
	service := todosvc.NewService(todoRepo)
	handler := todogrpc.NewHandler(service)

	if err := server.Run(ctx, server.Config{Addr: cfg.GRPCAddr}, func(s *grpc.Server) {
		gen.RegisterTodoServiceServer(s, handler)
	}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func loadConfig() (config.Config, error) {
	path := env("CONFIG_PATH", "configs/local.yml")
	cfg, err := config.Load(path)
	if err != nil {
		return config.Config{}, err
	}
	if v := os.Getenv("GRPC_ADDR"); v != "" {
		cfg.GRPCAddr = v
	}
	if v := os.Getenv("POSTGRES_DSN"); v != "" {
		cfg.PostgresDSN = v
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
