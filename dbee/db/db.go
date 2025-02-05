package db

import (
	"database/sql"
	"fmt"
	"os"
)

type DbConfig struct {
	host     string
	port     string
	user     string
	password string
	dbname   string
}

func NewDBConfig() DbConfig {
	return DbConfig{
		host:     os.Getenv("DB_HOST"),
		port:     os.Getenv("DB_PORT"),
		user:     os.Getenv("DB_USER"),
		password: os.Getenv("DB_PASSWORD"),
		dbname:   os.Getenv("DB_NAME"),
	}
}

func (config DbConfig) Open() (*sql.DB, error) {
	connstring := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.host,
		config.port,
		config.user,
		config.password,
		config.dbname)

	// open database
	db, err := sql.Open("postgres", connstring)
	if err != nil {
		return nil, err
	}

	// check db
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
