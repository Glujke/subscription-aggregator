package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestSubscription_Validate(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
	end := MonthYear{Month: 12, Year: 2025}

	tests := []struct {
		name    string
		sub     Subscription
		wantErr bool
	}{
		{
			name: "корректная подписка",
			sub: Subscription{
				ID:          uuid.New(),
				ServiceName: "Yandex Plus",
				Price:       400,
				UserID:      userID,
				StartDate:   MonthYear{Month: 7, Year: 2025},
			},
		},
		{
			name: "корректная с датой окончания",
			sub: Subscription{
				ID:          uuid.New(),
				ServiceName: "Yandex Plus",
				Price:       400,
				UserID:      userID,
				StartDate:   MonthYear{Month: 7, Year: 2025},
				EndDate:     &end,
			},
		},
		{
			name: "пустое название сервиса",
			sub: Subscription{
				ServiceName: "",
				Price:       400,
				UserID:      userID,
				StartDate:   MonthYear{Month: 7, Year: 2025},
			},
			wantErr: true,
		},
		{
			name: "нулевая цена",
			sub: Subscription{
				ServiceName: "Yandex Plus",
				Price:       0,
				UserID:      userID,
				StartDate:   MonthYear{Month: 7, Year: 2025},
			},
			wantErr: true,
		},
		{
			name: "отрицательная цена",
			sub: Subscription{
				ServiceName: "Yandex Plus",
				Price:       -100,
				UserID:      userID,
				StartDate:   MonthYear{Month: 7, Year: 2025},
			},
			wantErr: true,
		},
		{
			name: "дата окончания раньше начала",
			sub: Subscription{
				ServiceName: "Yandex Plus",
				Price:       400,
				UserID:      userID,
				StartDate:   MonthYear{Month: 7, Year: 2025},
				EndDate:     &MonthYear{Month: 6, Year: 2025},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.sub.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatal("ожидалась ошибка")
				}
				return
			}
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
		})
	}
}
