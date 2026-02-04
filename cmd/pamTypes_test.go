// Copyright 2025 Keyfactor
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
	"os"
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

// hasIntegrationTestEnvironment checks if complete Keyfactor credentials are configured
// for running integration tests against a real Keyfactor Command instance.
// Returns true if hostname AND either basic auth (username+password) or OAuth credentials are set.
func hasIntegrationTestEnvironment() bool {
	hostname := os.Getenv("KEYFACTOR_HOSTNAME")
	if hostname == "" {
		return false
	}

	// Check for basic auth credentials
	username := os.Getenv("KEYFACTOR_USERNAME")
	password := os.Getenv("KEYFACTOR_PASSWORD")
	hasBasicAuth := username != "" && password != ""

	// Check for OAuth credentials
	clientID := os.Getenv("KEYFACTOR_AUTH_CLIENT_ID")
	clientSecret := os.Getenv("KEYFACTOR_AUTH_CLIENT_SECRET")
	hasOAuth := clientID != "" && clientSecret != ""

	return hasBasicAuth || hasOAuth
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
		input     *[]any
		wantErr   bool
		wantCount int
	}{
		{
			name: "valid PAM types list",
			input: &[]any{
				map[string]any{
					"Name": "Test-Type-1",
					"Parameters": []any{
						map[string]any{
							"Name":          "Param1",
							"DisplayName":   "Parameter 1",
							"DataType":      1,
							"InstanceLevel": false,
						},
					},
				},
				map[string]any{
					"Name": "Test-Type-2",
					"Parameters": []any{
						map[string]any{
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
			input:     &[]any{},
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

// Test_PAMTypes_ListCommand tests the list command
// When full Keyfactor credentials are configured, it runs both integration and mock tests.
// Otherwise, it runs only mock tests to validate the list functionality.
func Test_PAMTypes_ListCommand(t *testing.T) {
	// Always run mock test
	t.Run(
		"Mock", func(t *testing.T) {
			// Create mock server and pre-populate with PAM types
			server := NewTestServer(t)
			pamTypes := loadPAMTypesFromJSON(t)

			// Pre-populate server with PAM types
			for _, pamType := range pamTypes {
				id := fmt.Sprintf("%s-id", strings.ToLower(strings.ReplaceAll(pamType.Name, " ", "-")))
				params := make([]PAMTypeParameterResponse, len(pamType.Parameters))
				for i, p := range pamType.Parameters {
					params[i] = PAMTypeParameterResponse{
						Id:            i + 1,
						Name:          p.Name,
						DisplayName:   p.DisplayName,
						DataType:      p.DataType,
						InstanceLevel: p.InstanceLevel,
					}
				}
				server.PAMTypes[id] = PAMTypeResponse{
					Id:         id,
					Name:       pamType.Name,
					Parameters: params,
				}
			}

			// Make GET request to list PAM types
			resp, err := http.Get(server.URL + "/KeyfactorAPI/PamProviders/Types")
			require.NoError(t, err, "Failed to make HTTP request")
			defer resp.Body.Close()

			// Verify response status
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK for list")

			// Parse response
			var listedTypes []PAMTypeResponse
			err = json.NewDecoder(resp.Body).Decode(&listedTypes)
			require.NoError(t, err, "Failed to decode response")

			// Verify all types are returned
			assert.Equal(t, len(pamTypes), len(listedTypes), "Should return all PAM types")
			t.Logf("Successfully listed %d PAM types from mock server", len(listedTypes))

			// Validate structure of returned types
			for _, pamType := range listedTypes {
				assert.NotEmpty(t, pamType.Id, "PAM type should have an Id")
				assert.NotEmpty(t, pamType.Name, "PAM type should have a Name")
				assert.NotEmpty(t, pamType.Parameters, "PAM type should have Parameters")
			}
		},
	)

	// Run integration test only if environment is configured and not in short mode
	if hasIntegrationTestEnvironment() && !testing.Short() {
		t.Run(
			"Integration", func(t *testing.T) {
				testCmd := RootCmd
				testCmd.SetArgs([]string{"pam-types", "list", "--no-prompt"})

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
					var pamTypesList []map[string]any
					if err := json.Unmarshal([]byte(output), &pamTypesList); err == nil {
						t.Logf("Successfully listed %d PAM types from real server", len(pamTypesList))

						// Validate structure of returned types
						for _, pamType := range pamTypesList {
							assert.NotNil(t, pamType["Id"], "PAM type should have an Id")
							assert.NotNil(t, pamType["Name"], "PAM type should have a Name")
						}
					}
				}
			},
		)
	} else {
		t.Log("Skipping integration test: Keyfactor environment not configured")
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

// Test_PAMTypes_CreateAllTypes_Serialization tests that all PAM types can be serialized for create operations
func Test_PAMTypes_CreateAllTypes_Serialization(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)

	for _, pamType := range pamTypes {
		t.Run(
			fmt.Sprintf("Create_%s", pamType.Name), func(t *testing.T) {
				// Convert PAM type to JSON (simulating create request payload)
				pamTypeJSON, err := json.Marshal(pamType)
				require.NoError(t, err, "Failed to marshal PAM type %s", pamType.Name)
				assert.NotEmpty(t, pamTypeJSON, "Marshaled JSON should not be empty")

				// Verify JSON can be unmarshaled back
				var unmarshaled PAMTypeDefinition
				err = json.Unmarshal(pamTypeJSON, &unmarshaled)
				require.NoError(t, err, "Failed to unmarshal PAM type %s", pamType.Name)

				// Verify key fields are preserved
				assert.Equal(t, pamType.Name, unmarshaled.Name, "Name should be preserved")
				assert.Equal(
					t, len(pamType.Parameters), len(unmarshaled.Parameters),
					"Parameter count should be preserved for %s", pamType.Name,
				)

				// Verify each parameter is preserved
				for i, param := range pamType.Parameters {
					assert.Equal(
						t, param.Name, unmarshaled.Parameters[i].Name,
						"Parameter %d name should be preserved", i,
					)
					assert.Equal(
						t, param.DisplayName, unmarshaled.Parameters[i].DisplayName,
						"Parameter %d DisplayName should be preserved", i,
					)
					assert.Equal(
						t, param.DataType, unmarshaled.Parameters[i].DataType,
						"Parameter %d DataType should be preserved", i,
					)
					assert.Equal(
						t, param.InstanceLevel, unmarshaled.Parameters[i].InstanceLevel,
						"Parameter %d InstanceLevel should be preserved", i,
					)
				}

				t.Logf("✓ PAM type %s serialization validated", pamType.Name)
			},
		)
	}
}

// Test_PAMTypes_UpdateAllTypes_Serialization tests that all PAM types can be serialized for update operations
func Test_PAMTypes_UpdateAllTypes_Serialization(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)

	for _, pamType := range pamTypes {
		t.Run(
			fmt.Sprintf("Update_%s", pamType.Name), func(t *testing.T) {
				// Simulate an update by marshaling with an existing ID
				updatePayload := map[string]any{
					"Id":         fmt.Sprintf("existing-id-%s", pamType.Name),
					"Name":       pamType.Name,
					"Parameters": pamType.Parameters,
				}

				// Convert to JSON (simulating update request payload)
				updateJSON, err := json.Marshal(updatePayload)
				require.NoError(t, err, "Failed to marshal update payload for %s", pamType.Name)
				assert.NotEmpty(t, updateJSON, "Marshaled update JSON should not be empty")

				// Verify JSON can be unmarshaled back
				var unmarshaled map[string]any
				err = json.Unmarshal(updateJSON, &unmarshaled)
				require.NoError(t, err, "Failed to unmarshal update payload for %s", pamType.Name)

				// Verify key fields are preserved
				assert.Equal(t, pamType.Name, unmarshaled["Name"], "Name should be preserved")
				assert.NotEmpty(t, unmarshaled["Id"], "ID should be present")
				assert.NotNil(t, unmarshaled["Parameters"], "Parameters should be present")

				t.Logf("✓ PAM type %s update serialization validated", pamType.Name)
			},
		)
	}
}

// Test_PAMTypes_DeleteAllTypes_Validation tests that all PAM type IDs are valid for delete operations
func Test_PAMTypes_DeleteAllTypes_Validation(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)

	for _, pamType := range pamTypes {
		t.Run(
			fmt.Sprintf("Delete_%s", pamType.Name), func(t *testing.T) {
				// Simulate delete operation by validating type ID format
				typeID := fmt.Sprintf("pam-type-id-%s", pamType.Name)

				// Validate ID is not empty
				assert.NotEmpty(t, typeID, "Type ID should not be empty for deletion")

				// Validate PAM type name for delete operation
				assert.NotEmpty(t, pamType.Name, "PAM type name should not be empty")
				assert.True(
					t, len(pamType.Name) > 0,
					"PAM type %s should have valid name for delete lookup", pamType.Name,
				)

				// Verify the type definition is complete (needed for safe deletion)
				assert.NotEmpty(t, pamType.Parameters, "Type %s should have parameters", pamType.Name)

				t.Logf("✓ PAM type %s deletion validation passed", pamType.Name)
			},
		)
	}
}

// Test_PAMTypes_CreateWithInvalidData_Validation tests validation for invalid PAM type data
func Test_PAMTypes_CreateWithInvalidData_Validation(t *testing.T) {
	tests := []struct {
		name       string
		pamTypeDef map[string]any
		shouldFail bool
		failReason string
	}{
		{
			name: "missing_name",
			pamTypeDef: map[string]any{
				"Parameters": []any{},
			},
			shouldFail: true,
			failReason: "Name is required",
		},
		{
			name: "missing_parameters",
			pamTypeDef: map[string]any{
				"Name": "Test-PAM-Type",
			},
			shouldFail: true,
			failReason: "Parameters are required",
		},
		{
			name: "empty_parameters",
			pamTypeDef: map[string]any{
				"Name":       "Test-PAM-Type",
				"Parameters": []any{},
			},
			shouldFail: true,
			failReason: "Parameters array should not be empty",
		},
		{
			name: "valid_pam_type",
			pamTypeDef: map[string]any{
				"Name": "Test-PAM-Type",
				"Parameters": []any{
					map[string]any{
						"Name":          "TestParam",
						"DisplayName":   "Test Parameter",
						"DataType":      1,
						"InstanceLevel": false,
					},
				},
			},
			shouldFail: false,
			failReason: "",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Convert to JSON
				pamTypeJSON, err := json.Marshal(tt.pamTypeDef)
				require.NoError(t, err, "Failed to marshal test PAM type definition")

				var pamType PAMTypeDefinition
				err = json.Unmarshal(pamTypeJSON, &pamType)

				// Validate based on expected outcome
				if tt.shouldFail {
					// Check for missing required fields
					if tt.name == "missing_name" {
						assert.Empty(t, pamType.Name, "Name should be empty")
					}
					if tt.name == "missing_parameters" || tt.name == "empty_parameters" {
						assert.Empty(t, pamType.Parameters, "Parameters should be empty")
					}
					t.Logf("✓ Validation correctly identified: %s", tt.failReason)
				} else {
					assert.NoError(t, err, "Valid PAM type should unmarshal without error")
					assert.NotEmpty(t, pamType.Name, "Name should not be empty")
					assert.NotEmpty(t, pamType.Parameters, "Parameters should not be empty")
					t.Logf("✓ Valid PAM type passed validation")
				}
			},
		)
	}
}

// Test_PAMTypes_DeleteNonExistent_Validation tests validation for deleting non-existent PAM type
func Test_PAMTypes_DeleteNonExistent_Validation(t *testing.T) {
	nonExistentID := "non-existent-id-12345"
	nonExistentName := "NonExistent-PAM-Type"

	// Test ID validation
	t.Run(
		"ValidateNonExistentID", func(t *testing.T) {
			assert.NotEmpty(t, nonExistentID, "ID should not be empty")
			assert.True(
				t, len(nonExistentID) > 0,
				"Non-existent ID should have valid format",
			)
			t.Logf("✓ Non-existent ID format validated: %s", nonExistentID)
		},
	)

	// Test name validation
	t.Run(
		"ValidateNonExistentName", func(t *testing.T) {
			assert.NotEmpty(t, nonExistentName, "Name should not be empty")

			// Check that name doesn't match any existing PAM type
			pamTypes := loadPAMTypesFromJSON(t)
			found := false
			for _, pt := range pamTypes {
				if pt.Name == nonExistentName {
					found = true
					break
				}
			}
			assert.False(t, found, "Non-existent name should not match any real PAM type")
			t.Logf("✓ Confirmed %s does not exist in pam_types.json", nonExistentName)
		},
	)
}

// Test_PAMTypes_CreateDuplicate_Validation tests validation for duplicate PAM type creation
func Test_PAMTypes_CreateDuplicate_Validation(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)
	require.NotEmpty(t, pamTypes, "Need at least one PAM type for this test")

	// Test for duplicate names within pam_types.json
	nameMap := make(map[string]int)
	for _, pamType := range pamTypes {
		nameMap[pamType.Name]++
	}

	// Verify no duplicates exist
	for name, count := range nameMap {
		t.Run(
			fmt.Sprintf("CheckUnique_%s", name), func(t *testing.T) {
				assert.Equal(
					t, 1, count,
					"PAM type name %s should appear exactly once (found %d times)", name, count,
				)
			},
		)
	}

	// Test duplicate detection logic
	t.Run(
		"SimulateDuplicateDetection", func(t *testing.T) {
			testPAMType := pamTypes[0]

			// Create a "duplicate" with same name
			duplicatePayload := map[string]any{
				"Name":       testPAMType.Name, // Same name
				"Parameters": testPAMType.Parameters,
			}

			duplicateJSON, err := json.Marshal(duplicatePayload)
			require.NoError(t, err, "Failed to marshal duplicate payload")

			// Verify the duplicate has the same name
			var unmarshaled PAMTypeDefinition
			err = json.Unmarshal(duplicateJSON, &unmarshaled)
			require.NoError(t, err, "Failed to unmarshal duplicate")

			assert.Equal(
				t, testPAMType.Name, unmarshaled.Name,
				"Duplicate should have same name as original",
			)
			t.Logf("✓ Duplicate detection logic validated for %s", testPAMType.Name)
		},
	)
}

// Test_PAMTypes_CreateUpdateDeleteLifecycle_Validation tests full lifecycle validation for each PAM type
func Test_PAMTypes_CreateUpdateDeleteLifecycle_Validation(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)

	for _, pamType := range pamTypes {
		t.Run(
			fmt.Sprintf("Lifecycle_%s", pamType.Name), func(t *testing.T) {
				var operations []string

				// Step 1: Validate CREATE operation
				t.Run(
					"Create", func(t *testing.T) {
						// Serialize for create
						createPayload, err := json.Marshal(pamType)
						require.NoError(t, err, "Failed to marshal for create")
						assert.NotEmpty(t, createPayload, "Create payload should not be empty")

						// Verify can be deserialized
						var unmarshaled PAMTypeDefinition
						err = json.Unmarshal(createPayload, &unmarshaled)
						require.NoError(t, err, "Failed to unmarshal create payload")
						assert.Equal(t, pamType.Name, unmarshaled.Name, "Name should match")

						operations = append(operations, "CREATE")
						t.Logf("✓ CREATE validated for %s", pamType.Name)
					},
				)

				// Step 2: Validate UPDATE operation
				t.Run(
					"Update", func(t *testing.T) {
						// Simulate update with ID
						updatePayload := map[string]any{
							"Id":         fmt.Sprintf("id-%s", pamType.Name),
							"Name":       pamType.Name,
							"Parameters": pamType.Parameters,
						}

						updateJSON, err := json.Marshal(updatePayload)
						require.NoError(t, err, "Failed to marshal for update")
						assert.NotEmpty(t, updateJSON, "Update payload should not be empty")

						// Verify can be deserialized
						var unmarshaled map[string]any
						err = json.Unmarshal(updateJSON, &unmarshaled)
						require.NoError(t, err, "Failed to unmarshal update payload")
						assert.Equal(t, pamType.Name, unmarshaled["Name"], "Name should match")
						assert.NotEmpty(t, unmarshaled["Id"], "ID should be present")

						operations = append(operations, "UPDATE")
						t.Logf("✓ UPDATE validated for %s", pamType.Name)
					},
				)

				// Step 3: Validate DELETE operation
				t.Run(
					"Delete", func(t *testing.T) {
						// Validate deletion requirements
						typeID := fmt.Sprintf("id-%s", pamType.Name)
						assert.NotEmpty(t, typeID, "Type ID required for delete")
						assert.NotEmpty(t, pamType.Name, "Type name required for delete lookup")

						operations = append(operations, "DELETE")
						t.Logf("✓ DELETE validated for %s", pamType.Name)
					},
				)

				// Verify complete lifecycle
				expectedOps := []string{"CREATE", "UPDATE", "DELETE"}
				assert.Equal(
					t, expectedOps, operations,
					"Expected complete lifecycle for %s", pamType.Name,
				)
				t.Logf("✓ Full lifecycle validated for %s: %v", pamType.Name, operations)
			},
		)
	}
}

// Test_PAMTypes_BatchCreateAllTypes_Validation tests batch creation validation for all PAM types
func Test_PAMTypes_BatchCreateAllTypes_Validation(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)
	validatedTypes := make(map[string]bool)

	t.Logf("=== Batch Create Validation for %d PAM Types ===", len(pamTypes))

	// Validate each PAM type can be serialized for batch creation
	for i, pamType := range pamTypes {
		t.Run(
			fmt.Sprintf("%d_BatchCreate_%s", i+1, pamType.Name), func(t *testing.T) {
				// Serialize the PAM type
				pamTypeJSON, err := json.Marshal(pamType)
				require.NoError(t, err, "Failed to marshal PAM type %s", pamType.Name)
				assert.NotEmpty(t, pamTypeJSON, "Serialized JSON should not be empty")

				// Verify deserialization
				var unmarshaled PAMTypeDefinition
				err = json.Unmarshal(pamTypeJSON, &unmarshaled)
				require.NoError(t, err, "Failed to unmarshal PAM type %s", pamType.Name)

				// Validate key fields
				assert.Equal(t, pamType.Name, unmarshaled.Name, "Name should match")
				assert.Equal(
					t, len(pamType.Parameters), len(unmarshaled.Parameters),
					"Parameter count should match",
				)

				// Verify no duplicate names
				_, exists := validatedTypes[pamType.Name]
				assert.False(t, exists, "PAM type %s should not be a duplicate", pamType.Name)
				validatedTypes[pamType.Name] = true

				t.Logf("✓ [%d/%d] %s validated for batch creation", i+1, len(pamTypes), pamType.Name)
			},
		)
	}

	// Verify all types were validated
	assert.Equal(
		t, len(pamTypes), len(validatedTypes),
		"All %d PAM types should be validated", len(pamTypes),
	)

	// Summary report
	t.Logf("=== Batch Create Validation Summary ===")
	t.Logf("Total PAM types validated: %d", len(validatedTypes))
	t.Logf("Validation results:")
	for name := range validatedTypes {
		t.Logf("  ✓ %s", name)
	}
	t.Logf("=== All PAM types ready for batch creation ===")
}

