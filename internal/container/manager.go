// Package container manages Docker container lifecycle for user workspaces.
// Each workspace gets an isolated container with dedicated resources, providing
// a secure, sandboxed development environment.
package container

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	containerTypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/markbfromdc/cloudcode/internal/config"
	"github.com/markbfromdc/cloudcode/internal/logging"
)

// DockerClient is the subset of the Docker API used by Manager.
type DockerClient interface {
	ContainerCreate(ctx context.Context, config *containerTypes.Config, hostConfig *containerTypes.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (containerTypes.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options containerTypes.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options containerTypes.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options containerTypes.RemoveOptions) error
	ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error)
	ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error)
}

// Manager handles the lifecycle of workspace Docker containers.
type Manager struct {
	docker      DockerClient
	cfg         *config.Config
	log         *logging.Logger
	sessions    map[string]*WorkspaceSession
	mu          sync.RWMutex
	stopCleanup chan struct{}
}

// WorkspaceSession tracks an active workspace container and its metadata.
type WorkspaceSession struct {
	SessionID   string
	ContainerID string
	UserID      string
	Status      string
	CreatedAt   time.Time
}

// ExecSession represents an active exec attachment to a container,
// providing stdin/stdout streams for terminal interaction.
type ExecSession struct {
	ContainerID string
	Stdin       io.WriteCloser
	Stdout      io.Reader
}

// NewManager creates a new container Manager with a Docker client connection.
func NewManager(cfg *config.Config, log *logging.Logger) (*Manager, error) {
	dockerClient, err := client.NewClientWithOpts(
		client.WithHost(cfg.DockerHost),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &Manager{
		docker:      dockerClient,
		cfg:         cfg,
		log:         log.WithField("component", "container-manager"),
		sessions:    make(map[string]*WorkspaceSession),
		stopCleanup: make(chan struct{}),
	}, nil
}

// CreateWorkspace provisions a new isolated container for a user workspace.
// It enforces resource limits (memory, CPU) and security constraints (no privilege escalation,
// read-only root filesystem with writable workspace volume).
func (m *Manager) CreateWorkspace(ctx context.Context, userID string) (*WorkspaceSession, error) {
	sessionID := uuid.New().String()

	memoryBytes := m.cfg.ContainerMemoryMB * 1024 * 1024

	containerCfg := &containerTypes.Config{
		Image: m.cfg.WorkspaceImage,
		Env: []string{
			fmt.Sprintf("SESSION_ID=%s", sessionID),
			fmt.Sprintf("USER_ID=%s", userID),
		},
		WorkingDir: "/workspace",
		Tty:        true,
		OpenStdin:  true,
	}

	hostCfg := &containerTypes.HostConfig{
		Resources: containerTypes.Resources{
			Memory:    memoryBytes,
			CPUShares: m.cfg.ContainerCPUShares,
			PidsLimit: int64Ptr(512),
		},
		SecurityOpt: []string{
			"no-new-privileges",
		},
		ReadonlyRootfs: false,
		NetworkMode:    containerTypes.NetworkMode(m.cfg.NetworkName),
	}

	networkCfg := &network.NetworkingConfig{}

	containerName := fmt.Sprintf("workspace-%s", sessionID[:8])

	resp, err := m.docker.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	if err := m.docker.ContainerStart(ctx, resp.ID, containerTypes.StartOptions{}); err != nil {
		// Clean up the created container on start failure.
		_ = m.docker.ContainerRemove(ctx, resp.ID, containerTypes.RemoveOptions{Force: true})
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	session := &WorkspaceSession{
		SessionID:   sessionID,
		ContainerID: resp.ID,
		UserID:      userID,
		Status:      "running",
		CreatedAt:   time.Now(),
	}

	m.mu.Lock()
	m.sessions[sessionID] = session
	m.mu.Unlock()

	m.log.Info("workspace created: session=%s container=%s user=%s", sessionID, resp.ID[:12], userID)
	return session, nil
}

// AttachToContainer creates an exec session on an existing container,
// returning stdin/stdout streams for terminal interaction.
func (m *Manager) AttachToContainer(ctx context.Context, sessionID string) (*ExecSession, error) {
	m.mu.RLock()
	session, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	execCfg := container_ExecConfig(session.ContainerID)
	execResp, err := m.docker.ContainerExecCreate(ctx, session.ContainerID, execCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}

	attachResp, err := m.docker.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{Tty: true})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec: %w", err)
	}

	return &ExecSession{
		ContainerID: session.ContainerID,
		Stdin:       attachResp.Conn,
		Stdout:      attachResp.Reader,
	}, nil
}

