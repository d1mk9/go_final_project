package database

import (
	"database/sql"
	"finalProject/configs"
	"finalProject/internal/tasks"

	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const DbPath = "scheduler.db"

type TaskStore struct {
	DB *sql.DB
}

func (at TaskStore) NewTaskStore(db *sql.DB) TaskStore {
	return TaskStore{db}
}

func (at TaskStore) OpenDB(dbpath string) TaskStore {
	db, err := sql.Open("sqlite", dbpath)
	if err != nil {
		fmt.Println(err)
		return TaskStore{}
	}
	return TaskStore{DB: db}
}

func (at TaskStore) UpdateTask(t configs.Task) error {
	query := `UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?`

	result, err := at.DB.Exec(query, t.Date, t.Title, t.Comment, t.Repeat, t.ID)
	if err != nil {
		return fmt.Errorf(`{"error":"Задача с таким id не найдена"}`)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(`{"error":"Не удалось посчитать измененные строки"}`)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(`{"error":"Задача с таким id не найдена"}`)
	}

	return nil
}

func (at TaskStore) DeleteTask(id string) error {
	deleteQuery := `DELETE FROM scheduler WHERE id = ?`
	res, err := at.DB.Exec(deleteQuery, id)
	if err != nil {
		return fmt.Errorf(`{"error":"Ошибка удаления"}`)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf(`{"error":"Ошибка подсчета"}`)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(`{"error":"Задача не найдена"}`)
	}

	return nil
}

func (at TaskStore) AddTask(t configs.Task) (string, error) {
	var err error
	today := time.Now()

	if t.Title == "" {
		return "", fmt.Errorf(`{"error":"Не указан заголовок задачи"}`)
	}

	if t.Date == "" {
		t.Date = today.Format(configs.DateFormat)
	}

	_, err = time.Parse(configs.DateFormat, t.Date)
	if err != nil {
		return "", fmt.Errorf(`{"error":"Некорректный формат даты"}`)
	}

	if t.Date < today.Format(configs.DateFormat) {
		if t.Repeat != "" {
			nextDate, err := tasks.NextDate(time.Now(), t.Date, t.Repeat)
			if err != nil {
				return "", fmt.Errorf(`{"error":"Некорректное правило повторения"}`)
			}
			t.Date = nextDate
		} else {
			t.Date = today.Format(configs.DateFormat)
		}
	}

	query := `insert into scheduler (date, title, comment, repeat) values (?, ?, ?, ?)`
	res, err := at.DB.Exec(query, t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		return "", fmt.Errorf(`{"error":"Не удалось добавить задачу"}`)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf(`{"error":"Не удалось вернуть id новой задачи"}`)
	}

	return fmt.Sprintf("%d", id), nil
}

func (at TaskStore) GetTasks() ([]configs.Task, error) {
	now := time.Now().Format(configs.DateFormat)

	var t configs.Task
	var tasks []configs.Task
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date >= ? ORDER BY date ASC LIMIT ?`

	rows, err := at.DB.Query(query, now, configs.MaxTasks)
	if err != nil {
		return []configs.Task{}, fmt.Errorf(`{"error":"ошибка запроса"}`)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err = rows.Err(); err != nil {
			return []configs.Task{}, fmt.Errorf(`{"error":"Ошибка распознавания данных"}`)
		}
		tasks = append(tasks, t)
	}

	if len(tasks) == 0 {
		tasks = []configs.Task{}
	}

	return tasks, nil
}

func (at TaskStore) GetTask(id string) (configs.Task, error) {
	var t configs.Task
	if id == "" {
		return configs.Task{}, fmt.Errorf(`{"error":"Не указан id"}`)
	}

	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	row := at.DB.QueryRow(query, id)
	err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		return configs.Task{}, err
	}
	return t, nil
}

func (at TaskStore) DoneTask(id string) error {
	today := time.Now()

	t, err := at.GetTask(id)
	if err != nil {
		return fmt.Errorf(`{"error":"Ошибка GetTask"}`)
	}
	if t.Repeat == "" {
		err := at.DeleteTask(id)
		if err != nil {
			return fmt.Errorf(`{"error":"Ошибка DeleteTask"}`)
		}
	} else {
		nextDate, err := tasks.NextDate(today, t.Date, t.Repeat)
		if err != nil {
			return fmt.Errorf(`{"error":"Ошибка NextDate"}`)
		}

		t.Date = nextDate
		err = at.UpdateTask(t)
		if err != nil {
			return fmt.Errorf(`{"error":"Ошибка UpdateTask"}`)
		}
	}
	return nil
}
