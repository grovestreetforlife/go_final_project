package main

import (
	"strconv"
	"strings"
	"time"
)

type Database interface {
	AddTask(task *Task) (uint64, error)
	GetTaskById(id uint64) (*TaskJSON, error)
	GetTasks() (*TaskList, error)
	UpdateTask(task *Task) error
	DeleteTask(id uint64) error
	ValidTaskAndModify(t *Task) (*Task, error)
	NextDate(now time.Time, date string, repeat string) (string, error)
}

type Storage struct {
	db *SQLDatabase
}

func New(db *SQLDatabase) *Storage {
	return &Storage{
		db: db,
	}
}

func (m *Storage) GetTaskById(id uint64) (*Task, error) {
	return m.db.GetTaskById(id)
}

func (m *Storage) AddTask(task *Task) (uint64, error) {
	return m.db.AddTask(task)
}

func (m *Storage) GetTasks() (*TaskList, error) {
	return m.db.GetTasks()
}

func (m *Storage) UpdateTask(task *Task) error {
	return m.db.UpdateTask(task)
}

func (m *Storage) DeleteTask(id uint64) error {
	return m.db.DeleteTask(id)
}

func (m *Storage) DoneTask(id uint64) error {
	task, err := m.db.GetTaskById(uint64(id))
	if err != nil {
		return err
	}

	if task.Repeat == "" {
		return m.db.DeleteTask(task.ID)
	}

	task.Date, err = m.NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		return err
	}

	err = m.db.UpdateTask(task)
	if err != nil {
		return err
	}

	return nil

}

func (m *Storage) ValidTaskAndModify(t *Task) (*Task, error) {
	if strings.TrimSpace(t.Title) == "" {
		return &Task{}, ErrEmptyTitle
	}

	now := time.Now()

	if t.Date == "" {
		t.Date = now.Format("20060102")
	}

	_, err := time.Parse("20060102", t.Date)
	if err != nil {
		return &Task{}, ErrBadDate
	}

	if t.Date < now.Format("20060102") {
		if t.Repeat == "" {
			t.Date = now.Format("20060102")
		} else {
			t.Date, err = m.NextDate(now, t.Date, t.Repeat)
			if err != nil {
				return &Task{}, err
			}
		}

	}

	return t, nil
}

func (m *Storage) NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", ErrBadVal
	}

	if repeat[0] != 'd' && repeat[0] != 'y' {
		return "", ErrBadVal
	}

	if repeat[0] == 'd' && len(repeat) < 3 {
		return "", ErrBadVal
	}

	switch repeat[0] {

	case 'd':

		repDays, err := strconv.Atoi(repeat[2:])
		if err != nil {
			return "", err
		}
		if repDays < 1 || repDays > 400 {
			return "", ErrBadVal
		}

		planDate, err := time.Parse("20060102", date)
		if err != nil {
			return "", err
		}

		for planDate.Before(now) || date >= planDate.Format("20060102") {
			planDate = planDate.AddDate(0, 0, repDays)
		}

		return planDate.Format("20060102"), nil

	case 'y':
		planDate, err := time.Parse("20060102", date)
		if err != nil {
			return "", err
		}

		for planDate.Before(now) || date >= planDate.Format("20060102") {
			planDate = planDate.AddDate(1, 0, 0)
		}

		return planDate.Format("20060102"), nil

	default:
		return "", ErrBadVal
	}
}
