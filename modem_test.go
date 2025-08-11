package main

import (
	"bufio"
	"fmt"
	"log"
	"strings"
	"time"

	serial "go.bug.st/serial"
)

func main() {
	portName := "/dev/ttyUSB0"
	baudRate := 115200

	fmt.Printf("Testing modem on port %s at %d baud...\n", portName, baudRate)

	mode := &serial.Mode{
		BaudRate: baudRate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		log.Fatalf("Failed to open port: %v", err)
	}
	defer port.Close()

	fmt.Println("Port opened successfully")

	// Test basic AT commands
	commands := []struct {
		command  string
		expected string
		desc     string
	}{
		{"AT", "OK", "Basic AT test"},
		{"ATE0", "OK", "Disable echo"},
		{"AT+CPIN?", "OK", "Check SIM card"},
		{"AT+CREG?", "OK", "Check network registration"},
		{"AT+CSQ", "OK", "Check signal quality"},
		{"AT+CMGF?", "OK", "Check SMS mode"},
		{"AT+CMGF=1", "OK", "Set SMS to text mode"},
		{"AT+CMGF=0", "OK", "Set SMS to PDU mode"},
	}

	for _, cmd := range commands {
		fmt.Printf("\n--- Testing: %s ---\n", cmd.desc)
		fmt.Printf("Command: %s\n", cmd.command)

		// Send command
		_, err := port.Write([]byte(cmd.command + "\r\n"))
		if err != nil {
			fmt.Printf("Failed to write: %v\n", err)
			continue
		}

		// Read response
		response, err := readUntil(port, 5*time.Second, cmd.expected)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Printf("Response: %s\n", response)
		} else {
			fmt.Printf("Success: %s\n", response)
		}
	}

	// Test SMS sending in text mode
	fmt.Printf("\n--- Testing SMS in text mode ---\n")

	// Set to text mode
	port.Write([]byte("AT+CMGF=1\r\n"))
	readUntil(port, 5*time.Second, "OK")

	// Try to send SMS
	port.Write([]byte("AT+CMGS=\"0123456789\"\r\n"))
	response, err := readUntil(port, 5*time.Second, ">")
	if err != nil {
		fmt.Printf("Failed to get prompt: %v\n", err)
	} else {
		fmt.Printf("Got prompt: %s\n", response)

		// Send message
		port.Write([]byte("Test message\x1A"))
		response, err = readUntil(port, 10*time.Second, "OK")
		if err != nil {
			fmt.Printf("Failed to send SMS: %v\n", err)
			fmt.Printf("Response: %s\n", response)
		} else {
			fmt.Printf("SMS sent successfully: %s\n", response)
		}
	}
}

func readUntil(port serial.Port, timeout time.Duration, expected string) (string, error) {
	deadline := time.Now().Add(timeout)
	var response strings.Builder
	reader := bufio.NewReader(port)

	for {
		if time.Now().After(deadline) {
			return response.String(), fmt.Errorf("timeout waiting for: %s", expected)
		}

		port.SetReadTimeout(100 * time.Millisecond)
		b, err := reader.ReadByte()
		if err != nil {
			continue
		}

		response.WriteByte(b)
		text := response.String()

		if strings.Contains(strings.ToUpper(text), strings.ToUpper(expected)) {
			return text, nil
		}
	}
}
