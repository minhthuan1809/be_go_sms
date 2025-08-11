package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"myproject/internal/service"
	"myproject/internal/types"
	"myproject/internal/utils"
)

// SMSHandler handles SMS-related HTTP requests
type SMSHandler struct {
	smsService *service.SMSService
	version    string
}

// NewSMSHandler creates a new SMS handler
func NewSMSHandler(version string) *SMSHandler {
	return &SMSHandler{
		smsService: service.NewSMSService(),
		version:    version,
	}
}

// HandleRoot handles the root endpoint
func (h *SMSHandler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"service": "SMS Gateway API",
		"version": h.version,
		"endpoints": map[string]string{
			"POST /send":       "Send SMS message",
			"GET /health":      "Service health check",
			"GET /port/status": "Check port availability",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(response)
}

// HandleHealth handles health check requests
func (h *SMSHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, types.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   h.version,
	})
}

// HandlePortStatus handles port status check requests
func (h *SMSHandler) HandlePortStatus(w http.ResponseWriter, r *http.Request) {
	port := r.URL.Query().Get("port")
	if port == "" {
		port = "/dev/ttyUSB0" // default for Linux
	}

	available, err := h.smsService.CheckPortAvailability(port)
	response := types.PortStatusResponse{
		Port:      port,
		Available: available,
	}
	if err != nil {
		response.Error = err.Error()
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

// Version returns the current version
func (h *SMSHandler) Version() string {
	return h.version
}

// HandleSendSMS handles SMS sending requests
func (h *SMSHandler) HandleSendSMS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		utils.WriteJSON(w, http.StatusMethodNotAllowed, types.SendSMSResponse{
			Success: false,
			Error:   "Method not allowed. Use POST.",
		})
		return
	}

	var req types.SendSMSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, types.SendSMSResponse{
			Success: false,
			Error:   "Invalid JSON body: " + err.Error(),
		})
		return
	}

	// Validate and set defaults
	if err := utils.ValidateRequest(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, types.SendSMSResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	startTime := time.Now()

	// Send SMS with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Timeout)*time.Second)
	defer cancel()

	steps, messageID, err := h.smsService.SendSMSViaAT(ctx, req.Port, req.BaudRate, req.To, req.Message)
	duration := time.Since(startTime)

	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, types.SendSMSResponse{
			Success:  false,
			Steps:    steps,
			Error:    err.Error(),
			Duration: duration.String(),
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.SendSMSResponse{
		Success:   true,
		Steps:     steps,
		MessageID: messageID,
		Duration:  duration.String(),
	})
}
