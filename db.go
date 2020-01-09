package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// Storage ...
type Storage struct {
	DB *sql.DB
}

type storageCredentials struct {
	user     string
	password string
	host     string
	port     string
	dbName   string
	ssl      string
}

// Init ...
func (storage *Storage) Init() error {
	dbCreds := getEnv()
	err := storage.Connect(dbCreds)
	if err != nil {
		return err
	}

	return nil
}

// Connect ...
func (storage *Storage) Connect(creds storageCredentials) error {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		creds.user,
		creds.password,
		creds.host,
		creds.port,
		creds.dbName,
		creds.ssl,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}
	log.Println("db Connected")

	storage.DB = db
	return nil
}

func getEnv() storageCredentials {
	user := os.Getenv("DB_USER")
	if len(user) == 0 {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if len(password) == 0 {
		password = "some_secret_password"
	}
	host := os.Getenv("DB_HOST")
	if len(host) == 0 {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if len(port) == 0 {
		port = "5432"
	}
	dbName := os.Getenv("DB_NAME")
	if len(dbName) == 0 {
		dbName = "gotest"
	}
	ssl := os.Getenv("DB_SSL")
	if len(ssl) == 0 {
		ssl = "disable"
	}

	return storageCredentials{
		user:     user,
		password: password,
		host:     host,
		port:     port,
		dbName:   dbName,
		ssl:      ssl,
	}
}
