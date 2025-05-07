package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "modernc.org/sqlite"
)

type Purchase struct {
	Product    string
	Price      string
	PriceFloat float64
}

type Receipt struct {
	Date      string     `json:"date"`
	Purchases []Purchase `json:"-"`
	Amount    string     `json:"amount"`
}

type rawJsonData struct {
	Date   string            `json:"date"`
	Values map[string]string `json:"values"`
	Amount string            `json:"amount"`
}

type Category struct {
	ID       int
	Category string
}

func AddReceipt(receipt Receipt, db *sql.DB) (int64, error) {
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

func AddPurchaseList(purchases []Purchase, receiptId int64, ctx context.Context, tx *sql.Tx) error {
	for _, purchase := range purchases {
		_, err := AddPurchase(purchase, receiptId, ctx, tx)
		if err != nil {
			log.Fatal("Error adding a purchase: ", err)
			return err
		}
	}

	return nil
}

func AddPurchase(purchase Purchase, receiptId int64, ctx context.Context, tx *sql.Tx) (int64, error) {
	id, err := tx.ExecContext(ctx, "INSERT INTO Purchases (name, price, receiptId) VALUES (?, ?, ?)", purchase.Product, purchase.Price, receiptId)
	if err != nil {
		log.Fatal("Error inserting purchase into database: ", err)
	}

	return id.LastInsertId()
}

func AssignPurchases() {
	db, err := OpenDatabase()
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	rows, err := db.Query("SELECT * FROM Purchases WHERE categoryId IS NULL")
	if err != nil {
		log.Fatal("Error reading from Purchases table: ", err)
	}
	defer rows.Close()

	var categories = GetAllCategories(db)

	fmt.Println("Assign the purchase to one of these categories: ")
	for _, category := range *categories {
		fmt.Printf("ID: %d, Category: %s \n", category.ID, category.Category)
	}
}

func GetAllCategories(db *sql.DB) *[]Category {
	rows, err := db.Query("SELECT * FROM Categories")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var c Category

		err := rows.Scan(&c.ID, &c.Category)
		if err != nil {
			fmt.Println(err)
		}

		categories = append(categories, c)

	}

	return &categories
}

func ReadReceipts() {
	db, err := OpenDatabase()
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	fileContent, err := os.ReadFile("output.json")
	if err != nil {
		fmt.Println("Error reading output file:", err)
		return
	}

	var raw rawJsonData
	err = json.Unmarshal(fileContent, &raw)
	if err != nil {
		log.Fatal("Error unmarshalling JSON:", err)
	}

	var data Receipt
	data.Date = raw.Date

	for product, price := range raw.Values {
		var newPurchase Purchase
		newPurchase.Product = product
		newPurchase.Price = price
		newPurchase.PriceFloat, err = strconv.ParseFloat(price, 64)

		if err == nil {
			data.Purchases = append(data.Purchases, newPurchase)

		}
	}

	var sum float64
	for i := range data.Purchases {
		fmt.Println(data.Purchases[i].Product + " " + data.Purchases[i].Price + " " + fmt.Sprintf("%.2f", sum))
		sum += data.Purchases[i].PriceFloat
	}

	id, err := AddReceipt(data, db)
	if err != nil {
		log.Fatal("Error adding receipt: ", err)
	}

	fmt.Printf("Added receipt with id %d", id)

}

func OpenDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "test_database.db")
	return db, err
}

func main() {

	operation := "assign"

	if operation == "assign" {
		AssignPurchases()
	} else if operation == "read" {
		ReadReceipts()
	}

}
