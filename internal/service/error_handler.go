package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrorHandler handles SMS error processing
type ErrorHandler struct{}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// ProcessSMSResponse processes the final SMS response and extracts errors
func (h *ErrorHandler) ProcessSMSResponse(response string) (string, error) {
	// Check for errors
	if strings.Contains(strings.ToUpper(response), "ERROR") {
		// Extract CMS error code if available
		if strings.Contains(response, "+CMS ERROR") {
			// Try to extract error code
			re := regexp.MustCompile(`\+CMS ERROR:\s*(\d+)`)
			matches := re.FindStringSubmatch(response)
			if len(matches) > 1 {
				errorCode := matches[1]
				errorMsg := h.getCMSErrorMessage(errorCode)

				// Check for specific error codes related to balance
				switch errorCode {
				case "1", "3", "8", "10", "21", "27", "28", "30", "38", "41", "42", "47", "50":
					return "", fmt.Errorf("SIM hết tiền hoặc số dư không đủ để gửi SMS (Error %s: %s)", errorCode, errorMsg)
				default:
					return "", fmt.Errorf("CMS ERROR %s: %s", errorCode, errorMsg)
				}
			} else {
				// No specific error code, check for common patterns
				return "", h.analyzeGenericError(response)
			}
		}
		return "", errors.New("modem returned ERROR")
	}

	// Check for successful OK
	if !strings.Contains(strings.ToUpper(response), "OK") {
		return "", errors.New("SMS send did not complete successfully")
	}

	// Extract message ID if available
	messageID := h.extractMessageID(response)
	return messageID, nil
}

// analyzeGenericError analyzes generic error responses
func (h *ErrorHandler) analyzeGenericError(response string) error {
	lowerResp := strings.ToLower(response)
	if strings.Contains(lowerResp, "sim") ||
		strings.Contains(lowerResp, "balance") ||
		strings.Contains(lowerResp, "insufficient") ||
		strings.Contains(lowerResp, "credit") {
		return fmt.Errorf("SIM hết tiền hoặc số dư không đủ để gửi SMS")
	} else if strings.Contains(lowerResp, "network") ||
		strings.Contains(lowerResp, "signal") {
		return fmt.Errorf("Lỗi mạng hoặc tín hiệu yếu")
	} else if strings.Contains(lowerResp, "number") ||
		strings.Contains(lowerResp, "phone") {
		return fmt.Errorf("Số điện thoại không đúng hoặc không tồn tại")
	} else {
		// Generic error with suggestions
		return fmt.Errorf("Lỗi gửi SMS. Nguyên nhân có thể: SIM hết tiền, số điện thoại sai, hoặc lỗi mạng")
	}
}

// extractMessageID extracts message ID from successful response
func (h *ErrorHandler) extractMessageID(response string) string {
	if strings.Contains(response, "+CMGS:") {
		re := regexp.MustCompile(`\+CMGS:\s*(\d+)`)
		matches := re.FindStringSubmatch(response)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// getCMSErrorMessage returns human-readable message for CMS error codes
func (h *ErrorHandler) getCMSErrorMessage(code string) string {
	switch code {
	case "1":
		return "Unassigned (unallocated) number"
	case "3":
		return "No route to destination"
	case "8":
		return "Operator determined barring"
	case "10":
		return "Call barred"
	case "21":
		return "Short message transfer rejected"
	case "27":
		return "Destination out of service"
	case "28":
		return "Unidentified subscriber"
	case "29":
		return "Facility rejected"
	case "30":
		return "Unknown subscriber"
	case "38":
		return "Network out of order"
	case "41":
		return "Temporary failure"
	case "42":
		return "Congestion"
	case "47":
		return "Resources unavailable"
	case "50":
		return "Requested facility not subscribed"
	case "69":
		return "Requested facility not implemented"
	case "81":
		return "Invalid short message transfer reference value"
	case "95":
		return "Invalid message, unspecified"
	case "96":
		return "Invalid mandatory information"
	case "97":
		return "Message type non-existent or not implemented"
	case "98":
		return "Message not compatible with short message protocol state"
	case "99":
		return "Information element non-existent or not implemented"
	case "111":
		return "Protocol error, unspecified"
	case "127":
		return "Interworking, unspecified"
	default:
		return "Unknown error code"
	}
}
