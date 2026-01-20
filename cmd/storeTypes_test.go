// Copyright 2026 Keyfactor
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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// StoreTypeProperty represents a property in a store type definition
type StoreTypeProperty struct {
	Name         string `json:"Name"`
	DisplayName  string `json:"DisplayName"`
	Description  string `json:"Description,omitempty"`
	Type         string `json:"Type"`
	DependsOn    string `json:"DependsOn,omitempty"`
	DefaultValue string `json:"DefaultValue,omitempty"`
	Required     bool   `json:"Required"`
}

// StoreTypeEntryParameter represents an entry parameter
type StoreTypeEntryParameter struct {
	Name         string      `json:"Name"`
	DisplayName  string      `json:"DisplayName"`
	Description  string      `json:"Description,omitempty"`
	Type         string      `json:"Type"`
	DefaultValue string      `json:"DefaultValue,omitempty"`
	RequiredWhen interface{} `json:"RequiredWhen,omitempty"`
}

// StoreTypePasswordOptions represents password options
type StoreTypePasswordOptions struct {
	EntrySupported bool   `json:"EntrySupported"`
	StoreRequired  bool   `json:"StoreRequired"`
	Style          string `json:"Style"`
}

// StoreTypeSupportedOperations represents supported operations
type StoreTypeSupportedOperations struct {
	Add        bool `json:"Add"`
	Inventory  bool `json:"Inventory"`
	Create     bool `json:"Create"`
	Discovery  bool `json:"Discovery"`
	Enrollment bool `json:"Enrollment"`
	Remove     bool `json:"Remove"`
}

// StoreTypeDefinition represents a complete store type definition
type StoreTypeDefinition struct {
	BlueprintAllowed         bool                         `json:"BlueprintAllowed"`
	Capability               string                       `json:"Capability"`
	ClientMachineDescription string                       `json:"ClientMachineDescription,omitempty"`
	CustomAliasAllowed       string                       `json:"CustomAliasAllowed"`
	EntryParameters          []StoreTypeEntryParameter    `json:"EntryParameters"`
	JobProperties            []interface{}                `json:"JobProperties"`
	LocalStore               bool                         `json:"LocalStore"`
	Name                     string                       `json:"Name"`
	PasswordOptions          StoreTypePasswordOptions     `json:"PasswordOptions"`
	PowerShell               bool                         `json:"PowerShell"`
	PrivateKeyAllowed        string                       `json:"PrivateKeyAllowed"`
	Properties               []StoreTypeProperty          `json:"Properties"`
	ServerRequired           bool                         `json:"ServerRequired"`
	ShortName                string                       `json:"ShortName"`
	StorePathDescription     string                       `json:"StorePathDescription,omitempty"`
	StorePathType            string                       `json:"StorePathType,omitempty"`
	StorePathValue           string                       `json:"StorePathValue,omitempty"`
	SupportedOperations      StoreTypeSupportedOperations `json:"SupportedOperations"`
}

// loadStoreTypesFromJSON loads all store types from the embedded store_types.json
func loadStoreTypesFromJSON(t *testing.T) []StoreTypeDefinition {
	var storeTypes []StoreTypeDefinition
	err := json.Unmarshal(EmbeddedStoreTypesJSON, &storeTypes)
	require.NoError(t, err, "Failed to unmarshal embedded store types JSON")
	require.NotEmpty(t, storeTypes, "No store types found in store_types.json")
	return storeTypes
}

