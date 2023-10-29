package utils

import (
	"encoding/json"
	"net/http"
	"strings"
)

func WithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	return json.NewEncoder(w).Encode(payload)
}

// remove extra slashes in the end
func StripSlashes(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if len(path) > 1 && strings.HasSuffix(path, "/") {
			path = strings.TrimSuffix(path, "/")
		}

		r.URL.Path = path

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
