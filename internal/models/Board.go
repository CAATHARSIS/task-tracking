package models

import "time"

/*
Поле name обязательно для заполнения и должно быть длиной от 3 до 50 символов
Поле userId - внешний ключ для связи с пользователем
*/
type Board struct {
	ID         int       `json:"id"`
	Name       string    `json:"name" validate:"required,min=3,max=50"`
	UserID     int       `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdateddAt time.Time `json:"updated_at"`
}

// Отношение многие ко многим
type BoardTask struct {
	BoardID int `json:"board_id"`
	TaskID  int `json:"task_id"`
}

type BoardRequest struct {
	Name       string    `json:"name" validate:"required,min=3,max=50"`
}