package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"netlink/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) InitSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id VARCHAR(64) PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);`
	_, err := r.pool.Exec(ctx, schema)
	return err
}

func (r *SessionRepository) Create(userID int, duration time.Duration) (*models.Session, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	sessionID := hex.EncodeToString(bytes)
	expiresAt := time.Now().Add(duration)

	_, err := r.pool.Exec(
		context.Background(),
		"INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)",
		sessionID, userID, expiresAt,
	)
	if err != nil {
		return nil, err
	}

	return &models.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}

func (r *SessionRepository) GetByID(sessionID string) (*models.Session, error) {
	var session models.Session
	err := r.pool.QueryRow(
		context.Background(),
		"SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = $1 AND expires_at > $2",
		sessionID, time.Now(),
	).Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) Delete(sessionID string) error {
	_, err := r.pool.Exec(context.Background(), "DELETE FROM sessions WHERE id = $1", sessionID)
	return err
}

func (r *SessionRepository) DeleteByUserID(userID int) error {
	_, err := r.pool.Exec(context.Background(), "DELETE FROM sessions WHERE user_id = $1", userID)
	return err
}

func (r *SessionRepository) CleanExpired() error {
	_, err := r.pool.Exec(context.Background(), "DELETE FROM sessions WHERE expires_at < $1", time.Now())
	return err
}
