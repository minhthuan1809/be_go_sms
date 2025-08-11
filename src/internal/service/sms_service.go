package service

import (
	"context"
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
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Set defaults if not provided
	if req.Port == "" {
		req.Port = s.config.Modem.DefaultPort
	}
	if req.BaudRate == 0 {
		req.BaudRate = s.config.Modem.DefaultBaudRate
	}
	if req.Timeout == 0 {
		req.Timeout = s.config.SMS.DefaultTimeout
	}

	startTime := time.Now()

	// Send SMS using the SMS client
	steps, messageID, err := s.smsClient.SendViaPDU(ctx, req.Port, req.BaudRate, req.To, req.Message)
	duration := time.Since(startTime)

	response := &model.SendSMSResponse{
		Steps:    steps,
		Duration: duration.String(),
	}

	if err != nil {
		response.Success = false
		response.Error = err.Error()
		return response, err
	}

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
