package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"subscription-aggregator/internal/domain"
	"subscription-aggregator/internal/repository"
	"subscription-aggregator/internal/service"
)

func TestService_Create(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}
	svc := service.New(repo, fixedNow())

	input := domain.CreateSubscriptionInput{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   domain.MonthYear{Month: 7, Year: 2025},
	}

	got, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if got.ID == uuid.Nil {
		t.Fatal("идентификатор должен быть задан")
	}
	if len(repo.items) != 1 {
		t.Fatalf("в репозитории %d записей, ожидали 1", len(repo.items))
	}
}

func TestService_GetByID_notFound(t *testing.T) {
	t.Parallel()

	svc := service.New(&fakeRepository{}, fixedNow())

	_, err := svc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("ожидали ErrNotFound, получили %v", err)
	}
}

func TestService_Update(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}
	svc := service.New(repo, fixedNow())

	created, err := svc.Create(context.Background(), domain.CreateSubscriptionInput{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.New(),
		StartDate:   domain.MonthYear{Month: 7, Year: 2025},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	newPrice := 500
	updated, err := svc.Update(context.Background(), created.ID, domain.UpdateSubscriptionPatch{
		Price: &newPrice,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Price != 500 {
		t.Fatalf("получили %d, ожидали 500", updated.Price)
	}
}

func TestService_Delete(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}
	svc := service.New(repo, fixedNow())

	created, err := svc.Create(context.Background(), domain.CreateSubscriptionInput{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.New(),
		StartDate:   domain.MonthYear{Month: 7, Year: 2025},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(repo.items) != 0 {
		t.Fatalf("в репозитории %d записей, ожидали 0", len(repo.items))
	}
}

func TestService_List_normalizesPagination(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}
	svc := service.New(repo, fixedNow())

	_, err := svc.List(context.Background(), nil, 0, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if repo.lastList.Limit != 20 {
		t.Fatalf("limit: получили %d, ожидали 20", repo.lastList.Limit)
	}

	_, err = svc.List(context.Background(), nil, 500, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if repo.lastList.Limit != 100 {
		t.Fatalf("limit: получили %d, ожидали 100", repo.lastList.Limit)
	}
}

func TestService_CalculateCost_overlapDefault(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
	repo := &fakeRepository{
		filtered: []domain.Subscription{{
			Price:     400,
			UserID:    userID,
			StartDate: domain.MonthYear{Month: 7, Year: 2025},
		}},
	}
	svc := service.New(repo, fixedNow())

	got, err := svc.CalculateCost(context.Background(), service.CostRequest{
		From: domain.MonthYear{Month: 1, Year: 2025},
	})
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if got.TotalCost != 4800 {
		t.Fatalf("TotalCost: получили %d, ожидали 4800", got.TotalCost)
	}
}

func TestService_CalculateCost_sumStrategy(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{
		filtered: []domain.Subscription{{
			Price:     400,
			StartDate: domain.MonthYear{Month: 1, Year: 2025},
			EndDate:   &domain.MonthYear{Month: 12, Year: 2025},
		}},
	}
	svc := service.New(repo, fixedNow())

	got, err := svc.CalculateCost(context.Background(), service.CostRequest{
		From:     domain.MonthYear{Month: 1, Year: 2025},
		To:       ptrMonth(12, 2025),
		Strategy: service.CostStrategySum,
	})
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if got.TotalCost != 400 {
		t.Fatalf("TotalCost: получили %d, ожидали 400", got.TotalCost)
	}
}

func TestService_CalculateCost_invalidPeriod(t *testing.T) {
	t.Parallel()

	svc := service.New(&fakeRepository{}, fixedNow())

	_, err := svc.CalculateCost(context.Background(), service.CostRequest{
		From: domain.MonthYear{Month: 8, Year: 2025},
		To:   ptrMonth(7, 2025),
	})
	if !errors.Is(err, domain.ErrInvalidPeriod) {
		t.Fatalf("ожидали ErrInvalidPeriod, получили %v", err)
	}
}

func fixedNow() service.NowFunc {
	return func() domain.MonthYear {
		return domain.MonthYear{Month: 6, Year: 2026}
	}
}

type fakeRepository struct {
	items    map[uuid.UUID]domain.Subscription
	lastList repository.ListFilter
	filtered []domain.Subscription
}

func (f *fakeRepository) Create(_ context.Context, sub domain.Subscription) error {
	if f.items == nil {
		f.items = make(map[uuid.UUID]domain.Subscription)
	}
	f.items[sub.ID] = sub
	return nil
}

func (f *fakeRepository) GetByID(_ context.Context, id uuid.UUID) (domain.Subscription, error) {
	sub, ok := f.items[id]
	if !ok {
		return domain.Subscription{}, repository.ErrNotFound
	}
	return sub, nil
}

func (f *fakeRepository) Update(_ context.Context, sub domain.Subscription) error {
	if _, ok := f.items[sub.ID]; !ok {
		return repository.ErrNotFound
	}
	f.items[sub.ID] = sub
	return nil
}

func (f *fakeRepository) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := f.items[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.items, id)
	return nil
}

func (f *fakeRepository) List(_ context.Context, filter repository.ListFilter) ([]domain.Subscription, error) {
	f.lastList = filter
	return nil, nil
}

func (f *fakeRepository) ListByFilters(_ context.Context, _ repository.CostFilter) ([]domain.Subscription, error) {
	return f.filtered, nil
}
