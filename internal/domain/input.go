package domain

import "github.com/google/uuid"

// CreateSubscriptionInput — данные для создания подписки.
type CreateSubscriptionInput struct {
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   MonthYear
	EndDate     *MonthYear
}

// Validate проверяет поля запроса на создание.
func (in CreateSubscriptionInput) Validate() error {
	sub := Subscription{
		ServiceName: in.ServiceName,
		Price:       in.Price,
		UserID:      in.UserID,
		StartDate:   in.StartDate,
		EndDate:     in.EndDate,
	}
	return sub.Validate()
}

// ToSubscription создаёт подписку с новым идентификатором.
func (in CreateSubscriptionInput) ToSubscription() (Subscription, error) {
	if err := in.Validate(); err != nil {
		return Subscription{}, err
	}

	return Subscription{
		ID:          uuid.New(),
		ServiceName: in.ServiceName,
		Price:       in.Price,
		UserID:      in.UserID,
		StartDate:   in.StartDate,
		EndDate:     in.EndDate,
	}, nil
}

// UpdateSubscriptionPatch — частичное обновление подписки.
type UpdateSubscriptionPatch struct {
	ServiceName *string
	Price       *int
	StartDate   *MonthYear
	EndDate     *MonthYear
}

func (p UpdateSubscriptionPatch) isEmpty() bool {
	return p.ServiceName == nil && p.Price == nil && p.StartDate == nil && p.EndDate == nil
}

// Apply применяет патч к существующей подписке.
func (p UpdateSubscriptionPatch) Apply(sub Subscription) (Subscription, error) {
	if p.isEmpty() {
		return Subscription{}, ErrEmptyPatch
	}

	updated := sub

	if p.ServiceName != nil {
		updated.ServiceName = *p.ServiceName
	}
	if p.Price != nil {
		updated.Price = *p.Price
	}
	if p.StartDate != nil {
		updated.StartDate = *p.StartDate
	}
	if p.EndDate != nil {
		updated.EndDate = p.EndDate
	}

	if err := updated.Validate(); err != nil {
		return Subscription{}, err
	}

	return updated, nil
}
