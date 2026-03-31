// Package main is the entry point for the Cloud IDE backend server.
// It initializes configuration, container management, WebSocket hub,
// file tree API, and starts the HTTP server.
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/markbfromdc/cloudcode/internal/api"
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

	// Start session cleanup loop.
	containerMgr.StartCleanupLoop(5*time.Minute, time.Duration(cfg.ContainerTimeoutMin)*time.Minute)

	// Initialize WebSocket hub and start its event loop.
	hub := ws.NewHub(log)
	go hub.Run()

	// Initialize API handlers.
	fileHandler := api.NewFileTreeHandler(log)
	gitHandler := api.NewGitHandler(log)
	fileOpsHandler := api.NewFileOpsHandler(log)

	// Set up authenticated HTTP routes.
	appMux := http.NewServeMux()

	// Workspace management API.
	appMux.HandleFunc("/api/v1/workspaces", func(w http.ResponseWriter, r *http.Request) {
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

	appMux.HandleFunc("/api/v1/workspaces/stop", func(w http.ResponseWriter, r *http.Request) {
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

	// File tree API - route by session ID pattern.
	appMux.HandleFunc("/api/v1/workspaces/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/files") {
			fileHandler.HandleListFiles(w, r)
		} else if strings.HasSuffix(path, "/files/content") {
			if r.Method == http.MethodGet {
				fileHandler.HandleReadFile(w, r)
			} else if r.Method == http.MethodPut {
				fileHandler.HandleWriteFile(w, r)
			} else {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		}
	})

	// Git operations API.
	appMux.HandleFunc("/api/v1/git/status", gitHandler.HandleGitStatus)
	appMux.HandleFunc("/api/v1/git/log", gitHandler.HandleGitLog)
	appMux.HandleFunc("/api/v1/git/branches", gitHandler.HandleGitBranches)
	appMux.HandleFunc("/api/v1/git/commit", gitHandler.HandleGitCommit)
	appMux.HandleFunc("/api/v1/git/stage", gitHandler.HandleGitStage)
	appMux.HandleFunc("/api/v1/git/init", gitHandler.HandleGitInit)

	// File create/delete/rename operations.
	appMux.HandleFunc("/api/v1/files/create", fileOpsHandler.HandleCreateFile)
	appMux.HandleFunc("/api/v1/files/delete", fileOpsHandler.HandleDeleteFile)
	appMux.HandleFunc("/api/v1/files/rename", fileOpsHandler.HandleRenameFile)

	// WebSocket terminal endpoint.
	wsHandler := ws.NewHandler(hub, cfg, containerMgr, log)
	appMux.Handle("/ws/terminal", wsHandler)

	// Apply auth middleware to the app routes.
	authMiddleware := middleware.Auth(cfg.JWTSecret, log)
	authedHandler := authMiddleware(appMux)

	// Create root mux with /health outside auth.
	rootMux := http.NewServeMux()

	// Health check endpoint (unauthenticated).
	rootMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "healthy",
			"active_sessions":   hub.ActiveSessions(),
			"active_workspaces": containerMgr.ActiveWorkspaces(),
			"timestamp":         time.Now().UTC().Format(time.RFC3339),
		})
	})

	// All other routes go through auth.
	rootMux.Handle("/", authedHandler)

	// Apply outer middleware stack: logging -> rate limiter -> request ID -> cors -> handler
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	corsMiddleware := middleware.CORS(cfg.AllowedOrigins)
	logMiddleware := middleware.RequestLogger(log)

	// Chain: logging -> rate limiter -> request ID -> cors -> handler
	handler := logMiddleware(rateLimiter.Middleware(middleware.RequestID(corsMiddleware(rootMux))))

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

	shutdownTimeout := time.Duration(cfg.ShutdownTimeoutSec) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server shutdown error: %v", err)
	}

	hub.Stop()
	log.Info("websocket hub stopped")

	if err := containerMgr.Shutdown(ctx); err != nil {
		log.Error("container manager shutdown error: %v", err)
	}
	log.Info("container manager stopped")

	rateLimiter.Stop()

	log.Info("server stopped")
}
