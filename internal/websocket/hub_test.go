package websocket

import (
	"testing"
	"time"

	"github.com/markbfromdc/cloudcode/internal/logging"
)

func TestNewHub(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	hub := NewHub(log)

	if hub == nil {
		t.Fatal("expected non-nil hub")
	}
	if hub.clients == nil {
		t.Fatal("expected initialized clients map")
	}
	if hub.ActiveSessions() != 0 {
		t.Errorf("expected 0 active sessions, got %d", hub.ActiveSessions())
	}
}

func TestHubRegisterUnregister(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	hub := NewHub(log)
	go hub.Run()

	client := &Client{
		SessionID: "test-session-1",
		UserID:    "user-1",
		send:      make(chan []byte, 256),
	}

	hub.Register(client)
	time.Sleep(50 * time.Millisecond) // Allow goroutine to process.

	if hub.ActiveSessions() != 1 {
		t.Errorf("expected 1 active session, got %d", hub.ActiveSessions())
	}

	found, ok := hub.GetClient("test-session-1")
	if !ok {
		t.Fatal("expected to find registered client")
	}
	if found.SessionID != "test-session-1" {
		t.Errorf("expected session ID 'test-session-1', got %q", found.SessionID)
	}

	hub.Unregister(client)
	time.Sleep(50 * time.Millisecond)

	if hub.ActiveSessions() != 0 {
		t.Errorf("expected 0 active sessions after unregister, got %d", hub.ActiveSessions())
	}

	_, ok = hub.GetClient("test-session-1")
	if ok {
		t.Error("expected client to be removed after unregister")
	}
}

func TestHubGetClientNotFound(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	hub := NewHub(log)

	_, ok := hub.GetClient("nonexistent")
	if ok {
		t.Error("expected false for nonexistent client")
	}
}
