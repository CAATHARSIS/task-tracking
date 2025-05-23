package task_repo

import (
	"database/sql"
	"errors"
	"time"

	"github.com/CAATHARSIS/task-tracking/internal/models"
)

type TaskPostgresRepo struct {
	db *sql.DB
}

func NewTaskPostgresRepo(db *sql.DB) *TaskPostgresRepo {
	return &TaskPostgresRepo{db: db}
}

func (r *TaskPostgresRepo) Create(task *models.Task) error {
	query := `
		INSERT INTO tasks (title, description, status, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	now := time.Now()
	err := r.db.QueryRow(
		query,
		task.Title,
		task.Description,
		task.Status,
		task.UserID,
		now,
		now,
	).Scan(&task.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *TaskPostgresRepo) GetById(id int) (*models.Task, error) {
	query := `
		SELECT id, title, description, status, user_id, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.db.QueryRow(query, id)

	task := &models.Task{}
	err := row.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.UserID,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("task not found")
		}

		return nil, err
	}

	return task, nil
}

func (r *TaskPostgresRepo) Update(task *models.Task) error {
	query := `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.Exec(
		query,
		task.Title,
		task.Description,
		task.Status,
		time.Now(),
		task.ID,
	)

	return err
}

func (r *TaskPostgresRepo) Delete(id int) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *TaskPostgresRepo) ListByUser(userID int) ([]*models.Task, error) {
	query := `
		SELECT id, title, description, status, user_id, created_at, updated_at
		FROM tasks
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.UserID,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}
