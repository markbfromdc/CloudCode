package container

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/docker/docker/api/types"
	containerTypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/markbfromdc/cloudcode/internal/config"
	"github.com/markbfromdc/cloudcode/internal/logging"
)

// mockDockerClient implements the DockerClient interface for testing.
type mockDockerClient struct {
	createFunc   func(ctx context.Context, cfg *containerTypes.Config, hostCfg *containerTypes.HostConfig, netCfg *network.NetworkingConfig, platform *ocispec.Platform, name string) (containerTypes.CreateResponse, error)
	startFunc    func(ctx context.Context, containerID string, options containerTypes.StartOptions) error
	stopFunc     func(ctx context.Context, containerID string, options containerTypes.StopOptions) error
	removeFunc   func(ctx context.Context, containerID string, options containerTypes.RemoveOptions) error
	execCreateFn func(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error)
	execAttachFn func(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error)
}

func (m *mockDockerClient) ContainerCreate(ctx context.Context, cfg *containerTypes.Config, hostCfg *containerTypes.HostConfig, netCfg *network.NetworkingConfig, platform *ocispec.Platform, name string) (containerTypes.CreateResponse, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, cfg, hostCfg, netCfg, platform, name)
	}
	return containerTypes.CreateResponse{ID: "mock-container-id-0123456789"}, nil
}

func (m *mockDockerClient) ContainerStart(ctx context.Context, containerID string, options containerTypes.StartOptions) error {
	if m.startFunc != nil {
		return m.startFunc(ctx, containerID, options)
	}
	return nil
}

func (m *mockDockerClient) ContainerStop(ctx context.Context, containerID string, options containerTypes.StopOptions) error {
	if m.stopFunc != nil {
		return m.stopFunc(ctx, containerID, options)
	}
	return nil
}

func (m *mockDockerClient) ContainerRemove(ctx context.Context, containerID string, options containerTypes.RemoveOptions) error {
	if m.removeFunc != nil {
		return m.removeFunc(ctx, containerID, options)
	}
	return nil
}

func (m *mockDockerClient) ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error) {
	if m.execCreateFn != nil {
		return m.execCreateFn(ctx, container, config)
	}
	return types.IDResponse{ID: "mock-exec-id"}, nil
}

func (m *mockDockerClient) ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error) {
	if m.execAttachFn != nil {
		return m.execAttachFn(ctx, execID, config)
	}
	return types.HijackedResponse{}, nil
}

func newTestManager(mock DockerClient) *Manager {
	log := logging.New(nil, logging.INFO)
	cfg := &config.Config{
		WorkspaceImage:     "test-image:latest",
		ContainerMemoryMB:  512,
		ContainerCPUShares: 1024,
		NetworkName:        "test-net",
	}
	return &Manager{
		docker:   mock,
		cfg:      cfg,
		log:      log.WithField("component", "test"),
		sessions: make(map[string]*WorkspaceSession),
	}
}

func TestGetSession(t *testing.T) {
	mgr := newTestManager(&mockDockerClient{})

	_, ok := mgr.GetSession("nonexistent")
	if ok {
		t.Error("expected session not found")
	}

	mgr.sessions["abc"] = &WorkspaceSession{SessionID: "abc", ContainerID: "cid", UserID: "u1", Status: "running"}

	s, ok := mgr.GetSession("abc")
	if !ok {
		t.Fatal("expected session found")
	}
	if s.UserID != "u1" {
		t.Errorf("expected user u1, got %s", s.UserID)
	}
}

func TestActiveWorkspaces(t *testing.T) {
	mgr := newTestManager(&mockDockerClient{})

	if mgr.ActiveWorkspaces() != 0 {
		t.Errorf("expected 0, got %d", mgr.ActiveWorkspaces())
	}

	mgr.sessions["a"] = &WorkspaceSession{}
	mgr.sessions["b"] = &WorkspaceSession{}

	if mgr.ActiveWorkspaces() != 2 {
		t.Errorf("expected 2, got %d", mgr.ActiveWorkspaces())
	}
}

func TestCreateWorkspaceSuccess(t *testing.T) {
	mock := &mockDockerClient{
		createFunc: func(_ context.Context, cfg *containerTypes.Config, hostCfg *containerTypes.HostConfig, _ *network.NetworkingConfig, _ *ocispec.Platform, name string) (containerTypes.CreateResponse, error) {
			if hostCfg.Resources.Memory != 512*1024*1024 {
				t.Errorf("expected 512MB memory, got %d", hostCfg.Resources.Memory)
			}
			if hostCfg.Resources.CPUShares != 1024 {
				t.Errorf("expected 1024 CPU shares, got %d", hostCfg.Resources.CPUShares)
			}
			if cfg.Image != "test-image:latest" {
				t.Errorf("expected test-image:latest, got %s", cfg.Image)
			}
			if !strings.HasPrefix(name, "workspace-") {
				t.Errorf("expected name to start with workspace-, got %s", name)
			}
			return containerTypes.CreateResponse{ID: "created-1234567890ab"}, nil
		},
	}

	mgr := newTestManager(mock)
	session, err := mgr.CreateWorkspace(context.Background(), "user-42")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.UserID != "user-42" {
		t.Errorf("expected user-42, got %s", session.UserID)
	}
	if session.ContainerID != "created-1234567890ab" {
		t.Errorf("expected created-1234567890ab, got %s", session.ContainerID)
	}
	if session.Status != "running" {
		t.Errorf("expected running status, got %s", session.Status)
	}
	if session.SessionID == "" {
		t.Error("expected session ID to be set")
	}
	if mgr.ActiveWorkspaces() != 1 {
		t.Errorf("expected 1 active workspace, got %d", mgr.ActiveWorkspaces())
	}
}