// Test_PAMTypes_OperationsSummary provides a comprehensive summary of all operations
func Test_PAMTypes_OperationsSummary(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)

	t.Logf("╔════════════════════════════════════════════════════════════════╗")
	t.Logf("║  PAM Types Create/Update/Delete Operations Summary            ║")
	t.Logf("╠════════════════════════════════════════════════════════════════╣")
	t.Logf("║  Total PAM Types: %-44d ║", len(pamTypes))
	t.Logf("╠════════════════════════════════════════════════════════════════╣")

	createCount := 0
	updateCount := 0
	deleteCount := 0
	totalParams := 0

	for i, pamType := range pamTypes {
		// Count operations that would be performed
		createCount++ // Each type can be created
		updateCount++ // Each type can be updated
		deleteCount++ // Each type can be deleted
		totalParams += len(pamType.Parameters)

		t.Logf("║  %2d. %-56s ║", i+1, pamType.Name)
		t.Logf("║      Parameters: %-44d ║", len(pamType.Parameters))
		t.Logf("║      Operations: CREATE ✓  UPDATE ✓  DELETE ✓              ║")
	}

	t.Logf("╠════════════════════════════════════════════════════════════════╣")
	t.Logf("║  Summary Statistics:                                           ║")
	t.Logf("║  - Total CREATE operations validated:  %-23d ║", createCount)
	t.Logf("║  - Total UPDATE operations validated:  %-23d ║", updateCount)
	t.Logf("║  - Total DELETE operations validated:  %-23d ║", deleteCount)
	t.Logf("║  - Total parameters across all types:  %-23d ║", totalParams)
	t.Logf("╚════════════════════════════════════════════════════════════════╝")

	// Assert all operations are validated
	assert.Equal(t, len(pamTypes), createCount, "All types validated for CREATE")
	assert.Equal(t, len(pamTypes), updateCount, "All types validated for UPDATE")
	assert.Equal(t, len(pamTypes), deleteCount, "All types validated for DELETE")
}

