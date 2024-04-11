package main

import (
	"database/sql"
	"go_final_project/models"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SQLDatabase struct {
	DB *sql.DB
}

type Database interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	AddTask(task *models.Task) (int64, error)
	GetTaskById(id string) (*models.Task, error)
	GetTasks() (*models.TaskList, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id string) error
	ValidTask(t *models.Task) (*models.Task, error)
}

func (db *SQLDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.Query(query, args...)
}

func (db *SQLDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.DB.Exec(query, args...)
}

func NewDatabase() (*SQLDatabase, error) {
	sqlDB, err := sql.Open("sqlite3", DBName)
	if err != nil {
		return nil, err
	}
	return &SQLDatabase{DB: sqlDB}, nil
}

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
			return err
		}
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

func createDB() error {
	db, err := OpenDB()
	if err != nil {
		return ErrOpenDB
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT, date TEXT, title TEXT, comment TEXT, repeat VARCHAR(128));`)
	if err != nil {
		return ErrCreateDB
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_date ON scheduler (date);`)
	if err != nil {
		return ErrCreateIdx
	}

	return nil
}

func (s *SQLDatabase) Close() error {
	return s.DB.Close()
}

func (s *SQLDatabase) AddTask(task *models.Task) (int64, error) {
	res, err := s.DB.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, ErrSqlExec
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, ErrSqlExec
	}
	return id, nil
}

func (s *SQLDatabase) GetTaskById(id string) (*models.Task, error) {
	var t models.Task
	err := s.DB.QueryRow("SELECT * FROM scheduler WHERE id=?", id).Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		return nil, ErrSqlExec
	}
	return &t, nil
}

func (s *SQLDatabase) GetTasks() (*models.TaskList, error) {
	var tl models.TaskList
	rows, err := s.DB.Query(`SELECT * FROM scheduler ORDER BY date ASC LIMIT 50`)
	if err != nil {
		return nil, ErrSqlExec
	}
	for rows.Next() {
		var t models.Task
		err = rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, ErrSqlExec
		}
		tl.Tasks = append(tl.Tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, ErrSqlExec
	}
	if err := rows.Close(); err != nil {
		return nil, ErrSqlExec
	}
	return &tl, nil
}

func (s *SQLDatabase) UpdateTask(task *models.Task) error {
	// Функция должна обновлять задачу в базе данных по ID и возвращать ошибку, если задача не найдена
	stmt, err := s.DB.Prepare("UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?")
	if err != nil {
		return ErrSqlExec
	}

	res, err := stmt.Exec(task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return ErrSqlExec
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return ErrSqlExec
	}

	if rowsAffected == 0 {
		return ErrSearchTask
	}
	return nil
}

func (s *SQLDatabase) DeleteTask(id string) error {
	stmt, err := s.DB.Prepare("DELETE FROM scheduler WHERE id=?")
	if err != nil {
		return ErrSqlExec
	}

	res, err := stmt.Exec(id)
	if err != nil {
		return ErrSqlExec
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return ErrSqlExec
	}

	if rowsAffected == 0 {
		return ErrSearchTask
	}
	return nil
}

func (s *SQLDatabase) ValidTask(t *models.Task) (*models.Task, error) {

	if strings.TrimSpace(t.Title) == "" {
		return &models.Task{}, ErrEmptyTitle
	}

	now := time.Now()

	if t.Date == "" {
		t.Date = now.Format("20060102")
	}

	_, err := time.Parse("20060102", t.Date)
	if err != nil {
		return &models.Task{}, ErrBadDate
	}

	if t.Date < now.Format("20060102") {
		if t.Repeat == "" {
			t.Date = now.Format("20060102")
		} else {
			t.Date, err = NextDate(now, t.Date, t.Repeat)
			if err != nil {
				return &models.Task{}, err
			}
		}

	}

	return t, nil
}
