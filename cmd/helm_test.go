package cmd

import (
	"strings"
	"testing"
)

func TestHelm(t *testing.T) {
	t.Run("Test helm command", func(t *testing.T) {
		// The helm command doesn't have any flags or a RunE function, so the output should be the same as the help menu.
		cmd := NewCmdHelm()

		t.Logf("Testing %q", cmd.Use)

		helmNoFlag, err := testExecuteCommand(t, cmd, "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		helmHelp, err := testExecuteCommand(t, cmd, "-h")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		diff := strings.Compare(string(helmNoFlag), string(helmHelp))
		if diff != 0 {
			t.Errorf("Expected helmNoFlag to equal helmHelp, but got: %v", diff)
		}
	})
}
