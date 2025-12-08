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
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// StoreTypeTestServer represents a mock Keyfactor API server for testing store types
type StoreTypeTestServer struct {
	*httptest.Server
	StoreTypes map[int]StoreTypeDefinition // id -> StoreType
	NextID     int
	Calls      []StoreTypeAPICall
}

// StoreTypeAPICall represents an API call made to the test server
type StoreTypeAPICall struct {
	Method string
	Path   string
	Body   string
}

// NewStoreTypeTestServer creates a new mock Keyfactor API server for store types
func NewStoreTypeTestServer(t *testing.T) *StoreTypeTestServer {
	ts := &StoreTypeTestServer{
		StoreTypes: make(map[int]StoreTypeDefinition),
		NextID:     1,
		Calls:      []StoreTypeAPICall{},
	}

	mux := http.NewServeMux()

	// GET /KeyfactorAPI/CertificateStoreTypes - List all store types
	mux.HandleFunc(
		"/KeyfactorAPI/CertificateStoreTypes", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				ts.Calls = append(ts.Calls, StoreTypeAPICall{Method: "GET", Path: r.URL.Path})

				// Return all store types
				types := make([]map[string]interface{}, 0, len(ts.StoreTypes))
				for id, storeType := range ts.StoreTypes {
					typeMap := map[string]interface{}{
						"StoreType":                id,
						"Name":                     storeType.Name,
						"ShortName":                storeType.ShortName,
						"Capability":               storeType.Capability,
						"LocalStore":               storeType.LocalStore,
						"SupportedOperations":      storeType.SupportedOperations,
						"Properties":               storeType.Properties,
						"EntryParameters":          storeType.EntryParameters,
						"PasswordOptions":          storeType.PasswordOptions,
						"StorePathType":            storeType.StorePathType,
						"StorePathValue":           storeType.StorePathValue,
						"PrivateKeyAllowed":        storeType.PrivateKeyAllowed,
						"JobProperties":            storeType.JobProperties,
						"ServerRequired":           storeType.ServerRequired,
						"PowerShell":               storeType.PowerShell,
						"BlueprintAllowed":         storeType.BlueprintAllowed,
						"CustomAliasAllowed":       storeType.CustomAliasAllowed,
						"ClientMachineDescription": storeType.ClientMachineDescription,
						"StorePathDescription":     storeType.StorePathDescription,
					}
					types = append(types, typeMap)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(types)
				return
			}

			// POST /KeyfactorAPI/CertificateStoreTypes - Create store type
			if r.Method == http.MethodPost {
				body, _ := io.ReadAll(r.Body)
				ts.Calls = append(
					ts.Calls, StoreTypeAPICall{
						Method: "POST",
						Path:   r.URL.Path,
						Body:   string(body),
					},
				)

				var createReq StoreTypeDefinition
				if err := json.Unmarshal(body, &createReq); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{"Message": "Invalid request body"})
					return
				}

				// Check for duplicate ShortName
				for _, existing := range ts.StoreTypes {
					if existing.ShortName == createReq.ShortName {
						w.WriteHeader(http.StatusConflict)
						json.NewEncoder(w).Encode(
							map[string]string{
								"Message": fmt.Sprintf(
									"Store type with ShortName '%s' already exists",
									createReq.ShortName,
								),
							},
						)
						return
					}
				}

				// Create new store type
				id := ts.NextID
				ts.NextID++
				ts.StoreTypes[id] = createReq

				// Return response
				response := map[string]interface{}{
					"StoreType":                id,
					"Name":                     createReq.Name,
					"ShortName":                createReq.ShortName,
					"Capability":               createReq.Capability,
					"LocalStore":               createReq.LocalStore,
					"SupportedOperations":      createReq.SupportedOperations,
					"Properties":               createReq.Properties,
					"EntryParameters":          createReq.EntryParameters,
					"PasswordOptions":          createReq.PasswordOptions,
					"StorePathType":            createReq.StorePathType,
					"StorePathValue":           createReq.StorePathValue,
					"PrivateKeyAllowed":        createReq.PrivateKeyAllowed,
					"JobProperties":            createReq.JobProperties,
					"ServerRequired":           createReq.ServerRequired,
					"PowerShell":               createReq.PowerShell,
					"BlueprintAllowed":         createReq.BlueprintAllowed,
					"CustomAliasAllowed":       createReq.CustomAliasAllowed,
					"ClientMachineDescription": createReq.ClientMachineDescription,
					"StorePathDescription":     createReq.StorePathDescription,
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
				return
			}

			w.WriteHeader(http.StatusMethodNotAllowed)
		},
	)

	// GET /KeyfactorAPI/CertificateStoreTypes/{id} - Get store type by ID
	// DELETE /KeyfactorAPI/CertificateStoreTypes/{id} - Delete store type
	mux.HandleFunc(
		"/KeyfactorAPI/CertificateStoreTypes/", func(w http.ResponseWriter, r *http.Request) {
			// Extract ID from path
			pathParts := strings.Split(r.URL.Path, "/")
			idStr := pathParts[len(pathParts)-1]
			id, err := strconv.Atoi(idStr)

			if r.Method == http.MethodGet {
				ts.Calls = append(ts.Calls, StoreTypeAPICall{Method: "GET", Path: r.URL.Path})

				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{"Message": "Invalid ID format"})
					return
				}

				storeType, exists := ts.StoreTypes[id]
				if !exists {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]string{"Message": "Store type not found"})
					return
				}

				response := map[string]interface{}{
					"StoreType":                id,
					"Name":                     storeType.Name,
					"ShortName":                storeType.ShortName,
					"Capability":               storeType.Capability,
					"LocalStore":               storeType.LocalStore,
					"SupportedOperations":      storeType.SupportedOperations,
					"Properties":               storeType.Properties,
					"EntryParameters":          storeType.EntryParameters,
					"PasswordOptions":          storeType.PasswordOptions,
					"StorePathType":            storeType.StorePathType,
					"StorePathValue":           storeType.StorePathValue,
					"PrivateKeyAllowed":        storeType.PrivateKeyAllowed,
					"JobProperties":            storeType.JobProperties,
					"ServerRequired":           storeType.ServerRequired,
					"PowerShell":               storeType.PowerShell,
					"BlueprintAllowed":         storeType.BlueprintAllowed,
					"CustomAliasAllowed":       storeType.CustomAliasAllowed,
					"ClientMachineDescription": storeType.ClientMachineDescription,
					"StorePathDescription":     storeType.StorePathDescription,
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
				return
			}

			if r.Method == http.MethodDelete {
				ts.Calls = append(ts.Calls, StoreTypeAPICall{Method: "DELETE", Path: r.URL.Path})

				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{"Message": "Invalid ID format"})
					return
				}

				if _, exists := ts.StoreTypes[id]; !exists {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]string{"Message": "Store type not found"})
					return
				}

				delete(ts.StoreTypes, id)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			w.WriteHeader(http.StatusMethodNotAllowed)
		},
	)

	ts.Server = httptest.NewServer(mux)
	t.Cleanup(
		func() {
			ts.Close()
		},
	)

	return ts
}

