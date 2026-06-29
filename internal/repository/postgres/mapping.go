package postgres

import (
	"time"

	"subscription-aggregator/internal/domain"
)

// MonthYearToDate преобразует месяц в DATE (первое число).
func MonthYearToDate(my domain.MonthYear) time.Time {
	return time.Date(my.Year, time.Month(my.Month), 1, 0, 0, 0, 0, time.UTC)
}

// DateToMonthYear преобразует DATE в доменный месяц.
func DateToMonthYear(d time.Time) domain.MonthYear {
	d = d.UTC()
	return domain.MonthYear{Month: int(d.Month()), Year: d.Year()}
}

// DateToMonthYearPtr преобразует nullable DATE в доменный месяц.
func DateToMonthYearPtr(d *time.Time) *domain.MonthYear {
	if d == nil {
		return nil
	}
	my := DateToMonthYear(*d)
	return &my
}
