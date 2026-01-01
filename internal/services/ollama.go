package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// OllamaRequest represents the request structure for Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents the response from Ollama API
type OllamaResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	DoneReason         string `json:"done_reason"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

// CallOllama sends a prompt to your local Ollama instance
func CallOllama(modelName string, prompt string) (string, error) {
	// Create request payload
	reqBody := OllamaRequest{
		Model:  modelName,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	// Send request to Ollama API
	resp, err := http.Post("http://localhost:11434/api/generate",
		"application/json",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error calling Ollama API: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	// Parse response
	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		// Try to clean the response if it's malformed
		cleanBody := bytes.TrimSpace(body)
		if len(cleanBody) > 0 {
			if err := json.Unmarshal(cleanBody, &ollamaResp); err != nil {
				return "", fmt.Errorf("error unmarshaling response: %w", err)
			}
		} else {
			return "", fmt.Errorf("error unmarshaling response: %w", err)
		}
	}

	response := RemoveThinkTags(ollamaResp.Response)

	return response, nil
}

// RemoveThinkTagContent removes content inside <think> tags from the response
func RemoveThinkTags(response string) string {
	// Remove all content inside <think> tags (handle multiple pairs)
	for {
		start := strings.Index(response, "<think>")
		if start == -1 {
			break
		}
		end := strings.Index(response[start:], "</think>")
		if end != -1 {
			response = response[:start] + response[start+end+len("</think>"):]
		} else {
			response = response[:start]
			break
		}
	}

	// Remove any remaining <think> tags
	response = strings.ReplaceAll(response, "<think>", "")
	response = strings.ReplaceAll(response, "</think>", "")
	// Trim leading and trailing whitespace
	response = strings.TrimSpace(response)
	return response
}

// ParseLLMResponse attempts to parse the LLM response into a structured format
func ParseLLMResponse(response string) (int, error) {
	// First try: standard JSON parsing
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err == nil {
		if id, ok := result["ID"].(float64); ok {
			idInt := int(id)
			if idInt >= 1 && id == float64(idInt) {
				return idInt, nil
			}
		}
	}

	// Second try: extract JSON from markdown code blocks
	jsonBlockPattern := "```json\\s*({.*?})\\s*```"
	re := regexp.MustCompile(jsonBlockPattern)
	matches := re.FindStringSubmatch(response)
	if len(matches) > 1 {
		var extractedResult map[string]interface{}
		if err := json.Unmarshal([]byte(matches[1]), &extractedResult); err == nil {
			if id, ok := extractedResult["ID"].(float64); ok {
				idInt := int(id)
				if idInt >= 1 && id == float64(idInt) {
					return idInt, nil
				}
			}
		}
	}

	// Third try: look for any JSON-like pattern using regex
	jsonPattern := "{\\s*['\"]?ID['\"]?\\s*:\\s*(-?[0-9]+)\\s*}"
	re = regexp.MustCompile(jsonPattern)
	matches = re.FindStringSubmatch(response)
	if len(matches) > 1 {
		id, err := strconv.Atoi(matches[1])
		if err == nil && id >= 1 {
			return id, nil
		}
	}

	// Fourth try: look for ID followed by a number
	idPattern := "ID\\s*:?\\s*(-?[0-9]+)"
	re = regexp.MustCompile(idPattern)
	matches = re.FindStringSubmatch(response)
	if len(matches) > 1 {
		id, err := strconv.Atoi(matches[1])
		if err == nil && id >= 1 {
			return id, nil
		}
	}

	// Finally: attempt to find any number in the response as a last resort
	// Use negative lookbehind to avoid matching numbers preceded by minus sign
	numPattern := "(?:^|[^-])([0-9]+)"
	re = regexp.MustCompile(numPattern)
	matches = re.FindStringSubmatch(response)
	if len(matches) > 1 {
		id, err := strconv.Atoi(matches[1])
		if err == nil && id >= 1 && id <= 16 {
			return id, nil
		}
	}

	return 0, fmt.Errorf("could not extract a valid category ID from response: %s", response)
}
