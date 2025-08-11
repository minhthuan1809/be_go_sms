package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	serial "go.bug.st/serial"
)

// ATCommandHandler handles AT command operations
type ATCommandHandler struct{}

// NewATCommandHandler creates a new AT command handler
func NewATCommandHandler() *ATCommandHandler {
	return &ATCommandHandler{}
}

// writeAndRead sends a command and reads the response
func (h *ATCommandHandler) writeAndRead(ctx context.Context, port serial.Port, cmd string, expectPrompts ...string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if !strings.HasSuffix(cmd, "\r") {
		cmd += "\r"
	}

	if _, err := port.Write([]byte(cmd)); err != nil {
		return "", fmt.Errorf("failed to write command '%s': %w", strings.TrimSpace(cmd), err)
	}

	// Wait a bit for modem to process
	time.Sleep(200 * time.Millisecond)

	// Use the existing readUntilWithContext from SMS service
	smsService := &SMSService{}
	resp, err := smsService.readUntilWithContext(ctx, port, 5*time.Second, expectPrompts...)
	return resp, err
}

// TestModemConnection tests basic modem connectivity
func (h *ATCommandHandler) TestModemConnection(ctx context.Context, port serial.Port) (string, error) {
	return h.writeAndRead(ctx, port, "AT")
}

// DisableEcho disables command echo
func (h *ATCommandHandler) DisableEcho(ctx context.Context, port serial.Port) (string, error) {
	return h.writeAndRead(ctx, port, "ATE0")
}

// CheckSignalQuality checks signal strength
func (h *ATCommandHandler) CheckSignalQuality(ctx context.Context, port serial.Port) (string, error) {
	return h.writeAndRead(ctx, port, "AT+CSQ")
}

// CheckSIMStatus checks SIM card status
func (h *ATCommandHandler) CheckSIMStatus(ctx context.Context, port serial.Port) (string, error) {
	return h.writeAndRead(ctx, port, "AT+CPIN?")
}

// CheckNetworkRegistration checks network registration status
func (h *ATCommandHandler) CheckNetworkRegistration(ctx context.Context, port serial.Port) (string, error) {
	return h.writeAndRead(ctx, port, "AT+CREG?")
}

// CheckSMSCenter checks SMS service center
func (h *ATCommandHandler) CheckSMSCenter(ctx context.Context, port serial.Port) (string, error) {
	return h.writeAndRead(ctx, port, "AT+CSCA?")
}

// CheckSMSMemory checks SMS memory status
func (h *ATCommandHandler) CheckSMSMemory(ctx context.Context, port serial.Port) (string, error) {
	return h.writeAndRead(ctx, port, "AT+CPMS?")
}

// CheckOperator checks operator name
func (h *ATCommandHandler) CheckOperator(ctx context.Context, port serial.Port) (string, error) {
	return h.writeAndRead(ctx, port, "AT+COPS?")
}

// CheckBalance checks SIM balance using USSD
func (h *ATCommandHandler) CheckBalance(ctx context.Context, port serial.Port) (string, error) {
	return h.writeAndRead(ctx, port, "AT+CUSD=1,\"*101#\"")
}

// ResetModem resets the modem
func (h *ATCommandHandler) ResetModem(ctx context.Context, port serial.Port) (string, error) {
	return h.writeAndRead(ctx, port, "AT+CFUN=1,1")
}

// SetSMSTextMode sets SMS to text mode
func (h *ATCommandHandler) SetSMSTextMode(ctx context.Context, port serial.Port) (string, error) {
	return h.writeAndRead(ctx, port, "AT+CMGF=1")
}

// SetSMSParameters sets SMS text mode parameters
func (h *ATCommandHandler) SetSMSParameters(ctx context.Context, port serial.Port) (string, error) {
	return h.writeAndRead(ctx, port, "AT+CSMP=17,167,0,0")
}

// SetSMSPDUMode sets SMS to PDU mode
func (h *ATCommandHandler) SetSMSPDUMode(ctx context.Context, port serial.Port) (string, error) {
	return h.writeAndRead(ctx, port, "AT+CMGF=0")
}

// InitiateSMSSend initiates SMS sending
func (h *ATCommandHandler) InitiateSMSSend(ctx context.Context, port serial.Port, phoneNumber string) (string, error) {
	cmd := fmt.Sprintf("AT+CMGS=\"%s\"", phoneNumber)
	return h.writeAndRead(ctx, port, cmd, "> ", ">")
}

// InitiatePDUSMSSend initiates PDU SMS sending
func (h *ATCommandHandler) InitiatePDUSMSSend(ctx context.Context, port serial.Port, messageLength int) (string, error) {
	cmd := fmt.Sprintf("AT+CMGS=%d", messageLength)
	return h.writeAndRead(ctx, port, cmd, "> ", ">")
}
