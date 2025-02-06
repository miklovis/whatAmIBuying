package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	cmd := exec.Command("python", "C:\\Users\\arnas\\OneDrive\\Documents\\Receipts\\preprocess_image.py")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error preprocessing image:", err)
		return
	}

	// Read the output file
	data, err := ioutil.ReadFile("C:\\Users\\arnas\\OneDrive\\Documents\\Receipts\\output.txt")
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
	fmt.Println(content)

	// Print the total sum
	fmt.Printf("Total Sum: %.2f\n", totalSum)
}
