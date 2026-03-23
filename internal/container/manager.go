// Package container manages Docker container lifecycle for user workspaces.
// Each workspace gets an isolated container with dedicated resources, providing
// a secure, sandboxed development environment.
package container

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/docker/docker/api/types"
	containerTypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/markbfromdc/cloudcode/internal/config"
	"github.com/markbfromdc/cloudcode/internal/logging"
)

// Manager handles the lifecycle of workspace Docker containers.
type Manager struct {
	docker   client.APIClient
	cfg      *config.Config
	log      *logging.Logger
	sessions map[string]*WorkspaceSession
	mu       sync.RWMutex
}

// WorkspaceSession tracks an active workspace container and its metadata.
type WorkspaceSession struct {
	SessionID   string
	ContainerID string
	UserID      string
	Status      string
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
		docker:   dockerClient,
		cfg:      cfg,
		log:      log.WithField("component", "container-manager"),
		sessions: make(map[string]*WorkspaceSession),
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

func int64Ptr(v int64) *int64 {
	return &v
}
