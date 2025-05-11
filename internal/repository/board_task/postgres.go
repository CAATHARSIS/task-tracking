package board_task

import "database/sql"

type BoardTaskPostgresRepo struct {
	db *sql.DB
}

func NewBoardTaskPostgresRepo(db *sql.DB) *BoardTaskPostgresRepo {
	return &BoardTaskPostgresRepo{db: db}
}

func (r *BoardTaskPostgresRepo) AddTask(boardID, taskID int) error {
	query := `
		INSERT INTO board_tasks (board_id, task_id)
		VALUES ($1, $2)
		ON CONFLICT (board_id, task_id) DO NOTHING
	`

	_, err := r.db.Exec(query, boardID, taskID)
	return err
}

func (r *BoardTaskPostgresRepo) RemoveTask(boardId, taskID int) error {
	query := `DELETE FROM board_tasks WHERE board_id = $1 AND task_id = $2`

	_, err := r.db.Exec(query, boardId, taskID)
	return err
}

func (r *BoardTaskPostgresRepo) GetTasks(boardID int) ([]int, error) {
	query := `
		SELECT task_id
		FROM board_tasks
		WHERE board_id = $1
	`

	rows, err := r.db.Query(query, boardID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taskIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, id)
	}

	return taskIDs, nil
}

func (r *BoardTaskPostgresRepo) MoveTask(fromBoardID, toBoardID, taskID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(
		`DELETE FROM board_tasks WHERE board_id = $1 AND task_id = $2`,
		fromBoardID,
		taskID,
	); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(
		`INSERT INTO board_tasks (board_id, task_id) VALUES ($1, $2)`,
		toBoardID,
		taskID,
	); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *BoardTaskPostgresRepo) Exists(boardID, taskID int) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM board_tasks
		WHERE board_id = $1 AND task_id = $2
	`

	var count int
	err := r.db.QueryRow(query, boardID, taskID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}