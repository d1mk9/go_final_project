package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"finalProject/database"
	"finalProject/models"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const DateFormat = `20060102`

var t models.Task

type API struct {
	db database.TaskStore
}

func (api API) NewAPI(db database.TaskStore) API {
	return API{db}
}

func NextDate(now time.Time, dateStr string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("Правило повторения не указано")
	}

	date, err := time.Parse(DateFormat, dateStr)
	if err != nil {
		return "", fmt.Errorf("Неверный формат даты: %v", err)
	}

	parts := strings.Fields(repeat)
	rule := parts[0]

	var resultDate time.Time
	switch rule {
	case "":
		if date.Before(now) {
			resultDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		} else {
			resultDate = date
		}
	case "d":
		if len(parts) != 2 {
			return "", errors.New("Неверный формат повторения для 'd'")
		}

		daysToInt := make([]int, 0, 7)
		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			return "", errors.New("Неверное кол-во дней")
		}
		daysToInt = append(daysToInt, days)

		if daysToInt[0] == 1 {
			resultDate = date.AddDate(0, 0, 1)
		} else {
			resultDate = date.AddDate(0, 0, daysToInt[0])
			for resultDate.Before(now) {
				resultDate = resultDate.AddDate(0, 0, daysToInt[0])
			}
		}
	case "y":
		if len(parts) != 1 {
			return "", errors.New("Неверный формат повторения для 'y'")
		}

		resultDate = date.AddDate(1, 0, 0)
		for resultDate.Before(now) {
			resultDate = resultDate.AddDate(1, 0, 0)
		}
	default:
		return "", errors.New("Не поддерживаемый формат повторения")
	}

	return resultDate.Format(DateFormat), nil
}

func (api API) GetNextDate(w http.ResponseWriter, r *http.Request) {
	gNow, err := time.Parse(DateFormat, r.FormValue("now"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	gDate := r.FormValue("date")
	gRepeat := r.FormValue("repeat")
	newDate, err := NextDate(gNow, gDate, gRepeat)

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
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("read body ok")

	if err = json.Unmarshal(buf.Bytes(), &t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("JSON unmarshal ok ")

	if t.Title == "" {
		http.Error(w, `{"error": "title cannot be empty"}`, http.StatusBadRequest)
		return
	}

	if t.Date == "" {
		t.Date = time.Now().Format(DateFormat)
	} else {
		parsedDate, err := time.Parse(DateFormat, t.Date)
		if err != nil {
			http.Error(w, `{"error": "invalid date format, expected YYYYMMDD"}`, http.StatusBadRequest)
			return
		}
		today := time.Now()

		if parsedDate.Before(today) {
			if t.Repeat == "" {
				t.Date = today.Format(DateFormat)
			} else {
				nextDate, err := NextDate(today, t.Date, t.Repeat)
				if err != nil {
					http.Error(w, fmt.Sprintf(`{"error": "failed to calculate next date: %s"}`, err.Error()), http.StatusBadRequest)
					return
				}
				if nextDate > today.Format(DateFormat) {
					t.Date = today.Format(DateFormat)
				} else {
					t.Date = nextDate

				}
			}
		}
	}

	id, err := api.db.AddTask(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Формирование ответа
	w.Header().Set("Content-Type", "application/json, charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(fmt.Sprintf(`{"id": %d}`, id)))
	if err != nil {
		log.Println("unable to write:", err)
		return
	}

}

func (api API) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	response, err := api.db.GetTasks(t, w)
	if err != nil {
		models.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(models.ErrorResponse)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func (api API) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		err := errors.New("Пустой id")
		models.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(models.ErrorResponse)
		return
	}

	_, err := api.db.GetTask(t, id)
	if err != nil {
		err := errors.New("Такого id нет")
		models.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(models.ErrorResponse)
		return
	}

	// Возвращаем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(t)

}

func (api API) PutTaskHandler(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		err := errors.New("Ошибка чтения тела")
		models.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(models.ErrorResponse)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("Ошибка unmarshal JSON")
		return
	}
	defer r.Body.Close()

	if t.Title == "" {
		err := errors.New("Ошибка чтения заголовка")
		models.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(models.ErrorResponse)
		return
	}

	today := time.Now()

	if t.Date == "" {
		t.Date = today.Format(DateFormat)
	}

	parsedDate, err := time.Parse(DateFormat, t.Date)
	if err != nil {
		err := errors.New("Некорректный формат даты")
		models.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(models.ErrorResponse)
		return
	}

	if parsedDate.Before(today) {
		if t.Repeat == "" {
			t.Date = today.Format(DateFormat)
		} else {
			nextDate, err := NextDate(today, t.Date, t.Repeat)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "failed to calculate next date: %s"}`, err.Error()), http.StatusBadRequest)
				return
			}
			t.Date = nextDate
		}
	}

	// Формирование ответа
	w.Header().Set("Content-Type", "application/json, charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nil)
}

func (api API) DoneTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	_, err := api.db.DoneTask("select", t, id)
	if err != nil {
		err := errors.New("Такого id нет")
		models.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(models.ErrorResponse)
		return
	}

	if id == "" {
		err := errors.New("Пустой id")
		models.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(models.ErrorResponse)
		return
	}

	if t.Repeat != "" {
		nextDate, err := NextDate(time.Now(), t.Date, t.Repeat)
		if err != nil {
			err := errors.New("Ошибка nextdate")
			models.ErrorResponse.Error = err.Error()
			json.NewEncoder(w).Encode(models.ErrorResponse)
			return
		}

		t.Date = nextDate

		_, err = api.db.DoneTask("update", t, id)
		if err != nil {
			err := errors.New("Не удалось обновить дату задачи")
			models.ErrorResponse.Error = err.Error()
			json.NewEncoder(w).Encode(models.ErrorResponse)
			return
		}
	} else {
		_, err := api.db.DeleteTask(t.ID)
		if err != nil {
			err := errors.New("Ошибка удаления")
			models.ErrorResponse.Error = err.Error()
			json.NewEncoder(w).Encode(models.ErrorResponse)
			return
		}
	}
	// Возвращаем ответ
	w.Header().Set("Content-Type", "application/json, charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nil)
}

func (api API) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == "" {
		log.Printf("error id")
	}

	_, err := api.db.DeleteTask(id)
	if err != nil {
		if err.Error() == "задача не найдена" {
			http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		}
		return
	}

	// Возвращаем ответ
	w.Header().Set("Content-Type", "application/json, charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nil)
}
