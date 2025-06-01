package database

import (
	"database/sql"
	"fmt"
	"whatAmIBuying/internal/models"
)

func GetAllCategories(db *sql.DB) *[]models.Category {
	rows, err := db.Query("SELECT * FROM Categories")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var categories []models.Category

	for rows.Next() {
		var c models.Category

		err := rows.Scan(&c.ID, &c.Category)
		if err != nil {
			fmt.Println(err)
		}

		categories = append(categories, c)

	}

	return &categories
}
