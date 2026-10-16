package userrepo

import (
	"auth/internal/config"
	"auth/internal/models"
	"auth/internal/utils"
	"auth/pkg/logger"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type UserRepoPg struct {
	db  *pgxpool.Pool
	ctx context.Context
}

func NewUserRepoPg(ctx context.Context, c config.AuthPostgreConfig) (*UserRepoPg, error) {
	log := logger.GetLogger(ctx)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		c.UserName, c.Password, c.Host, c.Port, c.DbName)
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Error(ctx, "db connection failed! check your db")
		return &UserRepoPg{}, err
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return &UserRepoPg{ctx: ctx, db: db}, nil
}

func (r *UserRepoPg) Pool() *pgxpool.Pool { return r.db }

func (r *UserRepoPg) Close() { r.db.Close() }

func (r *UserRepoPg) AddUser(email, password string) (*models.User, string, error) {
	log := logger.GetLogger(r.ctx)
	tx, err := r.db.Begin(r.ctx)
	if err != nil {
		log.Error(r.ctx, "transaction not started", zap.String("error", err.Error()))
		return nil, "", err
	}
	defer tx.Rollback(r.ctx)

	hashed_password, err := utils.HashPassword(password)
	if err != nil {
		log.Error(r.ctx, "error hashing password")
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

	if err = tx.Commit(r.ctx); err != nil {
		log.Error(r.ctx, "postgre error", zap.String("Error", err.Error()))
		return nil, "", err
	}
	log.Info(r.ctx, "user added")
	return &u, "", nil

}

func (r *UserRepoPg) Authenticate(email, password string) bool {
	log := logger.GetLogger(r.ctx)
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
		log.Error(r.ctx, "an error occured authenticating user")
		return false
	}

	return utils.VerifyPassword(u.PasswordHash, password)
}
