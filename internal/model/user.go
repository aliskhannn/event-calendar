package model

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID           uuid.UUID `json:"id"`         // уникальный идентификатор
	Email        string    `json:"email"`      // email пользователя
	Name         string    `json:"name"`       // имя пользователя
	PasswordHash string    `json:"-"`          // хеш пароля (не отдаём в JSON)
	CreatedAt    time.Time `json:"created_at"` // дата создания
	UpdatedAt    time.Time `json:"updated_at"` // дата обновления
}
