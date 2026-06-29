package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"subscription-aggregator/internal/domain"
	"subscription-aggregator/internal/repository"
)

func TestRepository_createAndGet(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupRepository(t)

	sub := newSubscription(t, "Yandex Plus", 400)

	if err := repo.Create(ctx, sub); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByID(ctx, sub.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	assertSubscriptionEqual(t, sub, got)
}

func TestRepository_getNotFound(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupRepository(t)

	_, err := repo.GetByID(ctx, uuid.New())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("ожидали ErrNotFound, получили %v", err)
	}
}

func TestRepository_update(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupRepository(t)

	sub := newSubscription(t, "Yandex Plus", 400)
	if err := repo.Create(ctx, sub); err != nil {
		t.Fatalf("create: %v", err)
	}

	sub.Price = 500
	sub.ServiceName = "Kinopoisk"

	if err := repo.Update(ctx, sub); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := repo.GetByID(ctx, sub.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Price != 500 || got.ServiceName != "Kinopoisk" {
		t.Fatalf("получили %+v, ожидали обновлённые поля", got)
	}
}

func TestRepository_updateNotFound(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupRepository(t)

	sub := newSubscription(t, "Yandex Plus", 400)
	sub.ID = uuid.New()

	err := repo.Update(ctx, sub)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("ожидали ErrNotFound, получили %v", err)
	}
}

func TestRepository_delete(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupRepository(t)

	sub := newSubscription(t, "Yandex Plus", 400)
	if err := repo.Create(ctx, sub); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.Delete(ctx, sub.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := repo.GetByID(ctx, sub.ID)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("ожидали ErrNotFound после удаления, получили %v", err)
	}
}

func TestRepository_deleteNotFound(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupRepository(t)

	err := repo.Delete(ctx, uuid.New())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("ожидали ErrNotFound, получили %v", err)
	}
}

func TestRepository_list(t *testing.T) {
	ctx := context.Background()
	repo, userID := setupRepository(t)

	otherUser := uuid.New()
	sub1 := newSubscriptionForUser(t, "Yandex Plus", 400, userID)
	sub2 := newSubscriptionForUser(t, "Kinopoisk", 300, userID)
	sub3 := newSubscriptionForUser(t, "Spotify", 200, otherUser)

	for _, sub := range []domain.Subscription{sub1, sub2, sub3} {
		if err := repo.Create(ctx, sub); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	list, err := repo.List(ctx, repository.ListFilter{
		UserID: &userID,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("получили %d записей, ожидали 2", len(list))
	}

	page, err := repo.List(ctx, repository.ListFilter{
		UserID: &userID,
		Limit:  1,
		Offset: 1,
	})
	if err != nil {
		t.Fatalf("list page: %v", err)
	}
	if len(page) != 1 {
		t.Fatalf("получили %d записей, ожидали 1", len(page))
	}
}

func TestRepository_listByFilters(t *testing.T) {
	ctx := context.Background()
	repo, userID := setupRepository(t)

	serviceName := "Yandex Plus"
	sub1 := newSubscriptionForUser(t, serviceName, 400, userID)
	sub2 := newSubscriptionForUser(t, "Kinopoisk", 300, userID)
	sub3 := newSubscriptionForUser(t, serviceName, 500, uuid.New())

	for _, sub := range []domain.Subscription{sub1, sub2, sub3} {
		if err := repo.Create(ctx, sub); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	list, err := repo.ListByFilters(ctx, repository.CostFilter{
		UserID:      &userID,
		ServiceName: &serviceName,
	})
	if err != nil {
		t.Fatalf("list by filters: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("получили %d записей, ожидали 1", len(list))
	}
	if list[0].ServiceName != serviceName {
		t.Fatalf("получили %q, ожидали %q", list[0].ServiceName, serviceName)
	}
}

func newSubscription(t *testing.T, serviceName string, price int) domain.Subscription {
	t.Helper()
	return newSubscriptionForUser(t, serviceName, price, uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"))
}

func newSubscriptionForUser(t *testing.T, serviceName string, price int, userID uuid.UUID) domain.Subscription {
	t.Helper()

	end := domain.MonthYear{Month: 12, Year: 2025}
	return domain.Subscription{
		ID:          uuid.New(),
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   domain.MonthYear{Month: 7, Year: 2025},
		EndDate:     &end,
	}
}

func assertSubscriptionEqual(t *testing.T, want, got domain.Subscription) {
	t.Helper()

	if got.ID != want.ID {
		t.Fatalf("ID: получили %v, ожидали %v", got.ID, want.ID)
	}
	if got.ServiceName != want.ServiceName {
		t.Fatalf("ServiceName: получили %q, ожидали %q", got.ServiceName, want.ServiceName)
	}
	if got.Price != want.Price {
		t.Fatalf("Price: получили %d, ожидали %d", got.Price, want.Price)
	}
	if got.UserID != want.UserID {
		t.Fatalf("UserID: получили %v, ожидали %v", got.UserID, want.UserID)
	}
	if got.StartDate != want.StartDate {
		t.Fatalf("StartDate: получили %v, ожидали %v", got.StartDate, want.StartDate)
	}
	if (got.EndDate == nil) != (want.EndDate == nil) {
		t.Fatalf("EndDate nil: получили %v, ожидали %v", got.EndDate, want.EndDate)
	}
	if got.EndDate != nil && want.EndDate != nil && *got.EndDate != *want.EndDate {
		t.Fatalf("EndDate: получили %v, ожидали %v", *got.EndDate, *want.EndDate)
	}
}
