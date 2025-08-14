package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}
	
	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

func WriteValidationError(w http.ResponseWriter, errors ValidationErrors) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	
	fields := make(map[string]string)
	for _, err := range errors {
		fields[err.Field] = err.Message
	}
	
	errorResp := ErrorResponse{
		Error:   "Validation Failed",
		Message: "Please check your input and try again",
		Fields:  fields,
	}
	
	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		log.Printf("Failed to encode validation error response: %v", err)
	}
}

func WriteHTMLError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(statusCode)
	
	html := `
		<div class="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
			<div class="flex items-center">
				<span class="material-icons text-red-600 mr-2">error</span>
				<div>
					<p class="font-medium text-red-800">Error</p>
					<p class="text-red-700">` + message + `</p>
				</div>
			</div>
		</div>
	`
	
	w.Write([]byte(html))
}

func LogError(operation string, err error) {
	log.Printf("Error in %s: %v", operation, err)
}

func RecoverFromPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				WriteError(w, http.StatusInternalServerError, "An unexpected error occurred")
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}