package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"finalProject/models"

	"fmt"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

const DbPath = "scheduler.db"
const MaxTasks = 10
const DateFormat = `20060102`

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

func (at TaskStore) UpdateTask(t models.Task) error {
	query := `UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?`
	result, err := at.DB.Exec(query, t.Date, t.Title, t.Comment, t.Repeat, t.ID)
	if err != nil {
		fmt.Println("Задача не найдена")
		return err
	}

	// Считаем измененные строки
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("Ошибка подсчета")
		return err
	}

	if rowsAffected == 0 {
		fmt.Println("Задача без изменений")
		return err
	}

	return nil
}

func (at TaskStore) DeleteTask(id string) (sql.Result, error) {
	deleteQuery := `DELETE FROM scheduler WHERE id = ?`
	res, err := at.DB.Exec(deleteQuery, id)
	if err != nil {
		fmt.Println("Ошибка выполнения запроса")
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		fmt.Println("Ошибка получения результата запроса")
		return nil, err
	}

	if rowsAffected == 0 {
		fmt.Println("Ошибка не найдена")
		return nil, err
	}

	return res, nil
}

func (at TaskStore) AddTask(t models.Task) (int64, error) {
	res, err := at.DB.Exec("insert into scheduler (date, title, comment, repeat) values (?, ?, ?, ?)",
		t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (at TaskStore) DoneTask(oper string, t models.Task, id string) (TaskStore, error) {
	if oper == "select" {
		query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
		row := at.DB.QueryRow(query, id)
		err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		return at, err
	}

	if oper == "update" {
		query := `UPDATE scheduler SET date = ? WHERE id = ?`
		_, err := at.DB.Exec(query, t.Date, id)
		return at, err
	}

	return at, nil
}

func (at TaskStore) GetTasks(t models.Task, w http.ResponseWriter) (models.TasksResponse, error) {
	now := time.Now().Format(DateFormat)

	var tasks []models.Task
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date >= ? ORDER BY date ASC LIMIT ?`

	rows, err := at.DB.Query(query, now, MaxTasks)
	if err != nil {
		err := errors.New("Ошибка запроса к базе данных")
		models.ErrorResponse.Error = err.Error()
		json.NewEncoder(w).Encode(models.ErrorResponse)
		return models.TasksResponse{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			err := errors.New("Ошибка распознавания данных")
			models.ErrorResponse.Error = err.Error()
			json.NewEncoder(w).Encode(models.ErrorResponse)
			return models.TasksResponse{}, err
		}
		tasks = append(tasks, t)
	}

	if len(tasks) == 0 {
		tasks = []models.Task{}
	}

	response := models.TasksResponse{
		Tasks: tasks,
	}

	return response, nil
}

func (at TaskStore) GetTask(t models.Task, id string) (TaskStore, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	row := at.DB.QueryRow(query, id)
	err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		return TaskStore{}, err
	}
	return at, nil
}
