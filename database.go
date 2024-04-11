package main

import (
	"database/sql"
	"go_final_project/models"
)

type SQLDatabase struct {
	*sql.DB
}

type Database interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	AddTask(task *models.Task) (int64, error)
	GetTaskById(id string) (*models.Task, error)
	GetTasks() (*models.TaskList, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id string) error
}

func NewDatabase() (*SQLDatabase, error) {
	sqlDB, err := sql.Open("sqlite3", DBName)
	if err != nil {
		return nil, err
	}
	if err := migrate(sqlDB); err != nil {
		return nil, err
	}
	return &SQLDatabase{sqlDB}, nil
}

func migrate(d *sql.DB) error {
	_, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT, date TEXT, title TEXT, comment TEXT, repeat VARCHAR(128));
		CREATE INDEX IF NOT EXISTS idx_date ON scheduler (date);`)
	if err != nil {
		return ErrCreateDB
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
