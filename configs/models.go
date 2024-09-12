package configs

const MaxTasks = 10
const DateFormat = `20060102`

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

type TaskResponseResult struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

var ErrorResponse struct {
	Error string `json:"error,omitempty"`
}
