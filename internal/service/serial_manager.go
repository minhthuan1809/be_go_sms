package service

import (
	"log"
	"time"

	serial "go.bug.st/serial"
)

// SerialManager handles serial port operations
type SerialManager struct{}

// NewSerialManager creates a new serial manager
func NewSerialManager() *SerialManager {
	return &SerialManager{}
}

// CheckPortAvailability checks if a serial port is available
func (m *SerialManager) CheckPortAvailability(portName string) (bool, error) {
	mode := &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
		DataBits: 8,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return false, err
	}
	defer port.Close()

	return true, nil
}

// OpenPort opens a serial port with specified configuration
func (m *SerialManager) OpenPort(portName string, baudRate int) (serial.Port, error) {
	mode := &serial.Mode{
		BaudRate: baudRate,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
		DataBits: 8,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}

	// Set read timeout
	if err := port.SetReadTimeout(2 * time.Second); err != nil {
		log.Printf("Warning: failed to set read timeout: %v", err)
	}

	return port, nil
}

// ClosePort safely closes a serial port
func (m *SerialManager) ClosePort(port serial.Port) error {
	if port != nil {
		if err := port.Close(); err != nil {
			log.Printf("Warning: failed to close port: %v", err)
			return err
		}
	}
	return nil
}

// LockPort locks the port mutex to prevent concurrent access
func (m *SerialManager) LockPort() {
	// portMutex is defined in sms_service.go
	portMutex.Lock()
}

// UnlockPort unlocks the port mutex
func (m *SerialManager) UnlockPort() {
	// portMutex is defined in sms_service.go
	portMutex.Unlock()
}