func TestCreateWorkspaceEnvVars(t *testing.T) {
	mock := &mockDockerClient{
		createFunc: func(_ context.Context, cfg *containerTypes.Config, _ *containerTypes.HostConfig, _ *network.NetworkingConfig, _ *ocispec.Platform, _ string) (containerTypes.CreateResponse, error) {
			var hasSessionEnv, hasUserEnv bool
			for _, e := range cfg.Env {
				if strings.HasPrefix(e, "SESSION_ID=") {
					hasSessionEnv = true
				}
				if e == "USER_ID=user-99" {
					hasUserEnv = true
				}
			}
			if !hasSessionEnv {
				t.Error("expected SESSION_ID env var")
			}
			if !hasUserEnv {
				t.Error("expected USER_ID=user-99 env var")
			}
			return containerTypes.CreateResponse{ID: "c1-envtest-longid"}, nil
		},
	}

	mgr := newTestManager(mock)
	_, err := mgr.CreateWorkspace(context.Background(), "user-99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateWorkspaceCreateFails(t *testing.T) {
	mock := &mockDockerClient{
		createFunc: func(_ context.Context, _ *containerTypes.Config, _ *containerTypes.HostConfig, _ *network.NetworkingConfig, _ *ocispec.Platform, _ string) (containerTypes.CreateResponse, error) {
			return containerTypes.CreateResponse{}, fmt.Errorf("docker create failed")
		},
	}

	mgr := newTestManager(mock)
	_, err := mgr.CreateWorkspace(context.Background(), "user-1")

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to create container") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateWorkspaceStartFailsCleansUp(t *testing.T) {
	var removeCalled bool
	mock := &mockDockerClient{
		startFunc: func(_ context.Context, _ string, _ containerTypes.StartOptions) error {
			return fmt.Errorf("docker start failed")
		},
		removeFunc: func(_ context.Context, _ string, _ containerTypes.RemoveOptions) error {
			removeCalled = true
			return nil
		},
	}

	mgr := newTestManager(mock)
	_, err := mgr.CreateWorkspace(context.Background(), "user-1")

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to start container") {
		t.Errorf("unexpected error: %v", err)
	}
	if !removeCalled {
		t.Error("expected container to be cleaned up after start failure")
	}
}

