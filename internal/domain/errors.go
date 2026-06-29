package domain

import "errors"

var (
	ErrInvalidMonthYear = errors.New("неверный формат месяца, ожидается MM-YYYY")
	ErrInvalidPeriod    = errors.New("начало периода не может быть позже конца")
	ErrEmptyServiceName = errors.New("название сервиса не может быть пустым")
	ErrInvalidPrice     = errors.New("цена должна быть больше нуля")
	ErrInvalidEndDate   = errors.New("дата окончания не может быть раньше даты начала")
	ErrEmptyPatch       = errors.New("патч должен содержать хотя бы одно поле")
)
