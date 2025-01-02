package db

import (
	cred "6/ex01/internal/credentials"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

var Db *gorm.DB

var Creds cred.Credentials

type Article struct {
	ID      uint   `gorm:"primaryKey"`
	Title   string `gorm:"not null"`
	Content string `gorm:"not null"`
}

func ConnectDB() error {
	var err error
	Creds, err = cred.LoadCredentials("/home/valero/Desktop/s21_Go/6/ex01/internal/credentials/admin_credentials.txt")
	if err != nil {
		log.Println(err)
		return err
	}

	dsn := fmt.Sprintf(
		"host=localhost user=%s password=%s dbname=%s port=5432 sslmode=disable",
		Creds.DBUsername, Creds.DBPassword, Creds.DBName,
	)
	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("Failed to connect to database:", err)
		return err
	}

	migrateDB()
	return nil
}

func migrateDB() {
	err := Db.AutoMigrate(&Article{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	fmt.Println("Database migrated successfully")
}
