package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"myproject/internal/utils"

	serial "go.bug.st/serial"
)

// SMSSender handles the main SMS sending logic
type SMSSender struct {
	atHandler      *ATCommandHandler
	serialManager  *SerialManager
	errorHandler   *ErrorHandler
	balanceChecker *BalanceChecker
}

// NewSMSSender creates a new SMS sender
func NewSMSSender() *SMSSender {
	return &SMSSender{
		atHandler:      NewATCommandHandler(),
		serialManager:  NewSerialManager(),
		errorHandler:   NewErrorHandler(),
		balanceChecker: NewBalanceChecker(),
	}
}

// SendSMS sends an SMS using the modular approach
func (s *SMSSender) SendSMS(ctx context.Context, portName string, baudRate int, to string, message string) ([]string, string, error) {
	// Lock port để tránh xung đột
	s.serialManager.LockPort()
	defer s.serialManager.UnlockPort()

	// Open serial port
	port, err := s.serialManager.OpenPort(portName, baudRate)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open serial port %s: %w", portName, err)
	}
	defer s.serialManager.ClosePort(port)

	var steps []string
	var messageID string

	// Step 1: Test modem connection
	log.Printf("Testing modem connection...")
	resp, err := s.atHandler.TestModemConnection(ctx, port)
	if err != nil {
		steps = append(steps, fmt.Sprintf("AT -> ERROR: %s", utils.Sanitize(err.Error())))
		return steps, messageID, fmt.Errorf("modem not responding: %w", err)
	}
	steps = append(steps, fmt.Sprintf("AT -> %s", utils.Sanitize(resp)))

	// Step 2: Disable echo
	resp, err = s.atHandler.DisableEcho(ctx, port)
	if err != nil {
		steps = append(steps, fmt.Sprintf("ATE0 -> ERROR: %s", utils.Sanitize(err.Error())))
		// Continue anyway, echo doesn't affect SMS sending
	} else {
		steps = append(steps, fmt.Sprintf("ATE0 -> %s", utils.Sanitize(resp)))
	}

	// Step 3: Check modem status
	steps = s.checkModemStatus(ctx, port, steps)

	// Step 4: Check balance
	steps, err = s.checkBalance(ctx, port, steps)
	if err != nil {
		return steps, messageID, err
	}

	// Step 5: Reset modem
	steps = s.resetModem(ctx, port, steps)

	// Step 6: Send SMS
	messageID, err = s.performSMSSend(ctx, port, to, message, steps)
	if err != nil {
		return steps, messageID, err
	}

	log.Printf("SMS sent successfully. Message ID: %s", messageID)
	return steps, messageID, nil
}

// checkModemStatus checks various modem status indicators
func (s *SMSSender) checkModemStatus(ctx context.Context, port serial.Port, steps []string) []string {
	// Check signal quality
	if resp, err := s.atHandler.CheckSignalQuality(ctx, port); err == nil {
		steps = append(steps, fmt.Sprintf("AT+CSQ -> %s", utils.Sanitize(resp)))
	}

	// Check SIM card status
	if resp, err := s.atHandler.CheckSIMStatus(ctx, port); err == nil {
		steps = append(steps, fmt.Sprintf("AT+CPIN? -> %s", utils.Sanitize(resp)))
	}

	// Check network registration
	if resp, err := s.atHandler.CheckNetworkRegistration(ctx, port); err == nil {
		steps = append(steps, fmt.Sprintf("AT+CREG? -> %s", utils.Sanitize(resp)))
	}

	// Check SMS service center
	if resp, err := s.atHandler.CheckSMSCenter(ctx, port); err == nil {
		steps = append(steps, fmt.Sprintf("AT+CSCA? -> %s", utils.Sanitize(resp)))
	}

	// Check SMS memory status
	if resp, err := s.atHandler.CheckSMSMemory(ctx, port); err == nil {
		steps = append(steps, fmt.Sprintf("AT+CPMS? -> %s", utils.Sanitize(resp)))
	}

	// Check operator name
	if resp, err := s.atHandler.CheckOperator(ctx, port); err == nil {
		steps = append(steps, fmt.Sprintf("AT+COPS? -> %s", utils.Sanitize(resp)))
	}

	return steps
}

// checkBalance checks SIM balance
func (s *SMSSender) checkBalance(ctx context.Context, port serial.Port, steps []string) ([]string, error) {
	// Check balance using USSD
	resp, err := s.atHandler.CheckBalance(ctx, port)
	if err == nil {
		steps = append(steps, fmt.Sprintf("AT+CUSD=1,\"*101#\" -> %s", utils.Sanitize(resp)))

		// Wait for USSD response
		time.Sleep(3 * time.Second)
		balanceResp, err := s.balanceChecker.CheckBalance(ctx, port, s.atHandler)
		if err == nil {
			steps = append(steps, fmt.Sprintf("Balance response -> %s", utils.Sanitize(balanceResp)))

			// Check if balance is insufficient
			if s.balanceChecker.IsBalanceInsufficient(balanceResp) {
				return steps, fmt.Errorf("SIM hết tiền hoặc số dư không đủ để gửi SMS")
			}
		}
	}

	return steps, nil
}

