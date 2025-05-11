package repository

import "github.com/CAATHARSIS/task-tracking/internal/models"

type TaskRepository interface {
	Create(task *models.Task) error
	GetById(id int) (*models.Task, error)
	Update(task *models.Task) error
	Delete(id int) error
	ListByUser(userID int) ([]*models.Task, error)
}

type UserRepository interface {
	Create(user *models.User) error
	GetById(id int) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id int) error
}

type BoardRepository interface {
	Create(board *models.Board) error
	GetById(id int) (*models.Board, error)
	GetByUser(userID int) (*models.Board, error)
	Update(*models.Board) error
	Delete(id int) error	
}

type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	GetByHash(tokenHash string) (*models.RefreshToken, error)
	DeleteByHash(tokenHash string) error
	DeleteAllForUser(userID int) error
	Exists(tokenHash string) (bool, error)
	RevokeExpires() (int64, error)
}

type BoardTaskRepository interface {
	AddTask(boardID, taskID int) error
	RemoveTask(boardID, taskID int) error
	GetTasks(boardID int) ([]int, error)
	MoveTask(fromBoardID, toBoardID, taskID int) error
	Exists(boardID, taskID int) (bool, error)
}
