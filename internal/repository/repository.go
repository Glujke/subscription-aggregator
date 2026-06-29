package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"subscription-aggregator/internal/domain"
)

var ErrNotFound = errors.New("запись не найдена")

// ListFilter — параметры постраничного списка.
type ListFilter struct {
	UserID *uuid.UUID
	Limit  int
	Offset int
}

// CostFilter — фильтры выборки для расчёта стоимости.
type CostFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
}

// Repository — хранилище подписок.
type Repository interface {
	Create(ctx context.Context, sub domain.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error)
	Update(ctx context.Context, sub domain.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ListFilter) ([]domain.Subscription, error)
	ListByFilters(ctx context.Context, filter CostFilter) ([]domain.Subscription, error)
}
