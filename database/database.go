package database

import (
	"database/sql"
	"finalProject/models"
	"fmt"

	_ "modernc.org/sqlite"
)

const DbPath = "scheduler.db"

type TaskStore struct {
	DB *sql.DB
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
		fmt.Println("Ощибка подсчета")
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
