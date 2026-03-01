package repository

import (
	"context"
	"errors"
	"time"

	"netlink/models"

	"github.com/jackc/pgx/v5"
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
		is_verified BOOLEAN DEFAULT FALSE,
		verification_token VARCHAR(255),
		reset_token VARCHAR(255),
		reset_token_expires TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_verification_token ON users(verification_token);
	CREATE INDEX IF NOT EXISTS idx_users_reset_token ON users(reset_token);

	-- Add columns if they don't exist (for existing tables)
	ALTER TABLE users ADD COLUMN IF NOT EXISTS is_verified BOOLEAN DEFAULT FALSE;
	ALTER TABLE users ADD COLUMN IF NOT EXISTS verification_token VARCHAR(255);
	ALTER TABLE users ADD COLUMN IF NOT EXISTS reset_token VARCHAR(255);
	ALTER TABLE users ADD COLUMN IF NOT EXISTS reset_token_expires TIMESTAMP;
	ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);
	ALTER TABLE users ADD COLUMN IF NOT EXISTS name VARCHAR(255);
	ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar VARCHAR(500);
	ALTER TABLE users ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMP;
	CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
	`
	_, err := r.pool.Exec(ctx, schema)
	return err
}

func (r *UserRepository) Create(email, passwordHash, verificationToken string, name, phone *string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(
		context.Background(),
		"INSERT INTO users (email, password_hash, verification_token, name, phone) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		email, passwordHash, verificationToken, name, phone,
	).Scan(&id)
	return id, err
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(
		context.Background(),
		"SELECT id, email, password_hash, is_verified, verification_token, reset_token, reset_token_expires, created_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsVerified, &user.VerificationToken, &user.ResetToken, &user.ResetTokenExpires, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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
		"SELECT id, email, password_hash, is_verified, verification_token, reset_token, reset_token_expires, created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsVerified, &user.VerificationToken, &user.ResetToken, &user.ResetTokenExpires, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByVerificationToken(token string) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(
		context.Background(),
		"SELECT id, email, password_hash, is_verified, verification_token, reset_token, reset_token_expires, created_at FROM users WHERE verification_token = $1",
		token,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsVerified, &user.VerificationToken, &user.ResetToken, &user.ResetTokenExpires, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByResetToken(token string) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(
		context.Background(),
		"SELECT id, email, password_hash, is_verified, verification_token, reset_token, reset_token_expires, created_at FROM users WHERE reset_token = $1",
		token,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsVerified, &user.VerificationToken, &user.ResetToken, &user.ResetTokenExpires, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateVerificationStatus(id int, isVerified bool) error {
	_, err := r.pool.Exec(
		context.Background(),
		"UPDATE users SET is_verified = $1, verification_token = NULL WHERE id = $2",
		isVerified, id,
	)
	return err
}

func (r *UserRepository) SetResetToken(id int, token *string, expires *time.Time) error {
	_, err := r.pool.Exec(
		context.Background(),
		"UPDATE users SET reset_token = $1, reset_token_expires = $2 WHERE id = $3",
		token, expires, id,
	)
	return err
}

func (r *UserRepository) UpdatePassword(id int, passwordHash string) error {
	_, err := r.pool.Exec(
		context.Background(),
		"UPDATE users SET password_hash = $1, reset_token = NULL, reset_token_expires = NULL WHERE id = $2",
		passwordHash, id,
	)
	return err
}

func (r *UserRepository) GetByPhone(phone string) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(
		context.Background(),
		`SELECT id, email, COALESCE(phone, ''), COALESCE(name, ''), password_hash, is_verified, created_at 
		 FROM users WHERE phone = $1`,
		phone,
	).Scan(&user.ID, &user.Email, &user.Phone, &user.Name, &user.PasswordHash, &user.IsVerified, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateProfile(id int, name, phone, avatar string) error {
	_, err := r.pool.Exec(
		context.Background(),
		`UPDATE users SET name = $1, phone = $2, avatar = $3 WHERE id = $4`,
		name, phone, avatar, id,
	)
	return err
}

func (r *UserRepository) SearchUsers(query string, limit int) ([]models.User, error) {
	rows, err := r.pool.Query(
		context.Background(),
		`SELECT id, email, COALESCE(phone, ''), COALESCE(name, ''), COALESCE(avatar, '')
		 FROM users 
		 WHERE email ILIKE $1 OR name ILIKE $1 OR phone ILIKE $1
		 LIMIT $2`,
		"%"+query+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Phone, &u.Name, &u.Avatar); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
