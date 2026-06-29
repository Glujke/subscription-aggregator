package domain

import (
	"testing"
	"time"
)

func TestMonthsInclusive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		from    MonthYear
		to      MonthYear
		want    int
		wantErr bool
	}{
		{
			name: "один месяц",
			from: MonthYear{Month: 7, Year: 2025},
			to:   MonthYear{Month: 7, Year: 2025},
			want: 1,
		},
		{
			name: "год",
			from: MonthYear{Month: 1, Year: 2025},
			to:   MonthYear{Month: 12, Year: 2025},
			want: 12,
		},
		{
			name: "через границу года",
			from: MonthYear{Month: 11, Year: 2024},
			to:   MonthYear{Month: 2, Year: 2025},
			want: 4,
		},
		{
			name:    "from позже to",
			from:    MonthYear{Month: 8, Year: 2025},
			to:      MonthYear{Month: 7, Year: 2025},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := MonthsInclusive(tt.from, tt.to)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ожидалась ошибка")
				}
				return
			}
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
			if got != tt.want {
				t.Fatalf("получили %d, ожидали %d", got, tt.want)
			}
		})
	}
}

func TestIntersectMonths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		aFrom, aTo MonthYear
		bFrom, bTo MonthYear
		want       int
	}{
		{
			name:  "полное совпадение",
			aFrom: MonthYear{Month: 1, Year: 2025},
			aTo:   MonthYear{Month: 3, Year: 2025},
			bFrom: MonthYear{Month: 1, Year: 2025},
			bTo:   MonthYear{Month: 3, Year: 2025},
			want:  3,
		},
		{
			name:  "частичное пересечение",
			aFrom: MonthYear{Month: 1, Year: 2025},
			aTo:   MonthYear{Month: 6, Year: 2025},
			bFrom: MonthYear{Month: 4, Year: 2025},
			bTo:   MonthYear{Month: 12, Year: 2025},
			want:  3,
		},
		{
			name:  "без пересечения",
			aFrom: MonthYear{Month: 1, Year: 2025},
			aTo:   MonthYear{Month: 3, Year: 2025},
			bFrom: MonthYear{Month: 5, Year: 2025},
			bTo:   MonthYear{Month: 7, Year: 2025},
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := IntersectMonths(tt.aFrom, tt.aTo, tt.bFrom, tt.bTo)
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
			if got != tt.want {
				t.Fatalf("получили %d, ожидали %d", got, tt.want)
			}
		})
	}
}

func TestMonthYearFromTime(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		t.Fatalf("не удалось загрузить таймзону: %v", err)
	}

	ts := time.Date(2025, 7, 15, 23, 30, 0, 0, loc)
	got := MonthYearFromTime(ts)
	want := MonthYear{Month: 7, Year: 2025}
	if got != want {
		t.Fatalf("получили %v, ожидали %v", got, want)
	}
}

func TestResolvePeriodEnd(t *testing.T) {
	t.Parallel()

	now := MonthYear{Month: 6, Year: 2026}
	explicit := MonthYear{Month: 12, Year: 2025}

	if got := ResolvePeriodEnd(nil, now); got != now {
		t.Fatalf("получили %v, ожидали текущий месяц %v", got, now)
	}
	if got := ResolvePeriodEnd(&explicit, now); got != explicit {
		t.Fatalf("получили %v, ожидали %v", got, explicit)
	}
}

func TestIntersectMonths_openEndSubscription(t *testing.T) {
	t.Parallel()

	now := MonthYear{Month: 6, Year: 2026}
	subStart := MonthYear{Month: 7, Year: 2025}
	periodFrom := MonthYear{Month: 1, Year: 2025}

	got, err := IntersectMonths(subStart, now, periodFrom, now)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if got != 12 {
		t.Fatalf("получили %d месяцев, ожидали 12", got)
	}
}

func TestSubscriptionOverlapsPeriod(t *testing.T) {
	t.Parallel()

	now := MonthYear{Month: 6, Year: 2026}
	end := MonthYear{Month: 12, Year: 2025}

	tests := []struct {
		name string
		sub  Subscription
		from MonthYear
		to   MonthYear
		want bool
	}{
		{
			name: "пересечение с открытым концом подписки",
			sub: Subscription{
				StartDate: MonthYear{Month: 7, Year: 2025},
				EndDate:   nil,
			},
			from: MonthYear{Month: 1, Year: 2025},
			to:   MonthYear{Month: 12, Year: 2025},
			want: true,
		},
		{
			name: "подписка закончилась до периода",
			sub: Subscription{
				StartDate: MonthYear{Month: 1, Year: 2025},
				EndDate:   &end,
			},
			from: MonthYear{Month: 1, Year: 2026},
			to:   MonthYear{Month: 6, Year: 2026},
			want: false,
		},
		{
			name: "закрытая подписка внутри периода",
			sub: Subscription{
				StartDate: MonthYear{Month: 4, Year: 2025},
				EndDate:   &end,
			},
			from: MonthYear{Month: 1, Year: 2025},
			to:   MonthYear{Month: 12, Year: 2025},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := SubscriptionOverlapsPeriod(tt.sub, tt.from, tt.to, now)
			if got != tt.want {
				t.Fatalf("получили %v, ожидали %v", got, tt.want)
			}
		})
	}
}
