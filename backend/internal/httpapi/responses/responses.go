package responses

import (
	"encoding/json"
	"net/http"
)

// JSON writes payload as a JSON response.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if payload == nil {
		return
	}

	// Headers are already sent; a failing encode can only be reported to the log.
	_ = json.NewEncoder(w).Encode(payload)
}

// Error writes a JSON error message.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}
