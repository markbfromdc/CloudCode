package websocket

import (
	"testing"
	"time"

	"github.com/markbfromdc/cloudcode/internal/logging"
)

func TestNewClient(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	hub := NewHub(log)

	client := NewClient(ClientConfig{
		SessionID:      "test-123",
		UserID:         "user-456",
		ContainerID:    "container-789",
		Conn:           nil,
		Hub:            hub,
		ContainerWriter: nil,
		PingInterval:   30 * time.Second,
		PongTimeout:    40 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxMessageSize: 65536,
		Log:            log,
	})

	if client.SessionID != "test-123" {
		t.Errorf("expected session ID 'test-123', got %q", client.SessionID)
	}
	if client.UserID != "user-456" {
		t.Errorf("expected user ID 'user-456', got %q", client.UserID)
	}
	if client.ContainerID != "container-789" {
		t.Errorf("expected container ID 'container-789', got %q", client.ContainerID)
	}
	if client.pingInterval != 30*time.Second {
		t.Errorf("expected 30s ping interval, got %v", client.pingInterval)
	}
	if client.maxMessageSize != 65536 {
		t.Errorf("expected 65536 max message size, got %d", client.maxMessageSize)
	}
}

func TestClientSendBufferFull(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	hub := NewHub(log)

	client := NewClient(ClientConfig{
		SessionID:      "test-buffer",
		UserID:         "user-1",
		Conn:           nil,
		Hub:            hub,
		PingInterval:   30 * time.Second,
		PongTimeout:    40 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxMessageSize: 65536,
		Log:            log,
	})

	// Fill the send buffer (capacity 256).
	for i := 0; i < 256; i++ {
		client.Send([]byte("x"))
	}

	// This should not block — message should be dropped.
	done := make(chan struct{})
	go func() {
		client.Send([]byte("overflow"))
		close(done)
	}()

	select {
	case <-done:
		// Success — didn't block.
	case <-time.After(time.Second):
		t.Fatal("Send() blocked on full buffer")
	}
}

func TestClientSendQueuesMessage(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	hub := NewHub(log)

	client := NewClient(ClientConfig{
		SessionID:      "test-queue",
		UserID:         "user-1",
		Conn:           nil,
		Hub:            hub,
		PingInterval:   30 * time.Second,
		PongTimeout:    40 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxMessageSize: 65536,
		Log:            log,
	})

	client.Send([]byte("hello"))

	select {
	case msg := <-client.send:
		if string(msg) != "hello" {
			t.Errorf("expected 'hello', got %q", string(msg))
		}
	default:
		t.Error("expected message in send channel")
	}
}
