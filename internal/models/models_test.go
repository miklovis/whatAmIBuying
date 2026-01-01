package models

import (
	"database/sql"
	"testing"
)

func TestPurchaseStruct(t *testing.T) {
	purchase := Purchase{
		Id:         1,
		Product:    "Test Product",
		Price:      "9.99",
		PriceFloat: 9.99,
		ReceiptId:  10,
		CategoryId: sql.NullInt64{Int64: 5, Valid: true},
	}

	if purchase.Id != 1 {
		t.Errorf("Expected Id = 1, got %d", purchase.Id)
	}
	if purchase.Product != "Test Product" {
		t.Errorf("Expected Product = 'Test Product', got '%s'", purchase.Product)
	}
	if purchase.PriceFloat != 9.99 {
		t.Errorf("Expected PriceFloat = 9.99, got %f", purchase.PriceFloat)
	}
	if !purchase.CategoryId.Valid || purchase.CategoryId.Int64 != 5 {
		t.Errorf("Expected CategoryId = 5, got %v", purchase.CategoryId)
	}
}

func TestReceiptStruct(t *testing.T) {
	purchases := []Purchase{
		{Product: "Item 1", Price: "5.00", PriceFloat: 5.00},
		{Product: "Item 2", Price: "3.50", PriceFloat: 3.50},
	}

	receipt := Receipt{
		Date:      "2026-01-01 10:00:00",
		Purchases: purchases,
		Amount:    "8.50",
	}

	if receipt.Date != "2026-01-01 10:00:00" {
		t.Errorf("Expected Date = '2026-01-01 10:00:00', got '%s'", receipt.Date)
	}
	if len(receipt.Purchases) != 2 {
		t.Errorf("Expected 2 purchases, got %d", len(receipt.Purchases))
	}
	if receipt.Amount != "8.50" {
		t.Errorf("Expected Amount = '8.50', got '%s'", receipt.Amount)
	}
}

func TestRawJsonDataStruct(t *testing.T) {
	values := map[string]string{
		"Milk":  "2.50",
		"Bread": "1.20",
	}

	rawData := RawJsonData{
		Date:   "2026-01-01 10:00:00",
		Values: values,
		Amount: "3.70",
	}

	if rawData.Date != "2026-01-01 10:00:00" {
		t.Errorf("Expected Date = '2026-01-01 10:00:00', got '%s'", rawData.Date)
	}
	if len(rawData.Values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(rawData.Values))
	}
	if rawData.Values["Milk"] != "2.50" {
		t.Errorf("Expected Milk = '2.50', got '%s'", rawData.Values["Milk"])
	}
}

func TestCategoryStruct(t *testing.T) {
	category := Category{
		ID:       1,
		Category: "Dairy",
	}

	if category.ID != 1 {
		t.Errorf("Expected ID = 1, got %d", category.ID)
	}
	if category.Category != "Dairy" {
		t.Errorf("Expected Category = 'Dairy', got '%s'", category.Category)
	}
}

func TestCategoryScoreStruct(t *testing.T) {
	score := CategoryScore{
		CategoryID: 5,
		Score:      0.85,
	}

	if score.CategoryID != 5 {
		t.Errorf("Expected CategoryID = 5, got %d", score.CategoryID)
	}
	if score.Score != 0.85 {
		t.Errorf("Expected Score = 0.85, got %f", score.Score)
	}
}

func TestNullCategoryId(t *testing.T) {
	// Test with null category
	purchase1 := Purchase{
		Id:         1,
		Product:    "Uncategorized Item",
		Price:      "5.99",
		ReceiptId:  10,
		CategoryId: sql.NullInt64{Valid: false},
	}

	if purchase1.CategoryId.Valid {
		t.Error("Expected CategoryId to be invalid (null)")
	}

	// Test with valid category
	purchase2 := Purchase{
		Id:         2,
		Product:    "Categorized Item",
		Price:      "3.99",
		ReceiptId:  10,
		CategoryId: sql.NullInt64{Int64: 3, Valid: true},
	}

	if !purchase2.CategoryId.Valid {
		t.Error("Expected CategoryId to be valid")
	}
	if purchase2.CategoryId.Int64 != 3 {
		t.Errorf("Expected CategoryId = 3, got %d", purchase2.CategoryId.Int64)
	}
}
