//go:build integration

package integration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	gen "todo/internal/gen/todo/v1"
	"todo/internal/server"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestTodoCRUD(t *testing.T) {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:55432/todos?sslmode=disable"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := composeUp(ctx); err != nil {
		t.Skipf("skip integration test: %v", err)
	}
	defer composeDown(context.Background())

	if err := waitForDB(ctx, dsn); err != nil {
		t.Fatalf("database not ready: %v", err)
	}

	addr := randomAddr()
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- server.Run(ctx, server.Config{
			Addr:        addr,
			DatabaseURL: dsn,
		})
	}()

	if err := waitForServer(ctx, addr); err != nil {
		t.Fatalf("server did not start: %v", err)
	}

	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		t.Fatalf("dial server: %v", err)
	}
	defer conn.Close()

	client := gen.NewTodoServiceClient(conn)

	createCtx, cancelCreate := context.WithTimeout(ctx, 5*time.Second)
	defer cancelCreate()
	created, err := client.CreateTodo(createCtx, &gen.CreateTodoRequest{
		Title:       "Write integration test",
		Description: "exercise CRUD path",
	})
	if err != nil {
		t.Fatalf("create todo: %v", err)
	}

	todoID := created.GetTodo().GetId()

	getCtx, cancelGet := context.WithTimeout(ctx, 5*time.Second)
	defer cancelGet()
	fetched, err := client.GetTodo(getCtx, &gen.GetTodoRequest{Id: todoID})
	if err != nil {
		t.Fatalf("get todo: %v", err)
	}
	if fetched.GetTitle() != "Write integration test" {
		t.Fatalf("unexpected title: %s", fetched.GetTitle())
	}

	updateCtx, cancelUpdate := context.WithTimeout(ctx, 5*time.Second)
	defer cancelUpdate()
	updated, err := client.UpdateTodo(updateCtx, &gen.UpdateTodoRequest{
		Id:          todoID,
		Title:       "Write integration tests",
		Description: "exercise CRUD path",
		Completed:   true,
	})
	if err != nil {
		t.Fatalf("update todo: %v", err)
	}
	if !updated.GetCompleted() {
		t.Fatalf("expected todo to be completed")
	}

	listCtx, cancelList := context.WithTimeout(ctx, 5*time.Second)
	defer cancelList()
	listResp, err := client.ListTodos(listCtx, &gen.ListTodosRequest{})
	if err != nil {
		t.Fatalf("list todos: %v", err)
	}
	if len(listResp.GetTodos()) == 0 {
		t.Fatalf("expected todos in list")
	}

	deleteCtx, cancelDelete := context.WithTimeout(ctx, 5*time.Second)
	defer cancelDelete()
	if _, err := client.DeleteTodo(deleteCtx, &gen.DeleteTodoRequest{Id: todoID}); err != nil {
		t.Fatalf("delete todo: %v", err)
	}

	cancel()
	if err := <-srvErr; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("server run error: %v", err)
	}
}

func composeUp(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "up", "-d", "db")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func composeDown(ctx context.Context) {
	cmd := exec.CommandContext(ctx, "docker", "compose", "down", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func waitForDB(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	for i := 0; i < 30; i++ {
		if err := db.PingContext(ctx); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return errors.New("database not reachable")
}

func waitForServer(ctx context.Context, addr string) error {
	dialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	_, err := grpc.DialContext(dialCtx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	return err
}

func randomAddr() string {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "127.0.0.1:50051"
	}
	addr := lis.Addr().String()
	_ = lis.Close()
	return addr
}
