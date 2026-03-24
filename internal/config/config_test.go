package config

import (
	"os"
	"testing"
)

func TestLoadRequiresJWTSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is not set")
	}
}

func TestLoadWithValidConfig(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HTTPPort != 8080 {
		t.Errorf("expected default HTTP port 8080, got %d", cfg.HTTPPort)
	}
	if cfg.GRPCPort != 9090 {
		t.Errorf("expected default gRPC port 9090, got %d", cfg.GRPCPort)
	}
	if cfg.ContainerMemoryMB != 4096 {
		t.Errorf("expected default container memory 4096MB, got %d", cfg.ContainerMemoryMB)
	}
}

func TestLoadCustomPorts(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("HTTP_PORT", "3000")
	os.Setenv("GRPC_PORT", "50051")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("GRPC_PORT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HTTPPort != 3000 {
		t.Errorf("expected HTTP port 3000, got %d", cfg.HTTPPort)
	}
	if cfg.GRPCPort != 50051 {
		t.Errorf("expected gRPC port 50051, got %d", cfg.GRPCPort)
	}
}

func TestHTTPAddr(t *testing.T) {
	cfg := &Config{HostAddr: "0.0.0.0", HTTPPort: 8080}
	expected := "0.0.0.0:8080"
	if addr := cfg.HTTPAddr(); addr != expected {
		t.Errorf("expected %q, got %q", expected, addr)
	}
}

func TestGRPCAddr(t *testing.T) {
	cfg := &Config{HostAddr: "0.0.0.0", GRPCPort: 9090}
	expected := "0.0.0.0:9090"
	if addr := cfg.GRPCAddr(); addr != expected {
		t.Errorf("expected %q, got %q", expected, addr)
	}
}

func TestTLSValidation(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("ENABLE_TLS", "true")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ENABLE_TLS")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when TLS enabled without cert/key paths")
	}
}
