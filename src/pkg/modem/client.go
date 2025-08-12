package modem

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"sms-gateway/src/internal/config"
	"sms-gateway/src/internal/model"

	serial "go.bug.st/serial"
)

// Client handles modem operations
type Client struct {
	config *config.Config
}

// NewClient creates a new modem client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
	}
}

// CheckPortStatus checks if a serial port is available
func (c *Client) CheckPortStatus(portName string) (*model.PortStatus, error) {
	status := &model.PortStatus{
		Port: portName,
	}

	// Try to open the port
	mode := &serial.Mode{
		BaudRate: c.config.Modem.DefaultBaudRate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		status.Available = false
		status.Error = err.Error()
		return status, nil
	}
	defer port.Close()

	status.Available = true

	// Try to fetch SIM balance via USSD if configured
	if ussd := strings.TrimSpace(c.config.Modem.BalanceUSSD); ussd != "" {
		// Keep it short to avoid blocking for too long
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		// Enable USSD and query; ignore errors but capture response
		_ = func() error {
			if _, err := c.sendATCommand(ctx, port, "AT+CUSD=1"); err != nil {
				return err
			}
			cmd := "AT+CUSD=1,\"" + ussd + "\",15"
			if _, err := c.sendATCommand(ctx, port, cmd); err != nil {
				return err
			}
			resp, err := c.readResponse(ctx, port, 8*time.Second)
			if err == nil && strings.TrimSpace(resp) != "" {
				status.Balance = strings.TrimSpace(resp)
			}
			return nil
		}()
	}
	return status, nil
}

// GetInfo gets modem information
func (c *Client) GetInfo(ctx context.Context, portName string, baudRate int) (*model.ModemInfo, error) {
	info := &model.ModemInfo{
		Port:     portName,
		BaudRate: baudRate,
	}

	mode := &serial.Mode{
		BaudRate: baudRate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return info, fmt.Errorf("failed to open port: %w", err)
	}
	defer port.Close()

	info.Connected = true

	// Get manufacturer
	if manufacturer, err := c.sendATCommand(ctx, port, "AT+CGMI"); err == nil {
		info.Manufacturer = strings.TrimSpace(manufacturer)
	}

	// Get model
	if model, err := c.sendATCommand(ctx, port, "AT+CGMM"); err == nil {
		info.Model = strings.TrimSpace(model)
	}

	// Get version
	if version, err := c.sendATCommand(ctx, port, "AT+CGMR"); err == nil {
		info.Version = strings.TrimSpace(version)
	}

	// Get IMEI
	if imei, err := c.sendATCommand(ctx, port, "AT+CGSN"); err == nil {
		info.IMEI = strings.TrimSpace(imei)
	}

	return info, nil
}

// ListPorts lists available serial ports
func (c *Client) ListPorts() ([]string, error) {
	return serial.GetPortsList()
}

// ListPortsWithInfo lists available serial ports with device information
func (c *Client) ListPortsWithInfo() ([]model.PortInfo, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}

	var portInfos []model.PortInfo
	
	// Set overall timeout for the entire operation
	overallCtx, overallCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer overallCancel()
	
	for _, port := range ports {
		// Check if overall timeout has been reached
		select {
		case <-overallCtx.Done():
			log.Printf("Overall timeout reached while scanning ports")
			goto done
		default:
		}
		
		info := model.PortInfo{Port: port}
		info.DeviceName = c.getDeviceName(port)

		// Try to open port with very short timeout
		mode := &serial.Mode{
			BaudRate: c.config.Modem.DefaultBaudRate,
			Parity:   serial.NoParity,
			DataBits: 8,
			StopBits: serial.OneStopBit,
		}
		
		// Use a very short timeout for port operations
		portCtx, portCancel := context.WithTimeout(overallCtx, 3*time.Second)
		
		p, err := serial.Open(port, mode)
		if err == nil {
			info.Available = true
			
			// Only get basic info quickly
			if desc := c.getBasicDeviceInfo(portCtx, p); desc != "" {
				info.Description = desc
			} else {
				info.Description = "USB Serial Device"
			}
			
			// Skip MSISDN, balance, and packages for faster response
			// These can be obtained via the detailed device info endpoint if needed
			
			p.Close()
		} else {
			info.Available = false
			info.Error = err.Error()
			info.Description = "Port unavailable"
		}
		
		portCancel()
		portInfos = append(portInfos, info)
	}

