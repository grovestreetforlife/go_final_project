package models

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type RespErr struct {
	Err string `json:"error"`
}
type RespId struct {
	Id int64 `json:"id"`
}

type TaskList struct {
	Tasks []Task `json:"tasks"`
}
