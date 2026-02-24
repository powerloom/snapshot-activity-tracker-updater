package dashboard

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Server is the dashboard API server
type Server struct {
	reader *Reader
	router *mux.Router
}

// NewServer creates a new dashboard server
func NewServer(dataMarket, dumpDir string) *Server {
	reader := NewReader(dataMarket, dumpDir)

	s := &Server{
		reader: reader,
		router: mux.NewRouter(),
	}

	s.setupRoutes()

	return s
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	api := s.router.PathPrefix("/api").Subrouter()

	// Health check
	api.HandleFunc("/health", s.handleHealth).Methods(http.MethodGet)

	// Dashboard summary
	api.HandleFunc("/dashboard/summary", s.handleDashboardSummary).Methods(http.MethodGet)

	// Network topology (PRIORITY)
	api.HandleFunc("/network/topology", s.handleNetworkTopology).Methods(http.MethodGet)

	// Epochs
	api.HandleFunc("/epochs", s.handleEpochs).Methods(http.MethodGet)
	api.HandleFunc("/epochs/{epochID}", s.handleEpochDetail).Methods(http.MethodGet)

	// Validators
	api.HandleFunc("/validators", s.handleValidators).Methods(http.MethodGet)
	api.HandleFunc("/validators/{validatorID}", s.handleValidatorDetail).Methods(http.MethodGet)

	// Slots
	api.HandleFunc("/slots", s.handleSlots).Methods(http.MethodGet)
	api.HandleFunc("/slots/{slotID}", s.handleSlotDetail).Methods(http.MethodGet)

	// Projects
	api.HandleFunc("/projects", s.handleProjects).Methods(http.MethodGet)

	// Timeline
	api.HandleFunc("/timeline", s.handleTimeline).Methods(http.MethodGet)
}

// GetRouter returns the HTTP router
func (s *Server) GetRouter() http.Handler {
	return s.router
}

// handleHealth handles the health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// handleDashboardSummary handles the dashboard summary endpoint
func (s *Server) handleDashboardSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := s.reader.GetDashboardSummary(r.Context())
	if err != nil {
		log.Printf("Error getting dashboard summary: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to get dashboard summary")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// handleNetworkTopology handles the network topology endpoint (PRIORITY)
func (s *Server) handleNetworkTopology(w http.ResponseWriter, r *http.Request) {
	topology, err := s.reader.GetNetworkTopology(r.Context())
	if err != nil {
		log.Printf("Error getting network topology: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to get network topology")
		return
	}

	writeJSON(w, http.StatusOK, topology)
}

// handleEpochs handles the epochs list endpoint
func (s *Server) handleEpochs(w http.ResponseWriter, r *http.Request) {
	epochs, err := s.reader.GetEpochs(r.Context())
	if err != nil {
		log.Printf("Error getting epochs: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to get epochs")
		return
	}

	writeJSON(w, http.StatusOK, epochs)
}

// handleEpochDetail handles the epoch detail endpoint
func (s *Server) handleEpochDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epochIDStr := vars["epochID"]

	epochID, err := strconv.ParseUint(epochIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid epoch ID")
		return
	}

	detail, err := s.reader.GetEpochDetail(r.Context(), epochID)
	if err != nil {
		log.Printf("Error getting epoch detail: %v", err)
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "failed to get epoch detail")
		}
		return
	}

	writeJSON(w, http.StatusOK, detail)
}

// handleValidators handles the validators list endpoint
func (s *Server) handleValidators(w http.ResponseWriter, r *http.Request) {
	validators, err := s.reader.GetValidators(r.Context())
	if err != nil {
		log.Printf("Error getting validators: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to get validators")
		return
	}

	writeJSON(w, http.StatusOK, validators)
}

// handleValidatorDetail handles the validator detail endpoint
func (s *Server) handleValidatorDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	validatorID := vars["validatorID"]

	detail, err := s.reader.GetValidatorDetail(r.Context(), validatorID)
	if err != nil {
		log.Printf("Error getting validator detail: %v", err)
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "failed to get validator detail")
		}
		return
	}

	writeJSON(w, http.StatusOK, detail)
}

// handleSlots handles the slots list endpoint
func (s *Server) handleSlots(w http.ResponseWriter, r *http.Request) {
	slots, err := s.reader.GetSlots(r.Context())
	if err != nil {
		log.Printf("Error getting slots: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to get slots")
		return
	}

	writeJSON(w, http.StatusOK, slots)
}

// handleSlotDetail handles the slot detail endpoint
func (s *Server) handleSlotDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slotID := vars["slotID"]

	detail, err := s.reader.GetSlotDetail(r.Context(), slotID)
	if err != nil {
		log.Printf("Error getting slot detail: %v", err)
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "failed to get slot detail")
		}
		return
	}

	writeJSON(w, http.StatusOK, detail)
}

// handleProjects handles the projects list endpoint
func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.reader.GetProjects(r.Context())
	if err != nil {
		log.Printf("Error getting projects: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to get projects")
		return
	}

	writeJSON(w, http.StatusOK, projects)
}

// handleTimeline handles the timeline endpoint
func (s *Server) handleTimeline(w http.ResponseWriter, r *http.Request) {
	timeline, err := s.reader.GetTimeline(r.Context())
	if err != nil {
		log.Printf("Error getting timeline: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to get timeline")
		return
	}

	writeJSON(w, http.StatusOK, timeline)
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

// SetIndexHandler sets a custom index handler for serving the frontend
func (s *Server) SetIndexHandler(handler http.Handler) {
	// Catch-all route for SPA - must be registered last
	s.router.PathPrefix("/").Handler(handler)
}

// ListenAndServe starts the HTTP server
func (s *Server) ListenAndServe(addr string) error {
	log.Printf("Dashboard API server starting on %s", addr)
	return http.ListenAndServe(addr, s.router)
}
