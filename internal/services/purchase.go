package services

import (
	"fmt"
	"log"
	"os"
	"whatAmIBuying/internal/database"
	"whatAmIBuying/internal/models"
)

func AssignPurchases() error {
	db, err := database.OpenDatabase()
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	var categories = database.GetAllCategories(db)

	if len(*categories) == 0 {
		return fmt.Errorf("no categories found in database")
	}

	// Find min and max category IDs
	minID := (*categories)[0].ID
	maxID := (*categories)[0].ID
	for _, category := range *categories {
		if category.ID < minID {
			minID = category.ID
		}
		if category.ID > maxID {
			maxID = category.ID
		}
	}

	fmt.Println("\nAssign the purchase to one of these categories: ")
	for _, category := range *categories {
		fmt.Printf("  [%d] %s\n", category.ID, category.Category)
	}
	fmt.Println()

	var purchasesWithNullCategoryId []models.Purchase
	purchasesWithNullCategoryId, err = database.GetUnassignedPurchases(db)
	if err != nil {
		return fmt.Errorf("getting unassigned purchases failed: %w", err)
	}

	if len(purchasesWithNullCategoryId) == 0 {
		fmt.Println("No unassigned purchases found.")
		return nil
	}

	// Create validator with stdin
	validator := NewInputValidator(os.Stdin)
	maxAttempts := 3

	for i, p := range purchasesWithNullCategoryId {
		fmt.Printf("[%d/%d] Which category does '%s' (£%s) belong to? ",
			i+1, len(purchasesWithNullCategoryId), p.Product, p.Price)

		id, err := validator.ReadCategoryID(minID, maxID, maxAttempts)
		if err != nil {
			return fmt.Errorf("failed to read category ID for '%s': %w", p.Product, err)
		}

		_, err = database.ChangePurchaseCategory(db, &id, &p.Id)
		if err != nil {
			return fmt.Errorf("changing purchase category failed: %w", err)
		}

		fmt.Printf("Assigned '%s' to category %d\n\n", p.Product, id)
	}

	fmt.Printf("Successfully assigned %d purchases!\n", len(purchasesWithNullCategoryId))
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
