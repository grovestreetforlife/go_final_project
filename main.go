package main

import (
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const (
	webDir = "./web"
	dbName = "./scheduler.db"
	port   = ":7540"
)

func main() {
	db, err := NewStorage()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	service := NewService(db)

	server := NewServer(service)
	if err = server.Start(); err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}

}
