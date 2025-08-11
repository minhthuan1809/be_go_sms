package service

import (
	"context"
	"strings"
	"time"

	serial "go.bug.st/serial"
)

// BalanceChecker handles SIM balance checking operations
type BalanceChecker struct{}

// NewBalanceChecker creates a new balance checker
func NewBalanceChecker() *BalanceChecker {
	return &BalanceChecker{}
}

// CheckBalance checks SIM balance using USSD command
func (b *BalanceChecker) CheckBalance(ctx context.Context, port serial.Port, atHandler *ATCommandHandler) (string, error) {
	// Send USSD command to check balance
	_, err := atHandler.CheckBalance(ctx, port)
	if err != nil {
		return "", err
	}

	// Wait for USSD response
	time.Sleep(3 * time.Second)

	// Read response using SMS service's readUntilWithContext
	smsService := &SMSService{}
	balanceResp, err := smsService.readUntilWithContext(ctx, port, 10*time.Second)
	if err != nil {
		return "", err
	}

	return balanceResp, nil
}

// IsBalanceInsufficient checks if the balance response indicates insufficient funds
func (b *BalanceChecker) IsBalanceInsufficient(balanceResponse string) bool {
	lowerResp := strings.ToLower(balanceResponse)

	// Check for various low balance indicators
	return strings.Contains(lowerResp, "het tien") ||
		strings.Contains(lowerResp, "hết tiền") ||
		strings.Contains(lowerResp, "insufficient") ||
		strings.Contains(lowerResp, "balance") ||
		strings.Contains(lowerResp, "0 vnd") ||
		strings.Contains(lowerResp, "0đ") ||
		strings.Contains(lowerResp, "credit") ||
		strings.Contains(lowerResp, "low balance")
}

// GetBalanceInfo extracts balance information from USSD response
func (b *BalanceChecker) GetBalanceInfo(balanceResponse string) string {
	// This is a simple implementation - you might want to add more sophisticated parsing
	// based on the actual USSD response format from your carrier
	return balanceResponse
}
