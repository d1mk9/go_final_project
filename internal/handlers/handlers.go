package handlers

import (
	"encoding/json"
	"errors"
	"finalProject/configs"

	"finalProject/internal/database"
	"finalProject/internal/tasks"

	"fmt"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

const DateFormat = `20060102`

type API struct {
	db database.TaskStore
}

func (api API) NewAPI(db database.TaskStore) API {
	return API{db}
}

func (api API) GetNextDate(w http.ResponseWriter, r *http.Request) {
	gNow, err := time.Parse(configs.DateFormat, r.FormValue("now"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	gDate := r.FormValue("date")
	gRepeat := r.FormValue("repeat")
	newDate, err := tasks.NextDate(gNow, gDate, gRepeat)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(newDate))
	if err != nil {
		log.Println("unable to write:", err)
		return
	}
}

func (api API) PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	var t configs.Task

	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		err := errors.New("Ошибка десериализации JSON")
		configs.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(configs.ErrorResponse)
		return
	}
	defer r.Body.Close()

	id, err := api.db.AddTask(t)
	if err != nil {
		err := errors.New("Не удалось добавить задачу")
		configs.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(configs.ErrorResponse)
		return
	}

	response := configs.TaskResponseResult{ID: id}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (api API) GetTasksHandler(w http.ResponseWriter, r *http.Request) {

	tasks, err := api.db.GetTasks()
	if err != nil {
		err := errors.New("Ошибка запроса к базе данных")
		configs.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(configs.ErrorResponse)
		return
	}

	response := configs.TasksResponse{
		Tasks: tasks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func (api API) GetTaskHandler(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")

	t, err := api.db.GetTask(id)
	if err != nil {
		err := errors.New("Такого id нет")
		configs.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(configs.ErrorResponse)
		return
	}

	// Возвращаем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(t)

}

func (api API) PutTaskHandler(w http.ResponseWriter, r *http.Request) {
	var t configs.Task

	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		err := errors.New("Ошибка десериализации JSON")
		configs.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(configs.ErrorResponse)
		return
	}

	defer r.Body.Close()

	if t.ID == "" {
		err := errors.New("id в теле не указан")
		configs.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(configs.ErrorResponse)
		return
	}

	if t.Title == "" {
		err := errors.New("Ошибка чтения title")
		configs.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(configs.ErrorResponse)
		return
	}

	today := time.Now()

	if t.Date == "" {
		t.Date = today.Format(DateFormat)
	}

	parsedDate, err := time.Parse(DateFormat, t.Date)
	if err != nil {
		err := errors.New("Некорректный формат даты")
		configs.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(configs.ErrorResponse)
		return
	}

	if parsedDate.Before(today) {
		if t.Repeat == "" {
			t.Date = today.Format(DateFormat)
		} else {
			nextDate, err := tasks.NextDate(today, t.Date, t.Repeat)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "failed to calculate next date: %s"}`, err.Error()), http.StatusBadRequest)
				return
			}
			t.Date = nextDate
		}
	}

	err = api.db.UpdateTask(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Формирование ответа
	w.Header().Set("Content-Type", "application/json, charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nil)
}

func (api API) DoneTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	err := api.db.DoneTask(id)
	if err != nil {
		err := errors.New("Задача не найдена")
		configs.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(configs.ErrorResponse)
		return
	}

	w.Header().Set("Content-Type", "application/json, charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{})
}

func (api API) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	err := api.db.DeleteTask(id)
	if err != nil {
		err := errors.New("Задача не найдена")
		configs.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(configs.ErrorResponse)
		return
	}

	// Возвращаем ответ
	w.Header().Set("Content-Type", "application/json, charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nil)
}
