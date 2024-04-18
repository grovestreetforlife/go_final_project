package main

import (
	"strconv"
	"strings"
	"time"
)

type Service struct {
	db *Storage
}

func NewService(db *Storage) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) GetTask(id string) (*Task, error) {
	return s.db.GetTaskById(id)
}

func (s *Service) AddTask(task *Task) (string, error) {
	return s.db.AddTask(task)
}

func (s *Service) GetTasks() (*TaskList, error) {
	return s.db.GetTasks()
}

func (s *Service) UpdateTask(task *Task) error {
	return s.db.UpdateTask(task)
}

func (s *Service) DeleteTask(id string) error {
	return s.db.DeleteTask(id)
}

func (s *Service) DoneTask(id string) error {
	task, err := s.db.GetTaskById(id)
	if err != nil {
		return err
	}

	if task.Repeat == "" {
		return s.db.DeleteTask(task.ID)
	}

	task.Date, err = s.NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		return err
	}

	err = s.db.UpdateTask(task)
	if err != nil {
		return err
	}

	return nil

}

func (s *Service) ValidTaskAndModify(t *Task) (*Task, error) {
	if strings.TrimSpace(t.Title) == "" {
		return nil, ErrEmptyTitle
	}

	now := time.Now()

	if t.Date == "" {
		t.Date = now.Format("20060102")
	}

	_, err := time.Parse("20060102", t.Date)
	if err != nil {
		return nil, ErrBadDate
	}

	if t.Date < now.Format("20060102") {
		if t.Repeat == "" {
			t.Date = now.Format("20060102")
		} else {
			t.Date, err = s.NextDate(now, t.Date, t.Repeat)
			if err != nil {
				return nil, err
			}
		}

	}

	return t, nil
}

func (s *Service) NextDate(now time.Time, date string, repeat string) (string, error) {
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