done:
	return portInfos, nil
}

// getDeviceName extracts device name from port path
func (c *Client) getDeviceName(port string) string {
	// For Linux: /dev/ttyUSB0 -> USB0
	if strings.HasPrefix(port, "/dev/ttyUSB") {
		return strings.TrimPrefix(port, "/dev/tty")
	}

	// For Windows: COM3 -> COM3
	if strings.HasPrefix(port, "COM") {
		return port
	}

	// For macOS: /dev/tty.usbserial-* -> usbserial-*
	if strings.Contains(port, "usbserial") {
		parts := strings.Split(port, "/")
		if len(parts) > 0 {
			return strings.TrimPrefix(parts[len(parts)-1], "tty.")
		}
	}

	return port
}

// getDeviceDescription gets device description based on port
func (c *Client) getDeviceDescription(port string) string {
	// Try to get modem info to determine device type
	mode := &serial.Mode{
		BaudRate: c.config.Modem.DefaultBaudRate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	portHandle, err := serial.Open(port, mode)
	if err != nil {
		return "Unknown device"
	}
	defer portHandle.Close()

	// Try to get manufacturer info
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if manufacturer, err := c.sendATCommand(ctx, portHandle, "AT+CGMI"); err == nil {
		manufacturer = strings.TrimSpace(manufacturer)
		if manufacturer != "" && !strings.Contains(manufacturer, "ERROR") {
			return manufacturer + " Modem"
		}
	}

	return "USB Serial Device"
}

// sendATCommand sends AT command and returns response
func (c *Client) sendATCommand(ctx context.Context, port serial.Port, command string) (string, error) {
	// Send command
	_, err := port.Write([]byte(command + "\r\n"))
	if err != nil {
		return "", err
	}

	// Read response with timeout
	timeout := time.Duration(5) * time.Second
	return c.readResponse(ctx, port, timeout)
}

// readResponse reads response from modem
func (c *Client) readResponse(ctx context.Context, port serial.Port, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	var response strings.Builder

	buffer := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return response.String(), ctx.Err()
		default:
		}

		if time.Now().After(deadline) {
			return response.String(), fmt.Errorf("timeout")
		}

		port.SetReadTimeout(100 * time.Millisecond)
		n, err := port.Read(buffer)
		if err != nil {
			continue
		}

		if n > 0 {
			response.Write(buffer[:n])
			resp := response.String()
			if strings.Contains(resp, "OK") || strings.Contains(resp, "ERROR") {
				return resp, nil
			}
		}
	}
}

func (c *Client) getDeviceDescriptionWithPort(ctx context.Context, port serial.Port) string {
	if manufacturer, err := c.sendATCommand(ctx, port, "AT+CGMI"); err == nil {
		m := strings.TrimSpace(manufacturer)
		if m != "" && !strings.Contains(m, "ERROR") {
			return m + " Modem"
		}
	}
	return "USB Serial Device"
}

// getBasicDeviceInfo gets basic device info quickly with short timeout
func (c *Client) getBasicDeviceInfo(ctx context.Context, port serial.Port) string {
	// Use very short timeout for basic info
	quickCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	
	if manufacturer, err := c.sendATCommand(quickCtx, port, "AT+CGMI"); err == nil {
		m := c.cleanATResponse(manufacturer)
		if m != "" && !strings.Contains(m, "ERROR") {
			return m + " Modem"
		}
	}
	return ""
}

