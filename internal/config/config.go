// Package config provides application configuration management for the Cloud IDE backend.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration values for the Cloud IDE backend server.
type Config struct {
	// Server settings
	HTTPPort  int
	GRPCPort  int
	HostAddr  string
	TLSCert   string
	TLSKey    string
	EnableTLS bool

	// WebSocket settings
	WSReadBufferSize  int
	WSWriteBufferSize int
	WSPingInterval    time.Duration
	WSPongTimeout     time.Duration
	WSWriteTimeout    time.Duration
	WSMaxMessageSize  int64

	// Container settings
	DockerHost          string
	WorkspaceImage      string
	ContainerMemoryMB   int64
	ContainerCPUShares  int64
	ContainerTimeoutMin int
	NetworkName         string

	// Security settings
	AllowedOrigins []string
	JWTSecret      string
	SessionTimeout time.Duration
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort:  getEnvInt("HTTP_PORT", 8080),
		GRPCPort:  getEnvInt("GRPC_PORT", 9090),
		HostAddr:  getEnvStr("HOST_ADDR", "0.0.0.0"),
		TLSCert:   getEnvStr("TLS_CERT_PATH", ""),
		TLSKey:    getEnvStr("TLS_KEY_PATH", ""),
		EnableTLS: getEnvBool("ENABLE_TLS", false),

		WSReadBufferSize:  getEnvInt("WS_READ_BUFFER_SIZE", 8192),
		WSWriteBufferSize: getEnvInt("WS_WRITE_BUFFER_SIZE", 8192),
		WSPingInterval:    time.Duration(getEnvInt("WS_PING_INTERVAL_SEC", 30)) * time.Second,
		WSPongTimeout:     time.Duration(getEnvInt("WS_PONG_TIMEOUT_SEC", 40)) * time.Second,
		WSWriteTimeout:    time.Duration(getEnvInt("WS_WRITE_TIMEOUT_SEC", 10)) * time.Second,
		WSMaxMessageSize:  int64(getEnvInt("WS_MAX_MESSAGE_SIZE", 65536)),

		DockerHost:          getEnvStr("DOCKER_HOST", "unix:///var/run/docker.sock"),
		WorkspaceImage:      getEnvStr("WORKSPACE_IMAGE", "cloudide-workspace:latest"),
		ContainerMemoryMB:   int64(getEnvInt("CONTAINER_MEMORY_MB", 4096)),
		ContainerCPUShares:  int64(getEnvInt("CONTAINER_CPU_SHARES", 2048)),
		ContainerTimeoutMin: getEnvInt("CONTAINER_TIMEOUT_MIN", 480),
		NetworkName:         getEnvStr("DOCKER_NETWORK", "cloudide-net"),

		AllowedOrigins: []string{getEnvStr("ALLOWED_ORIGINS", "https://ide.cloudcode.dev")},
		JWTSecret:      getEnvStr("JWT_SECRET", ""),
		SessionTimeout: time.Duration(getEnvInt("SESSION_TIMEOUT_HOURS", 24)) * time.Hour,
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// validate checks that required configuration values are present.
func (c *Config) validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}
	if c.EnableTLS && (c.TLSCert == "" || c.TLSKey == "") {
		return fmt.Errorf("TLS_CERT_PATH and TLS_KEY_PATH are required when TLS is enabled")
	}
	return nil
}

// HTTPAddr returns the full HTTP listen address.
func (c *Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.HostAddr, c.HTTPPort)
}

// GRPCAddr returns the full gRPC listen address.
func (c *Config) GRPCAddr() string {
	return fmt.Sprintf("%s:%d", c.HostAddr, c.GRPCPort)
}

func getEnvStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
