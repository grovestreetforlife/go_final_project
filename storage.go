package main

import (
	"database/sql"
	"strconv"
)

type Storage struct {
	db *sql.DB
}

func NewStorage() (*Storage, error) {
	sqlDB, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	if err := migrate(sqlDB); err != nil {
		return nil, err
	}
	return &Storage{sqlDB}, nil
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

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) AddTask(task *Task) (string, error) {
	res, err := s.db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return "", err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

func (s *Storage) GetTaskById(ids string) (*Task, error) {
	var t Task
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRow("SELECT * FROM scheduler WHERE id=?", id).Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Storage) GetTasks() (*TaskList, error) {
	var tl TaskList
	rows, err := s.db.Query(`SELECT * FROM scheduler ORDER BY date ASC LIMIT 50`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var t Task
		err = rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, err
		}
		tl.Tasks = append(tl.Tasks, t)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &tl, nil
}

func (s *Storage) UpdateTask(task *Task) error {

	stmt, err := s.db.Prepare("UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?")
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(task.Date, task.Title, task.Comment, task.Repeat, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRows
	}
	return nil
}

func (s *Storage) DeleteTask(ids string) error {
	stmt, err := s.db.Prepare("DELETE FROM scheduler WHERE id=?")
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return err
	}
	return nil
}
