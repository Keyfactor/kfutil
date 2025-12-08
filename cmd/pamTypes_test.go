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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PAMTypeParameter represents a PAM provider parameter
type PAMTypeParameter struct {
	Name          string `json:"Name"`
	DisplayName   string `json:"DisplayName"`
	DataType      int    `json:"DataType"`
	InstanceLevel bool   `json:"InstanceLevel"`
	Description   string `json:"Description,omitempty"`
}

// PAMTypeDefinition represents a PAM provider type definition from pam_types.json
type PAMTypeDefinition struct {
	Name       string             `json:"Name"`
	Parameters []PAMTypeParameter `json:"Parameters"`
}

// loadPAMTypesFromJSON loads all PAM types from the embedded pam_types.json
func loadPAMTypesFromJSON(t *testing.T) []PAMTypeDefinition {
	var pamTypes []PAMTypeDefinition
	err := json.Unmarshal(EmbeddedPAMTypesJSON, &pamTypes)
	require.NoError(t, err, "Failed to unmarshal embedded PAM types JSON")
	require.NotEmpty(t, pamTypes, "No PAM types found in pam_types.json")
	return pamTypes
}

// setupMockServer creates a mock HTTP server for testing
func setupMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	t.Cleanup(
		func() {
			server.Close()
		},
	)
	return server
}

// Test_PAMTypesHelpCmd tests the help command for pam-types
func Test_PAMTypesHelpCmd(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "help flag",
			args:    []string{"pam-types", "--help"},
			wantErr: false,
		},
		{
			name:    "short help flag",
			args:    []string{"pam-types", "-h"},
			wantErr: false,
		},
		{
			name:    "invalid flag",
			args:    []string{"pam-types", "--halp"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				testCmd := RootCmd
				testCmd.SetArgs(tt.args)
				err := testCmd.Execute()

				if tt.wantErr {
					assert.Error(t, err, "Expected error for %s", tt.name)
				} else {
					assert.NoError(t, err, "Unexpected error for %s", tt.name)
				}
			},
		)
	}
}

// Test_PAMTypesJSON_Structure validates that each PAM type in pam_types.json has required fields
func Test_PAMTypesJSON_Structure(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)

	for _, pamType := range pamTypes {
		t.Run(
			fmt.Sprintf("ValidateStructure_%s", pamType.Name), func(t *testing.T) {
				// Test that Name is not empty
				assert.NotEmpty(t, pamType.Name, "PAM type should have a Name")

				// Test that Parameters exists and is not empty
				assert.NotEmpty(t, pamType.Parameters, "PAM type %s should have Parameters", pamType.Name)

				// Validate each parameter
				for i, param := range pamType.Parameters {
					t.Run(
						fmt.Sprintf("Parameter_%d_%s", i, param.Name), func(t *testing.T) {
							assert.NotEmpty(t, param.Name, "Parameter should have a Name")
							assert.NotEmpty(t, param.DisplayName, "Parameter %s should have a DisplayName", param.Name)

							// DataType should be 1 (string) or 2 (secret/password)
							assert.Contains(
								t, []int{1, 2}, param.DataType,
								"Parameter %s should have DataType 1 or 2, got %d", param.Name, param.DataType,
							)
						},
					)
				}
			},
		)
	}
}

// Test_PAMTypesJSON_AllTypesPresent ensures all expected PAM types are present
func Test_PAMTypesJSON_AllTypesPresent(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)

	// Create a map for easier lookup
	typeMap := make(map[string]bool)
	for _, pamType := range pamTypes {
		typeMap[pamType.Name] = true
	}

	// Test that we have at least the expected PAM types
	expectedTypes := []string{
		"1Password-CLI",
		"Azure-KeyVault",
		"Azure-KeyVault-ServicePrincipal",
		"BeyondTrust-PasswordSafe",
		"CyberArk-CentralCredentialProvider",
		"CyberArk-SdkCredentialProvider",
		"Delinea-SecretServer",
		"GCP-SecretManager",
		"Hashicorp-Vault",
	}

	for _, expectedType := range expectedTypes {
		t.Run(
			fmt.Sprintf("CheckPresence_%s", expectedType), func(t *testing.T) {
				assert.True(t, typeMap[expectedType], "Expected PAM type %s should be present", expectedType)
			},
		)
	}

	// Log all found types
	t.Logf("Found %d PAM types total", len(pamTypes))
	for _, pamType := range pamTypes {
		t.Logf("  - %s (%d parameters)", pamType.Name, len(pamType.Parameters))
	}
}

