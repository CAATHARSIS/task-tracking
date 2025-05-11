package refresh_token

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/CAATHARSIS/task-tracking/internal/models"
)

type RefreshTokenPostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *RefreshTokenPostgresRepo {
	return &RefreshTokenPostgresRepo{db: db}
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (r *RefreshTokenPostgresRepo) Create(token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (token_hash, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(
		query,
		hashToken(token.TokenHash),
		token.UserID,
		token.ExpiresAt,
		time.Now(),
	)

	return err
}

func (r *RefreshTokenPostgresRepo) GetByHash(tokenHash string) (*models.RefreshToken, error) {
	query := `
		SELECT token_hash, user_id, expires_at, created_at
		FROM refresh_token
		WHERE token_hash = $1 AND expires_at > NOW
	`

	refreshToken := &models.RefreshToken{}
	err := r.db.QueryRow(query, hashToken(tokenHash)).Scan(
		&refreshToken.TokenHash,
		&refreshToken.UserID,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}

	return refreshToken, nil
}

func (r *RefreshTokenPostgresRepo) DeleteByHash(tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	_, err := r.db.Exec(query, hashToken(tokenHash))
	return err
}

func (r *RefreshTokenPostgresRepo) DeleteAllForUser(userID int) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *RefreshTokenPostgresRepo) Exists(tokenHash string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM refresh_tokens
		WHERE token_hash = $1 AND expires_at > NOW()
	`
	
	var count int
	err := r.db.QueryRow(query, hashToken(tokenHash)).Scan(&count)

	if count > 0 {
		return true, err
	}
	return false, err
}

func (r *RefreshTokenPostgresRepo) RevokeExpires() (int64, error) {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	result, err := r.db.Exec(query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}