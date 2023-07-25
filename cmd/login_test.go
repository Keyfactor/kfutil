package cmd

import (
	"testing"
)

func Test_LoginHelpCmd(t *testing.T) {
	// Test root help
	testCmd := RootCmd
	testCmd.SetArgs([]string{"login", "--help"})
	testCmd.Execute()

	// test root halp
	testCmd.SetArgs([]string{"login", "-h"})
	testCmd.Execute()

	// test root halp
	testCmd.SetArgs([]string{"login", "--halp"})
	//testCmd.Execute()
	// check if error was returned
	if err := testCmd.Execute(); err == nil {
		t.Errorf("RootCmd() = %v, shouldNotPass %v", err, true)
	}
}

func Test_LoginCmdNoPrompt(t *testing.T) {
	// Test logging in w/ no args and no config w/ prompt
	testCmd := RootCmd
	// Test logging in w/o args and w/o prompt
	testCmd.SetArgs([]string{"login", "--no-prompt"})
	noPromptErr := testCmd.Execute()
	if noPromptErr != nil {
		t.Errorf("RootCmd() = %v, shouldNotPass %v", noPromptErr, true)
	}

}
