package main

import (
	"database/sql"
)

type SQLDatabase struct {
	db *sql.DB
}

func NewDatabase() (*SQLDatabase, error) {
	sqlDB, err := sql.Open("sqlite3", dbName)
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
	return s.db.Close()
}

func (s *SQLDatabase) AddTask(task *Task) (uint64, error) {
	res, err := s.db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, ErrSqlExec
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, ErrSqlExec
	}
	return uint64(id), nil
}

func (s *SQLDatabase) GetTaskById(id uint64) (*Task, error) {
	var t Task
	err := s.db.QueryRow("SELECT * FROM scheduler WHERE id=?", id).Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		return &Task{}, ErrSqlExec
	}
	return &t, nil
}

func (s *SQLDatabase) GetTasks() (*TaskList, error) {
	var tl TaskList
	rows, err := s.db.Query(`SELECT * FROM scheduler ORDER BY date ASC LIMIT 50`)
	if err != nil {
		return nil, ErrSqlExec
	}
	for rows.Next() {
		var t Task
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

func (s *SQLDatabase) UpdateTask(task *Task) error {

	stmt, err := s.db.Prepare("UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?")
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
		return ErrRows
	}
	return nil
}

func (s *SQLDatabase) DeleteTask(id uint64) error {
	stmt, err := s.db.Prepare("DELETE FROM scheduler WHERE id=?")
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
		return ErrRows
	}
	return nil
}