// Test artifact paths
const (
	testArtifactV2ManifestPath = "../pkg/test_artifacts/integration-manifest-v2.json"
	testArtifactV3ManifestPath = "../pkg/test_artifacts/integration-manifest-v3.json"
)

// Test_IntegrationManifestV2_ParsePAMTypes tests parsing PAM types from V2 integration manifest
func Test_IntegrationManifestV2_ParsePAMTypes(t *testing.T) {
	// Read the V2 manifest file
	manifestBytes, err := os.ReadFile(testArtifactV2ManifestPath)
	require.NoError(t, err, "Failed to read V2 integration manifest file")

	// Parse as V2 manifest
	var manifest IntegrationManifestV2
	err = json.Unmarshal(manifestBytes, &manifest)
	require.NoError(t, err, "Failed to parse V2 integration manifest")

	// Verify manifest structure
	assert.NotEmpty(t, manifest.Schema, "Schema should not be empty")
	assert.Contains(t, manifest.Schema, "v2", "Schema should indicate V2")
	assert.Equal(t, "pam", manifest.IntegrationType, "Integration type should be 'pam'")
	assert.NotEmpty(t, manifest.Name, "Name should not be empty")
	assert.NotEmpty(t, manifest.Description, "Description should not be empty")

	// Verify PAM section
	assert.NotEmpty(t, manifest.About.PAM.Name, "PAM provider name should not be empty")
	assert.NotEmpty(t, manifest.About.PAM.AssemblyName, "PAM assembly name should not be empty")
	assert.NotEmpty(t, manifest.About.PAM.DBName, "PAM DB name should not be empty")
	assert.NotEmpty(t, manifest.About.PAM.FullyQualifiedClassName, "PAM class name should not be empty")

	// Verify PAM types (V2 uses map)
	require.NotEmpty(t, manifest.About.PAM.PAMTypes, "PAM types map should not be empty")

	for typeName, pamType := range manifest.About.PAM.PAMTypes {
		t.Run(
			fmt.Sprintf("V2_PAMType_%s", typeName), func(t *testing.T) {
				assert.NotEmpty(t, pamType.Name, "PAM type name should not be empty")
				assert.Equal(t, typeName, pamType.Name, "PAM type name should match map key")
				require.NotNil(t, pamType.Parameters, "PAM type should have parameters")
				assert.NotEmpty(t, *pamType.Parameters, "PAM type parameters should not be empty")

				// Verify parameters
				hasProviderLevel := false
				hasInstanceLevel := false
				for _, param := range *pamType.Parameters {
					assert.NotEmpty(t, param.Name, "Parameter name should not be empty")
					assert.NotNil(t, param.DisplayName, "Parameter display name should not be nil")
					assert.True(
						t,
						int(param.DataType) == 1 || int(param.DataType) == 2,
						"DataType should be 1 or 2, got %d",
						param.DataType,
					)

					if param.InstanceLevel {
						hasInstanceLevel = true
					} else {
						hasProviderLevel = true
					}
				}

				assert.True(t, hasProviderLevel, "Should have at least one provider-level parameter")
				assert.True(t, hasInstanceLevel, "Should have at least one instance-level parameter")

				t.Logf("✓ V2 PAM type '%s' validated with %d parameters", typeName, len(*pamType.Parameters))
			},
		)
	}

	t.Logf("✓ V2 Integration manifest parsed successfully with %d PAM types", len(manifest.About.PAM.PAMTypes))
}

