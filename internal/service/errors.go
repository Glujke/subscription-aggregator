package service

import "errors"

var (
	ErrNotFound         = errors.New("запись не найдена")
	ErrInvalidStrategy  = errors.New("недопустимая стратегия расчёта")
)
