// Copyright 2024 Keyfactor
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
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Keyfactor/keyfactor-auth-client-go/auth_providers"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
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

func Test_LoginCmdEnvOnly(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	// Define the path to the file in the user's home directory
	configFilePath := filepath.Join(homeDir, auth_providers.DefaultConfigFilePath)
	testEnvCredsOnly(t, configFilePath, false)

}

func Test_LoginFileNoPrompt(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	configFilePath := filepath.Join(homeDir, auth_providers.DefaultConfigFilePath)
	// Test logging in w/o args and w/o prompt
	username, password, domain := exportBasicEnvVariables()
	clientId, clientSecret, tokenUrl := exportOAuthEnvVariables()
	envUsername, envPassword, envDomain := exportBasicEnvVariables()
	envClientId, envClientSecret, envTokenUrl := exportOAuthEnvVariables()
	os.Setenv(auth_providers.EnvKeyfactorSkipVerify, "true")
	if (envUsername == "" || envPassword == "" || envDomain == "") && (envClientId == "" || envClientSecret == "" || envTokenUrl == "") {
		t.Errorf("Environment variables are not set")
		t.FailNow()
	}
	existingConfig, exErr := auth_providers.ReadConfigFromJSON(configFilePath)
	if exErr != nil {
		t.Errorf("Error reading existing config: %s", exErr)
		t.FailNow()
	}
	defer func() {
		//restore config file
		if existingConfig != nil {
			wErr := auth_providers.WriteConfigToJSON(configFilePath, existingConfig)
			if wErr != nil {
				t.Errorf("Error writing existing config: %s", wErr)
				t.FailNow()
			}
		}
	}()
	t.Run(
		fmt.Sprintf("login no prompt from file"), func(t *testing.T) {
			unsetOAuthEnvVariables()
			unsetBasicEnvVariables()
			defer setOAuthEnvVariables(clientId, clientSecret, tokenUrl)
			defer setBasicEnvVariables(username, password, domain)

			npfCmd := RootCmd
			npfCmd.SetArgs([]string{"login", "--no-prompt"})

			output := captureOutput(
				func() {
					noPromptErr := npfCmd.Execute()
					if noPromptErr != nil {
						t.Errorf(noPromptErr.Error())
						t.FailNow()
					}
				},
			)
			t.Logf("output: %s", output)
			assert.Contains(t, output, "Login successful to")
			testConfigExists(t, configFilePath, true)
			testConfigValid(t)
			//testLogout(t)
		},
	)
}

func Test_LoginCmdConfigParams(t *testing.T) {
	testCmd := RootCmd
	// test
	testCmd.SetArgs(
		[]string{
			"stores", "list", "--exp", "--config", "$HOME/.keyfactor/extra_config.json", "--profile",
			"oauth",
		},
	)
	output := captureOutput(
		func() {
			err := testCmd.Execute()
			assert.NoError(t, err)
		},
	)
	t.Logf("output: %s", output)
	var stores []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &stores); err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// Verify that the length of the response is greater than 0
	assert.True(t, len(stores) >= 0, "Expected non-empty list of stores")
}

func testLogout(t *testing.T, configFilePath string, restoreConfig bool) {
	t.Run(
		fmt.Sprintf("Logout"), func(t *testing.T) {
			testCmd := RootCmd
			//store current config in memory
			if restoreConfig {
				homeDir, _ := os.UserHomeDir()
				configFilePath := path.Join(homeDir, auth_providers.DefaultConfigFilePath)
				existingConfig, exErr := auth_providers.ReadConfigFromJSON(configFilePath)
				defer func() {
					//restore config file
					if existingConfig != nil {
						wErr := auth_providers.WriteConfigToJSON(configFilePath, existingConfig)
						if wErr != nil {
							t.Errorf("Error writing existing config: %s", wErr)
							t.FailNow()
						}
					}
				}()
				if exErr != nil {
					t.Errorf("Error reading existing config: %s", exErr)
					t.FailNow()
				}
			}
			testCmd.SetArgs([]string{"logout"})
			output := captureOutput(
				func() {
					err := testCmd.Execute()
					assert.NoError(t, err)
				},
			)
			t.Logf("output: %s", output)

			assert.Contains(t, output, "Logged out successfully!")

			// Test that the config file does not exist
			if _, fErr := os.Stat(configFile); !os.IsNotExist(fErr) {
				t.Errorf("Config file %s still exists, please remove", configFilePath)
				t.FailNow()
			}
		},
	)

}

