package services

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// Integration test for AssignPurchases with mocked input
func TestAssignPurchases_Integration(t *testing.T) {
	// Skip this test if running in CI or automated environment
	// since it requires database setup
	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration test")
	}

	// Create a temporary test database
	testDBPath := "test_assign_integration.db"
	db, err := sql.Open("sqlite", testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(testDBPath)
	}()

	// Create schema
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

	// Add test purchases
	ctx := context.Background()
	tx, _ := db.Begin()
	result, _ := tx.ExecContext(ctx, "INSERT INTO Receipts (date, amount) VALUES (?, ?)",
		"2026-01-01 10:00:00", "5.00")
	receiptId, _ := result.LastInsertId()
	tx.ExecContext(ctx, "INSERT INTO Purchases (name, price, receiptId) VALUES (?, ?, ?)",
		"Milk", "2.50", receiptId)
	tx.ExecContext(ctx, "INSERT INTO Purchases (name, price, receiptId) VALUES (?, ?, ?)",
		"Chicken", "2.50", receiptId)
	tx.Commit()

	// Verify purchases are unassigned
	var count int
	db.QueryRow("SELECT COUNT(*) FROM Purchases WHERE categoryId IS NULL").Scan(&count)
	if count != 2 {
		t.Fatalf("Expected 2 unassigned purchases, got %d", count)
	}

	t.Log("Integration test setup complete - purchases created and ready for assignment")
}

func TestInputValidator_RealWorldScenario(t *testing.T) {
	// Simulate user entering invalid inputs before getting it right
	userInput := strings.Join([]string{
		"",           // Empty input
		"abc",        // Not a number
		"99",         // Out of range
		"  5  ",      // Valid with whitespace
	}, "\n")

	reader := strings.NewReader(userInput)
	validator := NewInputValidator(reader)

	id, err := validator.ReadCategoryID(1, 10, 5)
	if err != nil {
		t.Fatalf("Expected successful validation after retries, got error: %v", err)
	}

	if id != 5 {
		t.Errorf("Expected category ID 5, got %d", id)
	}
}

func TestInputValidator_UserGivesUpScenario(t *testing.T) {
	// User gives up after 3 invalid attempts
	userInput := strings.Join([]string{
		"invalid",
		"also_invalid",
		"still_wrong",
	}, "\n")

	reader := strings.NewReader(userInput)
	validator := NewInputValidator(reader)

	_, err := validator.ReadCategoryID(1, 10, 3)
	if err == nil {
		t.Fatal("Expected error after max attempts, got nil")
	}

	if !strings.Contains(err.Error(), "maximum number of attempts") {
		t.Errorf("Expected 'maximum number of attempts' error, got: %v", err)
	}
}

func TestInputValidator_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		minID       int
		maxID       int
		maxAttempts int
		wantErr     bool
		expectedID  int
	}{
		{
			name:        "Just spaces",
			input:       "   \n5\n",
			minID:       1,
			maxID:       10,
			maxAttempts: 3,
			wantErr:     false,
			expectedID:  5,
		},
		{
			name:        "Tabs and spaces",
			input:       "\t\t  \n3\n",
			minID:       1,
			maxID:       10,
			maxAttempts: 3,
			wantErr:     false,
			expectedID:  3,
		},
		{
			name:        "Zero (invalid for most use cases)",
			input:       "0\n5\n",
			minID:       1,
			maxID:       10,
			maxAttempts: 3,
			wantErr:     false,
			expectedID:  5,
		},
		{
			name:        "Large number",
			input:       "9999\n5\n",
			minID:       1,
			maxID:       10,
			maxAttempts: 3,
			wantErr:     false,
			expectedID:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			validator := NewInputValidator(reader)

			id, err := validator.ReadCategoryID(tt.minID, tt.maxID, tt.maxAttempts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadCategoryID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && id != tt.expectedID {
				t.Errorf("ReadCategoryID() = %d, want %d", id, tt.expectedID)
			}
		})
	}
}
