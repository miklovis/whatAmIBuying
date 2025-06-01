package database

import (
	"context"
	"database/sql"
	"log"
	"whatAmIBuying/internal/models"
)

func AddReceipt(receipt models.Receipt, db *sql.DB) (int64, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := db.Conn(ctx)
	if err != nil {
		log.Fatal("Error opening connection to database: ", err)
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal("Error starting a transaction: ", err)
	}
	defer tx.Rollback()

	id, err := tx.ExecContext(ctx, "INSERT INTO Receipts (date, amount) VALUES (?, ?)", receipt.Date, receipt.Amount)
	if err != nil {
		log.Fatal("Error inserting receipt into database: ", err)
	}

	receiptId, err := id.LastInsertId()
	if err != nil {
		log.Fatal("Error getting last insert ID: ", err)
	}

	err = AddPurchaseList(receipt.Purchases, receiptId, ctx, tx)
	if err != nil {
		log.Fatal("Error adding purchase list to the database: ", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal("Error committing the transaction into the database: ", err)
	}

	return receiptId, err
}
