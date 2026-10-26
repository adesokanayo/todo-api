package todo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("todo not found")

type Todo struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS todos (
			id         BIGSERIAL PRIMARY KEY,
			title      TEXT NOT NULL CHECK (length(btrim(title)) > 0),
			done       BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return fmt.Errorf("migrate todos: %w", err)
	}
	return nil
}

func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Store) List(ctx context.Context) ([]Todo, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, done, created_at
		FROM todos
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

func (s *Store) Create(ctx context.Context, title string) (Todo, error) {
	var t Todo
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO todos (title)
		VALUES ($1)
		RETURNING id, title, done, created_at`, strings.TrimSpace(title)).
		Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if err != nil {
		return Todo{}, err
	}
	return t, nil
}

func (s *Store) Get(ctx context.Context, id int64) (Todo, error) {
	var t Todo
	err := s.db.QueryRowContext(ctx, `
		SELECT id, title, done, created_at
		FROM todos
		WHERE id = $1`, id).
		Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Todo{}, ErrNotFound
	}
	if err != nil {
		return Todo{}, err
	}
	return t, nil
}

func (s *Store) Update(ctx context.Context, id int64, title *string, done *bool) (Todo, error) {
	current, err := s.Get(ctx, id)
	if err != nil {
		return Todo{}, err
	}
	if title != nil {
		current.Title = strings.TrimSpace(*title)
	}
	if done != nil {
		current.Done = *done
	}

	var t Todo
	err = s.db.QueryRowContext(ctx, `
		UPDATE todos
		SET title = $2, done = $3
		WHERE id = $1
		RETURNING id, title, done, created_at`, id, current.Title, current.Done).
		Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Todo{}, ErrNotFound
	}
	if err != nil {
		return Todo{}, err
	}
	return t, nil
}

func (s *Store) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
