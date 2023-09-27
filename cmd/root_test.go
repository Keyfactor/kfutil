package cmd

import (
	"bytes"
	"github.com/spf13/cobra"
	"strings"
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

func testExecuteCommand(t *testing.T, cmd *cobra.Command, args ...string) (output []byte, err error) {
	t.Logf("Run \"%s %s\"", cmd.Use, strings.Join(args, " "))

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return buf.Bytes(), err
}
