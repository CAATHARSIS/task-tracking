package board_repo

import (
	"database/sql"
	"errors"
	"time"

	"github.com/CAATHARSIS/task-tracking/internal/models"
)

type BoardPostgresRepo struct {
	db *sql.DB
}

func NewBoardPostgresRepo(db *sql.DB) *BoardPostgresRepo {
	return &BoardPostgresRepo{db: db}
}

func (r *BoardPostgresRepo) Create(board *models.Board) error {
	query := `
		INSERT INTO boards (name, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	now := time.Now()
	err := r.db.QueryRow(
		query,
		board.Name,
		board.UserID,
		now,
		now,
	).Scan(&board.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *BoardPostgresRepo) GetById(id int) (*models.Board, error) {
	query := `
		SELECT id, name, user_id, created_at, updated_at
		FROM boards
		WHERE id = $1
	`

	row := r.db.QueryRow(query, id)
	board := &models.Board{}
	err := row.Scan(
		&board.ID,
		&board.Name,
		&board.UserID,
		&board.CreatedAt,
		&board.UpdateddAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("board not found")
		}

		return nil, err
	}

	return board, nil
}

func (r *BoardPostgresRepo) Update(board *models.Board) error {
	query := `
		UPDATE boards
		SET name = $1,
			updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(
		query,
		board.Name,
		board.UpdateddAt,
		board.ID,
	)

	return err
}

func (r *BoardPostgresRepo) Delete(id int) error {
	query := `DELETE FROM boards WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *BoardPostgresRepo) ListByUser(user_id int) ([]*models.Board, error) {
	query := `
		SELECT id, name, user_id, created_at
		FROM boards
		WHERE user_id = $1
	`

	rows, err := r.db.Query(query, user_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boards []*models.Board
	for rows.Next() {
		board := &models.Board{}
		err := rows.Scan(
			&board.ID,
			&board.Name,
			&board.UserID,
			&board.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		boards = append(boards, board)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return boards, nil
}