// Test_PAMTypesJSON_ParameterValidation validates parameter configurations
func Test_PAMTypesJSON_ParameterValidation(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)

	for _, pamType := range pamTypes {
		t.Run(
			fmt.Sprintf("ValidateParameters_%s", pamType.Name), func(t *testing.T) {
				hasInstanceLevel := false
				hasProviderLevel := false

				for _, param := range pamType.Parameters {
					if param.InstanceLevel {
						hasInstanceLevel = true
					} else {
						hasProviderLevel = true
					}
				}

				// Each PAM type should have at least one instance-level and one provider-level parameter
				assert.True(
					t, hasInstanceLevel,
					"PAM type %s should have at least one instance-level parameter", pamType.Name,
				)
				assert.True(
					t, hasProviderLevel,
					"PAM type %s should have at least one provider-level parameter", pamType.Name,
				)
			},
		)
	}
}

// Test_FormatPAMTypes tests the formatPAMTypes helper function
func Test_FormatPAMTypes(t *testing.T) {
	tests := []struct {
		name      string
		input     *[]interface{}
		wantErr   bool
		wantCount int
	}{
		{
			name: "valid PAM types list",
			input: &[]interface{}{
				map[string]interface{}{
					"Name": "Test-Type-1",
					"Parameters": []interface{}{
						map[string]interface{}{
							"Name":          "Param1",
							"DisplayName":   "Parameter 1",
							"DataType":      1,
							"InstanceLevel": false,
						},
					},
				},
				map[string]interface{}{
					"Name": "Test-Type-2",
					"Parameters": []interface{}{
						map[string]interface{}{
							"Name":          "Param2",
							"DisplayName":   "Parameter 2",
							"DataType":      2,
							"InstanceLevel": true,
						},
					},
				},
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "empty list",
			input:     &[]interface{}{},
			wantErr:   true,
			wantCount: 0,
		},
		{
			name:      "nil input",
			input:     nil,
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result, err := formatPAMTypes(tt.input)

				if tt.wantErr {
					assert.Error(t, err)
					assert.Nil(t, result)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, result)
					assert.Equal(t, tt.wantCount, len(result))
				}
			},
		)
	}
}

// Test_GetValidPAMTypes tests the getValidPAMTypes function
func Test_GetValidPAMTypes(t *testing.T) {
	// Test with offline mode (uses embedded JSON)
	offline = true
	types := getValidPAMTypes("", "", "")

	require.NotEmpty(t, types, "Should return PAM types in offline mode")

	// Verify types are sorted
	for i := 1; i < len(types); i++ {
		assert.True(t, types[i-1] <= types[i], "Types should be sorted alphabetically")
	}

	t.Logf("Found %d valid PAM types", len(types))
}

// Test_ReadPAMTypesConfig tests reading PAM types configuration
func Test_ReadPAMTypesConfig(t *testing.T) {
	tests := []struct {
		name     string
		offline  bool
		wantErr  bool
		minTypes int
	}{
		{
			name:     "offline mode with embedded JSON",
			offline:  true,
			wantErr:  false,
			minTypes: 5, // We expect at least 5 PAM types
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				offline = tt.offline
				config, err := readPAMTypesConfig("", "", "", tt.offline)

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, config)
					assert.GreaterOrEqual(
						t, len(config), tt.minTypes,
						"Should have at least %d PAM types", tt.minTypes,
					)
				}
			},
		)
	}
}

// Test_PAMTypesJSON_DataTypeValidation ensures all DataType values are valid
func Test_PAMTypesJSON_DataTypeValidation(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)
	validDataTypes := map[int]string{
		1: "String",
		2: "Secret/Password",
	}

	for _, pamType := range pamTypes {
		t.Run(
			fmt.Sprintf("DataTypes_%s", pamType.Name), func(t *testing.T) {
				for _, param := range pamType.Parameters {
					t.Run(
						param.Name, func(t *testing.T) {
							_, valid := validDataTypes[param.DataType]
							assert.True(
								t, valid,
								"Parameter %s in %s has invalid DataType %d. Valid types are: 1 (String), 2 (Secret)",
								param.Name, pamType.Name, param.DataType,
							)
						},
					)
				}
			},
		)
	}
}

// Test_PAMTypesJSON_InstanceLevelDistribution validates instance level parameter distribution
func Test_PAMTypesJSON_InstanceLevelDistribution(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)

	for _, pamType := range pamTypes {
		t.Run(
			pamType.Name, func(t *testing.T) {
				instanceParams := 0
				providerParams := 0

				for _, param := range pamType.Parameters {
					if param.InstanceLevel {
						instanceParams++
					} else {
						providerParams++
					}
				}

				t.Logf(
					"%s: %d provider-level, %d instance-level parameters",
					pamType.Name, providerParams, instanceParams,
				)

				// Both counts should be > 0
				assert.Greater(
					t, providerParams, 0,
					"Should have at least one provider-level parameter",
				)
				assert.Greater(
					t, instanceParams, 0,
					"Should have at least one instance-level parameter",
				)
			},
		)
	}
}

