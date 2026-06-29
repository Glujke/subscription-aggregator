package domain

import "github.com/google/uuid"

// Subscription — запись об онлайн-подписке пользователя.
type Subscription struct {
	ID          uuid.UUID
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   MonthYear
	EndDate     *MonthYear
}

// Validate проверяет корректность полей подписки.
func (s Subscription) Validate() error {
	if s.ServiceName == "" {
		return ErrEmptyServiceName
	}
	if s.Price <= 0 {
		return ErrInvalidPrice
	}
	if s.EndDate != nil && s.EndDate.Before(s.StartDate) {
		return ErrInvalidEndDate
	}
	return nil
}
