package main

import (
	"context"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"todo/internal/server"
)

func main() {
	addr := env("GRPC_ADDR", ":50051")
	dsn := env("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/todos?sslmode=disable")

	ctx := server.WaitForSignal(context.Background())
	if err := server.Run(ctx, server.Config{
		Addr:        addr,
		DatabaseURL: dsn,
	}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
