package services

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// InputValidator handles validation of user input
type InputValidator struct {
	reader io.Reader
}

// NewInputValidator creates a new input validator with the given reader
func NewInputValidator(reader io.Reader) *InputValidator {
	return &InputValidator{reader: reader}
}

// ValidateCategoryID checks if a category ID is within valid range
func ValidateCategoryID(id int, minID int, maxID int) error {
	if id < minID {
		return fmt.Errorf("category ID %d is below minimum allowed value %d", id, minID)
	}
	if id > maxID {
		return fmt.Errorf("category ID %d exceeds maximum allowed value %d", id, maxID)
	}
	return nil
}

// SanitizeInput trims whitespace and converts to lowercase for comparison
func SanitizeInput(input string) string {
	return strings.TrimSpace(input)
}

// ReadCategoryID reads and validates a category ID from user input
// Returns the validated ID or an error
func (v *InputValidator) ReadCategoryID(minID int, maxID int, maxAttempts int) (int, error) {
	scanner := bufio.NewScanner(v.reader)
	attempts := 0

	for attempts < maxAttempts {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return 0, fmt.Errorf("error reading input: %w", err)
			}
			return 0, fmt.Errorf("no input received")
		}

		input := SanitizeInput(scanner.Text())
		
		// Check for empty input
		if input == "" {
			attempts++
			if attempts < maxAttempts {
				fmt.Printf("Error: Input cannot be empty. Please try again (%d/%d): ", attempts, maxAttempts)
			}
			continue
		}

		// Try to parse as integer
		id, err := strconv.Atoi(input)
		if err != nil {
			attempts++
			if attempts < maxAttempts {
				fmt.Printf("Error: '%s' is not a valid number. Please enter a valid category ID (%d/%d): ", input, attempts, maxAttempts)
			}
			continue
		}

		// Validate the ID range
		if err := ValidateCategoryID(id, minID, maxID); err != nil {
			attempts++
			if attempts < maxAttempts {
				fmt.Printf("Error: %v. Please try again (%d/%d): ", err, attempts, maxAttempts)
			}
			continue
		}

		return id, nil
	}

	return 0, fmt.Errorf("maximum number of attempts (%d) exceeded", maxAttempts)
}

// IsValidCategoryIDString checks if a string can be parsed as a valid category ID
func IsValidCategoryIDString(input string, minID int, maxID int) (int, bool) {
	input = SanitizeInput(input)
	
	if input == "" {
		return 0, false
	}

	id, err := strconv.Atoi(input)
	if err != nil {
		return 0, false
	}

	if err := ValidateCategoryID(id, minID, maxID); err != nil {
		return 0, false
	}

	return id, true
}
