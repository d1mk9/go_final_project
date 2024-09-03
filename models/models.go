package models

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type TasksResponse struct {
	Tasks []Task `json:"tasks"`
}

var ErrorResponse struct {
	Error string `json:"error,omitempty"`
}
