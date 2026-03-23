// Package main is the entry point for the Cloud IDE backend server.
// It initializes configuration, container management, WebSocket hub,
// and starts both HTTP and gRPC servers.
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/markbfromdc/cloudcode/internal/config"
	"github.com/markbfromdc/cloudcode/internal/container"
	"github.com/markbfromdc/cloudcode/internal/logging"
	"github.com/markbfromdc/cloudcode/internal/middleware"
	ws "github.com/markbfromdc/cloudcode/internal/websocket"
)

func main() {
	log := logging.Default()
	log.Info("starting Cloud IDE backend server")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config: %v", err)
	}

	// Initialize container manager.
	containerMgr, err := container.NewManager(cfg, log)
	if err != nil {
		log.Fatal("failed to initialize container manager: %v", err)
	}

	// Initialize WebSocket hub and start its event loop.
	hub := ws.NewHub(log)
	go hub.Run()

	// Set up HTTP routes.
	mux := http.NewServeMux()

	// Health check endpoint.
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "healthy",
			"active_sessions":   hub.ActiveSessions(),
			"active_workspaces": containerMgr.ActiveWorkspaces(),
			"timestamp":         time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Workspace management API.
	mux.HandleFunc("/api/v1/workspaces", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(middleware.ContextKeyUserID).(string)
		if !ok {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		session, err := containerMgr.CreateWorkspace(r.Context(), userID)
		if err != nil {
			log.Error("failed to create workspace: %v", err)
			http.Error(w, `{"error":"failed to create workspace"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"session_id":   session.SessionID,
			"container_id": session.ContainerID,
			"status":       session.Status,
		})
	})

	mux.HandleFunc("/api/v1/workspaces/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		sessionID := r.URL.Query().Get("session_id")
		if sessionID == "" {
			http.Error(w, `{"error":"session_id required"}`, http.StatusBadRequest)
			return
		}

		if err := containerMgr.StopWorkspace(r.Context(), sessionID); err != nil {
			log.Error("failed to stop workspace: %v", err)
			http.Error(w, `{"error":"failed to stop workspace"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
	})

	// WebSocket terminal endpoint.
	wsHandler := ws.NewHandler(hub, cfg, containerMgr, log)
	mux.Handle("/ws/terminal", wsHandler)

	// Apply middleware stack.
	authMiddleware := middleware.Auth(cfg.JWTSecret, log)
	logMiddleware := middleware.RequestLogger(log)

	// Chain: logging -> auth -> handler
	handler := logMiddleware(authMiddleware(mux))

	// Create HTTP server with timeouts.
	server := &http.Server{
		Addr:         cfg.HTTPAddr(),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("HTTP server listening on %s", cfg.HTTPAddr())
		if cfg.EnableTLS {
			if err := server.ListenAndServeTLS(cfg.TLSCert, cfg.TLSKey); err != nil && err != http.ErrServerClosed {
				log.Fatal("HTTP server error: %v", err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal("HTTP server error: %v", err)
			}
		}
	}()

	<-stop
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server shutdown error: %v", err)
	}

	log.Info("server stopped")
}
