package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
// @Summary Send SMS message
// @Description Send an SMS message through the configured modem
// @Tags SMS
// @Accept json
// @Produce json
// @Param request body model.SendSMSRequest true "SMS request details"
// @Success 200 {object} model.SendSMSResponse "SMS sent successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.SendSMSResponse "Internal server error"
// @Router /api/v1/sms/send [post]
func (h *SMSHandler) HandleSendSMS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed. Use POST.")
		return
	}

	var req model.SendSMSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding JSON request: %v", err)
		h.writeError(w, http.StatusBadRequest, "Invalid JSON body: "+err.Error())
		return
	}

	// Log incoming request
	log.Printf("SMS request received - To: %s, Port: %s, Message length: %d",
		req.To, req.Port, len(req.Message))

	// Validate request
	if err := validation.ValidateSendSMSRequest(&req); err != nil {
		log.Printf("Validation error: %v", err)
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Send SMS
	response, err := h.smsService.SendSMS(r.Context(), &req)
	if err != nil {
		log.Printf("SMS service error: %v", err)
		// Return the response with error details
		utils.WriteJSON(w, http.StatusInternalServerError, response)
		return
	}

	log.Printf("SMS sent successfully - MessageID: %s, Duration: %s",
		response.MessageID, response.Duration)
	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleHealth handles health check requests
// @Summary Health check
// @Description Check the health status of the SMS Gateway service
// @Tags Health
// @Produce json
// @Success 200 {object} model.HealthResponse "Service health status"
// @Router /api/v1/health [get]
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
// @Summary Check port status
// @Description Check the status of a specific serial port and return SIM balance (if configured)
// @Tags Modem
// @Produce json
// @Param port query string false "Port name (defaults to configured default port)"
// @Success 200 {object} model.PortStatus "Port status information including balance if available"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/ports/status [get]
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
// @Summary Get modem information
// @Description Get detailed information about the modem
// @Tags Modem
// @Produce json
// @Param port query string false "Port name (defaults to configured default port)"
// @Success 200 {object} model.ModemInfo "Modem information"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/modem/info [get]
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
// @Summary List available ports
// @Description Get a list of all available serial ports with device information
// @Tags Modem
// @Produce json
// @Success 200 {object} model.SuccessResponse "List of available ports with device info"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/ports [get]
func (h *SMSHandler) HandleListPorts(w http.ResponseWriter, r *http.Request) {
	ports, err := h.smsService.ListPortsWithInfo()
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
// @Summary API information
// @Description Get information about the SMS Gateway API
// @Tags General
// @Produce json
// @Success 200 {object} map[string]interface{} "API information"
// @Router / [get]
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
			"GET /api/v1/device/info":  "Get detailed device information",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleDeviceInfo handles detailed device information requests
// @Summary Get detailed device information
// @Description Get comprehensive device information including phone number, balance, network type, and SIM details
// @Tags Device
// @Produce json
// @Param port query string false "Port name (defaults to configured default port)"
// @Param baud_rate query int false "Baud rate (defaults to configured default baud rate)"
// @Success 200 {object} model.DeviceInfo "Detailed device information"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/device/info [get]
func (h *SMSHandler) HandleDeviceInfo(w http.ResponseWriter, r *http.Request) {
	port := r.URL.Query().Get("port")
	if port == "" {
		port = h.config.Modem.DefaultPort
	}

	baudRate := h.config.Modem.DefaultBaudRate
	if baudRateStr := r.URL.Query().Get("baud_rate"); baudRateStr != "" {
		if br, err := strconv.Atoi(baudRateStr); err == nil && br > 0 {
			baudRate = br
		}
	}

	log.Printf("Getting device info for port: %s, baud rate: %d", port, baudRate)

	info, err := h.smsService.GetDeviceInfo(r.Context(), port, baudRate)
	if err != nil {
		log.Printf("Error getting device info: %v", err)
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("Device info retrieved successfully for port: %s", port)
	utils.WriteJSON(w, http.StatusOK, info)
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
