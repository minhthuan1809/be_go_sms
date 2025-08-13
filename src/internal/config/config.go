package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds application configuration
type Config struct {
	Version string
	Server  ServerConfig
	Modem   ModemConfig
	SMS     SMSConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Address      string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

// ModemConfig holds modem configuration
type ModemConfig struct {
	DefaultPort     string
	DefaultBaudRate int
	Timeout         time.Duration
	BalanceUSSD     string
	PackagesUSSD    string
}

// SMSConfig holds SMS configuration
type SMSConfig struct {
	MaxLength      int
	DefaultTimeout int
	RetryCount     int
	RetryDelay     time.Duration
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Address:      getEnv("SERVER_ADDRESS", ":3333"),
			ReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 10),
			WriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 10),
			IdleTimeout:  getEnvAsInt("SERVER_IDLE_TIMEOUT", 120),
		},
		Modem: ModemConfig{
			DefaultPort:     getEnv("MODEM_DEFAULT_PORT", "/dev/ttyUSB0"),
			DefaultBaudRate: getEnvAsInt("MODEM_DEFAULT_BAUDRATE", 115200),
			Timeout:         time.Duration(getEnvAsInt("MODEM_TIMEOUT", 30)) * time.Second,
			BalanceUSSD:     getEnv("MODEM_BALANCE_USSD", ""),
			PackagesUSSD:    getEnv("MODEM_PACKAGES_USSD", ""),
		},
		SMS: SMSConfig{
			MaxLength:      getEnvAsInt("SMS_MAX_LENGTH", 160),
			DefaultTimeout: getEnvAsInt("SMS_DEFAULT_TIMEOUT", 30),
			RetryCount:     getEnvAsInt("SMS_RETRY_COUNT", 3),
			RetryDelay:     time.Duration(getEnvAsInt("SMS_RETRY_DELAY", 2)) * time.Second,
		},
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
