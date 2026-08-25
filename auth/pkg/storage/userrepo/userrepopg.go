package userrepo

import (
	"auth/internal/models"
	"auth/internal/utils"
	"auth/pkg/logger"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type UserRepoPg struct {
	db *pgxpool.Pool
	ctx context.Context
}

func NewUserRepoPg(ctx context.Context, conn string) (*UserRepoPg, error) {
	log := logger.GetLogger(ctx)
	db, err := pgxpool.New(ctx, conn)
	if err != nil {
		log.Error(ctx, "db connection failed! check your db")
		return &UserRepoPg{}, err
	}
	return &UserRepoPg{ctx: ctx, db: db}, nil
}

func (r *UserRepoPg) AddUser(email, password string) (*models.User, string, error) {
	log := logger.GetLogger(r.ctx)
	tx, err := r.db.Begin(r.ctx)
	if err != nil {
		return nil, "", err
	}
	defer tx.Rollback(r.ctx)

	hashed_password, err := utils.HashPassword(password)
	if err != nil {
		return nil, "", err
	}
	
	var u models.User

	err = tx.QueryRow(r.ctx, `
	INSERT INTO users(email, password_hash)
	VALUES($1, $2)
	RETURNING id, email, password_hash, created_at
	`, email, hashed_password).
	Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)

	if err != nil {
		log.Error(r.ctx, "postgre error", zap.String("error", err.Error()))
		return nil, "", fmt.Errorf("postgre err: %w", err)
	}



	err = tx.Commit(r.ctx)
	if err != nil {
		log.Error(r.ctx, "postgre error", zap.String("Error", err.Error()))
		return nil, "", err
	}

	return &u, "", nil

}

func (r *UserRepoPg) Authenticate(email, password string) bool {
	var u models.User

	err := r.db.QueryRow(r.ctx, `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
	)

	if err != nil {
		return false
	}

	return utils.VerifyPassword(u.PasswordHash, password)
}