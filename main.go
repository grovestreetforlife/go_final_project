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
	err := CheckDB()
	if err != nil {
		log.Println("Failed to open DB", err)
	}
	err = Server()
	if err != nil {
		log.Println("Failed to start server", err)
	}
}
