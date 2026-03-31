package websocket

import (
	"io"
	"time"

	"github.com/gorilla/websocket"
	"github.com/markbfromdc/cloudcode/internal/logging"
)

// Client represents a single WebSocket connection from a browser terminal (xterm.js)
// to a workspace container's PTY. It handles bi-directional data flow and heartbeats.
type Client struct {
	// SessionID uniquely identifies this workspace session.
	SessionID string

	// UserID is the authenticated user who owns this session.
	UserID string

	// ContainerID is the Docker container backing this workspace.
	ContainerID string

	conn *websocket.Conn
	hub  *Hub
	send chan []byte
	log  *logging.Logger

	// containerWriter writes data from the WebSocket to the container's stdin.
	containerWriter io.WriteCloser

	// pingInterval is the interval between heartbeat pings.
	pingInterval time.Duration

	// pongTimeout is the maximum time to wait for a pong response.
	pongTimeout time.Duration

	// writeTimeout is the maximum time to wait for a write to complete.
	writeTimeout time.Duration

	// maxMessageSize is the maximum allowed message size from the client.
	maxMessageSize int64

	// cancelFunc cancels the context for the container output stream goroutine.
	cancelFunc func()
}

// ClientConfig holds the parameters needed to create a new Client.
type ClientConfig struct {
	SessionID       string
	UserID          string
	ContainerID     string
	Conn            *websocket.Conn
	Hub             *Hub
	ContainerWriter io.WriteCloser
	PingInterval    time.Duration
	PongTimeout     time.Duration
	WriteTimeout    time.Duration
	MaxMessageSize  int64
	Log             *logging.Logger
}

// NewClient creates a new WebSocket client with the given configuration.
func NewClient(cfg ClientConfig) *Client {
	return &Client{
		SessionID:       cfg.SessionID,
		UserID:          cfg.UserID,
		ContainerID:     cfg.ContainerID,
		conn:            cfg.Conn,
		hub:             cfg.Hub,
		send:            make(chan []byte, 256),
		containerWriter: cfg.ContainerWriter,
		pingInterval:    cfg.PingInterval,
		pongTimeout:     cfg.PongTimeout,
		writeTimeout:    cfg.WriteTimeout,
		maxMessageSize:  cfg.MaxMessageSize,
		log:             cfg.Log.WithField("session", cfg.SessionID),
	}
}

// ReadPump reads messages from the WebSocket connection and forwards them
// to the container's stdin. It runs in its own goroutine per connection.
//
// The pump sets read deadlines based on pongTimeout to detect dead connections.
// When the browser sends keystrokes via xterm.js, they arrive here and are
// written into the container's PTY.
func (c *Client) ReadPump() {
	defer func() {
		if c.cancelFunc != nil {
			c.cancelFunc()
		}
		c.hub.Unregister(c)
		c.conn.Close()
		if c.containerWriter != nil {
			c.containerWriter.Close()
		}
	}()

	c.conn.SetReadLimit(c.maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(c.pongTimeout)); err != nil {
		c.log.Error("failed to set initial read deadline: %v", err)
		return
	}

	c.conn.SetPongHandler(func(string) error {
		c.log.Debug("pong received")
		return c.conn.SetReadDeadline(time.Now().Add(c.pongTimeout))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				c.log.Error("unexpected websocket close: %v", err)
			}
			return
		}

		if c.containerWriter != nil {
			if _, err := c.containerWriter.Write(message); err != nil {
				c.log.Error("failed to write to container stdin: %v", err)
				return
			}
		}
	}
}

// WritePump sends messages from the container's stdout back to the WebSocket
// connection (and thus to xterm.js in the browser). It also handles heartbeat pings.
//
// A ping is sent at each pingInterval. If the client does not respond with a
// pong within pongTimeout, the connection is considered dead and is closed.
func (c *Client) WritePump() {
	ticker := time.NewTicker(c.pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
				c.log.Error("failed to set write deadline: %v", err)
				return
			}

			if !ok {
				// Hub closed the channel.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				c.log.Error("failed to get writer: %v", err)
				return
			}

			if _, err := w.Write(message); err != nil {
				c.log.Error("failed to write message: %v", err)
				return
			}

			// Batch any queued messages into the same write frame for efficiency.
			n := len(c.send)
			for i := 0; i < n; i++ {
				if _, err := w.Write(<-c.send); err != nil {
					c.log.Error("failed to write batched message: %v", err)
					return
				}
			}

			if err := w.Close(); err != nil {
				c.log.Error("failed to close writer: %v", err)
				return
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
				c.log.Error("failed to set ping write deadline: %v", err)
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.log.Error("ping failed: %v", err)
				return
			}
			c.log.Debug("ping sent")
		}
	}
}

// SetCancelFunc sets the cancel function that is called when ReadPump exits,
// allowing dependent goroutines (e.g., streamContainerOutput) to be notified.
func (c *Client) SetCancelFunc(cancel func()) {
	c.cancelFunc = cancel
}

// Send queues a message to be sent to the WebSocket client.
func (c *Client) Send(data []byte) {
	select {
	case c.send <- data:
	default:
		c.log.Warn("send buffer full, dropping message for session=%s", c.SessionID)
	}
}
