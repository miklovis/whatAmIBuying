package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	_ "modernc.org/sqlite"
)

/*func main() {
	cmd := exec.Command("python", "preprocess_image.py")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error preprocessing image:", err)
		return
	}

	// Read the output file
	data, err := ioutil.ReadFile("output.json")
	if err != nil {
		fmt.Println("Error reading output file:", err)
		return
	}

	// Convert data to string
	content := string(data)

	// Remove empty lines
	lines := strings.Split(content, "\n")
	var filteredLines []string
	var totalSum float64
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove letters after the prices
		cleanedLine := ""
		for _, char := range line {
			if (char >= '0' && char <= '9') || char == '.' || char == '-' {
				cleanedLine += string(char)
			} else {
				break
			}
		}

		// Convert to float
		price, err := strconv.ParseFloat(cleanedLine, 64)
		if err == nil {
			filteredLines = append(filteredLines, fmt.Sprintf("%.2f", price))
			totalSum += price
		}
	}
	content = strings.Join(filteredLines, "\n")

	// Print the modified contents of the output file
	fmt.Println(content),

	// Print the total sum
	fmt.Printf("Total Sum: %.2f\n", totalSum)
}*/

type Purchase struct {
	Product    string
	Price      string
	PriceFloat float64
}

type Category struct {
	ID       int
	Category string
}

func main() {
	var purchases []Purchase

	data, err := os.ReadFile("output.json")
	if err != nil {
		fmt.Println("Error reading output file:", err)
		return
	}

	var priceMap map[string]string
	err = json.Unmarshal(data, &priceMap)
	if err != nil {
		fmt.Println("Error decoding json:", err)
		return
	}

	for product, price := range priceMap {
		purchases = append(purchases, Purchase{Product: product, Price: price})

		var newPurchase Purchase
		newPurchase.Product = product
		newPurchase.Price = price
		newPurchase.PriceFloat, err = strconv.ParseFloat(price, 64)

		if err == nil {
			purchases = append(purchases, newPurchase)
		}

	}

	var sum float64
	for i := range purchases {
		fmt.Println(purchases[i].Product + " " + purchases[i].Price + " " + fmt.Sprintf("%.2f", sum))
		sum += purchases[i].PriceFloat
	}

	db, err := sql.Open("sqlite", "test_database.db")
	if err != nil {
		fmt.Println(err)
	}

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

	for _, category := range categories {
		fmt.Printf("%+v\n", category)
	}

	fmt.Printf("Total categories: %d", len(categories))
}
