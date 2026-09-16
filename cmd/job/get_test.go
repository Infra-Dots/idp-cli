package job

import "testing"

// A job parked in "Waiting Approval" never moves on its own — approval always
// comes from outside. Treating it as non-terminal made `--watch` poll until the
// CI runner gave up.
func TestWaitingApprovalIsTerminal(t *testing.T) {
	if !terminalStatuses[statusWaitingApproval] {
		t.Fatalf("%q must be terminal for --watch", statusWaitingApproval)
	}
	if failedStatuses[statusWaitingApproval] {
		t.Errorf("%q is not a failure — it needs a human, and gets its own exit code",
			statusWaitingApproval)
	}
}

// The literal spelling matters: unlike every other status this one has a space
// and capital letters.
func TestWaitingApprovalSpelling(t *testing.T) {
	if statusWaitingApproval != "Waiting Approval" {
		t.Errorf("statusWaitingApproval = %q, want %q", statusWaitingApproval, "Waiting Approval")
	}
}

func TestTerminalAndFailedStatuses(t *testing.T) {
	tests := []struct {
		status     string
		isTerminal bool
		isFailed   bool
	}{
		{"completed", true, false},
		{"applied", true, false},
		{"failed", true, true},
		{"rejected", true, true},
		{"cancelled", true, true},
		{"Waiting Approval", true, false},
		{"queueing", false, false},
		{"pending", false, false},
		{"running", false, false},
		{"runningApply", false, false},
		{"pendingApply", false, false},
		{"approved", false, false},
	}
	for _, tc := range tests {
		if got := terminalStatuses[tc.status]; got != tc.isTerminal {
			t.Errorf("terminal[%q] = %v, want %v", tc.status, got, tc.isTerminal)
		}
		if got := failedStatuses[tc.status]; got != tc.isFailed {
			t.Errorf("failed[%q] = %v, want %v", tc.status, got, tc.isFailed)
		}
	}
}

// The three outcomes must stay distinguishable by a CI script.
func TestWatchExitCodesAreDistinct(t *testing.T) {
	codes := map[int]string{
		exitJobFailed:        "failed",
		exitJobNeedsApproval: "needs approval",
		exitWatchTimeout:     "timeout",
	}
	if len(codes) != 3 {
		t.Error("watch exit codes collide; CI cannot tell the outcomes apart")
	}
	if exitJobFailed == 0 || exitJobNeedsApproval == 0 || exitWatchTimeout == 0 {
		t.Error("no watch failure mode may exit 0")
	}
}
