package modem

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"go.bug.st/serial"
)

// Modem represents a GSM modem connection
type Modem struct {
	port     serial.Port
	portName string
	baudRate int
}

// NewModem creates a new modem instance
func NewModem(portName string, baudRate int) *Modem {
	return &Modem{
		portName: portName,
		baudRate: baudRate,
	}
}

// Connect establishes connection to the modem
func (m *Modem) Connect() error {
	mode := &serial.Mode{
		BaudRate: m.baudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(m.portName, mode)
	if err != nil {
		return fmt.Errorf("failed to open port %s: %v", m.portName, err)
	}

	m.port = port
	return nil
}

// Close closes the modem connection
func (m *Modem) Close() error {
	if m.port != nil {
		return m.port.Close()
	}
	return nil
}

// SendCommand sends an AT command and returns the response
func (m *Modem) SendCommand(command string) (string, error) {
	if m.port == nil {
		return "", fmt.Errorf("modem not connected")
	}

	// Clear any pending data
	m.port.ResetInputBuffer()
	m.port.ResetOutputBuffer()

	// Send command
	cmd := command + "\r\n"
	_, err := m.port.Write([]byte(cmd))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %v", err)
	}

	// Wait a bit for the command to be processed
	time.Sleep(100 * time.Millisecond)

	// Read response
	scanner := bufio.NewScanner(m.port)
	var response strings.Builder
	
	for scanner.Scan() {
		line := scanner.Text()
		response.WriteString(line + "\n")
		
		// Stop reading if we get OK or ERROR
		if strings.Contains(line, "OK") || strings.Contains(line, "ERROR") {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	return response.String(), nil
}

// InitializeModem initializes the modem for SMS operations
func (m *Modem) InitializeModem() error {
	commands := []string{
		"AT",                    // Test AT
		"AT+CMGF=1",            // Set SMS text mode
		"AT+CSCS=\"GSM\"",      // Set character set to GSM
		"AT+CNMI=2,1,0,0,0",    // Set new message indication
	}

	for _, cmd := range commands {
		response, err := m.SendCommand(cmd)
		if err != nil {
			return fmt.Errorf("failed to send %s: %v", cmd, err)
		}
		
		if !strings.Contains(response, "OK") {
			return fmt.Errorf("command %s failed: %s", cmd, response)
		}
		
		log.Printf("Command %s successful", cmd)
	}

	return nil
}

// SendSMS sends an SMS message
func (m *Modem) SendSMS(phoneNumber, message string) error {
	// Set SMS text mode
	_, err := m.SendCommand("AT+CMGF=1")
	if err != nil {
		return fmt.Errorf("failed to set SMS text mode: %v", err)
	}

	// Set phone number
	setNumberCmd := fmt.Sprintf("AT+CMGS=\"%s\"", phoneNumber)
	response, err := m.SendCommand(setNumberCmd)
	if err != nil {
		return fmt.Errorf("failed to set phone number: %v", err)
	}

	// Check if we got the '>' prompt
	if !strings.Contains(response, ">") {
		return fmt.Errorf("did not get '>' prompt: %s", response)
	}

	// Send message content
	messageWithCtrlZ := message + string(26) // Ctrl+Z to send
	_, err = m.port.Write([]byte(messageWithCtrlZ))
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	// Wait for response
	time.Sleep(2 * time.Second)
	
	scanner := bufio.NewScanner(m.port)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "OK") {
			log.Printf("SMS sent successfully to %s", phoneNumber)
			return nil
		}
		if strings.Contains(line, "ERROR") {
			return fmt.Errorf("failed to send SMS: %s", line)
		}
	}

	return fmt.Errorf("no response received for SMS")
}

// GenerateOTP generates a random OTP of specified length
func GenerateOTP(length int) string {
	rand.Seed(time.Now().UnixNano())
	
	const digits = "0123456789"
	otp := make([]byte, length)
	
	for i := range otp {
		otp[i] = digits[rand.Intn(len(digits))]
	}
	
	return string(otp)
}

// SendOTP sends an OTP message
func (m *Modem) SendOTP(phoneNumber string, length int, messagePrefix string) (string, error) {
	// Generate OTP
	otp := GenerateOTP(length)
	
	// Create full message
	fullMessage := messagePrefix + otp
	
	// Send SMS
	err := m.SendSMS(phoneNumber, fullMessage)
	if err != nil {
		return "", err
	}
	
	return otp, nil
}
