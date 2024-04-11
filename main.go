package main

import (
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const (
	webDir = "./web"
	DBName = "./scheduler.db"
	Port   = ":7540"
)

func main() {
	db, err := NewDatabase() // Использование функции для создания базы данных
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.DB.Close()

	methods := NewDates(db)

	server := NewServer(db, methods)
	err = server.Start()
	if err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}

}
