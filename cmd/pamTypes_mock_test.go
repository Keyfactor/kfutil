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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer represents a mock Keyfactor API server for testing
type TestServer struct {
	*httptest.Server
	PAMTypes map[string]PAMTypeResponse // id -> PAMType
	Calls    []APICall
}

// APICall represents an API call made to the test server
type APICall struct {
	Method string
	Path   string
	Body   string
}

// PAMTypeResponse represents the API response structure
type PAMTypeResponse struct {
	Id         string                     `json:"Id"`
	Name       string                     `json:"Name"`
	Parameters []PAMTypeParameterResponse `json:"Parameters"`
}

// PAMTypeParameterResponse represents a parameter in the API response
type PAMTypeParameterResponse struct {
	Id            int    `json:"Id"`
	Name          string `json:"Name"`
	DisplayName   string `json:"DisplayName"`
	DataType      int    `json:"DataType"`
	InstanceLevel bool   `json:"InstanceLevel"`
}

// NewTestServer creates a new mock Keyfactor API server
func NewTestServer(t *testing.T) *TestServer {
	ts := &TestServer{
		PAMTypes: make(map[string]PAMTypeResponse),
		Calls:    []APICall{},
	}

	mux := http.NewServeMux()

	// GET /KeyfactorAPI/PamProviders/Types - List all PAM types
	mux.HandleFunc(
		"/KeyfactorAPI/PamProviders/Types", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				ts.Calls = append(ts.Calls, APICall{Method: "GET", Path: r.URL.Path})

				// Return all PAM types
				types := make([]PAMTypeResponse, 0, len(ts.PAMTypes))
				for _, pamType := range ts.PAMTypes {
					types = append(types, pamType)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(types)
				return
			}

			// POST /KeyfactorAPI/PamProviders/Types - Create PAM type
			if r.Method == http.MethodPost {
				body, _ := io.ReadAll(r.Body)
				ts.Calls = append(
					ts.Calls, APICall{
						Method: "POST",
						Path:   r.URL.Path,
						Body:   string(body),
					},
				)

				var createReq PAMTypeDefinition
				if err := json.Unmarshal(body, &createReq); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{"Message": "Invalid request body"})
					return
				}

				// Check for duplicate name
				for _, existing := range ts.PAMTypes {
					if existing.Name == createReq.Name {
						w.WriteHeader(http.StatusConflict)
						json.NewEncoder(w).Encode(
							map[string]string{
								"Message": fmt.Sprintf("PAM Provider Type '%s' already exists", createReq.Name),
							},
						)
						return
					}
				}

				// Create new PAM type
				id := fmt.Sprintf("%s-id", strings.ToLower(strings.ReplaceAll(createReq.Name, " ", "-")))
				params := make([]PAMTypeParameterResponse, len(createReq.Parameters))
				for i, p := range createReq.Parameters {
					params[i] = PAMTypeParameterResponse{
						Id:            i + 1,
						Name:          p.Name,
						DisplayName:   p.DisplayName,
						DataType:      p.DataType,
						InstanceLevel: p.InstanceLevel,
					}
				}

				pamType := PAMTypeResponse{
					Id:         id,
					Name:       createReq.Name,
					Parameters: params,
				}
				ts.PAMTypes[id] = pamType

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(pamType)
				return
			}

			w.WriteHeader(http.StatusMethodNotAllowed)
		},
	)

	// DELETE /KeyfactorAPI/PamProviders/Types/{id} - Delete PAM type
	mux.HandleFunc(
		"/KeyfactorAPI/PamProviders/Types/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete {
				// Extract ID from path
				pathParts := strings.Split(r.URL.Path, "/")
				id := pathParts[len(pathParts)-1]

				ts.Calls = append(ts.Calls, APICall{Method: "DELETE", Path: r.URL.Path})

				if _, exists := ts.PAMTypes[id]; !exists {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(
						map[string]string{
							"Message": "PAM Provider Type not found",
						},
					)
					return
				}

				delete(ts.PAMTypes, id)
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

// Test_PAMTypes_Mock_CreateAllTypes tests creating all PAM types via HTTP mock server
func Test_PAMTypes_Mock_CreateAllTypes(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)
	server := NewTestServer(t)

	for _, pamType := range pamTypes {
		t.Run(
			fmt.Sprintf("MockCreate_%s", pamType.Name), func(t *testing.T) {
				// Prepare request
				requestBody, err := json.Marshal(pamType)
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
				assert.Equal(
					t, len(pamType.Parameters), len(createdType.Parameters),
					"Parameter count should match",
				)

				// Verify parameters
				for i, param := range pamType.Parameters {
					assert.Equal(t, param.Name, createdType.Parameters[i].Name, "Parameter name should match")
					assert.Equal(
						t, param.DisplayName, createdType.Parameters[i].DisplayName,
						"Parameter DisplayName should match",
					)
					assert.Equal(
						t, param.DataType, createdType.Parameters[i].DataType,
						"Parameter DataType should match",
					)
					assert.Equal(
						t, param.InstanceLevel, createdType.Parameters[i].InstanceLevel,
						"Parameter InstanceLevel should match",
					)
				}

				// Verify API call was recorded
				assert.True(
					t, len(server.Calls) > 0,
					"At least one API call should be recorded",
				)
				lastCall := server.Calls[len(server.Calls)-1]
				assert.Equal(t, "POST", lastCall.Method, "Last call should be POST")
				assert.Contains(t, lastCall.Path, "/PamProviders/Types", "Path should contain /PamProviders/Types")

				t.Logf("✓ Successfully created %s via mock HTTP API", pamType.Name)
			},
		)
	}

	// Verify all types were created
	assert.Equal(
		t, len(pamTypes), len(server.PAMTypes),
		"All PAM types should be created in server",
	)
}

