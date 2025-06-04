package services

import (
	"fmt"
	"log"
	"whatAmIBuying/internal/database"
	"whatAmIBuying/internal/models"
)

func AssignPurchases() error {
	db, err := database.OpenDatabase()
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	var categories = database.GetAllCategories(db)

	fmt.Println("Assign the purchase to one of these categories: ")
	for _, category := range *categories {
		fmt.Printf("ID: %d, Category: %s \n", category.ID, category.Category)
	}

	var purchasesWithNullCategoryId []models.Purchase
	purchasesWithNullCategoryId, err = database.GetUnassignedPurchases(db)
	if err != nil {
		return fmt.Errorf("getting unassigned purchases failed: %w", err)
	}

	for _, p := range purchasesWithNullCategoryId {
		var id int

		fmt.Printf("Which category does %s bought for %s belong to? ", p.Product, p.Price)
		fmt.Scan(&id)

		_, err = database.ChangePurchaseCategory(db, &id, &p.Id)
		if err != nil {
			return fmt.Errorf("changing purchase category failed: %w", err)
		}
	}

	return nil
}

func TestLLM() {
	// Example: Call Ollama with your deepseek model
	db, err := database.OpenDatabase()
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	var purchasesWithNullCategoryId []models.Purchase
	purchasesWithNullCategoryId, err = database.GetUnassignedPurchases(db)
	if err != nil {
		fmt.Errorf("getting unassigned purchases failed: %w", err)
	}

	var categories = database.GetAllCategories(db)

	prompt := `You are an AI assistant that helps to categorize purchases. 

TASK: Categorize the following purchase into one of the available categories.

IMPORTANT INSTRUCTIONS:
1. Take your time to think carefully about what this product actually is.
2. Consider specific keywords and context clues in the purchase description.
3. If the item contains multiple ingredients or components, focus on the main ingredient.
4. For prepared foods, categorize based on the primary component.

REQUIRED RESPONSE FORMAT:
Your final answer MUST be provided in valid JSON format with a single 'ID' field containing the category ID as a number. Example: {"ID": 1}

DO NOT include any explanations, reasoning, or additional text in your output - ONLY the JSON object.

`
	var categoryListString string
	categoryListString = "Available categories: \n"
	for _, category := range *categories {
		categoryListString += fmt.Sprintf("ID: %d, Category: %s \n", category.ID, category.Category)
	}

	prompt += categoryListString + "\n"
	for _, p := range purchasesWithNullCategoryId {
		fmt.Println(prompt)
		response, err := CallOllama("deepseek-r1:7b", prompt+p.Product+" bought for "+p.Price)

		if err != nil {
			log.Printf("Error calling Ollama: %v", err)
			continue
		} else {
			id, err := ParseLLMResponse(response)
			if err != nil {
				log.Printf("Error parsing LLM response: %v", err)
				continue
			}
			categoryName, err := database.GetCategoryNameByID(db, id)
			if err != nil {
				log.Printf("Error getting category name by ID: %v", err)
				continue
			}
			fmt.Printf("Parsed category ID: %d, category name: %s for purchase %s bought for %s\n", id, categoryName, p.Product, p.Price)

			_, err = database.ChangePurchaseCategory(db, &id, &p.Id)
			if err != nil {
				log.Printf("Error changing purchase category: %v", err)
				continue
			}
		}

	}
}
