package service

import (
	"context"
	"log"
	"sync"
	"time"

	"sms-gateway/src/internal/config"
	"sms-gateway/src/internal/model"
	"sms-gateway/src/pkg/modem"
	"sms-gateway/src/pkg/sms"
)

var (
	portMutex = &sync.Mutex{} // Ensure no conflicts when using ports
)

// SMSService handles SMS operations
type SMSService struct {
	config      *config.Config
	modemClient *modem.Client
	smsClient   *sms.Client
	mutex       sync.RWMutex
}

// NewSMSService creates a new SMS service instance
func NewSMSService(cfg *config.Config) *SMSService {
	return &SMSService{
		config:      cfg,
		modemClient: modem.NewClient(cfg),
		smsClient:   sms.NewClient(cfg),
	}
}

// SendSMS sends an SMS message
func (s *SMSService) SendSMS(ctx context.Context, req *model.SendSMSRequest) (*model.SendSMSResponse, error) {
	log.Printf("Starting SMS send process - Port: %s, BaudRate: %d, To: %s, Mode: %s",
		req.Port, req.BaudRate, req.To, req.Mode)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Set defaults if not provided
	if req.Port == "" {
		req.Port = s.config.Modem.DefaultPort
		log.Printf("Using default port: %s", req.Port)
	}
	if req.BaudRate == 0 {
		req.BaudRate = s.config.Modem.DefaultBaudRate
		log.Printf("Using default baud rate: %d", req.BaudRate)
	}
	if req.Timeout == 0 {
		req.Timeout = s.config.SMS.DefaultTimeout
		log.Printf("Using default timeout: %d", req.Timeout)
	}
	if req.Mode == "" {
		req.Mode = "text" // Default to text mode
		log.Printf("Using default mode: %s", req.Mode)
	}

	startTime := time.Now()

	// Send SMS using the appropriate mode
	var steps []string
	var messageID string
	var err error

	if req.Mode == "pdu" {
		log.Printf("Calling SMS client SendViaPDU...")
		steps, messageID, err = s.smsClient.SendViaPDU(ctx, req.Port, req.BaudRate, req.To, req.Message)
	} else {
		log.Printf("Calling SMS client SendViaText...")
		steps, messageID, err = s.smsClient.SendViaText(ctx, req.Port, req.BaudRate, req.To, req.Message)
	}

	duration := time.Since(startTime)

	response := &model.SendSMSResponse{
		Steps:     steps,
		Duration:  duration.String(),
		Mode:      req.Mode,
		To:        req.To,
		Message:   req.Message,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err != nil {
		log.Printf("SMS client error: %v", err)
		response.Success = false
		response.Error = err.Error()
		return response, err
	}

	log.Printf("SMS sent successfully - MessageID: %s, Steps: %d, Duration: %v, Mode: %s",
		messageID, len(steps), duration, req.Mode)
	response.Success = true
	response.MessageID = messageID
	return response, nil
}

// CheckPortStatus checks if a serial port is available
func (s *SMSService) CheckPortStatus(portName string) (*model.PortStatus, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.modemClient.CheckPortStatus(portName)
}

// GetModemInfo gets modem information
func (s *SMSService) GetModemInfo(ctx context.Context, port string, baudRate int) (*model.ModemInfo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.modemClient.GetInfo(ctx, port, baudRate)
}

// ListPorts lists available serial ports
func (s *SMSService) ListPorts() ([]string, error) {
	return s.modemClient.ListPorts()
}

// ListPortsWithInfo lists available serial ports with device information
func (s *SMSService) ListPortsWithInfo() ([]model.PortInfo, error) {
	return s.modemClient.ListPortsWithInfo()
}
