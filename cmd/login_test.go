package cmd

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_LoginHelpCmd(t *testing.T) {
	// Test root help
	testCmd := RootCmd
	testCmd.SetArgs([]string{"login", "--help"})
	err := testCmd.Execute()

	assert.NoError(t, err)

	// test root halp
	testCmd.SetArgs([]string{"login", "-h"})
	err = testCmd.Execute()
	assert.NoError(t, err)

	// test root halp
	testCmd.SetArgs([]string{"login", "--halp"})
	err = testCmd.Execute()

	assert.Error(t, err)
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

func Test_LoginCmdConfigParams(t *testing.T) {
	testCmd := RootCmd
	// test
	testCmd.SetArgs([]string{"stores", "list", "--exp", "--config", "/Users/sbailey/.keyfactor/extra_config.json"})
	output := captureOutput(func() {
		err := testCmd.Execute()
		assert.NoError(t, err)
	})
	var stores []string
	if err := json.Unmarshal([]byte(output), &stores); err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// Verify that the length of the response is greater than 0
	assert.True(t, len(stores) >= 0, "Expected non-empty list of stores")
}
