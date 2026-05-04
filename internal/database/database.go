package database

import (
	"rkmin/internal/domain"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := domain.AutoMigrateModels(db); err != nil {
		return nil, err
	}
	return db, nil
}
