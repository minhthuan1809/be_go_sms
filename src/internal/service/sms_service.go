package service

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
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

// GetDeviceInfo gets comprehensive device information including SIM details
func (s *SMSService) GetDeviceInfo(ctx context.Context, port string, baudRate int) (*model.DeviceInfo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.modemClient.GetDeviceInfo(ctx, port, baudRate)
}

// GetAllDevicesInfo gets device information for all available USB ports with optimizations
func (s *SMSService) GetAllDevicesInfo(ctx context.Context) ([]model.DeviceInfo, error) {
	startTime := time.Now()
	log.Printf("Starting GetAllDevicesInfo operation")

	// Get list of available USB ports
	allPorts, err := s.modemClient.ListPorts()
	if err != nil {
		log.Printf("Failed to get ports list: %v", err)
		return nil, fmt.Errorf("failed to get ports: %w", err)
	}

	// Filter for USB ports only
	var usbPorts []string
	for _, port := range allPorts {
		if strings.Contains(port, "ttyUSB") {
			usbPorts = append(usbPorts, port)
		}
	}

	log.Printf("Found %d USB ports: %v", len(usbPorts), usbPorts)

	if len(usbPorts) == 0 {
		return []model.DeviceInfo{}, nil
	}

	// Create overall timeout context - increase timeout for multiple devices
	overallTimeout := time.Duration(len(usbPorts)*15) * time.Second // 15 seconds per device
	if overallTimeout > 60*time.Second {
		overallTimeout = 60 * time.Second // Max 60 seconds total
	}
	
	overallCtx, cancel := context.WithTimeout(ctx, overallTimeout)
	defer cancel()

	// Use semaphore to limit concurrent operations - REDUCE to 2 for USB bandwidth
	maxConcurrent := 2 // Reduced from 4 to avoid USB bandwidth issues
	if len(usbPorts) == 1 {
		maxConcurrent = 1
	}
	log.Printf("Using max concurrent operations: %d", maxConcurrent)

	semaphore := make(chan struct{}, maxConcurrent)
	results := make(chan struct {
		info model.DeviceInfo
		port string
		err  error
	}, len(usbPorts))

	// Start workers with staggered delays to reduce USB contention
	for i, port := range usbPorts {
		go func(portName string, index int) {
			// Stagger the start times to reduce USB bus contention
			delay := time.Duration(index*500) * time.Millisecond // 500ms delay between starts
			time.Sleep(delay)
			
			log.Printf("[%s] Worker %d starting after %v delay", portName, index+1, delay)
			
			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-overallCtx.Done():
				log.Printf("[%s] Worker %d cancelled before acquiring semaphore", portName, index+1)
				results <- struct {
					info model.DeviceInfo
					port string
					err  error
				}{model.DeviceInfo{Port: portName, Error: "cancelled"}, portName, fmt.Errorf("cancelled")}
				return
			}

			log.Printf("[%s] Worker %d acquired semaphore, getting device info", portName, index+1)
			workerStart := time.Now()

			// Create individual timeout for this device - longer for single device operations
			deviceTimeout := 12 * time.Second
			deviceCtx, deviceCancel := context.WithTimeout(overallCtx, deviceTimeout)
			defer deviceCancel()

			info, err := s.modemClient.GetDeviceInfo(deviceCtx, portName, s.config.Modem.DefaultBaudRate)
			workerDuration := time.Since(workerStart)
			
			if err != nil {
				log.Printf("[%s] Worker %d failed after %v: %v", portName, index+1, workerDuration, err)
				if info == nil {
					info = &model.DeviceInfo{Port: portName, Error: err.Error()}
				}
			} else {
				log.Printf("[%s] Worker %d completed successfully in %v (Phone: %s, Operator: %s)", 
					portName, index+1, workerDuration, info.PhoneNumber, info.Operator)
			}

			results <- struct {
				info model.DeviceInfo
				port string
				err  error
			}{*info, portName, err}
			
			log.Printf("[%s] Worker %d finished, releasing semaphore", portName, index+1)
		}(port, i)
	}

	// Collect results with detailed logging
	log.Printf("Starting to collect results from %d workers", len(usbPorts))
	var devicesInfo []model.DeviceInfo
	for i := 0; i < len(usbPorts); i++ {
		log.Printf("Waiting for result %d/%d", i+1, len(usbPorts))
		select {
		case res := <-results:
			log.Printf("Received result from %s: Success=%t, Phone=%s", 
				res.port, res.err == nil, res.info.PhoneNumber)
			devicesInfo = append(devicesInfo, res.info)
		case <-overallCtx.Done():
			log.Printf("Overall timeout reached while collecting results, returning %d devices", len(devicesInfo))
			goto done
		}
	}

done:
	totalDuration := time.Since(startTime)
	log.Printf("GetAllDevicesInfo completed in %v, returning %d devices", totalDuration, len(devicesInfo))
	
	// Sort results by port name for consistent ordering
	sort.Slice(devicesInfo, func(i, j int) bool {
		return devicesInfo[i].Port < devicesInfo[j].Port
	})
	
	return devicesInfo, nil
}
