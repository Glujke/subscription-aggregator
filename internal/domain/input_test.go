package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestCreateSubscriptionInput_Validate(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")

	tests := []struct {
		name    string
		input   CreateSubscriptionInput
		wantErr bool
	}{
		{
			name: "корректные данные",
			input: CreateSubscriptionInput{
				ServiceName: "Yandex Plus",
				Price:       400,
				UserID:      userID,
				StartDate:   MonthYear{Month: 7, Year: 2025},
			},
		},
		{
			name: "некорректная цена",
			input: CreateSubscriptionInput{
				ServiceName: "Yandex Plus",
				Price:       0,
				UserID:      userID,
				StartDate:   MonthYear{Month: 7, Year: 2025},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.input.Validate()
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

func TestCreateSubscriptionInput_ToSubscription(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
	input := CreateSubscriptionInput{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      userID,
		StartDate:   MonthYear{Month: 7, Year: 2025},
	}

	sub, err := input.ToSubscription()
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if sub.ID == uuid.Nil {
		t.Fatal("идентификатор должен быть сгенерирован")
	}
	if sub.ServiceName != input.ServiceName {
		t.Fatalf("получили %q, ожидали %q", sub.ServiceName, input.ServiceName)
	}
}

func TestUpdateSubscriptionPatch_Apply(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
	original := Subscription{
		ID:          uuid.New(),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      userID,
		StartDate:   MonthYear{Month: 7, Year: 2025},
	}

	newPrice := 500
	patch := UpdateSubscriptionPatch{Price: &newPrice}

	updated, err := patch.Apply(original)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if updated.Price != 500 {
		t.Fatalf("получили %d, ожидали 500", updated.Price)
	}
	if updated.ServiceName != original.ServiceName {
		t.Fatal("неизменённые поля должны сохраниться")
	}
}

func TestUpdateSubscriptionPatch_Apply_empty(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
	original := Subscription{
		ID:          uuid.New(),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      userID,
		StartDate:   MonthYear{Month: 7, Year: 2025},
	}

	_, err := UpdateSubscriptionPatch{}.Apply(original)
	if err == nil {
		t.Fatal("ожидалась ошибка при пустом патче")
	}
}

func TestUpdateSubscriptionPatch_Apply_invalidEndDate(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
	original := Subscription{
		ID:          uuid.New(),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      userID,
		StartDate:   MonthYear{Month: 7, Year: 2025},
	}

	end := MonthYear{Month: 6, Year: 2025}
	patch := UpdateSubscriptionPatch{EndDate: &end}

	_, err := patch.Apply(original)
	if err == nil {
		t.Fatal("ожидалась ошибка при end_date раньше start_date")
	}
}