// Test_StoreTypes_Mock_CreateAllTypes tests creating all store types via HTTP mock server
func Test_StoreTypes_Mock_CreateAllTypes(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)
	server := NewStoreTypeTestServer(t)

	for _, storeType := range storeTypes {
		t.Run(
			fmt.Sprintf("MockCreate_%s", storeType.ShortName), func(t *testing.T) {
				// Prepare request
				requestBody, err := json.Marshal(storeType)
				require.NoError(t, err, "Failed to marshal store type")

				// Make request to mock server
				resp, err := http.Post(
					server.URL+"/KeyfactorAPI/CertificateStoreTypes",
					"application/json",
					strings.NewReader(string(requestBody)),
				)
				require.NoError(t, err, "Failed to make HTTP request")
				defer resp.Body.Close()

				// Verify response status
				assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK for create")

				// Parse response
				var responseMap map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&responseMap)
				require.NoError(t, err, "Failed to decode response")

				// Verify created type
				assert.NotNil(t, responseMap["StoreType"], "Created type should have a StoreType ID")
				assert.Equal(t, storeType.Name, responseMap["Name"], "Name should match")
				assert.Equal(t, storeType.ShortName, responseMap["ShortName"], "ShortName should match")
				assert.Equal(t, storeType.Capability, responseMap["Capability"], "Capability should match")

				// Verify API call was recorded
				assert.True(
					t, len(server.Calls) > 0,
					"At least one API call should be recorded",
				)
				lastCall := server.Calls[len(server.Calls)-1]
				assert.Equal(t, "POST", lastCall.Method, "Last call should be POST")
				assert.Contains(
					t,
					lastCall.Path,
					"/CertificateStoreTypes",
					"Path should contain /CertificateStoreTypes",
				)

				t.Logf("✓ Successfully created %s via mock HTTP API", storeType.ShortName)
			},
		)
	}
}

