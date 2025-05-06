package models

// Поле refreshToken опционально
type Token struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}