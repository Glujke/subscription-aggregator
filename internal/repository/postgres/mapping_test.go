package postgres_test

import (
	"testing"
	"time"

	"subscription-aggregator/internal/domain"
	"subscription-aggregator/internal/repository/postgres"
)

func TestMonthYearToDate(t *testing.T) {
	t.Parallel()

	my := domain.MonthYear{Month: 7, Year: 2025}
	got := postgres.MonthYearToDate(my)
	want := time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC)

	if !got.Equal(want) {
		t.Fatalf("получили %v, ожидали %v", got, want)
	}
}

func TestDateToMonthYear(t *testing.T) {
	t.Parallel()

	d := time.Date(2025, time.July, 15, 12, 0, 0, 0, time.UTC)
	got := postgres.DateToMonthYear(d)
	want := domain.MonthYear{Month: 7, Year: 2025}

	if got != want {
		t.Fatalf("получили %v, ожидали %v", got, want)
	}
}

func TestDateToMonthYear_nilEndDate(t *testing.T) {
	t.Parallel()

	got := postgres.DateToMonthYearPtr(nil)
	if got != nil {
		t.Fatalf("ожидали nil, получили %v", got)
	}
}
