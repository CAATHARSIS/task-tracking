package models

import "time"

/*
Поле email обязательно для заполения и будет валидироваться по маске
Полу passwordHash не будет передоваться в json формат
*/
type User struct {
	ID int `json:"id"`
	Email string `json:"email" validate:"requrired,email"`
	PasswordHash string `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}