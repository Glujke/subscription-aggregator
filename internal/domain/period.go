package domain

// ResolvePeriodEnd возвращает конец периода: явный to или текущий месяц.
func ResolvePeriodEnd(to *MonthYear, now MonthYear) MonthYear {
	if to != nil {
		return *to
	}
	return now
}

// MonthsInclusive считает число месяцев в интервале [from, to] включительно.
func MonthsInclusive(from, to MonthYear) (int, error) {
	if from.After(to) {
		return 0, ErrInvalidPeriod
	}

	return (to.Year-from.Year)*12 + (to.Month-from.Month) + 1, nil
}

// IntersectMonths считает число месяцев пересечения двух включительных интервалов.
func IntersectMonths(aFrom, aTo, bFrom, bTo MonthYear) (int, error) {
	start := aFrom
	if bFrom.After(start) {
		start = bFrom
	}

	end := aTo
	if bTo.Before(end) {
		end = bTo
	}

	if start.After(end) {
		return 0, nil
	}

	return MonthsInclusive(start, end)
}

// SubscriptionOverlapsPeriod проверяет пересечение подписки с периодом.
// now — текущий месяц для открытого конца подписки без end_date.
func SubscriptionOverlapsPeriod(sub Subscription, from, to, now MonthYear) bool {
	subEnd := now
	if sub.EndDate != nil {
		subEnd = *sub.EndDate
	}

	months, err := IntersectMonths(sub.StartDate, subEnd, from, to)
	if err != nil {
		return false
	}

	return months > 0
}
