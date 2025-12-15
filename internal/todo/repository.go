package todo

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrNotFound = errors.New("todo not found")

type Repository struct {
	db *sql.DB
}

type Record struct {
	ID          string
	Title       string
	Description string
	Completed   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, title, description string) (Record, error) {
	now := time.Now().UTC()
	query := `
insert into todos (title, description, completed, created_at, updated_at)
values ($1, $2, false, $3, $3)
returning id, title, description, completed, created_at, updated_at`

	var rec Record
	if err := r.db.QueryRowContext(ctx, query, title, description, now).Scan(
		&rec.ID, &rec.Title, &rec.Description, &rec.Completed, &rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		return Record{}, err
	}
	return rec, nil
}

func (r *Repository) Get(ctx context.Context, id string) (Record, error) {
	query := `
select id, title, description, completed, created_at, updated_at
from todos
where id = $1`

	var rec Record
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rec.ID, &rec.Title, &rec.Description, &rec.Completed, &rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Record{}, ErrNotFound
		}
		return Record{}, err
	}
	return rec, nil
}

func (r *Repository) List(ctx context.Context) ([]Record, error) {
	query := `
select id, title, description, completed, created_at, updated_at
from todos
order by created_at desc`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Record
	for rows.Next() {
		var rec Record
		if err := rows.Scan(
			&rec.ID, &rec.Title, &rec.Description, &rec.Completed, &rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) Update(ctx context.Context, id, title, description string, completed bool) (Record, error) {
	now := time.Now().UTC()
	query := `
update todos
set title = $2, description = $3, completed = $4, updated_at = $5
where id = $1
returning id, title, description, completed, created_at, updated_at`

	var rec Record
	if err := r.db.QueryRowContext(ctx, query, id, title, description, completed, now).Scan(
		&rec.ID, &rec.Title, &rec.Description, &rec.Completed, &rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Record{}, ErrNotFound
		}
		return Record{}, err
	}
	return rec, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `delete from todos where id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