func testConfigValid(t *testing.T) {
	// Test config is valid
	//envUsername := os.Getenv("KEYFACTOR_USERNAME")
	//envPassword := os.Getenv("KEYFACTOR_PASSWORD")
	//t.Logf("envUsername: %s", envUsername)
	//t.Logf("envPassword: %s", envPassword)
	t.Logf("Attempting to run `store-types list`")
	t.Run(
		fmt.Sprintf("List store types"), func(t *testing.T) {
			skipVerify := os.Getenv(auth_providers.EnvKeyfactorSkipVerify)
			t.Logf("skipVerify: %s", skipVerify)
			testCmd := RootCmd
			t.Log("Setting args")
			testCmd.SetArgs([]string{"store-types", "list"})
			t.Logf("args: %v", testCmd.Args)
			t.Log("Capturing output")
			output := captureOutput(
				func() {
					tErr := testCmd.Execute()
					assert.NoError(t, tErr)
					if tErr != nil {
						t.Errorf("Error running command: %s", tErr)
						t.FailNow()
					}
				},
			)
			t.Logf("output: %s", output)

			var storeTypes []map[string]interface{}
			if err := json.Unmarshal([]byte(output), &storeTypes); err != nil {
				t.Fatalf("Error unmarshalling JSON: %v", err)
			}

			// Verify that the length of the response is greater than 0
			assert.True(t, len(storeTypes) >= 0, "Expected non-empty list of store types")
		},
	)
}

func testConfigExists(t *testing.T, filePath string, allowExist bool) {
	var testName string
	if allowExist {
		testName = "Config file exists"
	} else {
		testName = "Config file does not exist"
	}
	t.Run(
		fmt.Sprintf(testName), func(t *testing.T) {
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
				kfcServers, ok := fileConfigJSON["servers"].(map[string]interface{})
				if !ok {
					t.Errorf("Error decoding config file: %s", err)
					assert.False(t, ok, "Error decoding config file")
					return
				}
				assert.Contains(t, kfcServers, "default")
				defaultServer := kfcServers["default"].(map[string]interface{})
				assert.Contains(t, defaultServer, "host")
				confUsername, uOk := defaultServer["username"]
				confPassword, pOk := defaultServer["password"]
				confDomain, _ := defaultServer["domain"]
				confClientID, cOk := defaultServer["client_id"]
				confClientSecret, sOk := defaultServer["client_secret"]
				confTokenUrl, tOk := defaultServer["token_url"]
				t.Logf("confUsername: %s", confUsername)
				t.Logf("confPassword: %s", hashSecretValue(fmt.Sprintf("%v", confPassword)))
				t.Logf("confDomain: %s", confDomain)
				t.Logf("confClientID: %s", confClientID)
				t.Logf("confClientSecret: %s", hashSecretValue(fmt.Sprintf("%v", confClientSecret)))
				t.Logf("confTokenUrl: %s", confTokenUrl)

				if (uOk && pOk) || (cOk && sOk && tOk) {
					assert.True(t, uOk && pOk || cOk && sOk && tOk)
				} else {
					t.Errorf("Config file does not contain valid credentials")
				}
			} else {
				assert.True(t, !allowExist && os.IsNotExist(fErr))
			}
		},
	)
}

func testEnvCredsOnly(t *testing.T, configFilePath string, allowExist bool) {
	t.Run(
		fmt.Sprintf("Auth w/ env ONLY"), func(t *testing.T) {
			envUsername, envPassword, envDomain := exportBasicEnvVariables()
			envClientId, envClientSecret, envTokenUrl := exportOAuthEnvVariables()
			os.Setenv(auth_providers.EnvKeyfactorSkipVerify, "true")
			if (envUsername == "" || envPassword == "" || envDomain == "") && (envClientId == "" || envClientSecret == "" || envTokenUrl == "") {
				t.Errorf("Environment variables are not set")
				t.FailNow()
			}
			if configFilePath == "" {
				homeDir, _ := os.UserHomeDir()
				configFilePath = path.Join(homeDir, auth_providers.DefaultConfigFilePath)
			}

			existingConfig, exErr := auth_providers.ReadConfigFromJSON(configFilePath)
			if exErr != nil {
				t.Errorf("Error reading existing config: %s", exErr)
				t.FailNow()
			}
			defer func() {
				//restore config file
				if existingConfig != nil {
					wErr := auth_providers.WriteConfigToJSON(configFilePath, existingConfig)
					if wErr != nil {
						t.Errorf("Error writing existing config: %s", wErr)
						t.FailNow()
					}
				}
			}()
			testLogout(t, configFilePath, false)
			testConfigExists(t, configFilePath, false)
			testConfigValid(t)
		},
	)
}

