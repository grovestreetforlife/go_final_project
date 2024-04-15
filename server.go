package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type TodoList interface {
	AddTask(task *Task) (string, error)
	GetTaskById(id string) (*Task, error)
	GetTasks() (*TaskList, error)
	UpdateTask(task *Task) error
	DeleteTask(id string) error
	ValidTaskAndModify(t *Task) (*Task, error)
	DoneTask(id string) error
	NextDate(now time.Time, date string, repeat string) (string, error)
}

type Server struct {
	m TodoList
}

func NewServer(td TodoList) *Server {
	s := &Server{m: td}
	s.startHandlers()
	return s
}

func (s *Server) Start() error {
	log.Println("Listening on port" + port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) startHandlers() {
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	http.HandleFunc("GET /api/nextdate", s.nextDate)
	http.HandleFunc("GET /api/task", s.getTask)
	http.HandleFunc("GET /api/tasks", s.getAllTasks)

	http.HandleFunc("POST /api/task/done", s.doneTask)
	http.HandleFunc("POST /api/task", s.createTask)

	http.HandleFunc("PUT /api/task", s.updateTask)

	http.HandleFunc("DELETE /api/task", s.deleteTask)
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

func (s *Server) doneTask(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")

	err := s.m.DoneTask(id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}

func (s *Server) updateTask(w http.ResponseWriter, r *http.Request) {
	var t *Task

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	t, err = s.m.ValidTaskAndModify(t)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if err := s.m.UpdateTask(t); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, ErrSqlExec.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, ErrEmptyId.Error()), http.StatusBadRequest)
		return
	}

	t, err := s.m.GetTaskById(id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(t)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func (s *Server) getAllTasks(w http.ResponseWriter, r *http.Request) {
	var tl *TaskList

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	tl, err := s.m.GetTasks()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, ErrSqlExec.Error()), http.StatusInternalServerError)
		return
	}

	if tl.Tasks == nil {
		tl.Tasks = []Task{}
	}

	res, err := json.Marshal(tl)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, ErrBadFormat.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var task *Task

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	task, err = s.m.ValidTaskAndModify(task)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	id, err := s.m.AddTask(task)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"id":"%v"}`, id)))

}

func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")

	err := s.m.DeleteTask(id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, ErrSqlExec.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}
