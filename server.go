package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go_final_project/models"
)

func Server() error {
	log.Println("Listening on port" + Port)

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.Handle("/api/nextdate", http.HandlerFunc(nextDate))
	http.Handle("/api/task", http.HandlerFunc(task))
	http.Handle("/api/tasks", http.HandlerFunc(getAll))
	http.Handle("/api/task/done", http.HandlerFunc(taskDone))
	err := http.ListenAndServe(Port, nil)
	if err != nil {
		return err
	}
	return nil
}

func nextDate(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	nowTime, err := time.Parse("20060102", now)
	if err != nil {
		return
	}

	nextDate, err := NextDate(nowTime, date, repeat)
	if err != nil {
		return
	}

	w.Write([]byte(nextDate))
}

func task(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getById(w, r)
	case http.MethodPost:
		taskPost(w, r)
	case http.MethodPut:
		taskPut(w, r)
	case http.MethodDelete:
		taskDelete(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func taskDone(w http.ResponseWriter, r *http.Request) {
	db, err := OpenDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var t models.Task

	id := r.URL.Query().Get("id")
	if id == "" {
		respErr := models.RespErr{Err: ErrEmptyId.Error()}
		res, err := json.Marshal(respErr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(res))
		return
	}

	rows, err := db.Query(`SELECT * FROM scheduler WHERE id=:id`, sql.Named("id", id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		err = rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := rows.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if t.Repeat == "" {
		err := DeleteTask(db, t.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(`{}`))
		return
	}

	newDate, err := NextDate(time.Now(), t.Date, t.Repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Date = newDate
	err = UpdateTask(db, &t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{}`))
}

func taskPut(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	db, err := OpenDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, err = ValidTask(&task)
	if err != nil {
		respErr := models.RespErr{Err: ErrBadVal.Error()}
		res, err := json.Marshal(respErr)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(res))
		return
	}

	if err := UpdateTask(db, &task); err != nil {
		respErr := models.RespErr{Err: err.Error()}
		res, err := json.Marshal(respErr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(res)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func getById(w http.ResponseWriter, r *http.Request) {
	db, err := OpenDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var t models.Task
	id := r.URL.Query().Get("id")
	if id == "" {
		respErr := models.RespErr{Err: ErrEmptyId.Error()}
		res, err := json.Marshal(respErr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(res))
		return
	}
	rows, err := db.Query(`SELECT * FROM scheduler WHERE id=:id`, sql.Named("id", id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		err = rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := rows.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if t.ID == "" {
		respErr := models.RespErr{Err: ErrSearchTask.Error()}
		res, err := json.Marshal(respErr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(res))
		return
	}

	res, err := json.Marshal(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(res))
}

func getAll(w http.ResponseWriter, r *http.Request) {
	db, err := OpenDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var tl models.TaskList
	rows, err := db.Query(`SELECT * FROM scheduler ORDER BY date ASC LIMIT 50`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		var t models.Task
		err = rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tl.Tasks = append(tl.Tasks, t)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := rows.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if tl.Tasks == nil {
		tl.Tasks = []models.Task{}
	}

	res, err := json.Marshal(tl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(res))
}

func taskPost(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	db, err := OpenDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, err = ValidTask(&task)
	if err != nil {
		respErr := models.RespErr{Err: ErrBadVal.Error()}
		res, err := json.Marshal(respErr)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res)
		return
	}

	id, err := AddTask(db, &task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respId := models.RespId{Id: id}
	res, err := json.Marshal(respId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(res)

}

func taskDelete(w http.ResponseWriter, r *http.Request) {
	db, err := OpenDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	id := r.URL.Query().Get("id")
	if id == "" {
		respErr := models.RespErr{Err: ErrEmptyId.Error()}
		res, err := json.Marshal(respErr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(res))
		return
	}
	if _, err := strconv.Atoi(id); err != nil {
		respErr := models.RespErr{Err: ErrBadVal.Error()}
		res, err := json.Marshal(respErr)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res)
		return
	}

	err = DeleteTask(db, id)
	if err != nil {
		respErr := models.RespErr{Err: err.Error()}
		res, err := json.Marshal(respErr)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res)
		return
	}

	w.Write([]byte(`{}`))

}

func ValidTask(t *models.Task) (models.Task, error) {
	if strings.TrimSpace(t.Title) == "" {
		return models.Task{}, ErrEmptyTitle
	}

	now := time.Now()

	if t.Date == "" {
		t.Date = now.Format("20060102")
	}

	_, err := time.Parse("20060102", t.Date)
	if err != nil {
		return models.Task{}, ErrBadDate
	}

	if t.Date < now.Format("20060102") {
		if t.Repeat == "" {
			t.Date = now.Format("20060102")
		} else {
			t.Date, err = NextDate(now, t.Date, t.Repeat)
			if err != nil {
				return models.Task{}, err
			}
		}

	}

	return *t, nil
}

func UpdateTask(db *sql.DB, task *models.Task) error {
	// Функция должна обновлять задачу в базе данных по ID и возвращать ошибку, если задача не найдена
	stmt, err := db.Prepare("UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?")
	if err != nil {
		return ErrSqlExec
	}
	defer stmt.Close()

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

func DeleteTask(db *sql.DB, id string) error {
	stmt, err := db.Prepare("DELETE FROM scheduler WHERE id=?")
	if err != nil {
		return ErrSqlExec
	}
	defer stmt.Close()

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
