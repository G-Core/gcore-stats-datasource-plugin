package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type apiErrorShape struct {
	Error   string `json:"error"`
	Detail  string `json:"detail"`
	Message string `json:"message"`
}

func extractErrorMessage(body []byte) string {
	s := strings.TrimSpace(string(body))
	if s == "" {
		return ""
	}
	var parsed apiErrorShape
	if json.Unmarshal(body, &parsed) == nil {
		if parsed.Error != "" {
			return parsed.Error
		}
		if parsed.Detail != "" {
			return parsed.Detail
		}
		if parsed.Message != "" {
			return parsed.Message
		}
	}
	return s
}

func HandleAPIError(statusCode int, body []byte) error {
	if statusCode == http.StatusOK || statusCode == http.StatusNoContent {
		return nil
	}

	errorMsg := extractErrorMessage(body)

	switch statusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("bad request (400): %s. Please check your query parameters", errorMsg)
	case http.StatusUnauthorized:
		return fmt.Errorf("datasource authentication error: invalid API key. Please verify your credentials in the datasource settings")
	case http.StatusForbidden:
		return fmt.Errorf("forbidden (403): access denied. You may not have permissions to access this resource")
	case http.StatusNotFound:
		return fmt.Errorf("not found (404): resource not found. Check if the URL/resource exists")
	case http.StatusTooManyRequests:
		return fmt.Errorf("too many requests (429): rate limit exceeded. Please reduce query frequency")
	}

	if statusCode >= 500 {
		return fmt.Errorf("server error (%d): %s. This is an issue with the Gcore API", statusCode, errorMsg)
	}

	return fmt.Errorf("gcore api error (%d): %s", statusCode, errorMsg)
}

