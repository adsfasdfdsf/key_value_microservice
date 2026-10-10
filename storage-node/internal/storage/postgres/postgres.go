package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/interfaces"
)

type Storage struct{ db *pgxpool.Pool }

func New(ctx context.Context, dsn string) (*Storage, error) {
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	if err = db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Close() { s.db.Close() }

func (s *Storage) Add(ctx context.Context, user, key, value string) error {
	_, err := s.db.Exec(ctx, `INSERT INTO key_values ("user", "key", "value")
 VALUES ($1, $2, $3)
 ON CONFLICT ("user", "key") DO UPDATE SET "value" = EXCLUDED."value"`, user, key, value)
	return err
}

func (s *Storage) Get(ctx context.Context, user, key string) (string, error) {
	var value string
	err := s.db.QueryRow(ctx, `SELECT "value" FROM key_values WHERE "user" = $1 AND "key" = $2`, user, key).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", interfaces.ErrNotFound
	}
	return value, err
}

func (s *Storage) GetAll(ctx context.Context, user string) (map[string]string, error) {
	rows, err := s.db.Query(ctx, `SELECT "key", "value" FROM key_values WHERE "user" = $1`, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	result := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}