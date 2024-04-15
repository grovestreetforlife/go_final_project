package main

import (
	"fmt"
)

var (
	ErrBadDate    = fmt.Errorf("некорректная дата")
	ErrBadFormat  = fmt.Errorf("некорректный формат")
	ErrBadVal     = fmt.Errorf("некорректное значение")
	ErrBadTask    = fmt.Errorf("некорректная задача")
	ErrSearchTask = fmt.Errorf("задача не найдена")
	ErrRows       = fmt.Errorf("изменено 0 строк")

	ErrOpenDB    = fmt.Errorf("не удалось открыть базу данных")
	ErrCreateDB  = fmt.Errorf("не удалось создать базу данных")
	ErrSqlExec   = fmt.Errorf("не удалось выполнить запрос")
	ErrCreateIdx = fmt.Errorf("не удалось создать индекс")
	ErrMigrateDb = fmt.Errorf("не удалось выполнить миграцию бд")

	ErrEmptyTitle = fmt.Errorf("заголовок задачи не может быть пустым")
	ErrEmptyDate  = fmt.Errorf("дата задачи не может быть пустой")
	ErrSearch     = fmt.Errorf("ошибка в поиске")
	ErrEmptyId    = fmt.Errorf("не указан id")
)
