package main

import "strconv"

type Task struct {
	ID      uint64 `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskList struct {
	Tasks []Task
}

type TaskJSON struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskListJSON struct {
	Tasks []TaskJSON `json:"tasks"`
}

func TaskListToJSON(tl *TaskList) (*TaskListJSON, error) {
	tlJSON := &TaskListJSON{Tasks: make([]TaskJSON, len(tl.Tasks))}
	for i, task := range tl.Tasks {
		tlJSON.Tasks[i] = TaskJSON{
			ID:      strconv.FormatUint(task.ID, 10),
			Date:    task.Date,
			Title:   task.Title,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		}
	}
	return tlJSON, nil
}

func TaskListFromJSON(tlj *TaskListJSON) (*TaskList, error) {
	tl := &TaskList{Tasks: make([]Task, len(tlj.Tasks))}
	for i, taskJSON := range tlj.Tasks {
		id, err := strconv.ParseUint(taskJSON.ID, 10, 64)
		if err != nil {
			return nil, err
		}
		tl.Tasks[i] = Task{
			ID:      id,
			Date:    taskJSON.Date,
			Title:   taskJSON.Title,
			Comment: taskJSON.Comment,
			Repeat:  taskJSON.Repeat,
		}
	}
	return tl, nil
}

func TaskToJSON(t *Task) *TaskJSON {
	return &TaskJSON{
		ID:      strconv.FormatUint(t.ID, 10),
		Date:    t.Date,
		Title:   t.Title,
		Comment: t.Comment,
		Repeat:  t.Repeat,
	}
}

func TaskFromJSON(tj *TaskJSON) (*Task, error) {
	id, err := strconv.ParseUint(tj.ID, 10, 64)
	if err != nil {
		return &Task{}, ErrBadFormat
	}

	return &Task{
		ID:      id,
		Date:    tj.Date,
		Title:   tj.Title,
		Comment: tj.Comment,
		Repeat:  tj.Repeat,
	}, nil
}