// Test_PAMTypes_Mock_ListAllTypes tests listing all PAM types via HTTP mock server
func Test_PAMTypes_Mock_ListAllTypes(t *testing.T) {
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

	// Make GET request
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

	// Verify each type exists in response
	typeMap := make(map[string]bool)
	for _, typ := range listedTypes {
		typeMap[typ.Name] = true
	}

	for _, pamType := range pamTypes {
		assert.True(
			t, typeMap[pamType.Name],
			"PAM type %s should be in list response", pamType.Name,
		)
	}

	// Verify API call was recorded
	assert.True(t, len(server.Calls) > 0, "At least one API call should be recorded")
	lastCall := server.Calls[len(server.Calls)-1]
	assert.Equal(t, "GET", lastCall.Method, "Last call should be GET")

	t.Logf("✓ Successfully listed %d PAM types via mock HTTP API", len(listedTypes))
}

// Test_PAMTypes_Mock_DeleteAllTypes tests deleting all PAM types via HTTP mock server
func Test_PAMTypes_Mock_DeleteAllTypes(t *testing.T) {
	server := NewTestServer(t)
	pamTypes := loadPAMTypesFromJSON(t)

	// Pre-populate server with PAM types
	typeIDs := make([]string, 0, len(pamTypes))
	for _, pamType := range pamTypes {
		id := fmt.Sprintf("%s-id", strings.ToLower(strings.ReplaceAll(pamType.Name, " ", "-")))
		typeIDs = append(typeIDs, id)
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

	// Delete each type
	for i, id := range typeIDs {
		t.Run(
			fmt.Sprintf("MockDelete_%s", pamTypes[i].Name), func(t *testing.T) {
				// Make DELETE request
				req, err := http.NewRequest(
					"DELETE",
					server.URL+"/KeyfactorAPI/PamProviders/Types/"+id,
					nil,
				)
				require.NoError(t, err, "Failed to create DELETE request")

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err, "Failed to make HTTP request")
				defer resp.Body.Close()

				// Verify response status
				assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected 204 No Content for delete")

				// Verify type was removed from server
				_, exists := server.PAMTypes[id]
				assert.False(t, exists, "PAM type should be deleted from server")

				t.Logf("✓ Successfully deleted %s via mock HTTP API", pamTypes[i].Name)
			},
		)
	}

	// Verify all types were deleted
	assert.Equal(t, 0, len(server.PAMTypes), "All PAM types should be deleted from server")
}