// resetModem resets the modem
func (s *SMSSender) resetModem(ctx context.Context, port serial.Port, steps []string) []string {
	resp, err := s.atHandler.ResetModem(ctx, port)
	if err == nil {
		steps = append(steps, fmt.Sprintf("AT+CFUN=1,1 -> %s", utils.Sanitize(resp)))
		// Wait for modem to restart
		time.Sleep(5 * time.Second)

		// Test AT again after restart
		if resp, err := s.atHandler.TestModemConnection(ctx, port); err == nil {
			steps = append(steps, fmt.Sprintf("AT (after restart) -> %s", utils.Sanitize(resp)))
		}
	}

	return steps
}

// performSMSSend performs the actual SMS sending
func (s *SMSSender) performSMSSend(ctx context.Context, port serial.Port, to string, message string, steps []string) (string, error) {
	// Set SMS text mode
	resp, err := s.atHandler.SetSMSTextMode(ctx, port)
	if err != nil {
		steps = append(steps, fmt.Sprintf("AT+CMGF=1 -> ERROR: %s", utils.Sanitize(err.Error())))
		return "", fmt.Errorf("failed to set text mode: %w", err)
	}
	steps = append(steps, fmt.Sprintf("AT+CMGF=1 -> %s", utils.Sanitize(resp)))

	// Set SMS text mode parameters
	if resp, err := s.atHandler.SetSMSParameters(ctx, port); err == nil {
		steps = append(steps, fmt.Sprintf("AT+CSMP=17,167,0,0 -> %s", utils.Sanitize(resp)))
	}

	// Begin SMS composition
	cmgsCmd := fmt.Sprintf("AT+CMGS=\"%s\"", to)
	resp, err = s.atHandler.InitiateSMSSend(ctx, port, to)
	if err != nil {
		steps = append(steps, fmt.Sprintf("%s -> ERROR: %s", cmgsCmd, utils.Sanitize(err.Error())))
		return "", fmt.Errorf("failed to initiate SMS send: %w", err)
	}
	steps = append(steps, fmt.Sprintf("%s -> %s", cmgsCmd, utils.Sanitize(resp)))

	// Check if we got the prompt
	if !s.hasPrompt(resp) {
		// Try alternative approach - send SMS in PDU mode
		log.Printf("Trying PDU mode as fallback...")
		steps = append(steps, "Trying PDU mode as fallback...")

		// Set PDU mode
		if resp, err := s.atHandler.SetSMSPDUMode(ctx, port); err == nil {
			steps = append(steps, fmt.Sprintf("AT+CMGF=0 -> %s", utils.Sanitize(resp)))

			// Try PDU SMS send
			pduCmd := fmt.Sprintf("AT+CMGS=%d", len(message))
			if resp, err := s.atHandler.InitiatePDUSMSSend(ctx, port, len(message)); err == nil && s.hasPrompt(resp) {
				steps = append(steps, fmt.Sprintf("%s -> %s", pduCmd, utils.Sanitize(resp)))
				// Continue with PDU mode
			} else {
				return "", errors.New("modem did not provide SMS input prompt in both text and PDU modes")
			}
		} else {
			return "", errors.New("modem did not provide SMS input prompt")
		}
	}

	// Send message content
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	log.Printf("Sending message content...")
	if _, err := port.Write([]byte(message)); err != nil {
		steps = append(steps, fmt.Sprintf("Message content -> ERROR: %s", err.Error()))
		return "", fmt.Errorf("failed to write message: %w", err)
	}

	// Send Ctrl+Z to submit
	if _, err := port.Write([]byte{26}); err != nil { // 26 = Ctrl+Z
		steps = append(steps, fmt.Sprintf("Ctrl+Z -> ERROR: %s", err.Error()))
		return "", fmt.Errorf("failed to send Ctrl+Z: %w", err)
	}

	// Wait for final response
	log.Printf("Waiting for final response...")
	smsService := &SMSService{}
	finalResp, err := smsService.readUntilWithContext(ctx, port, 30*time.Second, "OK", "ERROR", "+CMGS:")
	steps = append(steps, fmt.Sprintf("Final response -> %s", utils.Sanitize(finalResp)))

	if err != nil {
		return "", fmt.Errorf("timeout waiting for final response: %w", err)
	}

	// Process the response
	messageID, err := s.errorHandler.ProcessSMSResponse(finalResp)
	return messageID, err
}

// hasPrompt checks if the response contains the SMS input prompt
func (s *SMSSender) hasPrompt(response string) bool {
	return strings.Contains(response, ">")
}