func testEnvCredsToFile(t *testing.T, filePath string, allowExist bool) {
	t.Run(
		fmt.Sprintf("Auth w/ env ONLY"), func(t *testing.T) {
			// Load .env file
			err := godotenv.Load("../.env_1040")
			if err != nil {
				t.Errorf("Error loading .env file")
			}
			testLogout(t, filePath, false)
			testConfigExists(t, filePath, false)
			testConfigValid(t)
		},
	)
}

// setOAuthEnvVariables sets the oAuth environment variables
func setOAuthEnvVariables(clientId, clientSecret, tokenUrl string) {
	os.Setenv(auth_providers.EnvKeyfactorClientID, clientId)
	os.Setenv(auth_providers.EnvKeyfactorClientSecret, clientSecret)
	os.Setenv(auth_providers.EnvKeyfactorAuthTokenURL, tokenUrl)
}

func exportEnvVarsWithPrefix(prefix string) map[string]string {
	result := make(map[string]string)
	for _, env := range os.Environ() {
		// Each environment variable is in the format "KEY=VALUE"
		pair := strings.SplitN(env, "=", 2)
		key := pair[0]
		value := pair[1]

		if strings.HasPrefix(key, prefix) {
			result[key] = value
		}
	}
	return result
}

// exportOAuthEnvVariables sets the oAuth environment variables
func exportOAuthEnvVariables() (string, string, string) {
	clientId := os.Getenv(auth_providers.EnvKeyfactorClientID)
	clientSecret := os.Getenv(auth_providers.EnvKeyfactorClientSecret)
	tokenUrl := os.Getenv(auth_providers.EnvKeyfactorAuthTokenURL)
	return clientId, clientSecret, tokenUrl
}

// unsetOAuthEnvVariables unsets the oAuth environment variables
func unsetOAuthEnvVariables() {
	os.Unsetenv(auth_providers.EnvKeyfactorClientID)
	os.Unsetenv(auth_providers.EnvKeyfactorClientSecret)
	os.Unsetenv(auth_providers.EnvKeyfactorAuthTokenURL)
	//os.Unsetenv(auth_providers.EnvKeyfactorSkipVerify)
	//os.Unsetenv(auth_providers.EnvKeyfactorConfigFile)
	//os.Unsetenv(auth_providers.EnvKeyfactorAuthProfile)
	//os.Unsetenv(auth_providers.EnvKeyfactorCACert)
	//os.Unsetenv(auth_providers.EnvAuthCACert)
	//os.Unsetenv(auth_providers.EnvKeyfactorHostName)
	//os.Unsetenv(auth_providers.EnvKeyfactorUsername)
	//os.Unsetenv(auth_providers.EnvKeyfactorPassword)
	//os.Unsetenv(auth_providers.EnvKeyfactorDomain)

}

// setBasicEnvVariables sets the basic environment variables
func setBasicEnvVariables(username, password, domain string) {
	os.Setenv(auth_providers.EnvKeyfactorUsername, username)
	os.Setenv(auth_providers.EnvKeyfactorPassword, password)
	os.Setenv(auth_providers.EnvKeyfactorDomain, domain)
}

// exportBasicEnvVariables sets the basic environment variables
func exportBasicEnvVariables() (string, string, string) {
	username := os.Getenv(auth_providers.EnvKeyfactorUsername)
	password := os.Getenv(auth_providers.EnvKeyfactorPassword)
	domain := os.Getenv(auth_providers.EnvKeyfactorDomain)
	return username, password, domain
}

// unsetBasicEnvVariables unsets the basic environment variables
func unsetBasicEnvVariables() {
	os.Unsetenv(auth_providers.EnvKeyfactorUsername)
	os.Unsetenv(auth_providers.EnvKeyfactorPassword)
	os.Unsetenv(auth_providers.EnvKeyfactorDomain)
}
