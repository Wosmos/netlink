package repository

import (
	"context"

	"go-to-do/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) InitSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		text TEXT NOT NULL,
		completed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_tasks_user ON tasks(user_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_user_created ON tasks(user_id, created_at DESC);`
	_, err := r.pool.Exec(ctx, schema)
	return err
}

func (r *TaskRepository) GetAllByUser(userID int) ([]models.Task, error) {
	rows, err := r.pool.Query(
		context.Background(),
		"SELECT id, user_id, text, completed, created_at FROM tasks WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.UserID, &task.Text, &task.Completed, &task.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *TaskRepository) Create(userID int, text string) error {
	_, err := r.pool.Exec(
		context.Background(),
		"INSERT INTO tasks (user_id, text) VALUES ($1, $2)",
		userID, text,
	)
	return err
}

func (r *TaskRepository) Toggle(id, userID int) error {
	_, err := r.pool.Exec(
		context.Background(),
		"UPDATE tasks SET completed = NOT completed WHERE id = $1 AND user_id = $2",
		id, userID,
	)
	return err
}

func (r *TaskRepository) Delete(id, userID int) error {
	_, err := r.pool.Exec(
		context.Background(),
		"DELETE FROM tasks WHERE id = $1 AND user_id = $2",
		id, userID,
	)
	return err
}
