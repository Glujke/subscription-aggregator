package domain

import (
	"encoding/json"
	"testing"
)

func TestParseMonthYear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    MonthYear
		wantErr bool
	}{
		{
			name:  "корректный формат",
			input: "07-2025",
			want:  MonthYear{Month: 7, Year: 2025},
		},
		{
			name:  "январь с ведущим нулём",
			input: "01-2024",
			want:  MonthYear{Month: 1, Year: 2024},
		},
		{
			name:    "месяц без ведущего нуля",
			input:   "7-2025",
			wantErr: true,
		},
		{
			name:    "неверный месяц",
			input:   "13-2025",
			wantErr: true,
		},
		{
			name:    "пустая строка",
			input:   "",
			wantErr: true,
		},
		{
			name:    "неверный разделитель",
			input:   "07/2025",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseMonthYear(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ожидалась ошибка")
				}
				return
			}
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
			if got != tt.want {
				t.Fatalf("получили %v, ожидали %v", got, tt.want)
			}
		})
	}
}

func TestMonthYear_String(t *testing.T) {
	t.Parallel()

	my := MonthYear{Month: 7, Year: 2025}
	if got := my.String(); got != "07-2025" {
		t.Fatalf("получили %q, ожидали %q", got, "07-2025")
	}
}

func TestMonthYear_Compare(t *testing.T) {
	t.Parallel()

	jan2024 := MonthYear{Month: 1, Year: 2024}
	mar2024 := MonthYear{Month: 3, Year: 2024}
	jan2025 := MonthYear{Month: 1, Year: 2025}

	if !jan2024.Before(mar2024) {
		t.Fatal("январь 2024 должен быть раньше марта 2024")
	}
	if mar2024.Before(jan2024) {
		t.Fatal("март 2024 не должен быть раньше января 2024")
	}
	if !mar2024.Before(jan2025) {
		t.Fatal("март 2024 должен быть раньше января 2025")
	}
	if !jan2024.Equal(jan2024) {
		t.Fatal("одинаковые значения должны быть равны")
	}
}

func TestMonthYear_JSON(t *testing.T) {
	t.Parallel()

	t.Run("десериализация", func(t *testing.T) {
		t.Parallel()

		var got MonthYear
		if err := json.Unmarshal([]byte(`"07-2025"`), &got); err != nil {
			t.Fatalf("неожиданная ошибка: %v", err)
		}
		want := MonthYear{Month: 7, Year: 2025}
		if got != want {
			t.Fatalf("получили %v, ожидали %v", got, want)
		}
	})

	t.Run("сериализация", func(t *testing.T) {
		t.Parallel()

		my := MonthYear{Month: 7, Year: 2025}
		data, err := json.Marshal(my)
		if err != nil {
			t.Fatalf("неожиданная ошибка: %v", err)
		}
		if string(data) != `"07-2025"` {
			t.Fatalf("получили %s, ожидали %q", data, `"07-2025"`)
		}
	})
}
