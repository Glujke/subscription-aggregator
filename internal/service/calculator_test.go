package service_test

import (
	"testing"

	"subscription-aggregator/internal/domain"
	"subscription-aggregator/internal/service"
)

func TestOverlapCalculator(t *testing.T) {
	t.Parallel()

	now := domain.MonthYear{Month: 6, Year: 2026}
	calc := service.OverlapCalculator{}

	tests := []struct {
		name     string
		subs     []domain.Subscription
		from     domain.MonthYear
		to       *domain.MonthYear
		want     int
		wantErr  bool
		wantSubs int
	}{
		{
			name: "одна подписка без конца",
			subs: []domain.Subscription{{
				ServiceName: "Yandex Plus",
				Price:       400,
				StartDate:   domain.MonthYear{Month: 7, Year: 2025},
			}},
			from:     domain.MonthYear{Month: 1, Year: 2025},
			to:       nil,
			want:     4800,
			wantSubs: 1,
		},
		{
			name: "подписка не пересекается с периодом",
			subs: []domain.Subscription{{
				Price:     400,
				StartDate: domain.MonthYear{Month: 1, Year: 2024},
				EndDate:   &domain.MonthYear{Month: 6, Year: 2024},
			}},
			from:     domain.MonthYear{Month: 1, Year: 2025},
			to:       ptrMonth(12, 2025),
			want:     0,
			wantSubs: 0,
		},
		{
			name: "две подписки overlap",
			subs: []domain.Subscription{
				{
					Price:     400,
					StartDate: domain.MonthYear{Month: 1, Year: 2025},
					EndDate:   &domain.MonthYear{Month: 3, Year: 2025},
				},
				{
					Price:     200,
					StartDate: domain.MonthYear{Month: 2, Year: 2025},
					EndDate:   &domain.MonthYear{Month: 2, Year: 2025},
				},
			},
			from:     domain.MonthYear{Month: 1, Year: 2025},
			to:       ptrMonth(3, 2025),
			want:     1400,
			wantSubs: 2,
		},
		{
			name:     "from позже to",
			subs:     nil,
			from:     domain.MonthYear{Month: 8, Year: 2025},
			to:       ptrMonth(7, 2025),
			wantErr:  true,
			wantSubs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := calc.Calculate(tt.subs, tt.from, tt.to, now)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ожидалась ошибка")
				}
				return
			}
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
			if got.TotalCost != tt.want {
				t.Fatalf("TotalCost: получили %d, ожидали %d", got.TotalCost, tt.want)
			}
			if got.SubscriptionsCount != tt.wantSubs {
				t.Fatalf("SubscriptionsCount: получили %d, ожидали %d", got.SubscriptionsCount, tt.wantSubs)
			}
		})
	}
}

func TestSumCalculator(t *testing.T) {
	t.Parallel()

	now := domain.MonthYear{Month: 6, Year: 2026}
	calc := service.SumCalculator{}

	subs := []domain.Subscription{
		{
			Price:     400,
			StartDate: domain.MonthYear{Month: 1, Year: 2025},
			EndDate:   &domain.MonthYear{Month: 12, Year: 2025},
		},
		{
			Price:     200,
			StartDate: domain.MonthYear{Month: 1, Year: 2026},
			EndDate:   &domain.MonthYear{Month: 1, Year: 2026},
		},
		{
			Price:     999,
			StartDate: domain.MonthYear{Month: 1, Year: 2020},
			EndDate:   &domain.MonthYear{Month: 1, Year: 2021},
		},
	}

	got, err := calc.Calculate(
		subs,
		domain.MonthYear{Month: 1, Year: 2025},
		ptrMonth(12, 2025),
		now,
	)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if got.TotalCost != 400 {
		t.Fatalf("TotalCost: получили %d, ожидали 400", got.TotalCost)
	}
	if got.SubscriptionsCount != 1 {
		t.Fatalf("SubscriptionsCount: получили %d, ожидали 1", got.SubscriptionsCount)
	}
}

func ptrMonth(month, year int) *domain.MonthYear {
	my := domain.MonthYear{Month: month, Year: year}
	return &my
}