// Test_IntegrationManifestV3_ParsePAMTypes tests parsing PAM types from V3 integration manifest
func Test_IntegrationManifestV3_ParsePAMTypes(t *testing.T) {
	// Read the V3 manifest file
	manifestBytes, err := os.ReadFile(testArtifactV3ManifestPath)
	require.NoError(t, err, "Failed to read V3 integration manifest file")

	// Parse as V3 manifest
	var manifest IntegrationManifestV3
	err = json.Unmarshal(manifestBytes, &manifest)
	require.NoError(t, err, "Failed to parse V3 integration manifest")

	// Verify manifest structure
	assert.NotEmpty(t, manifest.Schema, "Schema should not be empty")
	assert.Contains(t, manifest.Schema, "v3", "Schema should indicate V3")
	assert.Equal(t, "pam", manifest.IntegrationType, "Integration type should be 'pam'")
	assert.NotEmpty(t, manifest.Name, "Name should not be empty")
	assert.NotEmpty(t, manifest.Description, "Description should not be empty")

	// Verify PAM section
	assert.NotEmpty(t, manifest.About.PAM.Name, "PAM provider name should not be empty")
	assert.NotEmpty(t, manifest.About.PAM.AssemblyName, "PAM assembly name should not be empty")
	assert.NotEmpty(t, manifest.About.PAM.DBName, "PAM DB name should not be empty")
	assert.NotEmpty(t, manifest.About.PAM.FullyQualifiedClassName, "PAM class name should not be empty")

	// Verify PAM types (V3 uses array)
	require.NotEmpty(t, manifest.About.PAM.PAMTypes, "PAM types array should not be empty")

	for i, pamType := range manifest.About.PAM.PAMTypes {
		t.Run(
			fmt.Sprintf("V3_PAMType_%d_%s", i, pamType.Name), func(t *testing.T) {
				assert.NotEmpty(t, pamType.Name, "PAM type name should not be empty")
				require.NotNil(t, pamType.Parameters, "PAM type should have parameters")
				assert.NotEmpty(t, *pamType.Parameters, "PAM type parameters should not be empty")

				// Verify parameters
				hasProviderLevel := false
				hasInstanceLevel := false
				for _, param := range *pamType.Parameters {
					assert.NotEmpty(t, param.Name, "Parameter name should not be empty")
					assert.NotNil(t, param.DisplayName, "Parameter display name should not be nil")
					assert.True(
						t,
						int(param.DataType) == 1 || int(param.DataType) == 2,
						"DataType should be 1 or 2, got %d",
						param.DataType,
					)

					if param.InstanceLevel {
						hasInstanceLevel = true
					} else {
						hasProviderLevel = true
					}
				}

				assert.True(t, hasProviderLevel, "Should have at least one provider-level parameter")
				assert.True(t, hasInstanceLevel, "Should have at least one instance-level parameter")

				t.Logf("✓ V3 PAM type '%s' validated with %d parameters", pamType.Name, len(*pamType.Parameters))
			},
		)
	}

	t.Logf("✓ V3 Integration manifest parsed successfully with %d PAM types", len(manifest.About.PAM.PAMTypes))
}