func (c *Client) queryMsisdn(port serial.Port) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := c.sendATCommand(ctx, port, "AT+CNUM")
	if err != nil {
		return "", err
	}
	
	// Clean the response
	resp = c.cleanATResponse(resp)
	
	// Typical response: +CNUM: ,"+84123456789",145
	for _, line := range strings.Split(resp, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "+CNUM") || strings.Contains(line, "+CNUM:") {
			// Find all quoted strings in the line
			quotes := strings.Split(line, "\"")
			for i, part := range quotes {
				// Phone numbers typically start with + or digits
				if i%2 == 1 && (strings.HasPrefix(part, "+") || strings.HasPrefix(part, "0") || strings.HasPrefix(part, "8")) {
					// Validate it looks like a phone number
					if len(part) >= 10 && len(part) <= 15 {
						return strings.TrimSpace(part), nil
					}
				}
			}
		}
	}
	return "", nil
}

func (c *Client) runUSSD(port serial.Port, code string) (string, error) {
	// Enable USSD
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := c.sendATCommand(ctx, port, "AT+CUSD=1"); err != nil {
		return "", err
	}
	// Send USSD
	cmd := "AT+CUSD=1,\"" + code + "\",15"
	if _, err := c.sendATCommand(ctx, port, cmd); err != nil {
		return "", err
	}
	// Read response (may take a few seconds)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	resp, err := c.readResponse(ctx2, port, 10*time.Second)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp), nil
}

// GetDeviceInfo gets comprehensive device information including SIM details
func (c *Client) GetDeviceInfo(ctx context.Context, portName string, baudRate int) (*model.DeviceInfo, error) {
	info := &model.DeviceInfo{
		Port:      portName,
		BaudRate:  baudRate,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	mode := &serial.Mode{
		BaudRate: baudRate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		info.Error = fmt.Sprintf("Failed to open port: %v", err)
		return info, nil
	}
	defer port.Close()

	info.Connected = true

	// Get phone number (MSISDN)
	if phoneNumber, err := c.queryMsisdn(port); err == nil && phoneNumber != "" {
		info.PhoneNumber = phoneNumber
	}

	// Get manufacturer
	if manufacturer, err := c.sendATCommand(ctx, port, "AT+CGMI"); err == nil {
		info.Manufacturer = c.cleanATResponse(manufacturer)
	}

	// Get model
	if model, err := c.sendATCommand(ctx, port, "AT+CGMM"); err == nil {
		info.Model = c.cleanATResponse(model)
	}

	// Get version
	if version, err := c.sendATCommand(ctx, port, "AT+CGMR"); err == nil {
		info.Version = c.cleanATResponse(version)
	}

	// Get IMEI
	if imei, err := c.sendATCommand(ctx, port, "AT+CGSN"); err == nil {
		info.IMEI = c.cleanATResponse(imei)
	}

	// Get IMSI
	if imsi, err := c.sendATCommand(ctx, port, "AT+CIMI"); err == nil {
		info.IMSI = c.cleanATResponse(imsi)
	}

	// Get operator information
	if operator, err := c.sendATCommand(ctx, port, "AT+COPS?"); err == nil {
		info.Operator = c.parseOperator(operator)
	}

	// Get network registration status and technology
	if networkInfo, err := c.sendATCommand(ctx, port, "AT+CREG?"); err == nil {
		info.NetworkType = c.parseNetworkType(networkInfo)
	}

	// Get signal strength
	if signal, err := c.sendATCommand(ctx, port, "AT+CSQ"); err == nil {
		info.SignalLevel = c.parseSignalStrength(signal)
	}

	// Get balance via USSD if configured
	if ussd := strings.TrimSpace(c.config.Modem.BalanceUSSD); ussd != "" {
		if balance, err := c.runUSSD(port, ussd); err == nil && balance != "" {
			info.Balance = c.parseUSSDResponse(balance)
		}
	}

	return info, nil
}

// parseOperator extracts operator name from AT+COPS response
func (c *Client) parseOperator(response string) string {
	// Example: +COPS: 0,0,"Viettel",2
	response = c.cleanATResponse(response)
	
	for _, line := range strings.Split(response, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "+COPS:") {
			parts := strings.Split(line, ",")
			if len(parts) >= 3 {
				// Remove quotes from operator name
				operator := strings.Trim(parts[2], "\"")
				if operator != "" {
					return c.mapOperatorName(operator)
				}
			}
		}
	}
	return ""
}

