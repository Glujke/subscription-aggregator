package service

import (
	"github.com/google/uuid"

	"subscription-aggregator/internal/domain"
)

// CostStrategy — способ расчёта суммарной стоимости.
type CostStrategy string

const (
	CostStrategyOverlap CostStrategy = "overlap"
	CostStrategySum     CostStrategy = "sum"
)

// CostRequest — параметры расчёта стоимости подписок.
type CostRequest struct {
	From        domain.MonthYear
	To          *domain.MonthYear
	UserID      *uuid.UUID
	ServiceName *string
	Strategy    CostStrategy
}

// CostResult — результат расчёта стоимости.
type CostResult struct {
	TotalCost          int
	SubscriptionsCount int
}
