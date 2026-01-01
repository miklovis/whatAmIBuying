package services

import (
	"testing"
)

func TestRemoveThinkTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No think tags",
			input:    "This is a simple response",
			expected: "This is a simple response",
		},
		{
			name:     "Think tags present",
			input:    "<think>Some reasoning here</think>This is the answer",
			expected: "This is the answer",
		},
		{
			name:     "Think tags with whitespace",
			input:    "  <think>Reasoning</think>  Answer with spaces  ",
			expected: "Answer with spaces",
		},
		{
			name:     "Unclosed think tag",
			input:    "<think>Some reasoning without closing tag. This is the answer",
			expected: "",
		},
		{
			name:     "Multiple think tags",
			input:    "<think>First</think>Answer<think>Second</think>",
			expected: "Answer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveThinkTags(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveThinkTags() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseLLMResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{
			name:     "Standard JSON",
			input:    `{"ID": 5}`,
			expected: 5,
			wantErr:  false,
		},
		{
			name:     "JSON with spaces",
			input:    `{ "ID" : 3 }`,
			expected: 3,
			wantErr:  false,
		},
		{
			name:     "JSON in markdown code block",
			input:    "```json\n{\"ID\": 2}\n```",
			expected: 2,
			wantErr:  false,
		},
		{
			name:     "JSON-like pattern with single quotes",
			input:    "{'ID': 4}",
			expected: 4,
			wantErr:  false,
		},
		{
			name:     "Plain text with ID",
			input:    "The category ID is: 7",
			expected: 7,
			wantErr:  false,
		},
		{
			name:     "ID without colon",
			input:    "ID 6",
			expected: 6,
			wantErr:  false,
		},
		{
			name:     "Just a number in valid range",
			input:    "8",
			expected: 8,
			wantErr:  false,
		},
		{
			name:     "Number out of range",
			input:    "999",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "No valid ID",
			input:    "This has no category information",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "Complex response with JSON",
			input:    "After analyzing the purchase, I believe it belongs to category:\n```json\n{\"ID\": 1}\n```\nThis is because it's a dairy product.",
			expected: 1,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseLLMResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLLMResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseLLMResponse() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseLLMResponseEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Only whitespace",
			input:   "   \n\t  ",
			wantErr: true,
		},
		{
			name:    "Malformed JSON",
			input:   `{"ID": "not_a_number"}`,
			wantErr: true,
		},
		{
			name:    "Negative number",
			input:   `{"ID": -1}`,
			wantErr: true,
		},
		{
			name:    "Zero",
			input:   `{"ID": 0}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseLLMResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLLMResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
