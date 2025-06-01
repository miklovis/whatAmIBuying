package services

import (
	"fmt"
	"log"
	"whatAmIBuying/internal/database"
	"whatAmIBuying/internal/models"
)

func AssignPurchases() error {
	db, err := database.OpenDatabase()
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	rows, err := db.Query("SELECT * FROM Purchases WHERE categoryId IS NULL")
	if err != nil {
		log.Fatal("Error reading from Purchases table: ", err)
	}
	defer rows.Close()

	var categories = database.GetAllCategories(db)

	fmt.Println("Assign the purchase to one of these categories: ")
	for _, category := range *categories {
		fmt.Printf("ID: %d, Category: %s \n", category.ID, category.Category)
	}

	var purchasesWithNullCategoryId []models.Purchase

	for rows.Next() {
		var p models.Purchase

		err := rows.Scan(&p.Id, &p.Product, &p.Price, &p.ReceiptId, &p.CategoryId)
		if err != nil {
			fmt.Println(err)
		}

		purchasesWithNullCategoryId = append(purchasesWithNullCategoryId, p)

	}

	for _, p := range purchasesWithNullCategoryId {
		var id int

		fmt.Printf("Which category does %s bought for %s belong to? ", p.Product, p.Price)
		fmt.Scan(&id)

		_, err = database.ChangePurchaseCategory(db, &id, &p.Id)
		if err != nil {
			return fmt.Errorf("changing purchase category failed: %w", err)
		}
	}

	return nil
}
