package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"whatAmIBuying/internal/database"
	"whatAmIBuying/internal/models"
)

func ReadReceipts() {
	db, err := database.OpenDatabase()
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	fileContent, err := os.ReadFile("output.json")
	if err != nil {
		fmt.Println("Error reading output file:", err)
		return
	}

	var raw models.RawJsonData
	err = json.Unmarshal(fileContent, &raw)
	if err != nil {
		log.Fatal("Error unmarshalling JSON:", err)
	}

	var data models.Receipt
	data.Date = raw.Date

	for product, price := range raw.Values {
		var newPurchase models.Purchase
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

	id, err := database.AddReceipt(data, db)
	if err != nil {
		log.Fatal("Error adding receipt: ", err)
	}

	fmt.Printf("Added receipt with id %d", id)

}