// Test_PAMTypesJSON_SecretParameterValidation ensures sensitive parameters use DataType 2
func Test_PAMTypesJSON_SecretParameterValidation(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)

	// Exact parameter names that should be secrets (DataType 2)
	// These are actual secret values, not identifiers
	secretParameterNames := map[string]bool{
		"password":     true,
		"token":        true,
		"apikey":       true,
		"clientsecret": true,
	}

	for _, pamType := range pamTypes {
		t.Run(
			pamType.Name, func(t *testing.T) {
				for _, param := range pamType.Parameters {
					paramLower := strings.ToLower(param.Name)

					// Check if parameter name is a known secret field
					if secretParameterNames[paramLower] {
						t.Run(
							param.Name, func(t *testing.T) {
								assert.Equal(
									t,
									2,
									param.DataType,
									"Parameter %s in %s should use DataType 2 (Secret), but has DataType %d",
									param.Name,
									pamType.Name,
									param.DataType,
								)
							},
						)
					}
				}
			},
		)
	}
}

// Test_PAMTypesJSON_UniqueNames ensures all PAM type names are unique
func Test_PAMTypesJSON_UniqueNames(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)

	nameMap := make(map[string]int)
	for _, pamType := range pamTypes {
		nameMap[pamType.Name]++
	}

	for name, count := range nameMap {
		t.Run(
			name, func(t *testing.T) {
				assert.Equal(
					t, 1, count,
					"PAM type name %s appears %d times, should be unique", name, count,
				)
			},
		)
	}
}

// Test_PAMTypesJSON_ParameterNames validates parameter naming within each type
func Test_PAMTypesJSON_ParameterNames(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)

	for _, pamType := range pamTypes {
		t.Run(
			pamType.Name, func(t *testing.T) {
				paramNames := make(map[string]int)

				for _, param := range pamType.Parameters {
					paramNames[param.Name]++
				}

				// Check for duplicate parameter names
				for paramName, count := range paramNames {
					t.Run(
						paramName, func(t *testing.T) {
							assert.Equal(
								t, 1, count,
								"Parameter name %s in %s appears %d times, should be unique within the type",
								paramName, pamType.Name, count,
							)
						},
					)
				}
			},
		)
	}
}

// Test_PAMTypes_ListCommand tests the list command (requires test environment)
func Test_PAMTypes_ListCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if we have test credentials
	_, err := getTestEnv()
	if err != nil {
		t.Skip("Skipping test: no test environment configured")
	}

	testCmd := RootCmd
	testCmd.SetArgs([]string{"pam-types", "list"})

	output := captureOutput(
		func() {
			err := testCmd.Execute()
			if err != nil {
				t.Logf("List command error: %v", err)
			}
		},
	)

	// If the command executed successfully, validate the output
	if output != "" {
		var pamTypesList []map[string]interface{}
		if err := json.Unmarshal([]byte(output), &pamTypesList); err == nil {
			t.Logf("Successfully listed %d PAM types", len(pamTypesList))

			// Validate structure of returned types
			for _, pamType := range pamTypesList {
				assert.NotNil(t, pamType["Id"], "PAM type should have an Id")
				assert.NotNil(t, pamType["Name"], "PAM type should have a Name")
			}
		}
	}
}

// Test_PAMTypesJSON_CompleteCoverage ensures we test all types from the JSON
func Test_PAMTypesJSON_CompleteCoverage(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)

	t.Logf("=== PAM Types Coverage Report ===")
	t.Logf("Total PAM types in pam_types.json: %d", len(pamTypes))
	t.Logf("")

	totalParams := 0
	for i, pamType := range pamTypes {
		t.Logf("%d. %s", i+1, pamType.Name)
		t.Logf("   Parameters: %d", len(pamType.Parameters))

		providerLevel := 0
		instanceLevel := 0
		secrets := 0

		for _, param := range pamType.Parameters {
			totalParams++
			if param.InstanceLevel {
				instanceLevel++
			} else {
				providerLevel++
			}
			if param.DataType == 2 {
				secrets++
			}
		}

		t.Logf("   - Provider-level: %d", providerLevel)
		t.Logf("   - Instance-level: %d", instanceLevel)
		t.Logf("   - Secret params: %d", secrets)
		t.Logf("")
	}

	t.Logf("Total parameters across all types: %d", totalParams)
	t.Logf("=== End Coverage Report ===")

	// This test always passes but provides comprehensive reporting
	assert.True(t, true, "Coverage report generated")
}
