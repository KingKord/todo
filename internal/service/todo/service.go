// Package todo содержит бизнес-логику работы с задачами.
package todo

import (
	"context"
	"errors"
	"fmt"

	todorepo "todo/internal/todo"
)

// ErrValidation сигнализирует о нарушениях входных данных.
var ErrValidation = errors.New("validation error")

// Service инкапсулирует операции над задачами на уровне бизнес-логики.
type Service struct {
	repo *todorepo.Repository
}

// NewService создает сервис задач.
func NewService(repo *todorepo.Repository) *Service {
	return &Service{repo: repo}
}

// Create создаёт новую задачу.
func (s *Service) Create(ctx context.Context, title, description string) (todorepo.Record, error) {
	if title == "" {
		return todorepo.Record{}, fmt.Errorf("%w: title is required", ErrValidation)
	}
	return s.repo.Create(ctx, title, description)
}

// Get возвращает задачу по идентификатору.
func (s *Service) Get(ctx context.Context, id string) (todorepo.Record, error) {
	if id == "" {
		return todorepo.Record{}, fmt.Errorf("%w: id is required", ErrValidation)
	}
	return s.repo.Get(ctx, id)
}

// List возвращает все задачи.
func (s *Service) List(ctx context.Context) ([]todorepo.Record, error) {
	return s.repo.List(ctx)
}

// Update обновляет существующую задачу.
func (s *Service) Update(ctx context.Context, id, title, description string, completed bool) (todorepo.Record, error) {
	if id == "" {
		return todorepo.Record{}, fmt.Errorf("%w: id is required", ErrValidation)
	}
	return s.repo.Update(ctx, id, title, description, completed)
}

// Delete удаляет задачу по идентификатору.
func (s *Service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: id is required", ErrValidation)
	}
	return s.repo.Delete(ctx, id)
}
