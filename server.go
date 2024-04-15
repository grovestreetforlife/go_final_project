package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type TodoL interface {
	AddTask(task *Task) (uint64, error)
	GetTaskById(id uint64) (*Task, error)
	GetTasks() (*TaskList, error)
	UpdateTask(task *Task) error
	DeleteTask(id uint64) error
	ValidTaskAndModify(t *Task) (*Task, error)
	DoneTask(id uint64) error
	NextDate(now time.Time, date string, repeat string) (string, error)
}

type Server struct {
	m TodoL
}

func NewServer(td TodoL) *Server {
	return &Server{m: td}
}

func (s *Server) Start() error {
	log.Println("Listening on port" + port)

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.Handle("/api/nextdate", http.HandlerFunc(s.nextDate))
	http.Handle("/api/task", http.HandlerFunc(s.task))
	http.Handle("/api/tasks", http.HandlerFunc(s.getAll))
	http.Handle("/api/task/done", http.HandlerFunc(s.taskDone))
	err := http.ListenAndServe(port, nil)
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

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, ErrEmptyId.Error()), http.StatusInternalServerError)
		return
	}

	err = s.m.DoneTask(uint64(id))
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}

func (s *Server) taskPut(w http.ResponseWriter, r *http.Request) {
	var tj *TaskJSON

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := json.NewDecoder(r.Body).Decode(&tj)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	if tj.ID == "" {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, ErrEmptyId.Error()), http.StatusBadRequest)
		return
	}

	task, err := TaskFromJSON(tj)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	task, err = s.m.ValidTaskAndModify(task)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if err := s.m.UpdateTask(task); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}

func (s *Server) getById(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, ErrEmptyId.Error()), http.StatusBadRequest)
		return
	}

	t, err := s.m.GetTaskById(uint64(id))
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	tj := TaskToJSON(t)

	res, err := json.Marshal(tj)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func (s *Server) getAll(w http.ResponseWriter, r *http.Request) {
	var tl *TaskList

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	tl, err := s.m.GetTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if tl.Tasks == nil {
		tl.Tasks = []Task{}
	}

	tlJSON, err := TaskListToJSON(tl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(tlJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(res))
}

func (s *Server) taskPost(w http.ResponseWriter, r *http.Request) {
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

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"id":"%d"}`, id)))

}

func (s *Server) taskDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, ErrEmptyId.Error()), http.StatusInternalServerError)
		return
	}

	err = s.m.DeleteTask(uint64(id))
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}
