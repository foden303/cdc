package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/foden/cdc/pkg/dto/request"
)

func (s *AppServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, s.dashboard.Health(r.Context(), request.DashboardHealthRequest{}))
}

func (s *AppServer) handleDashboardSystemInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	inventory, err := s.dashboard.SystemInventory(r.Context(), request.DashboardSystemInventoryRequest{})
	if err != nil {
		slog.Error("dashboard inventory failed", "err", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to load dashboard inventory")
		return
	}

	writeJSON(w, http.StatusOK, inventory)
}

func (s *AppServer) handleDashboardLiveTelemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	telemetry, err := s.dashboard.LiveTelemetry(r.Context(), request.DashboardLiveTelemetryRequest{})
	if err != nil {
		slog.Error("dashboard live telemetry failed", "err", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to load dashboard live telemetry")
		return
	}

	writeJSON(w, http.StatusOK, telemetry)
}

func (s *AppServer) handleDashboardThroughputOverTime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	throughput, err := s.dashboard.ThroughputOverTime(r.Context(), request.DashboardThroughputOverTimeRequest{})
	if err != nil {
		slog.Error("dashboard throughput over time failed", "err", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to load dashboard throughput over time")
		return
	}

	writeJSON(w, http.StatusOK, throughput)
}

func writeJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Warn("failed to write JSON response", "err", err)
	}
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}
