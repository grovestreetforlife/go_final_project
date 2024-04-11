package main

import (
	"go_final_project/models"
	"strconv"
	"strings"
	"time"
)

type Meths interface {
	AddTask(task *models.Task) (int64, error)
	GetTaskById(id string) (*models.Task, error)
	GetTasks() (*models.TaskList, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id string) error
	ValidTask(t *models.Task) (*models.Task, error)
	NextDate(now time.Time, date string, repeat string) (string, error)
}

type Meth struct {
	db Database
}

func NewDates(conn Database) *Meth {
	return &Meth{
		db: conn,
	}
}

func (m *Meth) GetTaskById(id string) (*models.Task, error) {
	return m.db.GetTaskById(id)
}

func (m *Meth) AddTask(task *models.Task) (int64, error) {
	return m.db.AddTask(task)
}

func (m *Meth) GetTasks() (*models.TaskList, error) {
	return m.db.GetTasks()
}

func (m *Meth) UpdateTask(task *models.Task) error {
	return m.db.UpdateTask(task)
}

func (m *Meth) DeleteTask(id string) error {
	return m.db.DeleteTask(id)
}

func (m *Meth) ValidTask(t *models.Task) (*models.Task, error) {
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
			t.Date, err = m.NextDate(now, t.Date, t.Repeat)
			if err != nil {
				return &models.Task{}, err
			}
		}

	}

	return t, nil
}

func (m *Meth) NextDate(now time.Time, date string, repeat string) (string, error) {
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