// Test_StoreTypes_Mock_ListAllTypes tests listing all store types via HTTP mock server
func Test_StoreTypes_Mock_ListAllTypes(t *testing.T) {
	server := NewStoreTypeTestServer(t)
	storeTypes := loadStoreTypesFromJSON(t)

	// Pre-populate server with first 5 store types
	maxTypes := 500
	if len(storeTypes) < maxTypes {
		maxTypes = len(storeTypes)
	}

	for i := 0; i < maxTypes; i++ {
		server.StoreTypes[server.NextID] = storeTypes[i]
		server.NextID++
	}

	// Make GET request
	resp, err := http.Get(server.URL + "/KeyfactorAPI/CertificateStoreTypes")
	require.NoError(t, err, "Failed to make HTTP request")
	defer resp.Body.Close()

	// Verify response status
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK for list")

	// Parse response
	var listedTypes []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&listedTypes)
	require.NoError(t, err, "Failed to decode response")

	// Verify all types are returned
	assert.Equal(t, maxTypes, len(listedTypes), "Should return all store types")

	// Verify each type has required fields
	for _, typ := range listedTypes {
		assert.NotNil(t, typ["StoreType"], "Should have StoreType ID")
		assert.NotNil(t, typ["Name"], "Should have Name")
		assert.NotNil(t, typ["ShortName"], "Should have ShortName")
		assert.NotNil(t, typ["Capability"], "Should have Capability")
	}

	// Verify API call was recorded
	assert.True(t, len(server.Calls) > 0, "At least one API call should be recorded")
	lastCall := server.Calls[len(server.Calls)-1]
	assert.Equal(t, "GET", lastCall.Method, "Last call should be GET")

	t.Logf("✓ Successfully listed %d store types via mock HTTP API", len(listedTypes))
}

// Test_StoreTypes_Mock_GetByID tests getting a store type by ID
func Test_StoreTypes_Mock_GetByID(t *testing.T) {
	server := NewStoreTypeTestServer(t)
	storeTypes := loadStoreTypesFromJSON(t)
	require.NotEmpty(t, storeTypes, "Need at least one store type")

	// Pre-populate server
	storeType := storeTypes[0]
	id := server.NextID
	server.StoreTypes[id] = storeType
	server.NextID++

	// Make GET request
	resp, err := http.Get(server.URL + "/KeyfactorAPI/CertificateStoreTypes/" + strconv.Itoa(id))
	require.NoError(t, err, "Failed to make HTTP request")
	defer resp.Body.Close()

	// Verify response status
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK for get")

	// Parse response
	var responseMap map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseMap)
	require.NoError(t, err, "Failed to decode response")

	// Verify response
	assert.Equal(t, float64(id), responseMap["StoreType"], "StoreType ID should match")
	assert.Equal(t, storeType.ShortName, responseMap["ShortName"], "ShortName should match")

	t.Logf("✓ Successfully retrieved %s by ID via mock HTTP API", storeType.ShortName)
}

// Test_StoreTypes_Mock_DeleteAllTypes tests deleting store types via HTTP mock server
func Test_StoreTypes_Mock_DeleteAllTypes(t *testing.T) {
	server := NewStoreTypeTestServer(t)
	storeTypes := loadStoreTypesFromJSON(t)

	// Pre-populate server with first 5 store types
	maxTypes := 500
	if len(storeTypes) < maxTypes {
		maxTypes = len(storeTypes)
	}

	typeIDs := make([]int, 0, maxTypes)
	for i := 0; i < maxTypes; i++ {
		id := server.NextID
		typeIDs = append(typeIDs, id)
		server.StoreTypes[id] = storeTypes[i]
		server.NextID++
	}

	// Delete each type
	for i, id := range typeIDs {
		t.Run(
			fmt.Sprintf("MockDelete_%s", storeTypes[i].ShortName), func(t *testing.T) {
				// Make DELETE request
				req, err := http.NewRequest(
					"DELETE",
					server.URL+"/KeyfactorAPI/CertificateStoreTypes/"+strconv.Itoa(id),
					nil,
				)
				require.NoError(t, err, "Failed to create DELETE request")

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err, "Failed to make HTTP request")
				defer resp.Body.Close()

				// Verify response status
				assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected 204 No Content for delete")

				// Verify type was removed from server
				_, exists := server.StoreTypes[id]
				assert.False(t, exists, "Store type should be deleted from server")

				t.Logf("✓ Successfully deleted %s via mock HTTP API", storeTypes[i].ShortName)
			},
		)
	}

	// Verify all types were deleted
	assert.Equal(t, 0, len(server.StoreTypes), "All store types should be deleted from server")
}

