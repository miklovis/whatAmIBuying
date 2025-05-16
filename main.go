package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

type Purchase struct {
	Id         int
	Product    string
	Price      string
	PriceFloat float64
	ReceiptId  int
	CategoryId sql.NullInt64
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

type CategoryScore struct {
	CategoryID int
	Score      float64
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

func AssignPurchases() error {
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

	var purchasesWithNullCategoryId []Purchase

	for rows.Next() {
		var p Purchase

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

		_, err = ChangePurchaseCategory(db, &id, &p.Id)
		if err != nil {
			return fmt.Errorf("changing purchase category failed: %w", err)
		}
	}

	return nil
}

func ChangePurchaseCategory(db *sql.DB, categoryId *int, purchaseId *int) (sql.Result, error) {
	result, err := db.Exec("UPDATE Purchases SET categoryId = ? WHERE id = ?", categoryId, purchaseId)
	if err != nil {
		return nil, fmt.Errorf("Query failed: %w", err)
	}

	return result, nil
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

func PredictPurchases() {
	db, err := OpenDatabase()
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	targetTime := time.Date(2023, 12, 24, 15, 30, 0, 0, time.Local) // Example: Christmas Eve at 3:30 PM

	_, _ = getTimeBasedRecommendations(db, targetTime)
}

func getTimeBasedRecommendations(db *sql.DB, targetTime time.Time) ([]CategoryScore, error) {
	// Extract temporal components from target time
	targetHour := targetTime.Hour()
	targetWeekday := targetTime.Weekday()
	targetMonth := targetTime.Month()

	// Query to analyze purchase patterns
	query := `
		SELECT 
			p.id, 
			p.name,
			-- Time of day score (0-1 how close to target hour)
			(1 - ABS(? - HOUR(purchase_time)) / 12.0) * 0.3 AS hour_score,
			-- Day of week score (1 if match, 0 otherwise)
			CASE WHEN WEEKDAY(purchase_time) = ? THEN 0.3 ELSE 0 END AS weekday_score,
			-- Month score (1 if match, 0 otherwise)
			CASE WHEN MONTH(purchase_time) = ? THEN 0.4 ELSE 0 END AS month_score,
			-- Overall purchase frequency (normalized)
			COUNT(*) / (SELECT MAX(purchase_count) FROM (
				SELECT COUNT(*) AS purchase_count FROM purchases GROUP BY product_id
			) AS counts) * 0.5 AS frequency_score
		FROM purchases pu
		JOIN products p ON pu.product_id = p.id
		GROUP BY p.id, p.name
		ORDER BY (hour_score + weekday_score + month_score + frequency_score) DESC
		LIMIT 10
	`
	// SELECT p.id, p.name, p.price, p.receiptId, r.date FROM Purchases p left join Receipts r on p.receiptId = r.id WHERE p.categoryId IS NOT NULL

	/* SELECT 	p.id,
		p.name,
		(1 - ABS(? - HOUR(r.date)) / 12.0) * 0.3 AS hour_score,
		CASE WHEN WEEKDAY(r.date) = ? THEN 0.3 ELSE 0 END AS weekday_score,
		CASE WHEN MONTH(purchase_time) = ? THEN 0.4 ELSE 0 END AS month_score,
		-- Overall purchase frequency to be added
		p.price,
		p.receiptId,
		r.date
	FROM Purchases p
	left join Receipts r on p.receiptId = r.id
	WHERE p.categoryId IS NOT NULL
	ORDER BY (hour_score + weekday_score + month_score) DESC*/

	rows, err := db.Query(query, targetHour, targetWeekday, targetMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recommendations []CategoryScore
	for rows.Next() {
		var ps CategoryScore
		var hourScore, weekdayScore, monthScore, freqScore float64
		err := rows.Scan(&ps.CategoryID, &weekdayScore, &monthScore, &freqScore)
		if err != nil {
			return nil, err
		}
		ps.Score = hourScore + weekdayScore + monthScore + freqScore
		recommendations = append(recommendations, ps)
	}

	return recommendations, nil
}
func OpenDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "test_database.db")
	return db, err
}

func main() {
	assignFlag := flag.Bool("a", false, "assign mode (shorthand)")
	assignFlagLong := flag.Bool("assign", false, "assign mode")
	readFlag := flag.Bool("r", false, "read mode (shorthand)")
	readFlagLong := flag.Bool("read", false, "read mode")
	predictFlag := flag.Bool("p", false, "predict mode (shorthand)")
	predictFlagLong := flag.Bool("predict", false, "predict mode")

	flag.Parse()

	if *assignFlag || *assignFlagLong {
		fmt.Println("Assign mode activated")
		err := AssignPurchases()
		if err != nil {
			fmt.Errorf("Error assigning purchases: %w", err)
		}
	} else if *readFlag || *readFlagLong {
		fmt.Println("Read mode activated")
		ReadReceipts()
	} else if *predictFlag || *predictFlagLong {
		fmt.Println("Predict mode activated")
		PredictPurchases()
	}
}
