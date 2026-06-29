package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"subscription-aggregator/internal/domain"
	"subscription-aggregator/internal/service"
)

var (
	errInvalidUUID = errors.New("некорректный UUID")
	errInvalidBody = errors.New("некорректное тело запроса")
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, err error) {
	status, message := mapError(err)
	writeJSON(w, status, errorResponse{Error: message})
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, service.ErrInvalidStrategy),
		errors.Is(err, errInvalidUUID),
		errors.Is(err, errInvalidBody),
		errors.Is(err, domain.ErrInvalidMonthYear),
		errors.Is(err, domain.ErrInvalidPeriod),
		errors.Is(err, domain.ErrEmptyServiceName),
		errors.Is(err, domain.ErrInvalidPrice),
		errors.Is(err, domain.ErrInvalidEndDate),
		errors.Is(err, domain.ErrEmptyPatch):
		return http.StatusBadRequest, err.Error()
	default:
		return http.StatusInternalServerError, "внутренняя ошибка сервера"
	}
}