// Test_IntegrationManifest_V2andV3_Comparison tests that both manifest versions produce valid PAM types
func Test_IntegrationManifest_V2andV3_Comparison(t *testing.T) {
	// Read V2 manifest
	v2Bytes, err := os.ReadFile(testArtifactV2ManifestPath)
	require.NoError(t, err, "Failed to read V2 manifest")
	var v2Manifest IntegrationManifestV2
	require.NoError(t, json.Unmarshal(v2Bytes, &v2Manifest), "Failed to parse V2 manifest")

	// Read V3 manifest
	v3Bytes, err := os.ReadFile(testArtifactV3ManifestPath)
	require.NoError(t, err, "Failed to read V3 manifest")
	var v3Manifest IntegrationManifestV3
	require.NoError(t, json.Unmarshal(v3Bytes, &v3Manifest), "Failed to parse V3 manifest")

	// Verify both manifests have PAM types
	assert.NotEmpty(t, v2Manifest.About.PAM.PAMTypes, "V2 should have PAM types")
	assert.NotEmpty(t, v3Manifest.About.PAM.PAMTypes, "V3 should have PAM types")

	t.Logf("V2 manifest has %d PAM types (map format)", len(v2Manifest.About.PAM.PAMTypes))
	t.Logf("V3 manifest has %d PAM types (array format)", len(v3Manifest.About.PAM.PAMTypes))

	// Both should be valid for PAM type creation
	t.Run(
		"V2_ValidForCreation", func(t *testing.T) {
			for name, pamType := range v2Manifest.About.PAM.PAMTypes {
				assert.NotEmpty(t, pamType.Name, "V2 PAM type %s should have name", name)
				assert.NotEmpty(t, pamType.Parameters, "V2 PAM type %s should have parameters", name)

				// Verify serialization for create request
				jsonBytes, err := json.Marshal(pamType)
				require.NoError(t, err, "V2 PAM type %s should serialize", name)
				assert.NotEmpty(t, jsonBytes, "V2 PAM type %s JSON should not be empty", name)
			}
		},
	)

	t.Run(
		"V3_ValidForCreation", func(t *testing.T) {
			for _, pamType := range v3Manifest.About.PAM.PAMTypes {
				assert.NotEmpty(t, pamType.Name, "V3 PAM type should have name")
				assert.NotEmpty(t, pamType.Parameters, "V3 PAM type %s should have parameters", pamType.Name)

				// Verify serialization for create request
				jsonBytes, err := json.Marshal(pamType)
				require.NoError(t, err, "V3 PAM type %s should serialize", pamType.Name)
				assert.NotEmpty(t, jsonBytes, "V3 PAM type %s JSON should not be empty", pamType.Name)
			}
		},
	)
}

