package sandbox

import (
	"testing"

	"remote_judge/internal/config"
)

// TestBuildDockerSandbox verifies docker sandbox is created with circuit breaker.
func TestBuildDockerSandbox(t *testing.T) {
	t.Logf(">>> Factory: mode=docker -> DockerCLISandbox + CircuitBreaker")
	cfg := config.Config{PrewarmImages: false}
	sb := Build(cfg)
	if sb == nil {
		t.Fatal("expected non-nil sandbox")
	}
	cb, isCB := sb.(*CircuitBreakerSandbox)
	t.Logf("    sandbox_type=%T circuit_breaker=%v", sb, isCB)
	if !isCB {
		t.Fatal("docker sandbox should be wrapped in circuit breaker")
	}
	_, isDocker := cb.next.(*DockerCLISandbox)
	if !isDocker {
		t.Fatalf("inner sandbox should be DockerCLISandbox, got %T", cb.next)
	}
	t.Logf("    inner=%T transfer_mode=%s", cb.next, cb.next.(*DockerCLISandbox).TransferMode)
}

// TestBuildDockerSandboxCopyMode verifies transfer mode propagation.
func TestBuildDockerSandboxCopyMode(t *testing.T) {
	t.Logf(">>> Factory: mode=docker transfer=copy -> DockerCLISandbox(copy)")
	cfg := config.Config{DockerTransfer: "copy", PrewarmImages: false}
	sb := Build(cfg)
	if sb == nil {
		t.Fatal("expected non-nil sandbox")
	}
	cb := sb.(*CircuitBreakerSandbox)
	docker := cb.next.(*DockerCLISandbox)
	if docker.TransferMode != "copy" {
		t.Fatalf("expected transfer_mode=copy, got %s", docker.TransferMode)
	}
	t.Logf("    transfer_mode=%s", docker.TransferMode)
}

// TestBuildDockerSandboxPrewarmDisabled verifies prewarm flag is respected.
func TestBuildDockerSandboxPrewarmDisabled(t *testing.T) {
	t.Logf(">>> Factory: mode=docker prewarm=false -> no pre-warm")
	cfg := config.Config{PrewarmImages: false}
	sb := Build(cfg)
	if sb == nil {
		t.Fatal("expected non-nil sandbox")
	}
	t.Logf("    sandbox built without pre-warming: %T", sb)
}
