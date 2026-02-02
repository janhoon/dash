package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/janhoon/dash/backend/pkg/prometheus"
)

type PrometheusHandler struct {
	prometheusURL string
}

func NewPrometheusHandler(prometheusURL string) *PrometheusHandler {
	return &PrometheusHandler{
		prometheusURL: prometheusURL,
	}
}

type QueryRequest struct {
	Query string `json:"query"`
	Start int64  `json:"start"` // Unix timestamp in seconds
	End   int64  `json:"end"`   // Unix timestamp in seconds
	Step  int64  `json:"step"`  // Step interval in seconds
}

type ErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// Query executes a PromQL range query
func (h *PrometheusHandler) Query(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	query := r.URL.Query().Get("query")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status: "error",
			Error:  "query parameter is required",
		})
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	stepStr := r.URL.Query().Get("step")

	// Parse start time (default: 1 hour ago)
	var start time.Time
	if startStr != "" {
		startUnix, err := strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Status: "error",
				Error:  "invalid start timestamp",
			})
			return
		}
		start = time.Unix(startUnix, 0)
	} else {
		start = time.Now().Add(-1 * time.Hour)
	}

	// Parse end time (default: now)
	var end time.Time
	if endStr != "" {
		endUnix, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Status: "error",
				Error:  "invalid end timestamp",
			})
			return
		}
		end = time.Unix(endUnix, 0)
	} else {
		end = time.Now()
	}

	// Parse step (default: 15 seconds)
	var step time.Duration
	if stepStr != "" {
		stepSec, err := strconv.ParseInt(stepStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Status: "error",
				Error:  "invalid step value",
			})
			return
		}
		step = time.Duration(stepSec) * time.Second
	} else {
		step = 15 * time.Second
	}

	// Create Prometheus client
	client, err := prometheus.NewClient(h.prometheusURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status: "error",
			Error:  "failed to create Prometheus client: " + err.Error(),
		})
		return
	}

	// Execute query
	result, err := client.QueryRange(r.Context(), query, start, end, step)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status: "error",
			Error:  "query execution failed: " + err.Error(),
		})
		return
	}

	// Check if the query itself returned an error
	if result.Status == "error" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(result)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
