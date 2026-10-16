package requests

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const maxBodySize = 1 << 20 // 1 MiB

// DecodeJSON decodes a JSON request body into T, limiting the body size.
func DecodeJSON[T any](r *http.Request) (T, error) {
	var dto T

	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, maxBodySize))
	if err := decoder.Decode(&dto); err != nil {
		return dto, fmt.Errorf("decode request body: %w", err)
	}

	return dto, nil
}
