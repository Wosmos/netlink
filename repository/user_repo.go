package repository

import (
	"context"

	"go-to-do/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) InitSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`
	_, err := r.pool.Exec(ctx, schema)
	return err
}

func (r *UserRepository) Create(email, passwordHash string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(
		context.Background(),
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		email, passwordHash,
	).Scan(&id)
	return id, err
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(
		context.Background(),
		"SELECT id, email, password_hash, created_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(
		context.Background(),
		"SELECT id, email, password_hash, created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
