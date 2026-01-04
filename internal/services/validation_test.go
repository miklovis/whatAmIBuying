package services

import (
	"strings"
	"testing"
)

func TestValidateCategoryID(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		minID   int
		maxID   int
		wantErr bool
	}{
		{
			name:    "Valid ID in range",
			id:      5,
			minID:   1,
			maxID:   10,
			wantErr: false,
		},
		{
			name:    "Valid ID at minimum",
			id:      1,
			minID:   1,
			maxID:   10,
			wantErr: false,
		},
		{
			name:    "Valid ID at maximum",
			id:      10,
			minID:   1,
			maxID:   10,
			wantErr: false,
		},
		{
			name:    "ID below minimum",
			id:      0,
			minID:   1,
			maxID:   10,
			wantErr: true,
		},
		{
			name:    "ID above maximum",
			id:      11,
			minID:   1,
			maxID:   10,
			wantErr: true,
		},
		{
			name:    "Negative ID",
			id:      -5,
			minID:   1,
			maxID:   10,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCategoryID(tt.id, tt.minID, tt.maxID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCategoryID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No whitespace",
			input:    "5",
			expected: "5",
		},
		{
			name:     "Leading whitespace",
			input:    "  5",
			expected: "5",
		},
		{
			name:     "Trailing whitespace",
			input:    "5  ",
			expected: "5",
		},
		{
			name:     "Both sides whitespace",
			input:    "  5  ",
			expected: "5",
		},
		{
			name:     "Tabs and newlines",
			input:    "\t5\n",
			expected: "5",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only whitespace",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeInput() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIsValidCategoryIDString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		minID     int
		maxID     int
		expectedID int
		expectedValid bool
	}{
		{
			name:      "Valid number in range",
			input:     "5",
			minID:     1,
			maxID:     10,
			expectedID: 5,
			expectedValid: true,
		},
		{
			name:      "Valid number with whitespace",
			input:     "  3  ",
			minID:     1,
			maxID:     10,
			expectedID: 3,
			expectedValid: true,
		},
		{
			name:      "Invalid - not a number",
			input:     "abc",
			minID:     1,
			maxID:     10,
			expectedID: 0,
			expectedValid: false,
		},
		{
			name:      "Invalid - empty string",
			input:     "",
			minID:     1,
			maxID:     10,
			expectedID: 0,
			expectedValid: false,
		},
		{
			name:      "Invalid - below range",
			input:     "0",
			minID:     1,
			maxID:     10,
			expectedID: 0,
			expectedValid: false,
		},
		{
			name:      "Invalid - above range",
			input:     "11",
			minID:     1,
			maxID:     10,
			expectedID: 0,
			expectedValid: false,
		},
		{
			name:      "Invalid - negative number",
			input:     "-5",
			minID:     1,
			maxID:     10,
			expectedID: 0,
			expectedValid: false,
		},
		{
			name:      "Invalid - decimal number",
			input:     "5.5",
			minID:     1,
			maxID:     10,
			expectedID: 0,
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, valid := IsValidCategoryIDString(tt.input, tt.minID, tt.maxID)
			if valid != tt.expectedValid {
				t.Errorf("IsValidCategoryIDString() valid = %v, want %v", valid, tt.expectedValid)
			}
			if id != tt.expectedID {
				t.Errorf("IsValidCategoryIDString() id = %v, want %v", id, tt.expectedID)
			}
		})
	}
}

func TestReadCategoryID_ValidInput(t *testing.T) {
	input := "5\n"
	reader := strings.NewReader(input)
	validator := NewInputValidator(reader)

	id, err := validator.ReadCategoryID(1, 10, 3)
	if err != nil {
		t.Fatalf("ReadCategoryID() error = %v, want nil", err)
	}

	if id != 5 {
		t.Errorf("ReadCategoryID() = %d, want 5", id)
	}
}

