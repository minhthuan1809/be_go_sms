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
