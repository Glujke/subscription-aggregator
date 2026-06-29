package handler

import (
	"context"

	"github.com/google/uuid"

	"subscription-aggregator/internal/domain"
	"subscription-aggregator/internal/service"
)

// SubscriptionService — контракт бизнес-логики для HTTP-слоя.
type SubscriptionService interface {
	Create(ctx context.Context, input domain.CreateSubscriptionInput) (domain.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.UpdateSubscriptionPatch) (domain.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]domain.Subscription, error)
	CalculateCost(ctx context.Context, req service.CostRequest) (service.CostResult, error)
}