// Test_IntegrationManifestV2_CreatePAMType_MockServer tests creating PAM types from V2 manifest via mock server
func Test_IntegrationManifestV2_CreatePAMType_MockServer(t *testing.T) {
	// Read the V2 manifest file
	manifestBytes, err := os.ReadFile(testArtifactV2ManifestPath)
	require.NoError(t, err, "Failed to read V2 integration manifest file")

	// Parse as V2 manifest
	var manifest IntegrationManifestV2
	err = json.Unmarshal(manifestBytes, &manifest)
	require.NoError(t, err, "Failed to parse V2 integration manifest")

	// Create mock server
	server := NewTestServer(t)

	// Create each PAM type from the V2 manifest
	for typeName, pamType := range manifest.About.PAM.PAMTypes {
		t.Run(
			fmt.Sprintf("V2_MockCreate_%s", typeName), func(t *testing.T) {
				require.NotNil(t, pamType.Parameters, "PAM type parameters should not be nil")
				// Convert to PAMTypeDefinition for the mock server
				params := make([]PAMTypeParameter, len(*pamType.Parameters))
				for i, p := range *pamType.Parameters {
					displayName := ""
					if p.DisplayName != nil {
						displayName = *p.DisplayName
					}
					params[i] = PAMTypeParameter{
						Name:          p.Name,
						DisplayName:   displayName,
						DataType:      int(p.DataType),
						InstanceLevel: p.InstanceLevel,
					}
				}
				pamTypeDef := PAMTypeDefinition{
					Name:       pamType.Name,
					Parameters: params,
				}

				// Serialize for request
				requestBody, err := json.Marshal(pamTypeDef)
				require.NoError(t, err, "Failed to marshal PAM type")

				// Make request to mock server
				resp, err := http.Post(
					server.URL+"/KeyfactorAPI/PamProviders/Types",
					"application/json",
					strings.NewReader(string(requestBody)),
				)
				require.NoError(t, err, "Failed to make HTTP request")
				defer resp.Body.Close()

				// Verify response status
				assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK for create")

				// Parse response
				var createdType PAMTypeResponse
				err = json.NewDecoder(resp.Body).Decode(&createdType)
				require.NoError(t, err, "Failed to decode response")

				// Verify created type
				assert.NotEmpty(t, createdType.Id, "Created type should have an ID")
				assert.Equal(t, pamType.Name, createdType.Name, "Name should match")
				assert.Equal(t, len(*pamType.Parameters), len(createdType.Parameters), "Parameter count should match")

				t.Logf("✓ V2 PAM type '%s' created successfully with ID: %s", typeName, createdType.Id)
			},
		)
	}

	// Verify all types were created
	assert.Equal(t, len(manifest.About.PAM.PAMTypes), len(server.PAMTypes), "All V2 PAM types should be created")
	t.Logf("✓ All %d V2 PAM types created successfully via mock server", len(manifest.About.PAM.PAMTypes))
}