// Test_StoreTypes_Mock_CreateDuplicate tests creating duplicate store type
func Test_StoreTypes_Mock_CreateDuplicate(t *testing.T) {
	server := NewStoreTypeTestServer(t)
	storeTypes := loadStoreTypesFromJSON(t)
	require.NotEmpty(t, storeTypes, "Need at least one store type")

	storeType := storeTypes[0]

	// Create first time - should succeed
	requestBody, err := json.Marshal(storeType)
	require.NoError(t, err)

	resp1, err := http.Post(
		server.URL+"/KeyfactorAPI/CertificateStoreTypes",
		"application/json",
		strings.NewReader(string(requestBody)),
	)
	require.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusOK, resp1.StatusCode, "First create should succeed")

	// Create second time - should fail with conflict
	resp2, err := http.Post(
		server.URL+"/KeyfactorAPI/CertificateStoreTypes",
		"application/json",
		strings.NewReader(string(requestBody)),
	)
	require.NoError(t, err)
	defer resp2.Body.Close()

	// Verify conflict response
	assert.Equal(t, http.StatusConflict, resp2.StatusCode, "Second create should fail with 409 Conflict")

	var errorResp map[string]string
	json.NewDecoder(resp2.Body).Decode(&errorResp)
	assert.Contains(
		t, errorResp["Message"], "already exists",
		"Error message should indicate duplicate",
	)

	t.Logf("✓ Duplicate creation correctly rejected with 409 Conflict")
}

// Test_StoreTypes_Mock_DeleteNonExistent tests deleting non-existent store type
func Test_StoreTypes_Mock_DeleteNonExistent(t *testing.T) {
	server := NewStoreTypeTestServer(t)
	nonExistentID := 99999

	// Make DELETE request for non-existent type
	req, err := http.NewRequest(
		"DELETE",
		server.URL+"/KeyfactorAPI/CertificateStoreTypes/"+strconv.Itoa(nonExistentID),
		nil,
	)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify 404 response
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should return 404 Not Found")

	var errorResp map[string]string
	json.NewDecoder(resp.Body).Decode(&errorResp)
	assert.Contains(t, errorResp["Message"], "not found", "Error message should indicate not found")

	t.Logf("✓ Non-existent deletion correctly rejected with 404 Not Found")
}

// Test_StoreTypes_Mock_GetNonExistent tests getting non-existent store type
func Test_StoreTypes_Mock_GetNonExistent(t *testing.T) {
	server := NewStoreTypeTestServer(t)
	nonExistentID := 99999

	// Make GET request for non-existent type
	resp, err := http.Get(server.URL + "/KeyfactorAPI/CertificateStoreTypes/" + strconv.Itoa(nonExistentID))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify 404 response
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should return 404 Not Found")

	var errorResp map[string]string
	json.NewDecoder(resp.Body).Decode(&errorResp)
	assert.Contains(t, errorResp["Message"], "not found", "Error message should indicate not found")

	t.Logf("✓ Non-existent get correctly rejected with 404 Not Found")
}