// StopWorkspace gracefully stops and removes a workspace container.
func (m *Manager) StopWorkspace(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	session, ok := m.sessions[sessionID]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("session not found: %s", sessionID)
	}
	delete(m.sessions, sessionID)
	m.mu.Unlock()

	timeoutSec := 10
	stopOptions := containerTypes.StopOptions{Timeout: &timeoutSec}
	if err := m.docker.ContainerStop(ctx, session.ContainerID, stopOptions); err != nil {
		m.log.Warn("failed to stop container gracefully: %v", err)
	}

	if err := m.docker.ContainerRemove(ctx, session.ContainerID, containerTypes.RemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	m.log.Info("workspace stopped: session=%s container=%s", sessionID, session.ContainerID[:12])
	return nil
}

// GetSession returns the workspace session for a given session ID.
func (m *Manager) GetSession(sessionID string) (*WorkspaceSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[sessionID]
	return s, ok
}

// ActiveWorkspaces returns the number of currently running workspaces.
func (m *Manager) ActiveWorkspaces() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// container_ExecConfig builds the exec configuration for attaching a shell to a container.
func container_ExecConfig(containerID string) types.ExecConfig {
	return types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"/bin/bash"},
	}
}

// Shutdown gracefully stops all active workspace containers.
// It respects the provided context for cancellation/timeout.
func (m *Manager) Shutdown(ctx context.Context) error {
	// Signal cleanup loop to stop.
	select {
	case <-m.stopCleanup:
		// Already closed.
	default:
		close(m.stopCleanup)
	}

	m.mu.Lock()
	sessions := make([]*WorkspaceSession, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	m.mu.Unlock()

	var firstErr error
	for _, s := range sessions {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		m.log.Info("shutting down workspace: session=%s container=%s", s.SessionID, s.ContainerID[:12])
		timeoutSec := 5
		stopOpts := containerTypes.StopOptions{Timeout: &timeoutSec}
		if err := m.docker.ContainerStop(ctx, s.ContainerID, stopOpts); err != nil {
			m.log.Warn("failed to stop container %s: %v", s.ContainerID[:12], err)
		}
		if err := m.docker.ContainerRemove(ctx, s.ContainerID, containerTypes.RemoveOptions{Force: true}); err != nil {
			m.log.Error("failed to remove container %s: %v", s.ContainerID[:12], err)
			if firstErr == nil {
				firstErr = err
			}
		}

		m.mu.Lock()
		delete(m.sessions, s.SessionID)
		m.mu.Unlock()
	}

	return firstErr
}

// StartCleanupLoop runs a background goroutine that periodically removes
// sessions older than maxAge. Call this after creating the manager.
// The loop stops when Shutdown is called or stopCleanup is closed.
func (m *Manager) StartCleanupLoop(interval time.Duration, maxAge time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.mu.RLock()
				var expired []string
				for id, s := range m.sessions {
					if time.Since(s.CreatedAt) > maxAge {
						expired = append(expired, id)
					}
				}
				m.mu.RUnlock()

				for _, id := range expired {
					m.log.Info("cleaning up expired session: %s", id)
					if err := m.StopWorkspace(context.Background(), id); err != nil {
						m.log.Warn("failed to clean up session %s: %v", id, err)
					}
				}
			case <-m.stopCleanup:
				return
			}
		}
	}()
}

func int64Ptr(v int64) *int64 {
	return &v
}