func TestReadCategoryID_InvalidThenValid(t *testing.T) {
	input := "abc\n5\n"
	reader := strings.NewReader(input)
	validator := NewInputValidator(reader)

	id, err := validator.ReadCategoryID(1, 10, 3)
	if err != nil {
		t.Fatalf("ReadCategoryID() error = %v, want nil", err)
	}

	if id != 5 {
		t.Errorf("ReadCategoryID() = %d, want 5", id)
	}
}

func TestReadCategoryID_OutOfRangeThenValid(t *testing.T) {
	input := "20\n5\n"
	reader := strings.NewReader(input)
	validator := NewInputValidator(reader)

	id, err := validator.ReadCategoryID(1, 10, 3)
	if err != nil {
		t.Fatalf("ReadCategoryID() error = %v, want nil", err)
	}

	if id != 5 {
		t.Errorf("ReadCategoryID() = %d, want 5", id)
	}
}

func TestReadCategoryID_EmptyInput(t *testing.T) {
	input := "\n\n5\n"
	reader := strings.NewReader(input)
	validator := NewInputValidator(reader)

	id, err := validator.ReadCategoryID(1, 10, 3)
	if err != nil {
		t.Fatalf("ReadCategoryID() error = %v, want nil", err)
	}

	if id != 5 {
		t.Errorf("ReadCategoryID() = %d, want 5", id)
	}
}

func TestReadCategoryID_MaxAttemptsExceeded(t *testing.T) {
	input := "abc\ndef\nghi\njkl\n"
	reader := strings.NewReader(input)
	validator := NewInputValidator(reader)

	_, err := validator.ReadCategoryID(1, 10, 3)
	if err == nil {
		t.Fatal("ReadCategoryID() error = nil, want error")
	}

	expectedError := "maximum number of attempts (3) exceeded"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("ReadCategoryID() error = %v, want error containing %q", err, expectedError)
	}
}

func TestReadCategoryID_AllOutOfRange(t *testing.T) {
	input := "0\n20\n-5\n"
	reader := strings.NewReader(input)
	validator := NewInputValidator(reader)

	_, err := validator.ReadCategoryID(1, 10, 3)
	if err == nil {
		t.Fatal("ReadCategoryID() error = nil, want error")
	}
}

func TestReadCategoryID_NoInput(t *testing.T) {
	input := ""
	reader := strings.NewReader(input)
	validator := NewInputValidator(reader)

	_, err := validator.ReadCategoryID(1, 10, 3)
	if err == nil {
		t.Fatal("ReadCategoryID() error = nil, want error")
	}
}

func TestReadCategoryID_WithWhitespace(t *testing.T) {
	input := "  7  \n"
	reader := strings.NewReader(input)
	validator := NewInputValidator(reader)

	id, err := validator.ReadCategoryID(1, 10, 3)
	if err != nil {
		t.Fatalf("ReadCategoryID() error = %v, want nil", err)
	}

	if id != 7 {
		t.Errorf("ReadCategoryID() = %d, want 7", id)
	}
}

func TestReadCategoryID_BoundaryValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		minID    int
		maxID    int
		expected int
	}{
		{
			name:     "Minimum boundary",
			input:    "1\n",
			minID:    1,
			maxID:    10,
			expected: 1,
		},
		{
			name:     "Maximum boundary",
			input:    "10\n",
			minID:    1,
			maxID:    10,
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			validator := NewInputValidator(reader)

			id, err := validator.ReadCategoryID(tt.minID, tt.maxID, 3)
			if err != nil {
				t.Fatalf("ReadCategoryID() error = %v, want nil", err)
			}

			if id != tt.expected {
				t.Errorf("ReadCategoryID() = %d, want %d", id, tt.expected)
			}
		})
	}
}

func TestNewInputValidator(t *testing.T) {
	reader := strings.NewReader("test")
	validator := NewInputValidator(reader)

	if validator == nil {
		t.Fatal("NewInputValidator() returned nil")
	}

	if validator.reader != reader {
		t.Error("NewInputValidator() did not set reader correctly")
	}
}
