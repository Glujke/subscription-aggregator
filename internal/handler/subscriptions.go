package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"subscription-aggregator/internal/domain"
	"subscription-aggregator/internal/service"
)

// SubscriptionsHandler — HTTP-ручки подписок.
type SubscriptionsHandler struct {
	svc SubscriptionService
}

func NewSubscriptionsHandler(svc SubscriptionService) *SubscriptionsHandler {
	return &SubscriptionsHandler{svc: svc}
}

func (h *SubscriptionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errInvalidBody)
		return
	}

	sub, err := h.svc.Create(r.Context(), req.ToInput())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, NewSubscriptionResponse(sub))
}

func (h *SubscriptionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}

	sub, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, NewSubscriptionResponse(sub))
}

func (h *SubscriptionsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := parseOptionalUUIDQuery(r, "user_id")
	if err != nil {
		writeError(w, err)
		return
	}

	limit, err := parseIntQuery(r, "limit", 0)
	if err != nil {
		writeError(w, err)
		return
	}

	offset, err := parseIntQuery(r, "offset", 0)
	if err != nil {
		writeError(w, err)
		return
	}

	subs, err := h.svc.List(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := make([]SubscriptionResponse, 0, len(subs))
	for _, sub := range subs {
		resp = append(resp, NewSubscriptionResponse(sub))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *SubscriptionsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}

	var req UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errInvalidBody)
		return
	}

	sub, err := h.svc.Update(r.Context(), id, req.ToPatch())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, NewSubscriptionResponse(sub))
}

func (h *SubscriptionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SubscriptionsHandler) Cost(w http.ResponseWriter, r *http.Request) {
	fromRaw := r.URL.Query().Get("from")
	if fromRaw == "" {
		writeError(w, domain.ErrInvalidPeriod)
		return
	}

	from, err := domain.ParseMonthYear(fromRaw)
	if err != nil {
		writeError(w, err)
		return
	}

	var to *domain.MonthYear
	if toRaw := r.URL.Query().Get("to"); toRaw != "" {
		parsed, err := domain.ParseMonthYear(toRaw)
		if err != nil {
			writeError(w, err)
			return
		}
		to = &parsed
	}

	userID, err := parseOptionalUUIDQuery(r, "user_id")
	if err != nil {
		writeError(w, err)
		return
	}

	var serviceName *string
	if name := r.URL.Query().Get("service_name"); name != "" {
		serviceName = &name
	}

	strategy := service.CostStrategy(r.URL.Query().Get("strategy"))

	result, err := h.svc.CalculateCost(r.Context(), service.CostRequest{
		From:        from,
		To:          to,
		UserID:      userID,
		ServiceName: serviceName,
		Strategy:    strategy,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, NewCostResponse(result))
}

func parseUUIDParam(r *http.Request, key string) (uuid.UUID, error) {
	raw := chi.URLParam(r, key)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errInvalidUUID
	}
	return id, nil
}

func parseOptionalUUIDQuery(r *http.Request, key string) (*uuid.UUID, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, errInvalidUUID
	}
	return &id, nil
}

func parseIntQuery(r *http.Request, key string, fallback int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, domain.ErrInvalidPrice
	}
	return value, nil
}