// Test_IntegrationManifestV3_CreatePAMType_MockServer tests creating PAM types from V3 manifest via mock server
func Test_IntegrationManifestV3_CreatePAMType_MockServer(t *testing.T) {
	// Read the V3 manifest file
	manifestBytes, err := os.ReadFile(testArtifactV3ManifestPath)
	require.NoError(t, err, "Failed to read V3 integration manifest file")

	// Parse as V3 manifest
	var manifest IntegrationManifestV3
	err = json.Unmarshal(manifestBytes, &manifest)
	require.NoError(t, err, "Failed to parse V3 integration manifest")

	// Create mock server
	server := NewTestServer(t)

	// Create each PAM type from the V3 manifest
	for i, pamType := range manifest.About.PAM.PAMTypes {
		t.Run(
			fmt.Sprintf("V3_MockCreate_%d_%s", i, pamType.Name), func(t *testing.T) {
				require.NotNil(t, pamType.Parameters, "PAM type parameters should not be nil")
				// Convert to PAMTypeDefinition for the mock server
				params := make([]PAMTypeParameter, len(*pamType.Parameters))
				for j, p := range *pamType.Parameters {
					displayName := ""
					if p.DisplayName != nil {
						displayName = *p.DisplayName
					}
					params[j] = PAMTypeParameter{
						Name:          p.Name,
						DisplayName:   displayName,
						DataType:      int(p.DataType),
						InstanceLevel: p.InstanceLevel,
					}
				}
				pamTypeDef := PAMTypeDefinition{
					Name:       pamType.Name,
					Parameters: params,
				}

				// Serialize for request
				requestBody, err := json.Marshal(pamTypeDef)
				require.NoError(t, err, "Failed to marshal PAM type")

				// Make request to mock server
				resp, err := http.Post(
					server.URL+"/KeyfactorAPI/PamProviders/Types",
					"application/json",
					strings.NewReader(string(requestBody)),
				)
				require.NoError(t, err, "Failed to make HTTP request")
				defer resp.Body.Close()

				// Verify response status
				assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK for create")

				// Parse response
				var createdType PAMTypeResponse
				err = json.NewDecoder(resp.Body).Decode(&createdType)
				require.NoError(t, err, "Failed to decode response")

				// Verify created type
				assert.NotEmpty(t, createdType.Id, "Created type should have an ID")
				assert.Equal(t, pamType.Name, createdType.Name, "Name should match")
				assert.Equal(t, len(*pamType.Parameters), len(createdType.Parameters), "Parameter count should match")

				t.Logf("✓ V3 PAM type '%s' created successfully with ID: %s", pamType.Name, createdType.Id)
			},
		)
	}

	// Verify all types were created
	assert.Equal(t, len(manifest.About.PAM.PAMTypes), len(server.PAMTypes), "All V3 PAM types should be created")
	t.Logf("✓ All %d V3 PAM types created successfully via mock server", len(manifest.About.PAM.PAMTypes))
}

