package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// MonthYear представляет месяц и год в формате MM-YYYY.
type MonthYear struct {
	Month int
	Year  int
}

// ParseMonthYear разбирает строку формата MM-YYYY.
func ParseMonthYear(value string) (MonthYear, error) {
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return MonthYear{}, ErrInvalidMonthYear
	}

	if len(parts[0]) != 2 {
		return MonthYear{}, ErrInvalidMonthYear
	}

	month, err := strconv.Atoi(parts[0])
	if err != nil || month < 1 || month > 12 {
		return MonthYear{}, ErrInvalidMonthYear
	}

	year, err := strconv.Atoi(parts[1])
	if err != nil || year < 1 {
		return MonthYear{}, ErrInvalidMonthYear
	}

	return MonthYear{Month: month, Year: year}, nil
}

func (m MonthYear) String() string {
	return fmt.Sprintf("%02d-%04d", m.Month, m.Year)
}

func (m MonthYear) Before(other MonthYear) bool {
	return m.Compare(other) < 0
}

func (m MonthYear) After(other MonthYear) bool {
	return m.Compare(other) > 0
}

func (m MonthYear) Equal(other MonthYear) bool {
	return m.Compare(other) == 0
}

func (m MonthYear) Compare(other MonthYear) int {
	if m.Year != other.Year {
		return m.Year - other.Year
	}
	return m.Month - other.Month
}

func (m MonthYear) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.String() + `"`), nil
}

func (m *MonthYear) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return ErrInvalidMonthYear
	}

	parsed, err := ParseMonthYear(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}

	*m = parsed
	return nil
}

// MonthYearFromTime возвращает месяц и год момента времени в таймзоне Europe/Moscow.
func MonthYearFromTime(t time.Time) MonthYear {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		loc = time.UTC
	}

	local := t.In(loc)
	return MonthYear{Month: int(local.Month()), Year: local.Year()}
}
