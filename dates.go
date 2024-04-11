package main

import (
	"go_final_project/models"
	"strconv"
	"time"
)

type Meths interface {
	AddTask(task *models.Task) (int64, error)
	GetTaskById(id string) (*models.Task, error)
	GetTasks() (*models.TaskList, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id string) error
	ValidTask(t *models.Task) (*models.Task, error)
}

type Meth struct {
	meths Meths
}

func NewDates(db Meths) *Meth {
	return &Meth{
		meths: db,
	}
}

func (m *Meth) GetTaskById(id string) (*models.Task, error) {
	return m.meths.GetTaskById(id)
}

func (m *Meth) AddTask(task *models.Task) (int64, error) {
	return m.meths.AddTask(task)
}

func (m *Meth) GetTasks() (*models.TaskList, error) {
	return m.meths.GetTasks()
}

func (m *Meth) UpdateTask(task *models.Task) error {
	return m.meths.UpdateTask(task)
}

func (m *Meth) DeleteTask(id string) error {
	return m.meths.DeleteTask(id)
}

func (m *Meth) ValidTask(t *models.Task) (*models.Task, error) {
	return m.meths.ValidTask(t)
}

func NextDate(now time.Time, date string, repeat string) (string, error) {

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
