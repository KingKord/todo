package server

import (
	"context"
	"errors"

	gen "todo/internal/gen/todo/v1"
	todorepo "todo/internal/todo"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TodoService struct {
	gen.UnimplementedTodoServiceServer
	repo *todorepo.Repository
}

func NewTodoService(repo *todorepo.Repository) *TodoService {
	return &TodoService{repo: repo}
}

func (s *TodoService) CreateTodo(ctx context.Context, req *gen.CreateTodoRequest) (*gen.CreateTodoResponse, error) {
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	rec, err := s.repo.Create(ctx, req.GetTitle(), req.GetDescription())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create todo: %v", err)
	}

	return &gen.CreateTodoResponse{Todo: recordToProto(rec)}, nil
}

func (s *TodoService) GetTodo(ctx context.Context, req *gen.GetTodoRequest) (*gen.Todo, error) {
	rec, err := s.repo.Get(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, todorepo.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "todo not found")
		}
		return nil, status.Errorf(codes.Internal, "get todo: %v", err)
	}
	return recordToProto(rec), nil
}

func (s *TodoService) ListTodos(ctx context.Context, _ *gen.ListTodosRequest) (*gen.ListTodosResponse, error) {
	recs, err := s.repo.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list todos: %v", err)
	}
	out := make([]*gen.Todo, 0, len(recs))
	for _, rec := range recs {
		rec := rec
		out = append(out, recordToProto(rec))
	}
	return &gen.ListTodosResponse{Todos: out}, nil
}

func (s *TodoService) UpdateTodo(ctx context.Context, req *gen.UpdateTodoRequest) (*gen.Todo, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	rec, err := s.repo.Update(ctx, req.GetId(), req.GetTitle(), req.GetDescription(), req.GetCompleted())
	if err != nil {
		if errors.Is(err, todorepo.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "todo not found")
		}
		return nil, status.Errorf(codes.Internal, "update todo: %v", err)
	}
	return recordToProto(rec), nil
}

func (s *TodoService) DeleteTodo(ctx context.Context, req *gen.DeleteTodoRequest) (*gen.DeleteTodoResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if err := s.repo.Delete(ctx, req.GetId()); err != nil {
		if errors.Is(err, todorepo.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "todo not found")
		}
		return nil, status.Errorf(codes.Internal, "delete todo: %v", err)
	}
	return &gen.DeleteTodoResponse{}, nil
}

func recordToProto(rec todorepo.Record) *gen.Todo {
	return &gen.Todo{
		Id:          rec.ID,
		Title:       rec.Title,
		Description: rec.Description,
		Completed:   rec.Completed,
		CreatedAt:   rec.CreatedAt.Unix(),
		UpdatedAt:   rec.UpdatedAt.Unix(),
	}
}
