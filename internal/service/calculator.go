package service

import "subscription-aggregator/internal/domain"

// NowFunc возвращает текущий месяц для бизнес-расчётов.
type NowFunc func() domain.MonthYear

// CostCalculator считает стоимость по набору подписок и периоду.
type CostCalculator interface {
	Calculate(subs []domain.Subscription, from domain.MonthYear, to *domain.MonthYear, now domain.MonthYear) (CostResult, error)
}

// OverlapCalculator считает стоимость как сумму price × месяцы пересечения.
type OverlapCalculator struct{}

func (OverlapCalculator) Calculate(
	subs []domain.Subscription,
	from domain.MonthYear,
	to *domain.MonthYear,
	now domain.MonthYear,
) (CostResult, error) {
	periodEnd := domain.ResolvePeriodEnd(to, now)
	if from.After(periodEnd) {
		return CostResult{}, domain.ErrInvalidPeriod
	}

	total := 0
	count := 0

	for _, sub := range subs {
		if !domain.SubscriptionOverlapsPeriod(sub, from, periodEnd, now) {
			continue
		}

		subEnd := now
		if sub.EndDate != nil {
			subEnd = *sub.EndDate
		}

		months, err := domain.IntersectMonths(sub.StartDate, subEnd, from, periodEnd)
		if err != nil {
			return CostResult{}, err
		}

		total += months * sub.Price
		count++
	}

	return CostResult{TotalCost: total, SubscriptionsCount: count}, nil
}

// SumCalculator считает стоимость как сумму price подписок, пересекающих период.
type SumCalculator struct{}

func (SumCalculator) Calculate(
	subs []domain.Subscription,
	from domain.MonthYear,
	to *domain.MonthYear,
	now domain.MonthYear,
) (CostResult, error) {
	periodEnd := domain.ResolvePeriodEnd(to, now)
	if from.After(periodEnd) {
		return CostResult{}, domain.ErrInvalidPeriod
	}

	total := 0
	count := 0

	for _, sub := range subs {
		if !domain.SubscriptionOverlapsPeriod(sub, from, periodEnd, now) {
			continue
		}

		total += sub.Price
		count++
	}

	return CostResult{TotalCost: total, SubscriptionsCount: count}, nil
}
