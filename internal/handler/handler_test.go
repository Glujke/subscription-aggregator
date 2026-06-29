package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"subscription-aggregator/internal/domain"
	"subscription-aggregator/internal/handler"
	"subscription-aggregator/internal/service"
)

func TestCreateSubscription(t *testing.T) {
	t.Parallel()

	svc := &fakeService{
		createResult: sampleSubscription(),
	}
	router := handler.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), svc)

	body := `{
		"service_name": "Yandex Plus",
		"price": 400,
		"user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		"start_date": "07-2025"
	}`

	rr := execute(t, router, http.MethodPost, "/api/v1/subscriptions", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("статус: получили %d, ожидали %d, тело: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp handler.SubscriptionResponse
	decodeJSON(t, rr.Body, &resp)
	if resp.ServiceName != "Yandex Plus" {
		t.Fatalf("получили %q", resp.ServiceName)
	}
}

func TestGetSubscription(t *testing.T) {
	t.Parallel()

	sub := sampleSubscription()
	svc := &fakeService{getResult: sub}
	router := handler.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), svc)

	rr := execute(t, router, http.MethodGet, "/api/v1/subscriptions/"+sub.ID.String(), "")
	if rr.Code != http.StatusOK {
		t.Fatalf("статус: получили %d, тело: %s", rr.Code, rr.Body.String())
	}
}

func TestGetSubscription_notFound(t *testing.T) {
	t.Parallel()

	svc := &fakeService{getErr: service.ErrNotFound}
	router := handler.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), svc)

	rr := execute(t, router, http.MethodGet, "/api/v1/subscriptions/"+uuid.New().String(), "")
	if rr.Code != http.StatusNotFound {
		t.Fatalf("статус: получили %d, ожидали %d", rr.Code, http.StatusNotFound)
	}
}

func TestListSubscriptions(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
	svc := &fakeService{
		listResult: []domain.Subscription{sampleSubscription()},
	}
	router := handler.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), svc)

	rr := execute(t, router, http.MethodGet, "/api/v1/subscriptions?user_id="+userID.String()+"&limit=10&offset=0", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("статус: получили %d, тело: %s", rr.Code, rr.Body.String())
	}
	if svc.lastListUserID == nil || *svc.lastListUserID != userID {
		t.Fatal("user_id не передан в сервис")
	}
	if svc.lastListLimit != 10 {
		t.Fatalf("limit: получили %d, ожидали 10", svc.lastListLimit)
	}
}

func TestUpdateSubscription(t *testing.T) {
	t.Parallel()

	sub := sampleSubscription()
	svc := &fakeService{updateResult: sub}
	router := handler.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), svc)

	body := `{"price": 500}`
	rr := execute(t, router, http.MethodPatch, "/api/v1/subscriptions/"+sub.ID.String(), body)
	if rr.Code != http.StatusOK {
		t.Fatalf("статус: получили %d, тело: %s", rr.Code, rr.Body.String())
	}
}

func TestDeleteSubscription(t *testing.T) {
	t.Parallel()

	sub := sampleSubscription()
	svc := &fakeService{}
	router := handler.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), svc)

	rr := execute(t, router, http.MethodDelete, "/api/v1/subscriptions/"+sub.ID.String(), "")
	if rr.Code != http.StatusNoContent {
		t.Fatalf("статус: получили %d, ожидали %d", rr.Code, http.StatusNoContent)
	}
}

func TestCalculateCost(t *testing.T) {
	t.Parallel()

	svc := &fakeService{
		costResult: service.CostResult{TotalCost: 4800, SubscriptionsCount: 1},
	}
	router := handler.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), svc)

	rr := execute(t, router, http.MethodGet, "/api/v1/subscriptions/cost?from=01-2025&strategy=overlap", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("статус: получили %d, тело: %s", rr.Code, rr.Body.String())
	}

	var resp handler.CostResponse
	decodeJSON(t, rr.Body, &resp)
	if resp.TotalCost != 4800 {
		t.Fatalf("TotalCost: получили %d", resp.TotalCost)
	}
	if svc.lastCostReq.Strategy != service.CostStrategyOverlap {
		t.Fatalf("strategy: получили %q", svc.lastCostReq.Strategy)
	}
}

func TestHealth(t *testing.T) {
	t.Parallel()

	router := handler.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), &fakeService{})

	rr := execute(t, router, http.MethodGet, "/health", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("статус: получили %d", rr.Code)
	}
}

func TestSwaggerUI(t *testing.T) {
	t.Parallel()

	router := handler.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), &fakeService{})

	rr := execute(t, router, http.MethodGet, "/swagger/index.html", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("статус: получили %d", rr.Code)
	}
}

func execute(t *testing.T, router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}

	req := httptest.NewRequest(method, path, reader)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func decodeJSON(t *testing.T, r io.Reader, dest any) {
	t.Helper()

	if err := json.NewDecoder(r).Decode(dest); err != nil {
		t.Fatalf("не удалось разобрать JSON: %v", err)
	}
}

func sampleSubscription() domain.Subscription {
	end := domain.MonthYear{Month: 12, Year: 2025}
	return domain.Subscription{
		ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   domain.MonthYear{Month: 7, Year: 2025},
		EndDate:     &end,
	}
}

type fakeService struct {
	createResult   domain.Subscription
	getResult      domain.Subscription
	updateResult   domain.Subscription
	listResult     []domain.Subscription
	costResult     service.CostResult
	createErr      error
	getErr         error
	updateErr      error
	deleteErr      error
	listErr        error
	costErr        error
	lastListUserID *uuid.UUID
	lastListLimit  int
	lastListOffset int
	lastCostReq    service.CostRequest
}

func (f *fakeService) Create(_ context.Context, _ domain.CreateSubscriptionInput) (domain.Subscription, error) {
	if f.createErr != nil {
		return domain.Subscription{}, f.createErr
	}
	if f.createResult.ID != uuid.Nil {
		return f.createResult, nil
	}
	return sampleSubscription(), nil
}

func (f *fakeService) GetByID(_ context.Context, _ uuid.UUID) (domain.Subscription, error) {
	if f.getErr != nil {
		return domain.Subscription{}, f.getErr
	}
	return f.getResult, nil
}

func (f *fakeService) Update(_ context.Context, _ uuid.UUID, _ domain.UpdateSubscriptionPatch) (domain.Subscription, error) {
	if f.updateErr != nil {
		return domain.Subscription{}, f.updateErr
	}
	return f.updateResult, nil
}

func (f *fakeService) Delete(_ context.Context, _ uuid.UUID) error {
	return f.deleteErr
}

func (f *fakeService) List(_ context.Context, userID *uuid.UUID, limit, offset int) ([]domain.Subscription, error) {
	f.lastListUserID = userID
	f.lastListLimit = limit
	f.lastListOffset = offset
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func (f *fakeService) CalculateCost(_ context.Context, req service.CostRequest) (service.CostResult, error) {
	f.lastCostReq = req
	if f.costErr != nil {
		return service.CostResult{}, f.costErr
	}
	return f.costResult, nil
}
