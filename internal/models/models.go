package models

import (
	"database/sql"
	"time"
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

type RawJsonData struct {
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

type PurchaseRecord struct {
	Purchase    Purchase
	ReceiptDate time.Time
}
