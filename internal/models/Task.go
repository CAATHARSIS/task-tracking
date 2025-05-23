package models

import "time"

type TaskStatus string

// Возможные статусы
const (
	StatusToDo      TaskStatus = "todo"
	StatusInProgres TaskStatus = "in_progress"
	StatusDone      TaskStatus = "done"
)

/*
Поле title обязательно для заполнения и его длина должна быть между 3 и 100
Поле description может быть либо пустым, либо иметь максимальный размер до 500 символов
Поле status должно принимать одно из трёх константных значений
Поле userId - внешний ключ для связи с пользователем
*/
type Task struct {
	ID          int        `json:"id"`
	Title       string     `json:"title" validate:"required,min=3,max=100"`
	Description string     `json:"description,omitempty" validate:"max=500"`
	Status      TaskStatus `json:"status" validate:"oneof=todo in_progress done"`
	UserID      int        `json:"user_id" validate:"required"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TaskStatusUpdate struct {
	Status TaskStatus `json:"status" validate:"required,oneof=todo in_progress done"`
}

type TaskCreateRequest struct {
	Title       string     `json:"title" validate:"required,min=3,max=100"`
	Description string     `json:"description,omitempty" validate:"max=500"`
	Status      TaskStatus `json:"status" validate:"oneof=todo in_progress done"`
	UserID      int        `json:"user_id" validate:"required"`
}

func (s TaskStatus) IsValid() bool {
	switch s {
	case StatusToDo, StatusInProgres, StatusDone:
		return true
	default:
		return false
	}
}