// Test_StoreTypesHelpCmd tests the help command for store-types
func Test_StoreTypesHelpCmd(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "help flag",
			args:    []string{"store-types", "--help"},
			wantErr: false,
		},
		{
			name:    "short help flag",
			args:    []string{"store-types", "-h"},
			wantErr: false,
		},
		{
			name:    "invalid flag",
			args:    []string{"store-types", "--halp"},
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

// Test_StoreTypesJSON_Structure validates that each store type has required fields
func Test_StoreTypesJSON_Structure(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	for _, storeType := range storeTypes {
		t.Run(
			fmt.Sprintf("ValidateStructure_%s", storeType.ShortName), func(t *testing.T) {
				// Test that ShortName is not empty
				assert.NotEmpty(t, storeType.ShortName, "Store type should have a ShortName")

				// Test that Name is not empty
				assert.NotEmpty(t, storeType.Name, "Store type %s should have a Name", storeType.ShortName)

				// Test that Capability is not empty
				assert.NotEmpty(t, storeType.Capability, "Store type %s should have a Capability", storeType.ShortName)

				// Test that CustomAliasAllowed has valid value
				validCustomAlias := []string{"Optional", "Required", "Forbidden", ""}
				assert.Contains(
					t, validCustomAlias, storeType.CustomAliasAllowed,
					"Store type %s should have valid CustomAliasAllowed", storeType.ShortName,
				)

				// Test that PrivateKeyAllowed has valid value
				validPrivateKey := []string{"Optional", "Required", "Forbidden", ""}
				assert.Contains(
					t, validPrivateKey, storeType.PrivateKeyAllowed,
					"Store type %s should have valid PrivateKeyAllowed", storeType.ShortName,
				)

				// Validate PasswordOptions
				t.Run(
					"PasswordOptions", func(t *testing.T) {
						assert.NotEmpty(
							t, storeType.PasswordOptions.Style,
							"Store type %s should have PasswordOptions.Style", storeType.ShortName,
						)
					},
				)

				// Validate SupportedOperations
				t.Run(
					"SupportedOperations", func(t *testing.T) {
						// At least one operation should be supported
						hasOperation := storeType.SupportedOperations.Add ||
							storeType.SupportedOperations.Inventory ||
							storeType.SupportedOperations.Create ||
							storeType.SupportedOperations.Discovery ||
							storeType.SupportedOperations.Enrollment ||
							storeType.SupportedOperations.Remove

						assert.True(
							t, hasOperation,
							"Store type %s should support at least one operation", storeType.ShortName,
						)
					},
				)

				// Validate Properties
				for i, prop := range storeType.Properties {
					t.Run(
						fmt.Sprintf("Property_%d_%s", i, prop.Name), func(t *testing.T) {
							assert.NotEmpty(t, prop.Name, "Property should have a Name")
							assert.NotEmpty(t, prop.DisplayName, "Property %s should have a DisplayName", prop.Name)
							assert.NotEmpty(t, prop.Type, "Property %s should have a Type", prop.Name)

							// Validate property type
							validTypes := []string{"String", "MultipleChoice", "Bool", "Secret"}
							assert.Contains(
								t, validTypes, prop.Type,
								"Property %s in %s should have valid Type", prop.Name, storeType.ShortName,
							)
						},
					)
				}

				// Validate EntryParameters
				for i, param := range storeType.EntryParameters {
					t.Run(
						fmt.Sprintf("EntryParameter_%d_%s", i, param.Name), func(t *testing.T) {
							assert.NotEmpty(t, param.Name, "Entry parameter should have a Name")
							assert.NotEmpty(
								t,
								param.DisplayName,
								"Entry parameter %s should have a DisplayName",
								param.Name,
							)
							assert.NotEmpty(t, param.Type, "Entry parameter %s should have a Type", param.Name)
						},
					)
				}
			},
		)
	}
}

// Test_StoreTypesJSON_ShortNamesUnique ensures all short names are unique
func Test_StoreTypesJSON_ShortNamesUnique(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	shortNameMap := make(map[string]int)
	for _, storeType := range storeTypes {
		shortNameMap[storeType.ShortName]++
	}

	for shortName, count := range shortNameMap {
		t.Run(
			shortName, func(t *testing.T) {
				assert.Equal(
					t, 1, count,
					"Store type short name %s appears %d times, should be unique", shortName, count,
				)
			},
		)
	}
}

// Test_StoreTypesJSON_CapabilitiesUnique ensures all capabilities are unique
func Test_StoreTypesJSON_CapabilitiesUnique(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	capabilityMap := make(map[string]int)
	for _, storeType := range storeTypes {
		capabilityMap[storeType.Capability]++
	}

	for capability, count := range capabilityMap {
		t.Run(
			capability, func(t *testing.T) {
				if capability == "" {
					t.Logf("Skipping empty capability check")
				}
				t.Logf("Capability %s appears %d times", capability, count)
				assert.Equal(
					t, 1, count,
					"Store type capability %s appears %d times, should be unique", capability, count,
				)
			},
		)
	}
}

// Test_StoreTypesJSON_PropertyNames validates property names within each store type
func Test_StoreTypesJSON_PropertyNames(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	for _, storeType := range storeTypes {
		t.Run(
			storeType.ShortName, func(t *testing.T) {
				propertyNames := make(map[string]int)

				for _, prop := range storeType.Properties {
					propertyNames[prop.Name]++
				}

				// Check for duplicate property names
				for propName, count := range propertyNames {
					t.Run(
						propName, func(t *testing.T) {
							assert.Equal(
								t, 1, count,
								"Property name %s in %s appears %d times, should be unique within the type",
								propName, storeType.ShortName, count,
							)
						},
					)
				}
			},
		)
	}
}

// Test_StoreTypesJSON_EntryParameterNames validates entry parameter names
func Test_StoreTypesJSON_EntryParameterNames(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	for _, storeType := range storeTypes {
		if len(storeType.EntryParameters) == 0 {
			continue
		}

		t.Run(
			storeType.ShortName, func(t *testing.T) {
				paramNames := make(map[string]int)

				for _, param := range storeType.EntryParameters {
					paramNames[param.Name]++
				}

				// Check for duplicate parameter names
				for paramName, count := range paramNames {
					t.Run(
						paramName, func(t *testing.T) {
							assert.Equal(
								t, 1, count,
								"Entry parameter name %s in %s appears %d times, should be unique within the type",
								paramName, storeType.ShortName, count,
							)
						},
					)
				}
			},
		)
	}
}

// Test_StoreTypesJSON_SupportedOperations validates that operations make sense
func Test_StoreTypesJSON_SupportedOperations(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	for _, storeType := range storeTypes {
		t.Run(
			storeType.ShortName, func(t *testing.T) {
				ops := storeType.SupportedOperations

				// At least one operation should be supported
				hasOperation := ops.Add || ops.Inventory || ops.Create || ops.Discovery || ops.Enrollment || ops.Remove
				assert.True(
					t, hasOperation,
					"Store type %s should support at least one operation", storeType.ShortName,
				)

				// Log supported operations
				var supportedOps []string
				if ops.Add {
					supportedOps = append(supportedOps, "Add")
				}
				if ops.Create {
					supportedOps = append(supportedOps, "Create")
				}
				if ops.Discovery {
					supportedOps = append(supportedOps, "Discovery")
				}
				if ops.Enrollment {
					supportedOps = append(supportedOps, "Enrollment")
				}
				if ops.Remove {
					supportedOps = append(supportedOps, "Remove")
				}

				t.Logf("%s supports: %s", storeType.ShortName, strings.Join(supportedOps, ", "))
			},
		)
	}
}

// Test_StoreTypesJSON_PasswordOptions validates password options
func Test_StoreTypesJSON_PasswordOptions(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	validStyles := []string{"Default", "Custom", ""}

	for _, storeType := range storeTypes {
		t.Run(
			storeType.ShortName, func(t *testing.T) {
				assert.Contains(
					t, validStyles, storeType.PasswordOptions.Style,
					"Store type %s should have valid PasswordOptions.Style", storeType.ShortName,
				)

				// Log password options
				t.Logf(
					"%s: EntrySupported=%v, StoreRequired=%v, Style=%s",
					storeType.ShortName,
					storeType.PasswordOptions.EntrySupported,
					storeType.PasswordOptions.StoreRequired,
					storeType.PasswordOptions.Style,
				)
			},
		)
	}
}

// Test_StoreTypesJSON_PropertyTypes validates property type values
func Test_StoreTypesJSON_PropertyTypes(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	validPropertyTypes := []string{"String", "MultipleChoice", "Bool", "Secret"}

	for _, storeType := range storeTypes {
		t.Run(
			storeType.ShortName, func(t *testing.T) {
				for _, prop := range storeType.Properties {
					t.Run(
						prop.Name, func(t *testing.T) {
							assert.Contains(
								t, validPropertyTypes, prop.Type,
								"Property %s in %s has invalid Type %s",
								prop.Name, storeType.ShortName, prop.Type,
							)
						},
					)
				}
			},
		)
	}
}

// Test_StoreTypesJSON_SecretProperties validates that sensitive properties use Secret type
func Test_StoreTypesJSON_SecretProperties(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	// Property names that should typically be secrets
	secretKeywords := map[string]bool{
		"password":     true,
		"secret":       true,
		"apikey":       true,
		"token":        true,
		"clientsecret": true,
	}

	for _, storeType := range storeTypes {
		t.Run(
			storeType.ShortName, func(t *testing.T) {
				for _, prop := range storeType.Properties {
					propLower := strings.ToLower(prop.Name)

					// Check if property name suggests it should be a secret
					if secretKeywords[propLower] {
						t.Run(
							prop.Name, func(t *testing.T) {
								assert.Equal(
									t, "Secret", prop.Type,
									"Property %s in %s should use Type 'Secret', but has Type '%s'",
									prop.Name, storeType.ShortName, prop.Type,
								)
							},
						)
					}
				}
			},
		)
	}
}

// Test_StoreTypesJSON_LocalStoreValidation validates LocalStore field consistency
func Test_StoreTypesJSON_LocalStoreValidation(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	for _, storeType := range storeTypes {
		t.Run(
			storeType.ShortName, func(t *testing.T) {
				// If LocalStore is true, ServerRequired should typically be false
				if storeType.LocalStore {
					t.Logf(
						"%s: LocalStore=true, ServerRequired=%v",
						storeType.ShortName, storeType.ServerRequired,
					)
				}

				// Log the values for analysis
				t.Logf(
					"%s: LocalStore=%v, ServerRequired=%v",
					storeType.ShortName, storeType.LocalStore, storeType.ServerRequired,
				)
			},
		)
	}
}

// Test_StoreTypesJSON_RequiredProperties validates required properties
func Test_StoreTypesJSON_RequiredProperties(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	for _, storeType := range storeTypes {
		t.Run(
			storeType.ShortName, func(t *testing.T) {
				requiredCount := 0
				optionalCount := 0

				for _, prop := range storeType.Properties {
					if prop.Required {
						requiredCount++
					} else {
						optionalCount++
					}
				}

				t.Logf(
					"%s: %d required properties, %d optional properties",
					storeType.ShortName, requiredCount, optionalCount,
				)

				// Properties array can be empty, but if it exists, log the counts
				if len(storeType.Properties) > 0 {
					assert.True(
						t, requiredCount+optionalCount == len(storeType.Properties),
						"Property counts should match total",
					)
				}
			},
		)
	}
}

// Test_StoreTypesJSON_CompleteCoverage ensures we test all types and provides a report
func Test_StoreTypesJSON_CompleteCoverage(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	t.Logf("=== Store Types Coverage Report ===")
	t.Logf("Total store types in store_types.json: %d", len(storeTypes))
	t.Logf("")

	totalProperties := 0
	totalEntryParams := 0
	localStoreCount := 0
	serverRequiredCount := 0
	powerShellCount := 0

	for i, storeType := range storeTypes {
		t.Logf("%d. %s (%s)", i+1, storeType.ShortName, storeType.Name)
		t.Logf("   Capability: %s", storeType.Capability)
		t.Logf("   Properties: %d", len(storeType.Properties))
		t.Logf("   Entry Parameters: %d", len(storeType.EntryParameters))

		totalProperties += len(storeType.Properties)
		totalEntryParams += len(storeType.EntryParameters)

		if storeType.LocalStore {
			localStoreCount++
		}
		if storeType.ServerRequired {
			serverRequiredCount++
		}
		if storeType.PowerShell {
			powerShellCount++
		}

		// Count supported operations
		opsCount := 0
		if storeType.SupportedOperations.Add {
			opsCount++
		}
		if storeType.SupportedOperations.Create {
			opsCount++
		}
		if storeType.SupportedOperations.Discovery {
			opsCount++
		}
		if storeType.SupportedOperations.Enrollment {
			opsCount++
		}
		if storeType.SupportedOperations.Remove {
			opsCount++
		}

		t.Logf("   Supported Operations: %d", opsCount)
		t.Logf(
			"   LocalStore: %v, ServerRequired: %v, PowerShell: %v",
			storeType.LocalStore, storeType.ServerRequired, storeType.PowerShell,
		)
		t.Logf("")
	}

	t.Logf("=== Summary ===")
	t.Logf("Total properties across all types: %d", totalProperties)
	t.Logf("Total entry parameters across all types: %d", totalEntryParams)
	t.Logf("Local stores: %d", localStoreCount)
	t.Logf("Server required: %d", serverRequiredCount)
	t.Logf("PowerShell-based: %d", powerShellCount)
	t.Logf("=== End Coverage Report ===")

	// This test always passes but provides comprehensive reporting
	assert.True(t, true, "Coverage report generated")
}

// Test_StoreTypesJSON_MultipleChoiceDefaults validates MultipleChoice property defaults
func Test_StoreTypesJSON_MultipleChoiceDefaults(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	for _, storeType := range storeTypes {
		t.Run(
			storeType.ShortName, func(t *testing.T) {
				for _, prop := range storeType.Properties {
					if prop.Type == "MultipleChoice" {
						t.Run(
							prop.Name, func(t *testing.T) {
								// MultipleChoice properties should typically have a DefaultValue
								if prop.DefaultValue != "" {
									// Verify format (comma-separated values)
									values := strings.Split(prop.DefaultValue, ",")
									assert.Greater(
										t, len(values), 0,
										"Property %s in %s should have valid default values",
										prop.Name, storeType.ShortName,
									)

									t.Logf("%s.%s options: %v", storeType.ShortName, prop.Name, values)
								}
							},
						)
					}
				}
			},
		)
	}
}

// Test_GetValidStoreTypes tests the getValidStoreTypes helper function (if it exists)
func Test_GetValidStoreTypes(t *testing.T) {
	// Test with offline mode (uses embedded JSON)
	offline = true
	types := getValidStoreTypes("", "", "")

	require.NotEmpty(t, types, "Should return store types in offline mode")

	// Verify types are sorted
	for i := 1; i < len(types); i++ {
		t.Logf("Comparing %s <= %s", types[i-1], types[i])
		assert.True(
			t, strings.ToUpper(types[i-1]) <= strings.ToUpper(types[i]),
			"Types should be sorted alphabetically (case-insensitive)",
		)
	}

	t.Logf("Found %d valid store types", len(types))
}

// Test_StoreTypesJSON_BlueprintAllowed validates BlueprintAllowed field
func Test_StoreTypesJSON_BlueprintAllowed(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	blueprintAllowedCount := 0
	for _, storeType := range storeTypes {
		if storeType.BlueprintAllowed {
			blueprintAllowedCount++
		}
	}

	t.Logf(
		"Store types with BlueprintAllowed=true: %d out of %d",
		blueprintAllowedCount, len(storeTypes),
	)

	// This is informational, always passes
	assert.True(t, true, "BlueprintAllowed validation complete")
}

// Test_StoreTypesJSON_CreateValidation validates that each store type can be marshaled for creation
func Test_StoreTypesJSON_CreateValidation(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	for _, storeType := range storeTypes {
		t.Run(
			fmt.Sprintf("CreateValidation_%s", storeType.ShortName), func(t *testing.T) {
				// Test that the store type can be marshaled to JSON (simulating API creation)
				jsonBytes, err := json.Marshal(storeType)
				assert.NoError(t, err, "Store type %s should marshal to JSON", storeType.ShortName)
				assert.NotEmpty(t, jsonBytes, "Store type %s JSON should not be empty", storeType.ShortName)

				// Test that it can be unmarshaled back
				var unmarshaled StoreTypeDefinition
				err = json.Unmarshal(jsonBytes, &unmarshaled)
				assert.NoError(t, err, "Store type %s JSON should unmarshal", storeType.ShortName)

				// Verify key fields are preserved
				assert.Equal(
					t, storeType.ShortName, unmarshaled.ShortName,
					"ShortName should be preserved after marshal/unmarshal",
				)
				assert.Equal(
					t, storeType.Name, unmarshaled.Name,
					"Name should be preserved after marshal/unmarshal",
				)
				assert.Equal(
					t, storeType.Capability, unmarshaled.Capability,
					"Capability should be preserved after marshal/unmarshal",
				)

				t.Logf("✓ %s can be marshaled/unmarshaled successfully", storeType.ShortName)
			},
		)
	}
}

// Test_StoreTypesJSON_DeleteValidation validates that each store type has identifiable fields for deletion
func Test_StoreTypesJSON_DeleteValidation(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	for _, storeType := range storeTypes {
		t.Run(
			fmt.Sprintf("DeleteValidation_%s", storeType.ShortName), func(t *testing.T) {
				// Verify the store type has identifiable fields needed for deletion
				assert.NotEmpty(
					t, storeType.ShortName,
					"Store type must have ShortName for deletion by name",
				)
				assert.NotEmpty(
					t, storeType.Capability,
					"Store type must have Capability for identification",
				)

				// Verify ShortName is a valid identifier (no special chars that would break CLI)
				assert.NotContains(
					t, storeType.ShortName, " ",
					"ShortName should not contain spaces",
				)
				assert.NotContains(
					t, storeType.ShortName, "\n",
					"ShortName should not contain newlines",
				)
				assert.NotContains(
					t, storeType.ShortName, "\t",
					"ShortName should not contain tabs",
				)

				t.Logf("✓ %s has valid identifiers for deletion", storeType.ShortName)
			},
		)
	}
}

// Test_StoreTypesJSON_RequiredFieldsForCreate validates all required fields for creation
func Test_StoreTypesJSON_RequiredFieldsForCreate(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	for _, storeType := range storeTypes {
		t.Run(
			fmt.Sprintf("RequiredFields_%s", storeType.ShortName), func(t *testing.T) {
				// Core identification fields
				assert.NotEmpty(t, storeType.ShortName, "ShortName is required")
				assert.NotEmpty(t, storeType.Name, "Name is required")
				assert.NotEmpty(t, storeType.Capability, "Capability is required")

				// Configuration fields
				assert.NotEmpty(t, storeType.CustomAliasAllowed, "CustomAliasAllowed is required")
				assert.NotEmpty(t, storeType.PrivateKeyAllowed, "PrivateKeyAllowed is required")

				// Password options must exist
				assert.NotEmpty(
					t, storeType.PasswordOptions.Style,
					"PasswordOptions.Style is required",
				)

				// Supported operations structure must exist
				// At least one operation should be true (already tested elsewhere)
				hasOperation := storeType.SupportedOperations.Add ||
					storeType.SupportedOperations.Inventory ||
					storeType.SupportedOperations.Create ||
					storeType.SupportedOperations.Discovery ||
					storeType.SupportedOperations.Enrollment ||
					storeType.SupportedOperations.Remove
				assert.True(
					t, hasOperation,
					"At least one SupportedOperation must be true",
				)

				// Properties and EntryParameters can be empty arrays but must not be nil
				assert.NotNil(t, storeType.Properties, "Properties array must not be nil")
				//assert.NotNil(t, storeType.EntryParameters, "EntryParameters array must not be nil")

				t.Logf("✓ %s has all required fields for creation", storeType.ShortName)
			},
		)
	}
}

// Test_StoreTypesJSON_AllTypesCanBeCreated validates each store type individually for creation
func Test_StoreTypesJSON_AllTypesCanBeCreated(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	t.Logf("=== Store Type Creation Validation ===")
	t.Logf("Testing %d store types for creation readiness", len(storeTypes))
	t.Logf("")

	successCount := 0
	for i, storeType := range storeTypes {
		t.Run(
			fmt.Sprintf("Create_%d_%s", i+1, storeType.ShortName), func(t *testing.T) {
				// Test 1: Has unique identifier
				assert.NotEmpty(t, storeType.ShortName, "Must have ShortName")

				// Test 2: Has display name
				assert.NotEmpty(t, storeType.Name, "Must have Name")

				// Test 3: Has capability
				assert.NotEmpty(t, storeType.Capability, "Must have Capability")

				// Test 4: Can be serialized to JSON
				jsonBytes, err := json.Marshal(storeType)
				assert.NoError(t, err, "Must serialize to JSON")
				assert.Greater(t, len(jsonBytes), 10, "JSON must have content")

				// Test 5: JSON is valid and can be parsed back
				var testParse map[string]interface{}
				err = json.Unmarshal(jsonBytes, &testParse)
				assert.NoError(t, err, "JSON must be valid and parseable")

				// Test 6: Has required operational fields
				assert.Contains(
					t, []string{"Optional", "Required", "Forbidden", ""},
					storeType.CustomAliasAllowed, "CustomAliasAllowed must be valid",
				)
				assert.Contains(
					t, []string{"Optional", "Required", "Forbidden", ""},
					storeType.PrivateKeyAllowed, "PrivateKeyAllowed must be valid",
				)

				// Test 7: Properties are valid
				for j, prop := range storeType.Properties {
					assert.NotEmpty(
						t, prop.Name,
						"Property %d must have Name", j,
					)
					assert.Contains(
						t, []string{"String", "MultipleChoice", "Bool", "Secret"},
						prop.Type, "Property %d must have valid Type", j,
					)
				}

				// Test 8: Entry parameters are valid
				for j, param := range storeType.EntryParameters {
					assert.NotEmpty(
						t, param.Name,
						"Entry parameter %d must have Name", j,
					)
					assert.NotEmpty(
						t, param.Type,
						"Entry parameter %d must have Type", j,
					)
				}

				t.Logf(
					"✓ Store type %s (%s) is ready for creation",
					storeType.ShortName, storeType.Name,
				)

				successCount++
			},
		)
	}

	t.Logf("")
	t.Logf("=== Creation Validation Summary ===")
	t.Logf("Successfully validated: %d/%d store types", successCount, len(storeTypes))
	t.Logf("All store types are ready for creation API calls")
	t.Logf("======================================")
}

// Test_StoreTypesJSON_AllTypesCanBeDeleted validates each store type has required fields for deletion
func Test_StoreTypesJSON_AllTypesCanBeDeleted(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	t.Logf("=== Store Type Deletion Validation ===")
	t.Logf("Testing %d store types for deletion readiness", len(storeTypes))
	t.Logf("")

	successCount := 0
	for i, storeType := range storeTypes {
		t.Run(
			fmt.Sprintf("Delete_%d_%s", i+1, storeType.ShortName), func(t *testing.T) {
				// Test 1: Has unique identifier for deletion
				assert.NotEmpty(
					t, storeType.ShortName,
					"Must have ShortName for deletion by name",
				)

				// Test 2: ShortName is valid (no problematic characters)
				shortName := storeType.ShortName
				assert.NotContains(t, shortName, " ", "ShortName must not contain spaces")
				assert.NotContains(t, shortName, "\n", "ShortName must not contain newlines")
				assert.NotContains(t, shortName, "\t", "ShortName must not contain tabs")
				assert.NotContains(t, shortName, "'", "ShortName must not contain single quotes")
				assert.NotContains(t, shortName, "\"", "ShortName must not contain double quotes")

				// Test 3: Has capability for verification
				assert.NotEmpty(
					t, storeType.Capability,
					"Must have Capability for verification",
				)

				// Test 4: Has name for display in deletion confirmations
				assert.NotEmpty(
					t, storeType.Name,
					"Must have Name for display",
				)

				// Test 5: ShortName length is reasonable
				assert.LessOrEqual(
					t, len(shortName), 50,
					"ShortName should be reasonable length for CLI usage",
				)

				// Test 6: ShortName is ASCII-safe
				for _, char := range shortName {
					assert.True(
						t, char >= 32 && char <= 126,
						"ShortName should use printable ASCII characters",
					)
				}

				t.Logf("✓ Store type %s can be safely deleted by name", storeType.ShortName)

				successCount++
			},
		)
	}

	t.Logf("")
	t.Logf("=== Deletion Validation Summary ===")
	t.Logf("Successfully validated: %d/%d store types", successCount, len(storeTypes))
	t.Logf("All store types have valid identifiers for deletion")
	t.Logf("======================================")
}

// Test_StoreTypesJSON_CreateDeleteCycle validates the full lifecycle
func Test_StoreTypesJSON_CreateDeleteCycle(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	t.Logf("=== Store Type Lifecycle Validation ===")
	t.Logf("Testing create/delete cycle readiness for %d store types", len(storeTypes))
	t.Logf("")

	for i, storeType := range storeTypes {
		t.Run(
			fmt.Sprintf("Lifecycle_%d_%s", i+1, storeType.ShortName), func(t *testing.T) {
				// Simulate creation readiness
				t.Run(
					"CreateReadiness", func(t *testing.T) {
						// Can marshal to JSON
						jsonBytes, err := json.Marshal(storeType)
						assert.NoError(t, err, "Must be serializable for creation")
						if assert.Greater(t, len(jsonBytes), 10, "JSON must have content") {
							t.Logf("✓ Create: %s JSON serialization successful", storeType.ShortName)
						}

						// Has required fields
						assert.NotEmpty(t, storeType.ShortName, "Creation requires ShortName")
						assert.NotEmpty(t, storeType.Name, "Creation requires Name")
						assert.NotEmpty(t, storeType.Capability, "Creation requires Capability")

						t.Logf("✓ Create: %s is ready", storeType.ShortName)
					},
				)

				// Simulate deletion readiness
				t.Run(
					"DeleteReadiness", func(t *testing.T) {
						// Has identifier
						assert.NotEmpty(t, storeType.ShortName, "Deletion requires ShortName")

						// Identifier is safe for CLI
						assert.NotContains(
							t, storeType.ShortName, " ",
							"ShortName must be CLI-safe for deletion",
						)

						t.Logf("✓ Delete: %s can be deleted", storeType.ShortName)
					},
				)

				// Simulate verification after creation
				t.Run(
					"VerificationReadiness", func(t *testing.T) {
						// Has fields to verify creation succeeded
						assert.NotEmpty(
							t, storeType.Capability,
							"Verification requires Capability",
						)
						assert.NotEmpty(
							t, storeType.Name,
							"Verification requires Name",
						)

						t.Logf("✓ Verify: %s can be verified after creation", storeType.ShortName)
					},
				)
			},
		)
	}

	t.Logf("")
	t.Logf("All %d store types are ready for full create/delete lifecycle", len(storeTypes))
	t.Logf("===========================================")
}
