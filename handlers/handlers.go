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
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const DateFormat = `20060102`
const MaxTasks = 10

var ts database.TaskStore

func CreateDB() database.TaskStore {
	ts = ts.OpenDB(database.DbPath)
	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), database.DbPath)
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("DB creating")
			install = true
		} else {
			log.Println("DB create error")
			log.Fatal(err)
		}
	}

	log.Println("DB created before")

	install = true

	if install {
		CreateTable(ts)
		return ts
	} else {
		fmt.Println("Error create table")
		return ts
	}
}

func CreateTable(ts database.TaskStore) {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT,
		title TEXT,
		comment TEXT,
		repeat TEXT
	);`

	_, err := ts.DB.Exec(createTableSQL)
	if err != nil {
		fmt.Printf("error create table: %v", err)
	}

	createIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
	`

	_, err = ts.DB.Exec(createIndexSQL)
	if err != nil {
		fmt.Printf("error create index: %v", err)
	}

	fmt.Println("Table and index created successfully")
}

func GetNextDate(w http.ResponseWriter, r *http.Request) {
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

func PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("read body ok")

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("JSON unmarshal ok ")

	if task.Title == "" {
		http.Error(w, `{"error": "title cannot be empty"}`, http.StatusBadRequest)
		return
	}

	if task.Date == "" {
		task.Date = time.Now().Format(DateFormat)
	} else {
		parsedDate, err := time.Parse(DateFormat, task.Date)
		if err != nil {
			http.Error(w, `{"error": "invalid date format, expected YYYYMMDD"}`, http.StatusBadRequest)
			return
		}
		today := time.Now()

		if parsedDate.Before(today) {
			if task.Repeat == "" {
				task.Date = today.Format(DateFormat)
			} else {
				nextDate, err := NextDate(today, task.Date, task.Repeat)
				if err != nil {
					http.Error(w, fmt.Sprintf(`{"error": "failed to calculate next date: %s"}`, err.Error()), http.StatusBadRequest)
					return
				}
				if nextDate > today.Format(DateFormat) {
					task.Date = today.Format(DateFormat)
				} else {
					task.Date = nextDate

				}
			}
		}
	}

	id, err := ts.AddTask(task)
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

func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now().Format(DateFormat)

	var t models.Task
	var tasks []models.Task
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date >= ? ORDER BY date ASC LIMIT ?`

	rows, err := ts.DB.Query(query, now, MaxTasks)
	if err != nil {
		err := errors.New("Ошибка запроса к базе данных")
		models.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(models.ErrorResponse)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			err := errors.New("Ошибка распознавания данных")
			models.ErrorResponse.Error = err.Error()
			json.NewEncoder(w).Encode(models.ErrorResponse)
			return
		}
		tasks = append(tasks, t)
	}

	if len(tasks) == 0 {
		tasks = []models.Task{}
	}

	response := models.TasksResponse{
		Tasks: tasks,
	}

	// Возвращаем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	var t models.Task

	id := r.URL.Query().Get("id")

	if id == "" {
		err := errors.New("Пустой id")
		models.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(models.ErrorResponse)
		return
	}

	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	row := ts.DB.QueryRow(query, id)
	err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)

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

func PutTaskHandler(w http.ResponseWriter, r *http.Request) {
	var t models.Task
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

func DoneTaskHandler(w http.ResponseWriter, r *http.Request) {
	var t models.Task

	id := r.URL.Query().Get("id")

	if id == "" {
		err := errors.New("Пустой id")
		models.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(models.ErrorResponse)
		return
	}
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`

	today := time.Now()
	row := ts.DB.QueryRow(query, id)
	err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)

	if err != nil {
		err := errors.New("Такого id нет")
		models.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(models.ErrorResponse)
		return
	}

	if t.Repeat == "" {
		_, err := ts.DeleteTask(t.ID)
		if err != nil {
			return
		}
	} else {
		nextDate, err := NextDate(today, t.Date, t.Repeat)
		if err != nil {
			log.Println("Ошибка nextdata")
			return
		}

		t.Date = nextDate
	}

	// Возвращаем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))

}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		log.Printf("error id")
	}

	_, err := ts.DeleteTask(id)
	if err != nil {
		if err.Error() == "задача не найдена" {
			http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		}
		return
	}

	// Возвращаем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}
