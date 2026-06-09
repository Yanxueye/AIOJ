package domain

import "testing"

// TestIsTerminalStatus verifies that all terminal statuses are recognized.
func TestIsTerminalStatus(t *testing.T) {
	t.Logf(">>> Domain: IsTerminalStatus")
	terminal := []SubmissionStatus{
		StatusAccepted, StatusWrongAnswer, StatusCompileError,
		StatusRuntimeError, StatusTimeLimitExceeded,
		StatusMemoryLimitExceeded, StatusOutputLimitExceeded,
		StatusSystemError,
	}
	for _, s := range terminal {
		if !IsTerminalStatus(s) {
			t.Fatalf("expected %s to be terminal", s)
		}
	}
	nonTerminal := []SubmissionStatus{StatusPending, StatusQueueing, StatusCompiling, StatusRunning}
	for _, s := range nonTerminal {
		if IsTerminalStatus(s) {
			t.Fatalf("expected %s to be non-terminal", s)
		}
	}
	t.Logf("    terminal=%d non-terminal=%d all ok", len(terminal), len(nonTerminal))
}
