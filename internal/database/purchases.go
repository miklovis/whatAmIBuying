package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"whatAmIBuying/internal/models"
)

func AddPurchaseList(purchases []models.Purchase, receiptId int64, ctx context.Context, tx *sql.Tx) error {
	for _, purchase := range purchases {
		_, err := AddPurchase(purchase, receiptId, ctx, tx)
		if err != nil {
			log.Fatal("Error adding a purchase: ", err)
			return err
		}
	}

	return nil
}

func AddPurchase(purchase models.Purchase, receiptId int64, ctx context.Context, tx *sql.Tx) (int64, error) {
	id, err := tx.ExecContext(ctx, "INSERT INTO Purchases (name, price, receiptId) VALUES (?, ?, ?)", purchase.Product, purchase.Price, receiptId)
	if err != nil {
		log.Fatal("Error inserting purchase into database: ", err)
	}

	return id.LastInsertId()
}

func ChangePurchaseCategory(db *sql.DB, categoryId *int, purchaseId *int) (sql.Result, error) {
	result, err := db.Exec("UPDATE Purchases SET categoryId = ? WHERE id = ?", categoryId, purchaseId)
	if err != nil {
		return nil, fmt.Errorf("Query failed: %w", err)
	}

	return result, nil
}