// Test_StoreTypes_Mock_FullLifecycle tests full lifecycle for store types
func Test_StoreTypes_Mock_FullLifecycle(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)

	// Test first 3 store types
	maxTypes := 500
	if len(storeTypes) < maxTypes {
		maxTypes = len(storeTypes)
	}

	for i := 0; i < maxTypes; i++ {
		storeType := storeTypes[i]
		t.Run(
			fmt.Sprintf("MockLifecycle_%s", storeType.ShortName), func(t *testing.T) {
				server := NewStoreTypeTestServer(t)
				var createdID int

				// Step 1: CREATE
				t.Run(
					"Create", func(t *testing.T) {
						requestBody, err := json.Marshal(storeType)
						require.NoError(t, err)

						resp, err := http.Post(
							server.URL+"/KeyfactorAPI/CertificateStoreTypes",
							"application/json",
							strings.NewReader(string(requestBody)),
						)
						require.NoError(t, err)
						defer resp.Body.Close()

						assert.Equal(t, http.StatusOK, resp.StatusCode, "Create should return 200")

						var responseMap map[string]interface{}
						json.NewDecoder(resp.Body).Decode(&responseMap)
						createdID = int(responseMap["StoreType"].(float64))
						assert.NotZero(t, createdID, "Created ID should not be zero")
						assert.Equal(t, storeType.ShortName, responseMap["ShortName"], "ShortName should match")

						t.Logf("✓ Created %s with ID %d", storeType.ShortName, createdID)
					},
				)

				// Step 2: GET (verify exists)
				t.Run(
					"Get", func(t *testing.T) {
						resp, err := http.Get(server.URL + "/KeyfactorAPI/CertificateStoreTypes/" + strconv.Itoa(createdID))
						require.NoError(t, err)
						defer resp.Body.Close()

						assert.Equal(t, http.StatusOK, resp.StatusCode, "Get should return 200")

						var responseMap map[string]interface{}
						json.NewDecoder(resp.Body).Decode(&responseMap)
						assert.Equal(t, float64(createdID), responseMap["StoreType"], "ID should match")

						t.Logf("✓ Retrieved %s by ID", storeType.ShortName)
					},
				)

				// Step 3: LIST (verify in list)
				t.Run(
					"List", func(t *testing.T) {
						resp, err := http.Get(server.URL + "/KeyfactorAPI/CertificateStoreTypes")
						require.NoError(t, err)
						defer resp.Body.Close()

						assert.Equal(t, http.StatusOK, resp.StatusCode, "List should return 200")

						var types []map[string]interface{}
						json.NewDecoder(resp.Body).Decode(&types)
						assert.Greater(t, len(types), 0, "Should have at least one type")

						found := false
						for _, typ := range types {
							if int(typ["StoreType"].(float64)) == createdID {
								found = true
								break
							}
						}
						assert.True(t, found, "Created type should be in list")

						t.Logf("✓ Verified %s exists in list", storeType.ShortName)
					},
				)

				// Step 4: DELETE
				t.Run(
					"Delete", func(t *testing.T) {
						req, err := http.NewRequest(
							"DELETE",
							server.URL+"/KeyfactorAPI/CertificateStoreTypes/"+strconv.Itoa(createdID),
							nil,
						)
						require.NoError(t, err)

						resp, err := http.DefaultClient.Do(req)
						require.NoError(t, err)
						defer resp.Body.Close()

						assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Delete should return 204")

						// Verify deleted
						_, exists := server.StoreTypes[createdID]
						assert.False(t, exists, "Type should be deleted")

						t.Logf("✓ Deleted %s with ID %d", storeType.ShortName, createdID)
					},
				)

				// Verify API call sequence
				callMethods := []string{}
				for _, call := range server.Calls {
					callMethods = append(callMethods, call.Method)
				}
				expectedSequence := []string{"POST", "GET", "GET", "DELETE"}
				assert.Equal(
					t, expectedSequence, callMethods,
					"Expected POST -> GET -> GET -> DELETE sequence",
				)

				t.Logf("✓ Full lifecycle completed for %s", storeType.ShortName)
			},
		)
	}
}

// Test_StoreTypes_Mock_Summary provides comprehensive summary of mock tests
func Test_StoreTypes_Mock_Summary(t *testing.T) {
	storeTypes := loadStoreTypesFromJSON(t)
	server := NewStoreTypeTestServer(t)

	// Test first 10 store types
	maxTypes := 500
	if len(storeTypes) < maxTypes {
		maxTypes = len(storeTypes)
	}

	t.Logf("╔════════════════════════════════════════════════════════════════╗")
	t.Logf("║  Store Types Mock HTTP API Test Summary                       ║")
	t.Logf("╠════════════════════════════════════════════════════════════════╣")
	t.Logf("║  Mock Server URL: %-44s ║", server.URL)
	t.Logf("║  Total Store Types Available: %-32d ║", len(storeTypes))
	t.Logf("║  Store Types Tested: %-37d ║", maxTypes)
	t.Logf("╠════════════════════════════════════════════════════════════════╣")

	successCount := 0
	for i := 0; i < maxTypes; i++ {
		storeType := storeTypes[i]
		// Test create
		requestBody, _ := json.Marshal(storeType)
		resp, err := http.Post(
			server.URL+"/KeyfactorAPI/CertificateStoreTypes",
			"application/json",
			strings.NewReader(string(requestBody)),
		)

		success := "✓"
		if err != nil || resp.StatusCode != http.StatusOK {
			success = "✗"
		} else {
			successCount++
		}

		if resp != nil {
			resp.Body.Close()
		}

		t.Logf("║  %2d. %-50s %s    ║", i+1, storeType.ShortName, success)
	}

	t.Logf("╠════════════════════════════════════════════════════════════════╣")
	t.Logf("║  Results:                                                      ║")
	t.Logf("║  - Successful HTTP CREATE operations: %-23d ║", successCount)
	t.Logf("║  - Total API calls made:               %-23d ║", len(server.Calls))
	t.Logf("║  - Types stored in mock server:        %-23d ║", len(server.StoreTypes))
	t.Logf("╚════════════════════════════════════════════════════════════════╝")

	assert.Equal(t, maxTypes, successCount, "All types should be created successfully")
}
