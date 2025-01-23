// Package handlers contains HTTP handlers used in the project.
package handlers

import (
	"net/http"
	"strconv"
	"strings"
)

// UpdateHandler is the handler for the /update/ endpoint.
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	p := strings.Split(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/update/"), "/"), "/")

	if len(p) < 2 {
		http.Error(w, "Bad request, but the spec dictates a 404.", http.StatusNotFound)
		return
	}

	if len(p) != 3 {
		return
	}

	switch p[0] {
	case "gauge":
		handleGauge(w, r, p[1], p[2])
	case "counter":
		handleCounter(w, r, p[1], p[2])
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}

func handleGauge(w http.ResponseWriter, r *http.Request, name, value string) {
	_, err := strconv.ParseFloat(value, 64)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleCounter(w http.ResponseWriter, r *http.Request, name, value string) {
	_, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
