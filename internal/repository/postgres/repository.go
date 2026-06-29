package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"subscription-aggregator/internal/domain"
	"subscription-aggregator/internal/repository"
)

var _ repository.Repository = (*Repository)(nil)

// Repository — PostgreSQL-реализация хранилища подписок.
type Repository struct {
	pool *pgxpool.Pool
}

// New создаёт репозиторий поверх пула соединений.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, sub domain.Subscription) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		MonthYearToDate(sub.StartDate),
		monthYearToDatePtr(sub.EndDate),
	)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions
		WHERE id = $1`, id)

	sub, err := scanSubscription(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Subscription{}, repository.ErrNotFound
	}
	return sub, err
}

func (r *Repository) Update(ctx context.Context, sub domain.Subscription) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE subscriptions
		SET service_name = $2,
		    price = $3,
		    user_id = $4,
		    start_date = $5,
		    end_date = $6,
		    updated_at = now()
		WHERE id = $1`,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		MonthYearToDate(sub.StartDate),
		monthYearToDatePtr(sub.EndDate),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *Repository) List(ctx context.Context, filter repository.ListFilter) ([]domain.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions`
	args := make([]any, 0, 3)
	conditions := make([]string, 0, 1)

	if filter.UserID != nil {
		args = append(args, *filter.UserID)
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", len(args)))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY start_date DESC, id"

	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	if filter.Offset > 0 {
		args = append(args, filter.Offset)
		query += fmt.Sprintf(" OFFSET $%d", len(args))
	}

	return r.querySubscriptions(ctx, query, args...)
}

func (r *Repository) ListByFilters(ctx context.Context, filter repository.CostFilter) ([]domain.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions`
	args := make([]any, 0, 2)
	conditions := make([]string, 0, 2)

	if filter.UserID != nil {
		args = append(args, *filter.UserID)
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", len(args)))
	}
	if filter.ServiceName != nil {
		args = append(args, *filter.ServiceName)
		conditions = append(conditions, fmt.Sprintf("service_name = $%d", len(args)))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY start_date DESC, id"

	return r.querySubscriptions(ctx, query, args...)
}

func (r *Repository) querySubscriptions(ctx context.Context, query string, args ...any) ([]domain.Subscription, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subs := make([]domain.Subscription, 0)
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}

	return subs, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanSubscription(row scannable) (domain.Subscription, error) {
	var sub domain.Subscription
	var startDate time.Time
	var endDate *time.Time

	err := row.Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&startDate,
		&endDate,
	)
	if err != nil {
		return domain.Subscription{}, err
	}

	sub.StartDate = DateToMonthYear(startDate)
	sub.EndDate = DateToMonthYearPtr(endDate)

	return sub, nil
}

func monthYearToDatePtr(my *domain.MonthYear) *time.Time {
	if my == nil {
		return nil
	}
	d := MonthYearToDate(*my)
	return &d
}
