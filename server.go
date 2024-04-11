package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"go_final_project/models"
)

type Server struct {
	m Meths
}

func NewServer(meth Meths) *Server {
	return &Server{m: meth}
}

func (s *Server) Start() error {
	log.Println("Listening on port" + Port)

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.Handle("/api/nextdate", http.HandlerFunc(s.nextDate))
	http.Handle("/api/task", http.HandlerFunc(s.task))
	http.Handle("/api/tasks", http.HandlerFunc(s.getAll))
	http.Handle("/api/task/done", http.HandlerFunc(s.taskDone))
	err := http.ListenAndServe(Port, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) nextDate(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	nowTime, err := time.Parse("20060102", now)
	if err != nil {
		return
	}

	nextDate, err := s.m.NextDate(nowTime, date, repeat)
	if err != nil {
		return
	}

	w.Write([]byte(nextDate))
}

func (s *Server) task(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getById(w, r)
	case http.MethodPost:
		s.taskPost(w, r)
	case http.MethodPut:
		s.taskPut(w, r)
	case http.MethodDelete:
		s.taskDelete(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func (s *Server) taskDone(w http.ResponseWriter, r *http.Request) {

	var t *models.Task

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

	t, err := s.m.GetTaskById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if t.Repeat == "" {
		err := s.m.DeleteTask(t.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(`{}`))
		return
	}

	newDate, err := s.m.NextDate(time.Now(), t.Date, t.Repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Date = newDate
	err = s.m.UpdateTask(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{}`))
}

func (s *Server) taskPut(w http.ResponseWriter, r *http.Request) {
	var task *models.Task

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, err = s.m.ValidTask(task)
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

	if err := s.m.UpdateTask(task); err != nil {
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

func (s *Server) getById(w http.ResponseWriter, r *http.Request) {

	var t *models.Task
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

	t, err := s.m.GetTaskById(id)
	if err != nil {
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

func (s *Server) getAll(w http.ResponseWriter, r *http.Request) {

	var tl *models.TaskList
	tl, err := s.m.GetTasks()
	if err != nil {
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

func (s *Server) taskPost(w http.ResponseWriter, r *http.Request) {
	var task *models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, err = s.m.ValidTask(task)
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

	id, err := s.m.AddTask(task)
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

func (s *Server) taskDelete(w http.ResponseWriter, r *http.Request) {

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

	err := s.m.DeleteTask(id)
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
