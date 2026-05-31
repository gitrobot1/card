package model

import "time"

type User struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"size:32;uniqueIndex;not null" json:"username"`
	Nickname  string    `gorm:"size:32;not null" json:"nickname"`
	LastLogin time.Time `gorm:"not null" json:"last_login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
