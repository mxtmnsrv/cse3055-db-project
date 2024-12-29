package db

import (
	"database/sql"
	"log"

	_ "github.com/denisenkom/go-mssqldb" // Драйвер для SQL Server
)

var DB *sql.DB

func InitDB(connString string) {
	var err error
	DB, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Database connection is not alive: %v", err)
	}

	log.Println("Connected to the database successfully!")
}
