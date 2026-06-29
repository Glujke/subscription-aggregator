package handler

import (
	"subscription-aggregator/internal/domain"
	"subscription-aggregator/internal/service"

	"github.com/google/uuid"
)

// CreateSubscriptionRequest — тело запроса на создание подписки.
type CreateSubscriptionRequest struct {
	ServiceName string           `json:"service_name"`
	Price       int              `json:"price"`
	UserID      uuid.UUID        `json:"user_id"`
	StartDate   domain.MonthYear `json:"start_date"`
	EndDate     *domain.MonthYear `json:"end_date,omitempty"`
}

func (r CreateSubscriptionRequest) ToInput() domain.CreateSubscriptionInput {
	return domain.CreateSubscriptionInput{
		ServiceName: r.ServiceName,
		Price:       r.Price,
		UserID:      r.UserID,
		StartDate:   r.StartDate,
		EndDate:     r.EndDate,
	}
}

// UpdateSubscriptionRequest — тело PATCH-запроса.
type UpdateSubscriptionRequest struct {
	ServiceName *string           `json:"service_name,omitempty"`
	Price       *int              `json:"price,omitempty"`
	StartDate   *domain.MonthYear `json:"start_date,omitempty"`
	EndDate     *domain.MonthYear `json:"end_date,omitempty"`
}

func (r UpdateSubscriptionRequest) ToPatch() domain.UpdateSubscriptionPatch {
	return domain.UpdateSubscriptionPatch{
		ServiceName: r.ServiceName,
		Price:       r.Price,
		StartDate:   r.StartDate,
		EndDate:     r.EndDate,
	}
}

// SubscriptionResponse — подписка в ответе API.
type SubscriptionResponse struct {
	ID          uuid.UUID         `json:"id"`
	ServiceName string            `json:"service_name"`
	Price       int               `json:"price"`
	UserID      uuid.UUID         `json:"user_id"`
	StartDate   domain.MonthYear  `json:"start_date"`
	EndDate     *domain.MonthYear `json:"end_date,omitempty"`
}

func NewSubscriptionResponse(sub domain.Subscription) SubscriptionResponse {
	return SubscriptionResponse{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   sub.StartDate,
		EndDate:     sub.EndDate,
	}
}

// CostResponse — результат расчёта стоимости.
type CostResponse struct {
	TotalCost          int    `json:"total_cost"`
	SubscriptionsCount int    `json:"subscriptions_count"`
	Currency           string `json:"currency"`
}

func NewCostResponse(result service.CostResult) CostResponse {
	return CostResponse{
		TotalCost:          result.TotalCost,
		SubscriptionsCount: result.SubscriptionsCount,
		Currency:           "RUB",
	}
}