func TestStopWorkspaceSuccess(t *testing.T) {
	var stopCalled, removeCalled bool
	mock := &mockDockerClient{
		stopFunc: func(_ context.Context, id string, _ containerTypes.StopOptions) error {
			stopCalled = true
			if id != "container-xyz" {
				t.Errorf("expected container-xyz, got %s", id)
			}
			return nil
		},
		removeFunc: func(_ context.Context, _ string, _ containerTypes.RemoveOptions) error {
			removeCalled = true
			return nil
		},
	}

	mgr := newTestManager(mock)
	mgr.sessions["session-1"] = &WorkspaceSession{
		SessionID:   "session-1",
		ContainerID: "container-xyz",
		UserID:      "u1",
		Status:      "running",
	}

	err := mgr.StopWorkspace(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stopCalled {
		t.Error("expected stop to be called")
	}
	if !removeCalled {
		t.Error("expected remove to be called")
	}
	if mgr.ActiveWorkspaces() != 0 {
		t.Error("expected session to be removed")
	}
}

func TestStopWorkspaceNotFound(t *testing.T) {
	mgr := newTestManager(&mockDockerClient{})
	err := mgr.StopWorkspace(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "session not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStopWorkspaceRemoveFails(t *testing.T) {
	mock := &mockDockerClient{
		removeFunc: func(_ context.Context, _ string, _ containerTypes.RemoveOptions) error {
			return fmt.Errorf("remove failed")
		},
	}

	mgr := newTestManager(mock)
	mgr.sessions["s1"] = &WorkspaceSession{SessionID: "s1", ContainerID: "container-abc123", UserID: "u1"}

	err := mgr.StopWorkspace(context.Background(), "s1")
	if err == nil {
		t.Fatal("expected error on remove failure")
	}
	if !strings.Contains(err.Error(), "failed to remove container") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStopWorkspaceGracefulStopFails(t *testing.T) {
	// Even if graceful stop fails, remove should still be attempted.
	var removeCalled bool
	mock := &mockDockerClient{
		stopFunc: func(_ context.Context, _ string, _ containerTypes.StopOptions) error {
			return fmt.Errorf("stop timeout")
		},
		removeFunc: func(_ context.Context, _ string, _ containerTypes.RemoveOptions) error {
			removeCalled = true
			return nil
		},
	}

	mgr := newTestManager(mock)
	mgr.sessions["s1"] = &WorkspaceSession{SessionID: "s1", ContainerID: "container-abc123"}

	err := mgr.StopWorkspace(context.Background(), "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !removeCalled {
		t.Error("expected remove to be called even when stop fails")
	}
}

func TestAttachToContainerSuccess(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mock := &mockDockerClient{
		execCreateFn: func(_ context.Context, _ string, cfg types.ExecConfig) (types.IDResponse, error) {
			if !cfg.AttachStdin || !cfg.AttachStdout {
				t.Error("expected stdin/stdout to be attached")
			}
			return types.IDResponse{ID: "exec-1"}, nil
		},
		execAttachFn: func(_ context.Context, _ string, _ types.ExecStartCheck) (types.HijackedResponse, error) {
			return types.HijackedResponse{
				Conn:   client,
				Reader: bufio.NewReader(client),
			}, nil
		},
	}

	mgr := newTestManager(mock)
	mgr.sessions["s1"] = &WorkspaceSession{SessionID: "s1", ContainerID: "cid-1"}

	exec, err := mgr.AttachToContainer(context.Background(), "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.ContainerID != "cid-1" {
		t.Errorf("expected cid-1, got %s", exec.ContainerID)
	}
	if exec.Stdin == nil {
		t.Error("expected Stdin to be set")
	}
	if exec.Stdout == nil {
		t.Error("expected Stdout to be set")
	}
}

func TestAttachToContainerNotFound(t *testing.T) {
	mgr := newTestManager(&mockDockerClient{})
	_, err := mgr.AttachToContainer(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "session not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAttachToContainerExecCreateFails(t *testing.T) {
	mock := &mockDockerClient{
		execCreateFn: func(_ context.Context, _ string, _ types.ExecConfig) (types.IDResponse, error) {
			return types.IDResponse{}, fmt.Errorf("exec create failed")
		},
	}

	mgr := newTestManager(mock)
	mgr.sessions["s1"] = &WorkspaceSession{SessionID: "s1", ContainerID: "cid-1"}

	_, err := mgr.AttachToContainer(context.Background(), "s1")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to create exec") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAttachToContainerExecAttachFails(t *testing.T) {
	mock := &mockDockerClient{
		execAttachFn: func(_ context.Context, _ string, _ types.ExecStartCheck) (types.HijackedResponse, error) {
			return types.HijackedResponse{}, fmt.Errorf("attach failed")
		},
	}

	mgr := newTestManager(mock)
	mgr.sessions["s1"] = &WorkspaceSession{SessionID: "s1", ContainerID: "cid-1"}

	_, err := mgr.AttachToContainer(context.Background(), "s1")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to attach to exec") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConcurrentSessionAccess(t *testing.T) {
	mgr := newTestManager(&mockDockerClient{})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sid := fmt.Sprintf("session-%d", i)
			mgr.mu.Lock()
			mgr.sessions[sid] = &WorkspaceSession{SessionID: sid, ContainerID: fmt.Sprintf("c%d", i)}
			mgr.mu.Unlock()

			_, _ = mgr.GetSession(sid)
			_ = mgr.ActiveWorkspaces()
		}(i)
	}
	wg.Wait()

	if mgr.ActiveWorkspaces() != 50 {
		t.Errorf("expected 50 active workspaces, got %d", mgr.ActiveWorkspaces())
	}
}

func TestContainerExecConfig(t *testing.T) {
	cfg := container_ExecConfig("test-container")

	if !cfg.AttachStdin {
		t.Error("expected AttachStdin true")
	}
	if !cfg.AttachStdout {
		t.Error("expected AttachStdout true")
	}
	if !cfg.AttachStderr {
		t.Error("expected AttachStderr true")
	}
	if !cfg.Tty {
		t.Error("expected Tty true")
	}
	if len(cfg.Cmd) != 1 || cfg.Cmd[0] != "/bin/bash" {
		t.Errorf("expected [/bin/bash], got %v", cfg.Cmd)
	}
}

func TestInt64Ptr(t *testing.T) {
	p := int64Ptr(42)
	if *p != 42 {
		t.Errorf("expected 42, got %d", *p)
	}
}

func TestWorkspaceSessionFields(t *testing.T) {
	session := &WorkspaceSession{
		SessionID:   "s1",
		ContainerID: "c1",
		UserID:      "u1",
		Status:      "running",
	}
	if session.SessionID != "s1" || session.ContainerID != "c1" || session.UserID != "u1" || session.Status != "running" {
		t.Errorf("unexpected session fields: %+v", session)
	}
}
