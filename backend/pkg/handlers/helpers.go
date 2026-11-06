package handlers

import (
	"encoding/json"
	"net/http"

	"socialnetwork/pkg/utils"
)

// decodeJSON decodes a JSON request body into dst, writing a 400 on failure.
// It returns false when decoding failed so the caller can stop.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}
