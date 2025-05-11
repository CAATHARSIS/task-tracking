package user

import (
	"database/sql"
	"errors"
	"time"

	"github.com/CAATHARSIS/task-tracking/internal/models"
)

type UserPostgrtesRepo struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *UserPostgrtesRepo {
	return &UserPostgrtesRepo{db: db}
}

func (r *UserPostgrtesRepo) Create(user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		user.Email,
		user.PasswordHash,
		time.Now(),
	).Scan(&user.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *UserPostgrtesRepo) GetById(id int) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	row := r.db.QueryRow(query, id)
	user := &models.User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

func (r *UserPostgrtesRepo) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	row := r.db.QueryRow(query, email)
	user := &models.User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

func (r *UserPostgrtesRepo) Update(user *models.User) error {
	query := `
		UPDATE users
		SET email = $1
			password_hash = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(
		query,
		user.Email,
		user.PasswordHash,
		user.ID,
	)

	return err
}

func (r *UserPostgrtesRepo) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