// mapOperatorName maps operator codes to friendly names
func (c *Client) mapOperatorName(operator string) string {
	operatorMap := map[string]string{
		"45201": "Mobifone",
		"45202": "Vinaphone", 
		"45204": "Viettel",
		"45205": "Vietnamobile",
		"45207": "Gmobile",
		"45208": "Itelecom",
		"Viettel": "Viettel",
		"Mobifone": "Mobifone",
		"Vinaphone": "Vinaphone",
		"Vietnamobile": "Vietnamobile",
	}
	
	if friendlyName, exists := operatorMap[operator]; exists {
		return friendlyName
	}
	
	// If not found in map, return original but clean
	return operator
}

// cleanATResponse removes AT command artifacts from response
func (c *Client) cleanATResponse(response string) string {
	// Remove carriage returns and extra whitespace
	response = strings.ReplaceAll(response, "\r", "")
	response = strings.ReplaceAll(response, "\n\n", "\n")
	
	// Remove OK and ERROR from end
	response = strings.TrimSuffix(response, "\n\nOK")
	response = strings.TrimSuffix(response, "\nOK")
	response = strings.TrimSuffix(response, "OK")
	response = strings.TrimSuffix(response, "\n\nERROR")
	response = strings.TrimSuffix(response, "\nERROR")
	response = strings.TrimSuffix(response, "ERROR")
	
	return strings.TrimSpace(response)
}

// parseNetworkType extracts network type from AT+CREG response
func (c *Client) parseNetworkType(response string) string {
	// Example: +CREG: 0,1 or +CREG: 2,1,1A2B,C3D4,7
	response = c.cleanATResponse(response)
	for _, line := range strings.Split(response, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "+CREG:") {
			parts := strings.Split(line, ",")
			if len(parts) >= 5 {
				// The last parameter indicates access technology
				switch strings.TrimSpace(parts[4]) {
				case "0":
					return "GSM"
				case "2":
					return "UTRAN"
				case "3":
					return "GSM w/EGPRS"
				case "4":
					return "UTRAN w/HSDPA"
				case "5":
					return "UTRAN w/HSUPA"
				case "6":
					return "UTRAN w/HSDPA and HSUPA"
				case "7":
					return "E-UTRAN (LTE)"
				default:
					return "Unknown"
				}
			} else if len(parts) >= 2 {
				// Basic registration status
				status := strings.TrimSpace(parts[1])
				if status == "1" || status == "5" {
					return "GSM/GPRS"
				}
			}
		}
	}
	return "Unknown"
}

// parseSignalStrength extracts signal strength from AT+CSQ response
func (c *Client) parseSignalStrength(response string) int {
	// Example: +CSQ: 31,99
	response = c.cleanATResponse(response)
	for _, line := range strings.Split(response, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "+CSQ:") {
			parts := strings.Split(line, ",")
			if len(parts) >= 1 {
				rssiPart := strings.TrimSpace(strings.TrimPrefix(parts[0], "+CSQ:"))
				if rssi := strings.TrimSpace(rssiPart); rssi != "99" && rssi != "" {
					// Convert RSSI to dBm: -113 + (rssi * 2)
					var signal int
					if val, err := fmt.Sscanf(rssi, "%d", &signal); err == nil && val == 1 {
						if signal >= 0 && signal <= 31 {
							return -113 + (signal * 2)
						}
					}
				}
			}
		}
	}
	return 0
}

// parseUSSDResponse cleans up USSD response
func (c *Client) parseUSSDResponse(response string) string {
	// Remove AT command echoes and extract the actual response
	response = strings.ReplaceAll(response, "\r", "")
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for USSD response pattern
		if strings.Contains(line, "+CUSD:") {
			// Extract the message between quotes
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start >= 0 && end > start {
				return strings.TrimSpace(line[start+1 : end])
			}
		}
		// Sometimes the response is just plain text
		if !strings.Contains(line, "AT+") && !strings.Contains(line, "OK") &&
			!strings.Contains(line, "ERROR") && line != "" {
			return line
		}
	}

	return response
}
