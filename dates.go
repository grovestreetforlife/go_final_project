package main

import (
	"strconv"
	"time"
)

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