// Test_PAMTypes_Mock_CreateDuplicate tests creating duplicate PAM type
func Test_PAMTypes_Mock_CreateDuplicate(t *testing.T) {
	server := NewTestServer(t)
	pamTypes := loadPAMTypesFromJSON(t)
	require.NotEmpty(t, pamTypes, "Need at least one PAM type")

	pamType := pamTypes[0]

	// Create first time - should succeed
	requestBody, err := json.Marshal(pamType)
	require.NoError(t, err)

	resp1, err := http.Post(
		server.URL+"/KeyfactorAPI/PamProviders/Types",
		"application/json",
		strings.NewReader(string(requestBody)),
	)
	require.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusOK, resp1.StatusCode, "First create should succeed")

	// Create second time - should fail with conflict
	resp2, err := http.Post(
		server.URL+"/KeyfactorAPI/PamProviders/Types",
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

// Test_PAMTypes_Mock_DeleteNonExistent tests deleting non-existent PAM type
func Test_PAMTypes_Mock_DeleteNonExistent(t *testing.T) {
	server := NewTestServer(t)
	nonExistentID := "non-existent-id-12345"

	// Make DELETE request for non-existent type
	req, err := http.NewRequest(
		"DELETE",
		server.URL+"/KeyfactorAPI/PamProviders/Types/"+nonExistentID,
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

// Test_PAMTypes_Mock_FullLifecycle tests full lifecycle for each PAM type
func Test_PAMTypes_Mock_FullLifecycle(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)

	for _, pamType := range pamTypes {
		t.Run(
			fmt.Sprintf("MockLifecycle_%s", pamType.Name), func(t *testing.T) {
				server := NewTestServer(t)
				var createdID string

				// Step 1: CREATE
				t.Run(
					"Create", func(t *testing.T) {
						requestBody, err := json.Marshal(pamType)
						require.NoError(t, err)

						resp, err := http.Post(
							server.URL+"/KeyfactorAPI/PamProviders/Types",
							"application/json",
							strings.NewReader(string(requestBody)),
						)
						require.NoError(t, err)
						defer resp.Body.Close()

						assert.Equal(t, http.StatusOK, resp.StatusCode, "Create should return 200")

						var created PAMTypeResponse
						json.NewDecoder(resp.Body).Decode(&created)
						createdID = created.Id
						assert.NotEmpty(t, createdID, "Created ID should not be empty")
						assert.Equal(t, pamType.Name, created.Name, "Name should match")

						t.Logf("✓ Created %s with ID %s", pamType.Name, createdID)
					},
				)

				// Step 2: LIST (verify exists)
				t.Run(
					"List", func(t *testing.T) {
						resp, err := http.Get(server.URL + "/KeyfactorAPI/PamProviders/Types")
						require.NoError(t, err)
						defer resp.Body.Close()

						assert.Equal(t, http.StatusOK, resp.StatusCode, "List should return 200")

						var types []PAMTypeResponse
						json.NewDecoder(resp.Body).Decode(&types)
						assert.Greater(t, len(types), 0, "Should have at least one type")

						found := false
						for _, typ := range types {
							if typ.Id == createdID {
								found = true
								break
							}
						}
						assert.True(t, found, "Created type should be in list")

						t.Logf("✓ Verified %s exists in list", pamType.Name)
					},
				)

				// Step 3: DELETE
				t.Run(
					"Delete", func(t *testing.T) {
						req, err := http.NewRequest(
							"DELETE",
							server.URL+"/KeyfactorAPI/PamProviders/Types/"+createdID,
							nil,
						)
						require.NoError(t, err)

						resp, err := http.DefaultClient.Do(req)
						require.NoError(t, err)
						defer resp.Body.Close()

						assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Delete should return 204")

						// Verify deleted
						_, exists := server.PAMTypes[createdID]
						assert.False(t, exists, "Type should be deleted")

						t.Logf("✓ Deleted %s with ID %s", pamType.Name, createdID)
					},
				)

				// Verify API call sequence
				callMethods := []string{}
				for _, call := range server.Calls {
					callMethods = append(callMethods, call.Method)
				}
				expectedSequence := []string{"POST", "GET", "DELETE"}
				assert.Equal(
					t, expectedSequence, callMethods,
					"Expected POST -> GET -> DELETE sequence",
				)

				t.Logf("✓ Full lifecycle completed for %s", pamType.Name)
			},
		)
	}
}

// Test_PAMTypes_Mock_Summary provides comprehensive summary of mock tests
func Test_PAMTypes_Mock_Summary(t *testing.T) {
	pamTypes := loadPAMTypesFromJSON(t)
	server := NewTestServer(t)

	t.Logf("╔════════════════════════════════════════════════════════════════╗")
	t.Logf("║  PAM Types Mock HTTP API Test Summary                         ║")
	t.Logf("╠════════════════════════════════════════════════════════════════╣")
	t.Logf("║  Mock Server URL: %-44s ║", server.URL)
	t.Logf("║  Total PAM Types: %-44d ║", len(pamTypes))
	t.Logf("╠════════════════════════════════════════════════════════════════╣")

	successCount := 0
	for i, pamType := range pamTypes {
		// Test create
		requestBody, _ := json.Marshal(pamType)
		resp, err := http.Post(
			server.URL+"/KeyfactorAPI/PamProviders/Types",
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

		t.Logf("║  %2d. %-50s %s    ║", i+1, pamType.Name, success)
	}

	t.Logf("╠════════════════════════════════════════════════════════════════╣")
	t.Logf("║  Results:                                                      ║")
	t.Logf("║  - Successful HTTP CREATE operations: %-23d ║", successCount)
	t.Logf("║  - Total API calls made:               %-23d ║", len(server.Calls))
	t.Logf("║  - Types stored in mock server:        %-23d ║", len(server.PAMTypes))
	t.Logf("╚════════════════════════════════════════════════════════════════╝")

	assert.Equal(t, len(pamTypes), successCount, "All types should be created successfully")
}
