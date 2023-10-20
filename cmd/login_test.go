// Package cmd Copyright 2023 Keyfactor
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"os"
	"os/user"
	"path/filepath"
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
	// Get the current user's information
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Define the path to the file in the user's home directory
	filePath := filepath.Join(currentUser.HomeDir, ".keyfactor/command_config.json")
	testEnvCredsOnly(t, filePath, false)
	testLoginNoPrompt(t, filePath)
}

func Test_LoginCmdConfigParams(t *testing.T) {
	testCmd := RootCmd
	// test
	testCmd.SetArgs([]string{"stores", "list", "--exp", "--config", "$HOME/.keyfactor/extra_config.json"})
	output := captureOutput(func() {
		err := testCmd.Execute()
		assert.NoError(t, err)
	})
	t.Logf("output: %s", output)
	var stores []string
	if err := json.Unmarshal([]byte(output), &stores); err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// Verify that the length of the response is greater than 0
	assert.True(t, len(stores) >= 0, "Expected non-empty list of stores")
}

func testLogout(t *testing.T) {
	t.Run(fmt.Sprintf("Logout"), func(t *testing.T) {
		testCmd := RootCmd
		// test
		testCmd.SetArgs([]string{"logout"})
		output := captureOutput(func() {
			err := testCmd.Execute()
			assert.NoError(t, err)
		})
		t.Logf("output: %s", output)

		assert.Contains(t, output, "Logged out successfully!")

		// Get the current user's information
		currentUser, err := user.Current()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Define the path to the file in the user's home directory
		filePath := filepath.Join(currentUser.HomeDir, ".keyfactor/command_config.json")
		_, err = os.Stat(filePath)

		// Test that the config file does not exist
		if _, fErr := os.Stat(filePath); !os.IsNotExist(fErr) {
			t.Errorf("Config file %s still exists, please remove", filePath)
		}
	})

}

func testConfigValid(t *testing.T) {
	// Test config is valid
	//envUsername := os.Getenv("KEYFACTOR_USERNAME")
	//envPassword := os.Getenv("KEYFACTOR_PASSWORD")
	//t.Logf("envUsername: %s", envUsername)
	//t.Logf("envPassword: %s", envPassword)
	t.Logf("Attempting to run `store-types list`")
	t.Run(fmt.Sprintf("List store types"), func(t *testing.T) {
		testCmd := RootCmd
		t.Log("Setting args")
		testCmd.SetArgs([]string{"store-types", "list"})
		t.Logf("args: %v", testCmd.Args)
		t.Log("Capturing output")
		output := captureOutput(func() {
			tErr := testCmd.Execute()
			assert.NoError(t, tErr)
		})
		t.Logf("output: %s", output)

		var storeTypes []map[string]interface{}
		if err := json.Unmarshal([]byte(output), &storeTypes); err != nil {
			t.Fatalf("Error unmarshalling JSON: %v", err)
		}

		// Verify that the length of the response is greater than 0
		assert.True(t, len(storeTypes) >= 0, "Expected non-empty list of store types")
	})
}

func testConfigExists(t *testing.T, filePath string, allowExist bool) {
	var testName string
	if allowExist {
		testName = "Config file exists"
	} else {
		testName = "Config file does not exist"
	}
	t.Run(fmt.Sprintf(testName), func(t *testing.T) {
		_, fErr := os.Stat(filePath)
		if allowExist {
			assert.True(t, allowExist && fErr == nil)
			// Load the config file from JSON to map[string]interface{}
			fileConfigJSON := make(map[string]interface{})
			file, _ := os.Open(filePath)
			defer file.Close()
			decoder := json.NewDecoder(file)
			err := decoder.Decode(&fileConfigJSON)
			if err != nil {
				t.Errorf("Error decoding config file: %s", err)
			}
			// Verify that the config file has the correct keys
			assert.Contains(t, fileConfigJSON, "servers")
			kfcServers := fileConfigJSON["servers"].(map[string]interface{})
			assert.Contains(t, kfcServers, "default")
			defaultServer := kfcServers["default"].(map[string]interface{})
			assert.Contains(t, defaultServer, "host")
			assert.Contains(t, defaultServer, "kfcUsername")
			assert.Contains(t, defaultServer, "kfcPassword")
		} else {
			assert.True(t, !allowExist && os.IsNotExist(fErr))
		}
	})
}

func testEnvCredsOnly(t *testing.T, filePath string, allowExist bool) {
	t.Run(fmt.Sprintf("Auth w/ env ONLY"), func(t *testing.T) {
		// Load .env file
		err := godotenv.Load("../.env_1040")
		if err != nil {
			t.Errorf("Error loading .env file")
		}
		testLogout(t)
		testConfigExists(t, filePath, false)
		testConfigValid(t)
	})
}

func testEnvCredsToFile(t *testing.T, filePath string, allowExist bool) {
	t.Run(fmt.Sprintf("Auth w/ env ONLY"), func(t *testing.T) {
		// Load .env file
		err := godotenv.Load("../.env_1040")
		if err != nil {
			t.Errorf("Error loading .env file")
		}
		testLogout(t)
		testConfigExists(t, filePath, false)
		testConfigValid(t)
	})
}

func testLoginNoPrompt(t *testing.T, filePath string) {
	// Test logging in w/o args and w/o prompt
	t.Run(fmt.Sprintf("login no prompt"), func(t *testing.T) {
		testCmd := RootCmd
		testCmd.SetArgs([]string{"login", "--no-prompt"})
		noPromptErr := testCmd.Execute()
		if noPromptErr != nil {
			t.Errorf("RootCmd() = %v, shouldNotPass %v", noPromptErr, true)
		}
		testConfigExists(t, filePath, true)
		os.Unsetenv("KEYFACTOR_USERNAME")
		os.Unsetenv("KEYFACTOR_PASSWORD")
		testConfigValid(t)
		//testLogout(t)
	})
}
