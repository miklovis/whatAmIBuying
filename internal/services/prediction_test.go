package services

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDBForServices(t *testing.T) (*sql.DB, func()) {
	testDBPath := "test_services_temp.db"

	db, err := sql.Open("sqlite", testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS Receipts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		amount TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS Purchases (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		price TEXT NOT NULL,
		receiptId INTEGER NOT NULL,
		categoryId INTEGER,
		FOREIGN KEY(receiptId) REFERENCES Receipts(id)
	);

	CREATE TABLE IF NOT EXISTS Categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		Category TEXT NOT NULL UNIQUE
	);
	`

	_, err = db.Exec(createTables)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Seed categories
	categories := []string{"Dairy", "Meat", "Vegetables"}
	for _, cat := range categories {
		_, err := db.Exec("INSERT INTO Categories (Category) VALUES (?)", cat)
		if err != nil {
			t.Fatalf("Failed to seed categories: %v", err)
		}
	}

	cleanup := func() {
		db.Close()
		os.Remove(testDBPath)
	}

	return db, cleanup
}

func TestGetTimeBasedRecommendations(t *testing.T) {
	db, cleanup := setupTestDBForServices(t)
	defer cleanup()

	// Add test data - receipts with purchases
	ctx := context.Background()

	// Receipt 1: Monday, January
	tx, _ := db.Begin()
	result, _ := tx.ExecContext(ctx, "INSERT INTO Receipts (date, amount) VALUES (?, ?)",
		"2025-01-06 10:00:00", "10.00") // Monday
	receiptId1, _ := result.LastInsertId()
	tx.ExecContext(ctx, "INSERT INTO Purchases (name, price, receiptId, categoryId) VALUES (?, ?, ?, ?)",
		"Milk", "2.50", receiptId1, 1) // Dairy
	tx.ExecContext(ctx, "INSERT INTO Purchases (name, price, receiptId, categoryId) VALUES (?, ?, ?, ?)",
		"Chicken", "5.50", receiptId1, 2) // Meat
	tx.Commit()

	// Receipt 2: Tuesday, January
	tx, _ = db.Begin()
	result, _ = tx.ExecContext(ctx, "INSERT INTO Receipts (date, amount) VALUES (?, ?)",
		"2025-01-07 15:00:00", "8.00") // Tuesday
	receiptId2, _ := result.LastInsertId()
	tx.ExecContext(ctx, "INSERT INTO Purchases (name, price, receiptId, categoryId) VALUES (?, ?, ?, ?)",
		"Vegetables", "3.00", receiptId2, 3)
	tx.Commit()

	// Receipt 3: Monday, February
	tx, _ = db.Begin()
	result, _ = tx.ExecContext(ctx, "INSERT INTO Receipts (date, amount) VALUES (?, ?)",
		"2025-02-03 10:00:00", "12.00") // Monday
	receiptId3, _ := result.LastInsertId()
	tx.ExecContext(ctx, "INSERT INTO Purchases (name, price, receiptId, categoryId) VALUES (?, ?, ?, ?)",
		"Milk", "2.50", receiptId3, 1) // Dairy again
	tx.Commit()

	// Test prediction for a Monday in January
	targetTime := time.Date(2025, 1, 13, 10, 0, 0, 0, time.Local) // Monday

	scores, err := getTimeBasedRecommendations(db, targetTime)
	if err != nil {
		t.Fatalf("getTimeBasedRecommendations() error = %v", err)
	}

	if len(scores) == 0 {
		t.Fatal("Expected some category scores, got none")
	}

	// Verify we have scores for the categories
	categoryMap := make(map[int]float64)
	for _, score := range scores {
		categoryMap[score.CategoryID] = score.Score
	}

	// Dairy (category 1) should have a positive score since it was bought on Mondays
	if _, exists := categoryMap[1]; !exists {
		t.Error("Expected score for Dairy category (ID 1)")
	}

	// All scores should be positive
	for _, score := range scores {
		if score.Score < 0 {
			t.Errorf("Category %d has negative score: %f", score.CategoryID, score.Score)
		}
	}
}

func TestGetTimeBasedRecommendationsEmptyDB(t *testing.T) {
	db, cleanup := setupTestDBForServices(t)
	defer cleanup()

	targetTime := time.Date(2025, 1, 13, 10, 0, 0, 0, time.Local)

	scores, err := getTimeBasedRecommendations(db, targetTime)
	if err != nil {
		t.Fatalf("getTimeBasedRecommendations() error = %v", err)
	}

	if len(scores) != 0 {
		t.Errorf("Expected 0 scores for empty database, got %d", len(scores))
	}
}

func TestGetTimeBasedRecommendationsDifferentWeekdays(t *testing.T) {
	db, cleanup := setupTestDBForServices(t)
	defer cleanup()

	ctx := context.Background()

	// Add purchases on specific weekdays
	// Monday - Dairy
	tx, _ := db.Begin()
	result, _ := tx.ExecContext(ctx, "INSERT INTO Receipts (date, amount) VALUES (?, ?)",
		"2025-01-06 10:00:00", "5.00") // Monday
	receiptId, _ := result.LastInsertId()
	tx.ExecContext(ctx, "INSERT INTO Purchases (name, price, receiptId, categoryId) VALUES (?, ?, ?, ?)",
		"Milk", "2.50", receiptId, 1)
	tx.Commit()

	// Friday - Meat
	tx, _ = db.Begin()
	result, _ = tx.ExecContext(ctx, "INSERT INTO Receipts (date, amount) VALUES (?, ?)",
		"2025-01-10 17:00:00", "8.00") // Friday
	receiptId, _ = result.LastInsertId()
	tx.ExecContext(ctx, "INSERT INTO Purchases (name, price, receiptId, categoryId) VALUES (?, ?, ?, ?)",
		"Steak", "8.00", receiptId, 2)
	tx.Commit()

	// Test for Monday - should have stronger score for Dairy
	mondayTarget := time.Date(2025, 1, 13, 10, 0, 0, 0, time.Local) // Monday
	mondayScores, err := getTimeBasedRecommendations(db, mondayTarget)
	if err != nil {
		t.Fatalf("getTimeBasedRecommendations() error = %v", err)
	}

	mondayMap := make(map[int]float64)
	for _, score := range mondayScores {
		mondayMap[score.CategoryID] = score.Score
	}

	// Test for Friday - should have stronger score for Meat
	fridayTarget := time.Date(2025, 1, 17, 17, 0, 0, 0, time.Local) // Friday
	fridayScores, err := getTimeBasedRecommendations(db, fridayTarget)
	if err != nil {
		t.Fatalf("getTimeBasedRecommendations() error = %v", err)
	}

	fridayMap := make(map[int]float64)
	for _, score := range fridayScores {
		fridayMap[score.CategoryID] = score.Score
	}

	// On Monday, Dairy should have higher score than on Friday
	if mondayMap[1] <= fridayMap[1] {
		t.Error("Expected Dairy to have higher score on Monday than Friday")
	}

	// On Friday, Meat should have higher score than on Monday
	if fridayMap[2] <= mondayMap[2] {
		t.Error("Expected Meat to have higher score on Friday than Monday")
	}
}
