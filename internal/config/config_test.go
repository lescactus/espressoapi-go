package config

import (
	"os"
	"testing"
	"time"
)

func TestNewWithDefaults(t *testing.T) {
	config, err := New()
	if err != nil {
		t.Fatalf("Error creating new configuration: %v", err)
	}

	// Verify that the values are set to default values
	if config.ServerAddr != defaultServerAddr {
		t.Errorf("Expected ServerAddr to be %s, got %s", defaultServerAddr, config.ServerAddr)
	}
	if config.ServerReadTimeout != defaultServerReadTimeout {
		t.Errorf("Expected ServerReadTimeout to be %v, got %v", defaultServerReadTimeout, config.ServerReadTimeout)
	}
	if config.ServerReadHeaderTimeout != defaultServerReadHeaderTimeout {
		t.Errorf("Expected ServerReadHeaderTimeout to be %v, got %v", defaultServerReadHeaderTimeout, config.ServerReadHeaderTimeout)
	}
	if config.ServerWriteTimeout != defaultServerWriteTimeout {
		t.Errorf("Expected ServerWriteTimeout to be %v, got %v", defaultServerWriteTimeout, config.ServerWriteTimeout)
	}
	if config.ServerMaxRequestSize != defaultServerMaxRequestSize {
		t.Errorf("Expected ServerMaxRequestSize to be %d, got %d", defaultServerMaxRequestSize, config.ServerMaxRequestSize)
	}
	if config.LoggerLogLevel != defaultLoggerLogLevel {
		t.Errorf("Expected LoggerLogLevel to be %s, got %s", defaultLoggerLogLevel, config.LoggerLogLevel)
	}
	if config.LoggerDurationFieldUnit != defaultLoggerDurationFieldUnit {
		t.Errorf("Expected LoggerDurationFieldUnit to be %s, got %s", defaultLoggerDurationFieldUnit, config.LoggerDurationFieldUnit)
	}
	if config.LoggerFormat != defaultLoggerFormat {
		t.Errorf("Expected LoggerFormat to be %s, got %s", defaultLoggerFormat, config.LoggerFormat)
	}
}

func TestNewWithConfigFile(t *testing.T) {
	tempConfigFile := "config.yaml"
	defer func() {
		_ = os.Remove(tempConfigFile)
	}()

	yamlContent := `
server_addr: ":9090"
server_read_timeout: 15s
logger_log_level: "debug"
`
	err := os.WriteFile(tempConfigFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Error creating temporary config file: %v", err)
	}

	os.Setenv("CONFIG_FILE", tempConfigFile)
	defer os.Unsetenv("CONFIG_FILE")

	config, err := New()
	if err != nil {
		t.Fatalf("Error creating new configuration: %v", err)
	}

	// Verify that the values are set to the values from the config file
	if config.ServerAddr != ":9090" {
		t.Errorf("Expected ServerAddr to be :9090, got %s", config.ServerAddr)
	}
	if config.ServerReadTimeout != 15*time.Second {
		t.Errorf("Expected ServerReadTimeout to be %v, got %v", 15*time.Second, config.ServerReadTimeout)
	}
	if config.ServerReadHeaderTimeout != defaultServerReadHeaderTimeout {
		t.Errorf("Expected ServerReadHeaderTimeout to be %v, got %v", defaultServerReadHeaderTimeout, config.ServerReadHeaderTimeout)
	}
	if config.ServerWriteTimeout != defaultServerWriteTimeout {
		t.Errorf("Expected ServerWriteTimeout to be %v, got %v", defaultServerWriteTimeout, config.ServerWriteTimeout)
	}
	if config.ServerMaxRequestSize != defaultServerMaxRequestSize {
		t.Errorf("Expected ServerMaxRequestSize to be %d, got %d", defaultServerMaxRequestSize, config.ServerMaxRequestSize)
	}
	if config.LoggerLogLevel != "debug" {
		t.Errorf("Expected LoggerLogLevel to be debug, got %s", config.LoggerLogLevel)
	}
	if config.LoggerDurationFieldUnit != defaultLoggerDurationFieldUnit {
		t.Errorf("Expected LoggerDurationFieldUnit to be %s, got %s", defaultLoggerDurationFieldUnit, config.LoggerDurationFieldUnit)
	}
	if config.LoggerFormat != defaultLoggerFormat {
		t.Errorf("Expected LoggerFormat to be %s, got %s", defaultLoggerFormat, config.LoggerFormat)
	}
}

func TestNewWithInvalidConfigFile(t *testing.T) {
	tempConfigFile := "config.yaml"
	defer func() {
		_ = os.Remove(tempConfigFile)
	}()

	invalidYamlContent := `
invalid yaml
`
	err := os.WriteFile(tempConfigFile, []byte(invalidYamlContent), 0644)
	if err != nil {
		t.Fatalf("Error creating temporary invalid config file: %v", err)
	}

	os.Setenv("CONFIG_FILE", tempConfigFile)
	defer os.Unsetenv("CONFIG_FILE")

	_, err = New()
	if err == nil {
		t.Error("Expected error when creating new configuration with invalid config file, but got nil")
	}
}

func TestNewWithEnvVars(t *testing.T) {
	os.Setenv("SERVER_ADDR", ":9090")
	os.Setenv("SERVER_READ_TIMEOUT", "15s")
	os.Setenv("LOGGER_LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("SERVER_ADDR")
		os.Unsetenv("SERVER_READ_TIMEOUT")
		os.Unsetenv("LOGGER_LOG_LEVEL")
	}()

	config, err := New()
	if err != nil {
		t.Fatalf("Error creating new configuration: %v", err)
	}

	// Verify that the values are set to the values from environment variables
	if config.ServerAddr != ":9090" {
		t.Errorf("Expected ServerAddr to be :9090, got %s", config.ServerAddr)
	}
	if config.ServerReadTimeout != 15*time.Second {
		t.Errorf("Expected ServerReadTimeout to be %v, got %v", 15*time.Second, config.ServerReadTimeout)
	}
	if config.ServerReadHeaderTimeout != defaultServerReadHeaderTimeout {
		t.Errorf("Expected ServerReadHeaderTimeout to be %v, got %v", defaultServerReadHeaderTimeout, config.ServerReadHeaderTimeout)
	}
	if config.ServerWriteTimeout != defaultServerWriteTimeout {
		t.Errorf("Expected ServerWriteTimeout to be %v, got %v", defaultServerWriteTimeout, config.ServerWriteTimeout)
	}
	if config.ServerMaxRequestSize != defaultServerMaxRequestSize {
		t.Errorf("Expected ServerMaxRequestSize to be %d, got %d", defaultServerMaxRequestSize, config.ServerMaxRequestSize)
	}
	if config.LoggerLogLevel != "debug" {
		t.Errorf("Expected LoggerLogLevel to be debug, got %s", config.LoggerLogLevel)
	}
	if config.LoggerDurationFieldUnit != defaultLoggerDurationFieldUnit {
		t.Errorf("Expected LoggerDurationFieldUnit to be %s, got %s", defaultLoggerDurationFieldUnit, config.LoggerDurationFieldUnit)
	}
	if config.LoggerFormat != defaultLoggerFormat {
		t.Errorf("Expected LoggerFormat to be %s, got %s", defaultLoggerFormat, config.LoggerFormat)
	}
}
