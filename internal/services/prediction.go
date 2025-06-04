package services

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"
	"whatAmIBuying/internal/database"
	"whatAmIBuying/internal/models"
)

func PredictPurchases() {
	db, err := database.OpenDatabase()
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	targetTime := time.Date(2023, 12, 24, 15, 30, 0, 0, time.Local) // Example: Christmas Eve at 3:30 PM

	categoryScores, err := getTimeBasedRecommendations(db, targetTime)
	if err != nil {
		fmt.Println(err)
	}

	for _, cs := range categoryScores {
		fmt.Printf("Category ID: %d, score: %f \n", cs.CategoryID, cs.Score)
	}
}

func getTimeBasedRecommendations(db *sql.DB, targetTime time.Time) ([]models.CategoryScore, error) {
	rows, err := db.Query(`SELECT pu.Id, pu.name, pu.price, pu.categoryId, r.date
	FROM Purchases pu 
	JOIN Receipts r on pu.receiptId = r.Id
	WHERE pu.categoryId IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var purchases []models.PurchaseRecord
	for rows.Next() {
		var pr models.PurchaseRecord
		var dateStr string
		err = rows.Scan(&pr.Purchase.Id, &pr.Purchase.Product, &pr.Purchase.Price, &pr.Purchase.CategoryId, &dateStr)
		if err != nil {
			return nil, err
		}

		pr.ReceiptDate, err = time.Parse("2006-01-02 15:04:05", dateStr)
		if err != nil {
			// If standard format fails, try with microseconds
			pr.ReceiptDate, err = time.Parse("2006-01-02 15:04:05.000000", dateStr)
			if err != nil {
				return nil, err
			}
		}

		purchases = append(purchases, pr)
	}

	scores := make(map[int]*models.CategoryScore)
	targetWeekday := int(targetTime.Weekday())
	targetMonth := int(targetTime.Month())

	for _, p := range purchases {
		var catID = int(p.Purchase.CategoryId.Int64)
		if scores[catID] == nil {
			scores[catID] = &models.CategoryScore{
				CategoryID: catID,
				Score:      0,
			}
		}

		// Weekday score
		purchaseWeekday := int(p.ReceiptDate.Weekday())
		weekdayDist := math.Mod(float64(purchaseWeekday-targetWeekday+7), 7)
		weekdayScore := math.Exp(-math.Pow(weekdayDist, 2)/4.5) * 0.5

		// Month score
		purchaseMonth := int(p.ReceiptDate.Month())
		monthDist := math.Mod(float64(purchaseMonth-targetMonth+12), 12)
		if monthDist > 6 {
			monthDist = 12 - monthDist
		}
		monthScore := math.Exp(-math.Pow(monthDist, 2)/8.0) * 0.3

		freqScore := 0.2
		scores[catID].Score += weekdayScore + monthScore + freqScore
	}

	var categoryScores []models.CategoryScore
	for _, score := range scores {
		categoryScores = append(categoryScores, *score)
	}

	return categoryScores, nil
}
