package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"subscription-aggregator/internal/domain"
	"subscription-aggregator/internal/repository"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// Service — бизнес-логика подписок.
type Service struct {
	repo         repository.Repository
	now          NowFunc
	calculators  map[CostStrategy]CostCalculator
}

// New создаёт сервис подписок.
func New(repo repository.Repository, now NowFunc) *Service {
	return &Service{
		repo: repo,
		now:  now,
		calculators: map[CostStrategy]CostCalculator{
			CostStrategyOverlap: OverlapCalculator{},
			CostStrategySum:     SumCalculator{},
		},
	}
}

func (s *Service) Create(ctx context.Context, input domain.CreateSubscriptionInput) (domain.Subscription, error) {
	sub, err := input.ToSubscription()
	if err != nil {
		return domain.Subscription{}, err
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return domain.Subscription{}, err
	}

	return sub, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	return sub, mapNotFound(err)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, patch domain.UpdateSubscriptionPatch) (domain.Subscription, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Subscription{}, mapNotFound(err)
	}

	updated, err := patch.Apply(current)
	if err != nil {
		return domain.Subscription{}, err
	}

	if err := s.repo.Update(ctx, updated); err != nil {
		return domain.Subscription{}, mapNotFound(err)
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return mapNotFound(s.repo.Delete(ctx, id))
}

func (s *Service) List(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]domain.Subscription, error) {
	return s.repo.List(ctx, repository.ListFilter{
		UserID: userID,
		Limit:  normalizeLimit(limit),
		Offset: offset,
	})
}

func (s *Service) CalculateCost(ctx context.Context, req CostRequest) (CostResult, error) {
	strategy := req.Strategy
	if strategy == "" {
		strategy = CostStrategyOverlap
	}

	calc, ok := s.calculators[strategy]
	if !ok {
		return CostResult{}, ErrInvalidStrategy
	}

	subs, err := s.repo.ListByFilters(ctx, repository.CostFilter{
		UserID:      req.UserID,
		ServiceName: req.ServiceName,
	})
	if err != nil {
		return CostResult{}, err
	}

	return calc.Calculate(subs, req.From, req.To, s.now())
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return defaultListLimit
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

func mapNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
