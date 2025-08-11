package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"sms-gateway/src/internal/config"
	"sms-gateway/src/internal/model"
	"sms-gateway/src/internal/service"
	"sms-gateway/src/internal/utils"
	"sms-gateway/src/pkg/validation"
)

// SMSHandler handles SMS-related HTTP requests
type SMSHandler struct {
	config     *config.Config
	smsService *service.SMSService
	startTime  time.Time
}

// NewSMSHandler creates a new SMS handler
func NewSMSHandler(cfg *config.Config, smsService *service.SMSService) *SMSHandler {
	return &SMSHandler{
		config:     cfg,
		smsService: smsService,
		startTime:  time.Now(),
	}
}

// HandleSendSMS handles SMS sending requests
func (h *SMSHandler) HandleSendSMS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed. Use POST.")
		return
	}

	var req model.SendSMSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON body: "+err.Error())
		return
	}

	// Validate request
	if err := validation.ValidateSendSMSRequest(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Send SMS
	response, err := h.smsService.SendSMS(r.Context(), &req)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, response)
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleHealth handles health check requests
func (h *SMSHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(h.startTime)

	response := model.HealthResponse{
		Status:    "healthy",
		Version:   h.config.Version,
		Timestamp: time.Now().Format(time.RFC3339),
		Uptime:    uptime.String(),
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// HandlePortStatus handles port status check requests
func (h *SMSHandler) HandlePortStatus(w http.ResponseWriter, r *http.Request) {
	port := r.URL.Query().Get("port")
	if port == "" {
		port = h.config.Modem.DefaultPort
	}

	status, err := h.smsService.CheckPortStatus(port)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, status)
}

// HandleModemInfo handles modem information requests
func (h *SMSHandler) HandleModemInfo(w http.ResponseWriter, r *http.Request) {
	port := r.URL.Query().Get("port")
	if port == "" {
		port = h.config.Modem.DefaultPort
	}

	baudRate := h.config.Modem.DefaultBaudRate
	// You could add baudRate parameter parsing here if needed

	info, err := h.smsService.GetModemInfo(r.Context(), port, baudRate)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, info)
}

// HandleListPorts handles listing available ports
func (h *SMSHandler) HandleListPorts(w http.ResponseWriter, r *http.Request) {
	ports, err := h.smsService.ListPorts()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := model.SuccessResponse{
		Success:   true,
		Data:      ports,
		Message:   "Available ports retrieved successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleRoot handles the root endpoint
func (h *SMSHandler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service": "SMS Gateway API",
		"version": h.config.Version,
		"endpoints": map[string]string{
			"POST /api/v1/sms/send":    "Send SMS message",
			"GET /api/v1/health":       "Service health check",
			"GET /api/v1/ports":        "List available ports",
			"GET /api/v1/ports/status": "Check port status",
			"GET /api/v1/modem/info":   "Get modem information",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// writeError writes error response
func (h *SMSHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	response := model.ErrorResponse{
		Error:     message,
		Code:      statusCode,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	utils.WriteJSON(w, statusCode, response)
}
