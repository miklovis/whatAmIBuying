package database

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"whatAmIBuying/internal/models"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Create a temporary test database
	testDBPath := "test_temp.db"

	db, err := sql.Open("sqlite", testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create tables
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
	categories := []string{"Dairy", "Meat", "Vegetables", "Fruit", "Snacks"}
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

func TestOpenDatabase(t *testing.T) {
	db, err := OpenDatabase()
	if err != nil {
		t.Fatalf("OpenDatabase() error = %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

func TestAddReceipt(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	receipt := models.Receipt{
		Date:   "2026-01-01 10:00:00",
		Amount: "15.50",
		Purchases: []models.Purchase{
			{Product: "Milk", Price: "2.50", PriceFloat: 2.50},
			{Product: "Bread", Price: "1.20", PriceFloat: 1.20},
			{Product: "Eggs", Price: "3.00", PriceFloat: 3.00},
		},
	}

	id, err := AddReceipt(receipt, db)
	if err != nil {
		t.Fatalf("AddReceipt() error = %v", err)
	}

	if id <= 0 {
		t.Errorf("AddReceipt() returned invalid id = %d", id)
	}

	// Verify receipt was added
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM Receipts WHERE id = ?", id).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query receipt: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 receipt, got %d", count)
	}

	// Verify purchases were added
	err = db.QueryRow("SELECT COUNT(*) FROM Purchases WHERE receiptId = ?", id).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query purchases: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 purchases, got %d", count)
	}
}

func TestAddPurchase(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// First add a receipt
	receipt := models.Receipt{
		Date:      "2026-01-01 10:00:00",
		Amount:    "5.00",
		Purchases: []models.Purchase{},
	}
	receiptId, _ := AddReceipt(receipt, db)

	ctx := context.Background()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	purchase := models.Purchase{
		Product:    "Test Product",
		Price:      "5.99",
		PriceFloat: 5.99,
	}

	id, err := AddPurchase(purchase, receiptId, ctx, tx)
	if err != nil {
		t.Fatalf("AddPurchase() error = %v", err)
	}

	if id <= 0 {
		t.Errorf("AddPurchase() returned invalid id = %d", id)
	}

	tx.Commit()

	// Verify purchase was added
	var name string
	err = db.QueryRow("SELECT name FROM Purchases WHERE id = ?", id).Scan(&name)
	if err != nil {
		t.Fatalf("Failed to query purchase: %v", err)
	}
	if name != "Test Product" {
		t.Errorf("Expected 'Test Product', got '%s'", name)
	}
}

func TestGetUnassignedPurchases(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Add a receipt with purchases
	receipt := models.Receipt{
		Date:   "2026-01-01 10:00:00",
		Amount: "10.00",
		Purchases: []models.Purchase{
			{Product: "Unassigned Item 1", Price: "5.00"},
			{Product: "Unassigned Item 2", Price: "5.00"},
		},
	}
	AddReceipt(receipt, db)

	purchases, err := GetUnassignedPurchases(db)
	if err != nil {
		t.Fatalf("GetUnassignedPurchases() error = %v", err)
	}

	if len(purchases) != 2 {
		t.Errorf("Expected 2 unassigned purchases, got %d", len(purchases))
	}
}

func TestChangePurchaseCategory(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Add a receipt with a purchase
	receipt := models.Receipt{
		Date:   "2026-01-01 10:00:00",
		Amount: "5.00",
		Purchases: []models.Purchase{
			{Product: "Test Item", Price: "5.00"},
		},
	}
	AddReceipt(receipt, db)

	// Get the purchase
	purchases, _ := GetUnassignedPurchases(db)
	if len(purchases) == 0 {
		t.Fatal("No purchases found")
	}

	purchaseId := purchases[0].Id
	categoryId := 1 // Dairy

	result, err := ChangePurchaseCategory(db, &categoryId, &purchaseId)
	if err != nil {
		t.Fatalf("ChangePurchaseCategory() error = %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", rowsAffected)
	}

	// Verify category was updated
	var catId sql.NullInt64
	err = db.QueryRow("SELECT categoryId FROM Purchases WHERE id = ?", purchaseId).Scan(&catId)
	if err != nil {
		t.Fatalf("Failed to query purchase: %v", err)
	}
	if !catId.Valid || int(catId.Int64) != categoryId {
		t.Errorf("Expected categoryId = %d, got %d", categoryId, catId.Int64)
	}
}

func TestGetAllCategories(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	categories := GetAllCategories(db)

	if categories == nil {
		t.Fatal("GetAllCategories() returned nil")
	}

	if len(*categories) != 5 {
		t.Errorf("Expected 5 categories, got %d", len(*categories))
	}

	// Check first category
	if len(*categories) > 0 {
		firstCat := (*categories)[0]
		if firstCat.Category != "Dairy" {
			t.Errorf("Expected first category to be 'Dairy', got '%s'", firstCat.Category)
		}
	}
}

func TestGetCategoryNameByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name        string
		id          int
		wantName    string
		wantErr     bool
	}{
		{
			name:     "Valid category ID",
			id:       1,
			wantName: "Dairy",
			wantErr:  false,
		},
		{
			name:     "Invalid category ID",
			id:       999,
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := GetCategoryNameByID(db, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCategoryNameByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if name != tt.wantName {
				t.Errorf("GetCategoryNameByID() = %v, want %v", name, tt.wantName)
			}
		})
	}
}
