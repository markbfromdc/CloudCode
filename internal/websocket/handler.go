package websocket

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/markbfromdc/cloudcode/internal/config"
	"github.com/markbfromdc/cloudcode/internal/container"
	"github.com/markbfromdc/cloudcode/internal/logging"
)

// Handler manages HTTP-to-WebSocket upgrade requests and wires up the connection
// between the browser client and the container PTY.
type Handler struct {
	hub       *Hub
	upgrader  websocket.Upgrader
	cfg       *config.Config
	container *container.Manager
	log       *logging.Logger
}

// NewHandler creates a new WebSocket HTTP handler.
func NewHandler(hub *Hub, cfg *config.Config, cm *container.Manager, log *logging.Logger) *Handler {
	return &Handler{
		hub: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  cfg.WSReadBufferSize,
			WriteBufferSize: cfg.WSWriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				for _, allowed := range cfg.AllowedOrigins {
					if origin == allowed {
						return true
					}
				}
				return false
			},
		},
		cfg:       cfg,
		container: cm,
		log:       log.WithField("component", "ws-handler"),
	}
}

// ServeHTTP handles the WebSocket upgrade and establishes the bi-directional
// data flow between the browser terminal and the workspace container.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract session and user info from request context (set by auth middleware).
	sessionID := r.URL.Query().Get("session_id")
	userID := r.Context().Value(contextKeyUserID)
	if sessionID == "" || userID == nil {
		h.log.Warn("missing session_id or user_id in request")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		http.Error(w, "invalid user context", http.StatusInternalServerError)
		return
	}

	// Upgrade HTTP connection to WebSocket.
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("websocket upgrade failed: %v", err)
		return
	}

	// Attach to the container's exec session to get stdin writer and stdout reader.
	execSession, err := h.container.AttachToContainer(r.Context(), sessionID)
	if err != nil {
		h.log.Error("failed to attach to container: %v", err)
		conn.Close()
		return
	}

	client := NewClient(ClientConfig{
		SessionID:       sessionID,
		UserID:          userIDStr,
		ContainerID:     execSession.ContainerID,
		Conn:            conn,
		Hub:             h.hub,
		ContainerWriter: execSession.Stdin,
		PingInterval:    h.cfg.WSPingInterval,
		PongTimeout:     h.cfg.WSPongTimeout,
		WriteTimeout:    h.cfg.WSWriteTimeout,
		MaxMessageSize:  h.cfg.WSMaxMessageSize,
		Log:             h.log,
	})

	h.hub.Register(client)

	// Create a cancellable context that stops when the client disconnects.
	// ReadPump calls client.cancelFunc in its defer, which cancels this context.
	ctx, cancel := context.WithCancel(context.Background())
	client.SetCancelFunc(cancel)

	// Start the container stdout reader that feeds data to the client's send channel.
	go h.streamContainerOutput(ctx, client, execSession)

	// Start the read and write pumps.
	go client.WritePump()
	go client.ReadPump()
}

// streamContainerOutput reads from the container's stdout and sends it to the WebSocket client.
// It stops when the context is cancelled (client disconnect).
func (h *Handler) streamContainerOutput(ctx context.Context, client *Client, exec *container.ExecSession) {
	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			h.log.Info("container output stream cancelled for session=%s", client.SessionID)
			return
		default:
		}

		n, err := exec.Stdout.Read(buf)
		if err != nil {
			h.log.Info("container output stream ended for session=%s: %v", client.SessionID, err)
			return
		}
		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])
			client.Send(data)
		}
	}
}

type contextKey string

const contextKeyUserID contextKey = "user_id"
