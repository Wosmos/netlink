package repository

import (
	"context"

	"go-to-do/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NoteRepository struct {
	pool *pgxpool.Pool
}

func NewNoteRepository(pool *pgxpool.Pool) *NoteRepository {
	return &NoteRepository{pool: pool}
}

func (r *NoteRepository) InitSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS notes (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		title VARCHAR(255) NOT NULL DEFAULT '',
		content TEXT NOT NULL DEFAULT '',
		color VARCHAR(20) DEFAULT '#ffffff',
		pinned BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_notes_user ON notes(user_id);
	CREATE INDEX IF NOT EXISTS idx_notes_pinned ON notes(user_id, pinned DESC, updated_at DESC);`
	_, err := r.pool.Exec(ctx, schema)
	return err
}

func (r *NoteRepository) Create(userID int, title, content, color string) (*models.Note, error) {
	var note models.Note
	err := r.pool.QueryRow(
		context.Background(),
		`INSERT INTO notes (user_id, title, content, color) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id, user_id, title, content, color, pinned, created_at, updated_at`,
		userID, title, content, color,
	).Scan(&note.ID, &note.UserID, &note.Title, &note.Content, &note.Color, &note.Pinned, &note.CreatedAt, &note.UpdatedAt)
	return &note, err
}

func (r *NoteRepository) GetByID(id, userID int) (*models.Note, error) {
	var note models.Note
	err := r.pool.QueryRow(
		context.Background(),
		`SELECT id, user_id, title, content, color, pinned, created_at, updated_at 
		 FROM notes WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&note.ID, &note.UserID, &note.Title, &note.Content, &note.Color, &note.Pinned, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &note, nil
}

func (r *NoteRepository) GetAllByUser(userID int) ([]models.Note, error) {
	rows, err := r.pool.Query(
		context.Background(),
		`SELECT id, user_id, title, content, color, pinned, created_at, updated_at 
		 FROM notes WHERE user_id = $1 
		 ORDER BY pinned DESC, updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		if err := rows.Scan(&note.ID, &note.UserID, &note.Title, &note.Content, &note.Color, &note.Pinned, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (r *NoteRepository) Update(id, userID int, title, content, color string, pinned bool) error {
	_, err := r.pool.Exec(
		context.Background(),
		`UPDATE notes SET title = $1, content = $2, color = $3, pinned = $4, updated_at = CURRENT_TIMESTAMP 
		 WHERE id = $5 AND user_id = $6`,
		title, content, color, pinned, id, userID,
	)
	return err
}

func (r *NoteRepository) Delete(id, userID int) error {
	_, err := r.pool.Exec(
		context.Background(),
		`DELETE FROM notes WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	return err
}

func (r *NoteRepository) TogglePin(id, userID int) error {
	_, err := r.pool.Exec(
		context.Background(),
		`UPDATE notes SET pinned = NOT pinned, updated_at = CURRENT_TIMESTAMP 
		 WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	return err
}
