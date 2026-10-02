package tokenrepo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalidRefresh = errors.New("refresh token is missing or expired")

type TokenRepoPg struct{ db *pgxpool.Pool }

// New also creates the token table for databases initialized before token storage existed.
func New(ctx context.Context, db *pgxpool.Pool) (*TokenRepoPg, error) {
	_, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS jwt_tokens (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    access_token TEXT NOT NULL UNIQUE,
    refresh_token TEXT NOT NULL UNIQUE,
    access_expires_at TIMESTAMPTZ NOT NULL,
    refresh_expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`)
	if err != nil {
		return nil, err
	}
	return &TokenRepoPg{db: db}, nil
}

func (r *TokenRepoPg) Save(ctx context.Context, email, access, refresh string, accessExpiry, refreshExpiry time.Time) error {
	result, err := r.db.Exec(ctx, `
        INSERT INTO jwt_tokens(user_id, access_token, refresh_token, access_expires_at, refresh_expires_at)
        SELECT id, $2, $3, $4, $5 FROM users WHERE email = $1
    `, email, access, refresh, accessExpiry, refreshExpiry)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return errors.New("user not found")
	}
	return nil
}

// Rotate atomically consumes the old refresh token and stores the new pair.
func (r *TokenRepoPg) Rotate(ctx context.Context, email, oldRefresh, access, refresh string, accessExpiry, refreshExpiry time.Time) error {
	result, err := r.db.Exec(ctx, `
        UPDATE jwt_tokens SET access_token = $3, refresh_token = $4,
            access_expires_at = $5, refresh_expires_at = $6, created_at = NOW()
        WHERE refresh_token = $2 AND refresh_expires_at > NOW()
            AND user_id = (SELECT id FROM users WHERE email = $1)
    `, email, oldRefresh, access, refresh, accessExpiry, refreshExpiry)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrInvalidRefresh
	}
	return nil
}
