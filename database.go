package main

import (
	"database/sql"
	"go_final_project/models"
	"os"
	"path/filepath"
)

func CheckDB() error {
	appPath, err := os.Executable()
	if err != nil {
		return err
	}
	dbFile := filepath.Join(filepath.Dir(appPath), DBName)
	_, err = os.Stat(dbFile)
	if err != nil {
		err := createDB()
		if err != nil {
			return ErrCreateDB
		}
	}
	return nil
}

func createDB() error {
	db, err := sql.Open("sqlite3", DBName)
	if err != nil {
		return ErrOpenDB
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT, date TEXT, title TEXT, comment TEXT, repeat VARCHAR(128));`)
	if err != nil {
		return ErrSqlExec
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_date ON scheduler (date);`)
	if err != nil {
		return ErrSqlExec
	}

	return nil
}

func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", DBName)
	if err != nil {
		return nil, ErrOpenDB
	}

	return db, nil
}

func AddTask(db *sql.DB, task *models.Task) (int64, error) {
	res, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		task.Date, task.Title, task.Comment, task.Repeat)
	defer db.Close()
	if err != nil {
		return 0, ErrSqlExec
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, ErrSqlExec
	}
	return id, nil
}