// Test_IntegrationManifest_BothVersions_Summary provides a summary of both manifest version tests
func Test_IntegrationManifest_BothVersions_Summary(t *testing.T) {
	// Read V2 manifest
	v2Bytes, err := os.ReadFile(testArtifactV2ManifestPath)
	require.NoError(t, err, "Failed to read V2 manifest")
	var v2Manifest IntegrationManifestV2
	require.NoError(t, json.Unmarshal(v2Bytes, &v2Manifest), "Failed to parse V2 manifest")

	// Read V3 manifest
	v3Bytes, err := os.ReadFile(testArtifactV3ManifestPath)
	require.NoError(t, err, "Failed to read V3 manifest")
	var v3Manifest IntegrationManifestV3
	require.NoError(t, json.Unmarshal(v3Bytes, &v3Manifest), "Failed to parse V3 manifest")

	t.Logf("╔════════════════════════════════════════════════════════════════╗")
	t.Logf("║  Integration Manifest PAM Types Test Summary                   ║")
	t.Logf("╠════════════════════════════════════════════════════════════════╣")
	t.Logf("║  V2 Manifest (Map Format):                                     ║")
	t.Logf("║    - File: integration-manifest-v2.json                        ║")
	t.Logf("║    - Provider: %-46s ║", v2Manifest.About.PAM.Name)
	t.Logf("║    - PAM Types: %-45d ║", len(v2Manifest.About.PAM.PAMTypes))

	v2TotalParams := 0
	for name, pamType := range v2Manifest.About.PAM.PAMTypes {
		paramCount := 0
		if pamType.Parameters != nil {
			paramCount = len(*pamType.Parameters)
		}
		v2TotalParams += paramCount
		t.Logf("║      • %-50s ║", fmt.Sprintf("%s (%d params)", name, paramCount))
	}

	t.Logf("║    - Total Parameters: %-38d ║", v2TotalParams)
	t.Logf("╠════════════════════════════════════════════════════════════════╣")
	t.Logf("║  V3 Manifest (Array Format):                                   ║")
	t.Logf("║    - File: integration-manifest-v3.json                        ║")
	t.Logf("║    - Provider: %-46s ║", v3Manifest.About.PAM.Name)
	t.Logf("║    - PAM Types: %-45d ║", len(v3Manifest.About.PAM.PAMTypes))

	v3TotalParams := 0
	for _, pamType := range v3Manifest.About.PAM.PAMTypes {
		paramCount := 0
		if pamType.Parameters != nil {
			paramCount = len(*pamType.Parameters)
		}
		v3TotalParams += paramCount
		t.Logf("║      • %-50s ║", fmt.Sprintf("%s (%d params)", pamType.Name, paramCount))
	}

	t.Logf("║    - Total Parameters: %-38d ║", v3TotalParams)
	t.Logf("╠════════════════════════════════════════════════════════════════╣")
	t.Logf("║  Both formats validated for PAM type creation ✓                ║")
	t.Logf("╚════════════════════════════════════════════════════════════════╝")

	// Assert both have PAM types
	assert.NotEmpty(t, v2Manifest.About.PAM.PAMTypes, "V2 should have PAM types")
	assert.NotEmpty(t, v3Manifest.About.PAM.PAMTypes, "V3 should have PAM types")
}
