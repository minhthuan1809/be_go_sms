package modem

import (
	"context"
	"fmt"
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
	for _, port := range ports {
		info := model.PortInfo{Port: port}

		// Try to open port to query details
		mode := &serial.Mode{
			BaudRate: c.config.Modem.DefaultBaudRate,
			Parity:   serial.NoParity,
			DataBits: 8,
			StopBits: serial.OneStopBit,
		}
		p, err := serial.Open(port, mode)
		if err == nil {
			info.Available = true
			// Get device/manufacturer description
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			info.Description = c.getDeviceDescriptionWithPort(ctx, p)
			cancel()

			// Try to get MSISDN via AT+CNUM
			if msisdn, perr := c.queryMsisdn(p); perr == nil && msisdn != "" {
				info.Msisdn = msisdn
			}

			// Try to run USSD for balance/packages if configured
			if bal := strings.TrimSpace(c.config.Modem.BalanceUSSD); bal != "" {
				if resp, uerr := c.runUSSD(p, bal); uerr == nil && resp != "" {
					info.Balance = resp
				}
			}
			if pkg := strings.TrimSpace(c.config.Modem.PackagesUSSD); pkg != "" {
				if resp, uerr := c.runUSSD(p, pkg); uerr == nil && resp != "" {
					info.Packages = []string{resp}
				}
			}

			p.Close()
		} else {
			info.Available = false
			info.Error = err.Error()
		}

		info.DeviceName = c.getDeviceName(port)
		if info.Description == "" {
			info.Description = c.getDeviceDescription(port)
		}

		portInfos = append(portInfos, info)
	}

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

func (c *Client) queryMsisdn(port serial.Port) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := c.sendATCommand(ctx, port, "AT+CNUM")
	if err != nil {
		return "", err
	}
	// Typical response: +CNUM: ,"+84123456789",145
	resp = strings.ReplaceAll(resp, "\r", "")
	for _, line := range strings.Split(resp, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "+CNUM") || strings.Contains(line, "+CNUM:") {
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start >= 0 && end > start {
				return strings.TrimSpace(line[start+1 : end]), nil
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

// helper to read USSD response
type readerPort interface {
	Read([]byte) (int, error)
	SetReadTimeout(time.Duration)
}

func (c *Client) readResponse(p readerPort, ctx context.Context) (string, error) {
	deadline := time.Now().Add(10 * time.Second)
	var response strings.Builder
	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return response.String(), ctx.Err()
		default:
		}
		if time.Now().After(deadline) {
			return response.String(), fmt.Errorf("timeout")
		}
		p.SetReadTimeout(200 * time.Millisecond)
		n, err := p.Read(buf)
		if err != nil {
			continue
		}
		if n > 0 {
			response.Write(buf[:n])
			text := response.String()
			if strings.Contains(text, "+CUSD:") || strings.Contains(text, "OK") || strings.Contains(text, "ERROR") {
				return text, nil
			}
		}
	}
}
