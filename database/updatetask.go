package database

import (
	"database/sql"
	"errors"
	"finalProject/models"
)

func UpdateTask(db *sql.DB, t models.Task) error {
	query := `UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?`
	result, err := db.Exec(query, t.Date, t.Title, t.Comment, t.Repeat, t.ID)
	if err != nil {
		err := errors.New("Задача с таким id не найдена")
		return err
	}

	// Считаем измененные строки
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		err := errors.New("Не получилось посчитать измененные строки")
		return err
	}

	if rowsAffected == 0 {
		err := errors.New("Задача не изменена")
		return err
	}

	return nil

}
