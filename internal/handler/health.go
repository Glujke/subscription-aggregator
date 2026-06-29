package handler

import (
	"net/http"
)

// Health godoc
// @Summary      Проверка живости
// @Description  Возвращает статус работы сервиса
// @Tags         health
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Router       /health [get]
func Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}
