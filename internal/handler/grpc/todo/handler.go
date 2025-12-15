// Package todo содержит gRPC-обработчик для сервиса задач.
package todo

import (
	"context"
	"errors"

	gen "todo/internal/gen/todo/v1"
	todosvc "todo/internal/service/todo"
	todorepo "todo/internal/todo"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler реализует gRPC-методы сервиса задач.
type Handler struct {
	gen.UnimplementedTodoServiceServer
	service *todosvc.Service
}

// NewHandler создаёт gRPC-обработчик задач.
func NewHandler(service *todosvc.Service) *Handler {
	return &Handler{service: service}
}

// CreateTodo создаёт новую задачу.
func (h *Handler) CreateTodo(ctx context.Context, req *gen.CreateTodoRequest) (*gen.CreateTodoResponse, error) {
	rec, err := h.service.Create(ctx, req.GetTitle(), req.GetDescription())
	if err != nil {
		return nil, handleError(err)
	}
	return &gen.CreateTodoResponse{Todo: recordToProto(rec)}, nil
}

// GetTodo возвращает задачу по идентификатору.
func (h *Handler) GetTodo(ctx context.Context, req *gen.GetTodoRequest) (*gen.Todo, error) {
	rec, err := h.service.Get(ctx, req.GetId())
	if err != nil {
		return nil, handleError(err)
	}
	return recordToProto(rec), nil
}

// ListTodos возвращает список задач.
func (h *Handler) ListTodos(ctx context.Context, _ *gen.ListTodosRequest) (*gen.ListTodosResponse, error) {
	recs, err := h.service.List(ctx)
	if err != nil {
		return nil, handleError(err)
	}
	out := make([]*gen.Todo, 0, len(recs))
	for _, rec := range recs {
		rec := rec
		out = append(out, recordToProto(rec))
	}
	return &gen.ListTodosResponse{Todos: out}, nil
}

// UpdateTodo изменяет существующую задачу.
func (h *Handler) UpdateTodo(ctx context.Context, req *gen.UpdateTodoRequest) (*gen.Todo, error) {
	rec, err := h.service.Update(ctx, req.GetId(), req.GetTitle(), req.GetDescription(), req.GetCompleted())
	if err != nil {
		return nil, handleError(err)
	}
	return recordToProto(rec), nil
}

// DeleteTodo удаляет задачу.
func (h *Handler) DeleteTodo(ctx context.Context, req *gen.DeleteTodoRequest) (*gen.DeleteTodoResponse, error) {
	if err := h.service.Delete(ctx, req.GetId()); err != nil {
		return nil, handleError(err)
	}
	return &gen.DeleteTodoResponse{}, nil
}

func handleError(err error) error {
	switch {
	case errors.Is(err, todosvc.ErrValidation):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, todorepo.ErrNotFound):
		return status.Error(codes.NotFound, "todo not found")
	default:
		return status.Errorf(codes.Internal, "internal error: %v", err)
	}
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
