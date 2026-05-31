package database

import (
	"log"

	"github.com/time/card/backend/internal/model"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return err
	}
	log.Println("database migrated")
	return nil
}
