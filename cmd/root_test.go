package cmd

import (
	"testing"
)

func Test_RootCmd(t *testing.T) {
	// Test root help
	testCmd := RootCmd
	testCmd.SetArgs([]string{"--help"})
	testCmd.Execute()

	// test root halp
	testCmd.SetArgs([]string{"-h"})
	testCmd.Execute()

	// test root halp
	testCmd.SetArgs([]string{"--halp"})
	//testCmd.Execute()
	// check if error was returned
	if err := testCmd.Execute(); err == nil {
		t.Errorf("RootCmd() = %v, shouldNotPass %v", err, true)
	}
}
