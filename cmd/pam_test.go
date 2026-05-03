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
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PAMProviderResponse represents a PAM provider in API responses
type PAMProviderResponse struct {
	Id                      int                     `json:"Id"`
	Name                    string                  `json:"Name"`
	ProviderType            PAMProviderTypeInfo     `json:"ProviderType"`
	ProviderTypeParamValues []PAMProviderParamValue `json:"ProviderTypeParamValues"`
}

// PAMProviderTypeInfo represents the provider type info in a PAM provider
type PAMProviderTypeInfo struct {
	Id                 int                        `json:"Id"`
	Name               string                     `json:"Name"`
	ProviderTypeParams []PAMProviderTypeParamInfo `json:"ProviderTypeParams"`
}

// PAMProviderTypeParamInfo represents a parameter in a provider type
type PAMProviderTypeParamInfo struct {
	Id       int    `json:"Id"`
	Name     string `json:"Name"`
	DataType int    `json:"DataType"`
}

// PAMProviderParamValue represents a parameter value in a PAM provider
type PAMProviderParamValue struct {
	Id                int                      `json:"Id"`
	Value             string                   `json:"Value"`
	ProviderTypeParam PAMProviderTypeParamInfo `json:"ProviderTypeParam"`
}

// PAMProviderTestServer represents a mock Keyfactor API server for PAM provider testing
type PAMProviderTestServer struct {
	*httptest.Server
	PAMProviders map[int]PAMProviderResponse
	nextID       int
}

// NewPAMProviderTestServer creates a new mock server for PAM provider testing
func NewPAMProviderTestServer(t *testing.T) *PAMProviderTestServer {
	ts := &PAMProviderTestServer{
		PAMProviders: make(map[int]PAMProviderResponse),
		nextID:       1,
	}

	// Pre-populate with some test providers
	ts.PAMProviders[1] = PAMProviderResponse{
		Id:   1,
		Name: "Test-Provider-1",
		ProviderType: PAMProviderTypeInfo{
			Id:   1,
			Name: "Hashicorp-Vault",
			ProviderTypeParams: []PAMProviderTypeParamInfo{
				{Id: 1, Name: "Host", DataType: 1},
				{Id: 2, Name: "Token", DataType: 2},
				{Id: 3, Name: "MountPoint", DataType: 1},
				{Id: 4, Name: "SecretPath", DataType: 1},
			},
		},
		ProviderTypeParamValues: []PAMProviderParamValue{
			{
				Id:                1,
				Value:             "https://vault.example.com",
				ProviderTypeParam: PAMProviderTypeParamInfo{Id: 1, Name: "Host", DataType: 1},
			},
			{Id: 2, Value: "s.xxxxx", ProviderTypeParam: PAMProviderTypeParamInfo{Id: 2, Name: "Token", DataType: 2}},
		},
	}
	ts.PAMProviders[2] = PAMProviderResponse{
		Id:   2,
		Name: "Test-Provider-2",
		ProviderType: PAMProviderTypeInfo{
			Id:   2,
			Name: "Azure-KeyVault",
			ProviderTypeParams: []PAMProviderTypeParamInfo{
				{Id: 1, Name: "VaultName", DataType: 1},
				{Id: 2, Name: "TenantId", DataType: 1},
				{Id: 3, Name: "ClientId", DataType: 1},
				{Id: 4, Name: "ClientSecret", DataType: 2},
			},
		},
		ProviderTypeParamValues: []PAMProviderParamValue{
			{
				Id:                1,
				Value:             "my-keyvault",
				ProviderTypeParam: PAMProviderTypeParamInfo{Id: 1, Name: "VaultName", DataType: 1},
			},
		},
	}
	ts.nextID = 3

	mux := http.NewServeMux()

	// GET/POST /KeyfactorAPI/PamProviders - List or Create PAM providers
	mux.HandleFunc(
		"/KeyfactorAPI/PamProviders", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				providers := make([]PAMProviderResponse, 0, len(ts.PAMProviders))
				for _, p := range ts.PAMProviders {
					providers = append(providers, p)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(providers)
				return
			}

			if r.Method == http.MethodPost {
				var createReq PAMProviderResponse
				if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{"Message": "Invalid request body"})
					return
				}

				// Check for duplicate name
				for _, existing := range ts.PAMProviders {
					if existing.Name == createReq.Name {
						w.WriteHeader(http.StatusConflict)
						json.NewEncoder(w).Encode(
							map[string]string{
								"Message": fmt.Sprintf(
									"PAM Provider '%s' already exists",
									createReq.Name,
								),
							},
						)
						return
					}
				}

				// Assign new ID
				createReq.Id = ts.nextID
				ts.nextID++
				ts.PAMProviders[createReq.Id] = createReq

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(createReq)
				return
			}

			w.WriteHeader(http.StatusMethodNotAllowed)
		},
	)

	// GET/PUT/DELETE /KeyfactorAPI/PamProviders/{id}
	mux.HandleFunc(
		"/KeyfactorAPI/PamProviders/", func(w http.ResponseWriter, r *http.Request) {
			// Extract ID from path
			pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/KeyfactorAPI/PamProviders/"), "/")
			if len(pathParts) == 0 || pathParts[0] == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			id, err := strconv.Atoi(pathParts[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if r.Method == http.MethodGet {
				provider, exists := ts.PAMProviders[id]
				if !exists {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]string{"Message": "PAM Provider not found"})
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(provider)
				return
			}

			if r.Method == http.MethodPut {
				if _, exists := ts.PAMProviders[id]; !exists {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]string{"Message": "PAM Provider not found"})
					return
				}

				var updateReq PAMProviderResponse
				if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{"Message": "Invalid request body"})
					return
				}

				// Preserve the ID from the URL
				updateReq.Id = id
				ts.PAMProviders[id] = updateReq

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(updateReq)
				return
			}

			if r.Method == http.MethodDelete {
				if _, exists := ts.PAMProviders[id]; !exists {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]string{"Message": "PAM Provider not found"})
					return
				}
				delete(ts.PAMProviders, id)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			w.WriteHeader(http.StatusMethodNotAllowed)
		},
	)

	ts.Server = httptest.NewServer(mux)
	t.Cleanup(func() { ts.Server.Close() })

	return ts
}

func Test_PAMHelpCmd(t *testing.T) {
	defer resetRootCommandState()

	// Test root help
	testCmd := RootCmd
	testCmd.SetArgs([]string{"pam", "--help"})
	err := testCmd.Execute()

	assert.NoError(t, err)

	// test root halp
	testCmd.SetArgs([]string{"pam", "-h"})
	err = testCmd.Execute()
	assert.NoError(t, err)

	// test root halp
	testCmd.SetArgs([]string{"pam", "--halp"})
	err = testCmd.Execute()

	assert.Error(t, err)
	// check if error was returned
	if err := testCmd.Execute(); err == nil {
		t.Errorf("RootCmd() = %v, shouldNotPass %v", err, true)
	}
}

func Test_PAMListCmd(t *testing.T) {
	// Always run mock test
	t.Run(
		"Mock", func(t *testing.T) {
			server := NewPAMProviderTestServer(t)

			// Make request to mock server
			resp, err := http.Get(server.URL + "/KeyfactorAPI/PamProviders")
			require.NoError(t, err, "Failed to make HTTP request")
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

			var providers []PAMProviderResponse
			err = json.NewDecoder(resp.Body).Decode(&providers)
			require.NoError(t, err, "Failed to decode response")

			assert.Equal(t, 2, len(providers), "Should have 2 pre-populated providers")
			t.Logf("✓ Successfully listed %d PAM providers from mock server", len(providers))

			// Validate structure
			for _, p := range providers {
				assert.NotEmpty(t, p.Id, "Provider should have Id")
				assert.NotEmpty(t, p.Name, "Provider should have Name")
				assert.NotEmpty(t, p.ProviderType.Name, "Provider should have ProviderType.Name")
			}
		},
	)

	// Run integration test only if environment is configured
	if hasIntegrationTestEnvironment() {
		t.Run(
			"Integration", func(t *testing.T) {
				pamProviders, err := testListPamProviders(t)
				assert.NoError(t, err)
				if err != nil {
					t.Errorf("failed to list PAM providers: %v", err)
					return
				}

				if len(pamProviders) <= 0 {
					t.Errorf("0 PAM providers found, cannot test list")
				}
			},
		)
	} else {
		t.Log("Skipping integration test: Keyfactor environment not configured")
	}
}

func Test_PAMTypesListCmd(t *testing.T) {
	// This tests the deprecated command which returns a deprecation error
	testCmd := RootCmd
	var err error
	testCmd.SetArgs([]string{"pam", "types-list"})
	captureOutput(
		func() {
			err = testCmd.Execute()
		},
	)

	// The deprecated command returns an error with the deprecation message
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "this command is deprecated; use `pam types list`")
}

func Test_PAMGetCmd(t *testing.T) {
	// Always run mock test
	t.Run(
		"Mock", func(t *testing.T) {
			server := NewPAMProviderTestServer(t)

			// Test getting provider with ID 1
			resp, err := http.Get(server.URL + "/KeyfactorAPI/PamProviders/1")
			require.NoError(t, err, "Failed to make HTTP request")
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

			var provider PAMProviderResponse
			err = json.NewDecoder(resp.Body).Decode(&provider)
			require.NoError(t, err, "Failed to decode response")

			assert.Equal(t, 1, provider.Id, "Provider ID should be 1")
			assert.Equal(t, "Test-Provider-1", provider.Name, "Provider name should match")
			assert.Equal(t, "Hashicorp-Vault", provider.ProviderType.Name, "Provider type should match")
			t.Logf("✓ Successfully retrieved PAM provider '%s' from mock server", provider.Name)
		},
	)

	// Test getting non-existent provider
	t.Run(
		"Mock_NotFound", func(t *testing.T) {
			server := NewPAMProviderTestServer(t)

			resp, err := http.Get(server.URL + "/KeyfactorAPI/PamProviders/999")
			require.NoError(t, err, "Failed to make HTTP request")
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Expected 404 Not Found")
			t.Logf("✓ Correctly returned 404 for non-existent provider")
		},
	)

	// Run integration test only if environment is configured
	if hasIntegrationTestEnvironment() {
		t.Run(
			"Integration", func(t *testing.T) {
				testCmd := RootCmd
				pamProviders, err := testListPamProviders(t)
				assert.NoError(t, err)
				if err != nil {
					t.Fatalf("failed to list PAM providers: %v", err)
				}

				if len(pamProviders) > 0 {
					for _, p := range pamProviders {
						providerConfig := p.(map[string]any)
						assert.NotEmpty(t, providerConfig["Name"])
						assert.NotEmpty(t, providerConfig["Id"])
						assert.NotEmpty(t, providerConfig["ProviderType"])

						pTypeParams, _ := providerConfig["ProviderType"].(map[string]any)["ProviderTypeParams"].([]any)
						assert.GreaterOrEqual(t, len(pTypeParams), 0)
						if len(pTypeParams) > 0 {
							for _, param := range pTypeParams {
								assert.NotEmpty(t, param.(map[string]any)["Id"])
								assert.NotEmpty(t, param.(map[string]any)["Name"])
								assert.NotEmpty(t, param.(map[string]any)["DataType"])
							}
						}

						idInt := int(providerConfig["Id"].(float64))
						idStr := strconv.Itoa(idInt)
						testCmd.SetArgs([]string{"pam", "get", "--id", idStr})
						output := captureOutput(
							func() {
								err := testCmd.Execute()
								assert.NoError(t, err)
							},
						)
						var pamProvider any
						if err := json.Unmarshal([]byte(output), &pamProvider); err != nil {
							t.Fatalf("Error unmarshalling JSON: %v", err)
						}
						assert.NotEmpty(t, pamProvider.(map[string]any)["Name"])
						assert.NotEmpty(t, pamProvider.(map[string]any)["Id"])
						assert.NotEmpty(t, pamProvider.(map[string]any)["ProviderType"])
					}
				} else {
					t.Errorf("0 PAM providers found, cannot test get")
				}
			},
		)
	} else {
		t.Log("Skipping integration test: Keyfactor environment not configured")
	}
}

func Test_PAMTypesCreateCmd(t *testing.T) {
	// This tests the deprecated command which returns a deprecation error
	testCmd := RootCmd
	testCmd.SetArgs([]string{"pam", "types-create"})
	var err error
	captureOutput(
		func() {
			err = testCmd.Execute()
		},
	)

	// The deprecated command returns an error with the deprecation message
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "this command is deprecated; use `pam types create`")
}

func Test_PAMCreateCmd_Mock(t *testing.T) {
	// Test creating a new provider
	t.Run(
		"CreateNew", func(t *testing.T) {
			server := NewPAMProviderTestServer(t)

			newProvider := PAMProviderResponse{
				Name: "Mock-Test-Provider",
				ProviderType: PAMProviderTypeInfo{
					Id:   1,
					Name: "Hashicorp-Vault",
				},
				ProviderTypeParamValues: []PAMProviderParamValue{
					{
						Value: "https://vault.test.com",
						ProviderTypeParam: PAMProviderTypeParamInfo{
							Id:       1,
							Name:     "Host",
							DataType: 1,
						},
					},
				},
			}

			body, err := json.Marshal(newProvider)
			require.NoError(t, err)

			resp, err := http.Post(
				server.URL+"/KeyfactorAPI/PamProviders",
				"application/json",
				strings.NewReader(string(body)),
			)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

			var createdProvider PAMProviderResponse
			err = json.NewDecoder(resp.Body).Decode(&createdProvider)
			require.NoError(t, err)

			assert.Equal(t, 3, createdProvider.Id, "Should have been assigned ID 3")
			assert.Equal(t, "Mock-Test-Provider", createdProvider.Name)
			assert.Equal(t, "Hashicorp-Vault", createdProvider.ProviderType.Name)
			t.Logf("✓ Successfully created PAM provider '%s' with ID %d", createdProvider.Name, createdProvider.Id)

			// Verify provider is in the map
			assert.Contains(t, server.PAMProviders, 3, "Provider should exist in map")
		},
	)

	// Test creating duplicate provider
	t.Run(
		"CreateDuplicate", func(t *testing.T) {
			server := NewPAMProviderTestServer(t)

			// Try to create a provider with an existing name
			duplicateProvider := PAMProviderResponse{
				Name: "Test-Provider-1", // This name already exists
				ProviderType: PAMProviderTypeInfo{
					Id:   1,
					Name: "Hashicorp-Vault",
				},
			}

			body, err := json.Marshal(duplicateProvider)
			require.NoError(t, err)

			resp, err := http.Post(
				server.URL+"/KeyfactorAPI/PamProviders",
				"application/json",
				strings.NewReader(string(body)),
			)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusConflict, resp.StatusCode, "Expected 409 Conflict for duplicate name")
			t.Logf("✓ Correctly rejected duplicate provider creation")
		},
	)
}

func Test_PAMUpdateCmd_Mock(t *testing.T) {
	// Test updating existing provider
	t.Run(
		"UpdateExisting", func(t *testing.T) {
			server := NewPAMProviderTestServer(t)

			// Verify original name
			originalProvider := server.PAMProviders[1]
			assert.Equal(t, "Test-Provider-1", originalProvider.Name)

			// Update the provider
			updatedProvider := PAMProviderResponse{
				Id:   1,
				Name: "Updated-Provider-1",
				ProviderType: PAMProviderTypeInfo{
					Id:   1,
					Name: "Hashicorp-Vault",
				},
				ProviderTypeParamValues: []PAMProviderParamValue{
					{
						Id:    1,
						Value: "https://updated-vault.example.com",
						ProviderTypeParam: PAMProviderTypeParamInfo{
							Id:       1,
							Name:     "Host",
							DataType: 1,
						},
					},
				},
			}

			body, err := json.Marshal(updatedProvider)
			require.NoError(t, err)

			req, err := http.NewRequest(
				http.MethodPut,
				server.URL+"/KeyfactorAPI/PamProviders/1",
				strings.NewReader(string(body)),
			)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

			var returnedProvider PAMProviderResponse
			err = json.NewDecoder(resp.Body).Decode(&returnedProvider)
			require.NoError(t, err)

			assert.Equal(t, 1, returnedProvider.Id, "ID should be preserved")
			assert.Equal(t, "Updated-Provider-1", returnedProvider.Name)
			t.Logf("✓ Successfully updated PAM provider to '%s'", returnedProvider.Name)

			// Verify the update in the server's map
			assert.Equal(t, "Updated-Provider-1", server.PAMProviders[1].Name)
		},
	)

	// Test updating non-existent provider
	t.Run(
		"UpdateNonExistent", func(t *testing.T) {
			server := NewPAMProviderTestServer(t)

			updateReq := PAMProviderResponse{
				Id:   999,
				Name: "Non-Existent",
			}

			body, err := json.Marshal(updateReq)
			require.NoError(t, err)

			req, err := http.NewRequest(
				http.MethodPut,
				server.URL+"/KeyfactorAPI/PamProviders/999",
				strings.NewReader(string(body)),
			)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Expected 404 Not Found")
			t.Logf("✓ Correctly returned 404 for non-existent provider update")
		},
	)
}

func Test_PAMDeleteCmd_Mock(t *testing.T) {
	// Test deleting existing provider
	t.Run(
		"DeleteExisting", func(t *testing.T) {
			server := NewPAMProviderTestServer(t)

			// Verify provider exists first
			assert.Contains(t, server.PAMProviders, 1, "Provider 1 should exist before delete")

			// Delete provider
			req, err := http.NewRequest(http.MethodDelete, server.URL+"/KeyfactorAPI/PamProviders/1", nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected 204 No Content")

			// Verify provider is deleted
			assert.NotContains(t, server.PAMProviders, 1, "Provider 1 should be deleted")
			t.Logf("✓ Successfully deleted PAM provider from mock server")
		},
	)

	// Test deleting non-existent provider
	t.Run(
		"DeleteNonExistent", func(t *testing.T) {
			server := NewPAMProviderTestServer(t)

			req, err := http.NewRequest(http.MethodDelete, server.URL+"/KeyfactorAPI/PamProviders/999", nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Expected 404 Not Found")
			t.Logf("✓ Correctly returned 404 for non-existent provider deletion")
		},
	)
}

func Test_PAMCreateCmd(t *testing.T) {
	// Run integration test only if environment is configured
	if !hasIntegrationTestEnvironment() {
		t.Skip("Skipping integration test: Keyfactor environment not configured")
	}

	// get current working dir
	cwd, wdErr := os.Getwd()
	if wdErr != nil {
		cwd = "./"
	}
	t.Logf("cwd: %s", cwd)

	providerName := "Delinea-SecretServer-test"
	t.Logf("providerName: %s", providerName)
	inputFileName := path.Join(filepath.Dir(cwd), "artifacts/pam/pam-create-template.json")
	t.Logf("inputFileName: %s", inputFileName)
	invalidInputFileName := path.Join(filepath.Dir(cwd), "artifacts/pam/pam-create-invalid.json")
	t.Logf("invalidInputFileName: %s", invalidInputFileName)

	updatedFileName, fErr := testFormatPamCreateConfig(t, inputFileName, "", false)
	t.Logf("updatedFileName: %s", updatedFileName)
	assert.NoError(t, fErr)
	if fErr != nil {
		t.Errorf("failed to format PAM provider config file '%s': %v", inputFileName, fErr)
		return
	}

	// Test valid config file
	createResponse, err := testCreatePamProvider(t, updatedFileName, providerName, false)
	if err != nil && testCheckBug63171(err) {
		t.Skip("PAM Provider creation is not supported in Keyfactor Command version 12 and later")
	}
	assert.NoError(t, err)
	assert.NotNil(t, createResponse)
	if err != nil {
		t.Errorf("failed to create a PAM provider: %v", err)
		return
	}

	createdObject := createResponse.(map[string]any)
	createdId := int(createdObject["Id"].(float64))

	// Test creating same provider again
	_, err = testCreatePamProvider(t, inputFileName, providerName, true)
	assert.Error(t, err)
	if err == nil {
		t.Fatalf("this test should have failed to create a duplicate PAM provider: %v", err)
	}

	// Delete the provider we just created
	err = testDeletePamProvider(t, createdId, false)
	assert.NoError(t, err)
	if err != nil {
		t.Fatalf("failed to delete a PAM provider %d: %v", createdId, err)
	}

	// Test invalid config file
	_, err = testCreatePamProvider(t, invalidInputFileName, providerName, true)
	assert.Error(t, err)
	if err == nil {
		t.Fatalf("this test should have failed to create a PAM provider: %v", err)
	}

	// delete the updated file
	os.Remove(updatedFileName)
}

func Test_PAMUpdateCmd(t *testing.T) {
	// Run integration test only if environment is configured
	if !hasIntegrationTestEnvironment() {
		t.Skip("Skipping integration test: Keyfactor environment not configured")
	}

	// get current working dir
	cwd, wdErr := os.Getwd()
	if wdErr != nil {
		cwd = "./"
	}
	t.Logf("cwd: %s", cwd)

	providerName := "Delinea-SecretServer-test"
	t.Logf("providerName: %s", providerName)
	inputFileName := path.Join(filepath.Dir(cwd), "artifacts/pam/pam-create-template.json")
	t.Logf("inputFileName: %s", inputFileName)
	invalidInputFileName := path.Join(filepath.Dir(cwd), "artifacts/pam/pam-create-invalid.json")
	t.Logf("invalidInputFileName: %s", invalidInputFileName)

	// read input file into a map[string]any
	updatedFileName, fErr := testFormatPamCreateConfig(t, inputFileName, "", false)
	assert.NoError(t, fErr)
	if fErr != nil {
		t.Fatalf("failed to format PAM provider config file '%s': %v", inputFileName, fErr)
		return
	}
	// Create a provider to delete, doesn't matter if it fails, assume it exists then delete it
	testCreatePamProvider(t, updatedFileName, providerName, true)

	updatedFileName, fErr = testFormatPamCreateConfig(t, inputFileName, providerName, true)
	assert.NoError(t, fErr)
	if fErr != nil {
		t.Fatalf("failed to format PAM provider config file '%s': %v", inputFileName, fErr)
		return
	}

	testCmd := RootCmd
	testCmd.SetArgs([]string{"pam", "update", "--from-file", updatedFileName})
	output := captureOutput(
		func() {
			err := testCmd.Execute()
			if err != nil {
				if testCheckBug63171(err) {
					t.Skip("Updating PAM Providers is not supported in Keyfactor Command version 12 and later")
				}
				t.Errorf("failed to update a PAM provider: %v", err)
				t.FailNow()
			}
		},
	)

	var updateResponse any
	if err := json.Unmarshal([]byte(output), &updateResponse); err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}
	assert.NotNil(t, updateResponse)
	if updateResponse == nil {
		t.Errorf("failed to update a PAM provider")
		return
	}

	_, ok := updateResponse.(map[string]any)
	if !ok {
		t.Errorf("updateResponse is not a map[string]any")
		return
	}
	assert.NotEmpty(t, updateResponse.(map[string]any)["Id"])
	assert.NotEmpty(t, updateResponse.(map[string]any)["Name"])
	assert.Equal(t, updateResponse.(map[string]any)["Name"], providerName)
	assert.NotEmpty(t, updateResponse.(map[string]any)["ProviderType"])
	assert.NotEmpty(t, updateResponse.(map[string]any)["ProviderTypeParamValues"])

	// delete the pam provider we just created
	testDeletePamProvider(t, int(updateResponse.(map[string]any)["Id"].(float64)), false)

	// delete the updated file
	os.Remove(updatedFileName)
}

func Test_PAMDeleteCmd(t *testing.T) {
	// Run integration test only if environment is configured
	if !hasIntegrationTestEnvironment() {
		t.Skip("Skipping integration test: Keyfactor environment not configured")
	}

	// get current working dir
	cwd, wdErr := os.Getwd()
	if wdErr != nil {
		cwd = "./"
	}
	t.Logf("cwd: %s", cwd)

	providerName := "Delinea-SecretServer-test"
	t.Logf("providerName: %s", providerName)
	inputFileName := path.Join(filepath.Dir(cwd), "artifacts/pam/pam-create-template.json")
	t.Logf("inputFileName: %s", inputFileName)
	invalidInputFileName := path.Join(filepath.Dir(cwd), "artifacts/pam/pam-create-invalid.json")
	t.Logf("invalidInputFileName: %s", invalidInputFileName)

	// read input file into a map[string]any
	updatedFileName, fErr := testFormatPamCreateConfig(t, inputFileName, "", false)
	assert.NoError(t, fErr)
	if fErr != nil {
		t.Fatalf("failed to format PAM provider config file '%s': %v", inputFileName, fErr)
		return
	}
	// Create a provider to delete, doesn't matter if it fails, assume it exists then delete it
	_, cErr := testCreatePamProvider(t, updatedFileName, providerName, true)
	if cErr != nil && testCheckBug63171(cErr) {
		t.Skip("PAM Provider creation is not supported in Keyfactor Command version 12 and later")
	}

	// list providers
	providersList, err := testListPamProviders(t)
	assert.NoError(t, err)
	if err != nil {
		t.Fatalf("failed to list PAM providers: %v", err)
	}
	if len(providersList) > 0 {
		//find the one named providerName
		isDeleted := false
		for _, p := range providersList {
			providerConfig := p.(map[string]any)
			if providerConfig["Name"] == providerName {
				idInt := int(providerConfig["Id"].(float64))
				dErr := testDeletePamProvider(t, idInt, false)
				assert.NoError(t, dErr)
				isDeleted = true
				break
			}
		}
		if !isDeleted {
			t.Fatalf("failed to find PAM provider %s to delete", providerName)
		}
	} else {
		t.Fatalf("0 PAM providers found, cannot test delete")
	}
	// delete the updated file
	os.Remove(updatedFileName)
}

// Test_PAMProviderLifecycle_Mock tests the full lifecycle of a PAM provider using mock server
func Test_PAMProviderLifecycle_Mock(t *testing.T) {
	server := NewPAMProviderTestServer(t)

	// 1. List providers
	t.Run(
		"List", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/KeyfactorAPI/PamProviders")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var providers []PAMProviderResponse
			err = json.NewDecoder(resp.Body).Decode(&providers)
			require.NoError(t, err)

			assert.Equal(t, 2, len(providers), "Should start with 2 providers")
			t.Logf("✓ Listed %d providers", len(providers))
		},
	)

	// 2. Get specific provider
	t.Run(
		"Get", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/KeyfactorAPI/PamProviders/1")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var provider PAMProviderResponse
			err = json.NewDecoder(resp.Body).Decode(&provider)
			require.NoError(t, err)

			assert.Equal(t, "Test-Provider-1", provider.Name)
			t.Logf("✓ Got provider: %s", provider.Name)
		},
	)

	// 3. Delete provider
	t.Run(
		"Delete", func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, server.URL+"/KeyfactorAPI/PamProviders/1", nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
			t.Logf("✓ Deleted provider 1")
		},
	)

	// 4. Verify deletion
	t.Run(
		"VerifyDeletion", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/KeyfactorAPI/PamProviders/1")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
			t.Logf("✓ Verified provider 1 no longer exists")
		},
	)

	// 5. List again to confirm only 1 provider remains
	t.Run(
		"ListAfterDelete", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/KeyfactorAPI/PamProviders")
			require.NoError(t, err)
			defer resp.Body.Close()

			var providers []PAMProviderResponse
			err = json.NewDecoder(resp.Body).Decode(&providers)
			require.NoError(t, err)

			assert.Equal(t, 1, len(providers), "Should have 1 provider after deletion")
			t.Logf("✓ Confirmed %d provider remains", len(providers))
		},
	)
}

func testListPamProviders(t *testing.T) ([]any, error) {
	var output string
	var pamProviders []any
	var err error

	t.Run(
		"Listing PAM provider instances", func(t *testing.T) {
			testCmd := RootCmd
			testCmd.SetArgs([]string{"pam", "list"})
			output = captureOutput(
				func() {
					err = testCmd.Execute()
					assert.NoError(t, err)
				},
			)

			if err != nil {
				t.Errorf("failed to list PAM providers: %v", err)
				return
			}

			if err = json.Unmarshal([]byte(output), &pamProviders); err != nil {
				t.Fatalf("Error unmarshalling JSON: %v", err)
			}

			assert.GreaterOrEqual(t, len(pamProviders), 0)

			if len(pamProviders) > 0 {
				for _, p := range pamProviders {
					providerConfig := p.(map[string]any)
					assert.NotEmpty(t, providerConfig["Name"])
					assert.NotEmpty(t, providerConfig["Id"])
					assert.NotEmpty(t, providerConfig["ProviderType"])

					pTypeParams, _ := providerConfig["ProviderType"].(map[string]any)["ProviderTypeParams"].([]any)
					assert.GreaterOrEqual(t, len(pTypeParams), 0)
					if len(pTypeParams) > 0 {
						for _, param := range pTypeParams {
							assert.NotEmpty(t, param.(map[string]any)["Id"])
							assert.NotEmpty(t, param.(map[string]any)["Name"])
							assert.NotEmpty(t, param.(map[string]any)["DataType"])
						}
					}
				}
			} else {
				t.Errorf("0 PAM providers found, cannot test list")
				t.Fail()
			}

		},
	)
	if err != nil {
		t.Log(output)
		return nil, err
	}
	return pamProviders, nil
}

func testCreatePamProvider(t *testing.T, fileName string, providerName string, allowFail bool) (any, error) {
	var err error
	var createResponse any
	var testName string
	if allowFail {
		testName = fmt.Sprintf("Create PAM provider '%s' allow fail", providerName)
	} else {
		testName = fmt.Sprintf("Create PAM provider '%s'", providerName)
	}
	var bug63171 error
	t.Run(
		testName, func(t *testing.T) {
			testCmd := RootCmd

			args := []string{"pam", "create", "--from-file", fileName}
			t.Logf("args: %s", args)
			testCmd.SetArgs(args)
			t.Logf("fileName: %s", fileName)

			output := captureOutput(
				func() {
					err = testCmd.Execute()
					if !allowFail {
						if err != nil && testCheckBug63171(err) {
							bug63171 = err
							t.Skip("PAM Provider creation is not supported in Keyfactor Command version 12 and later")
						}
						assert.NoError(t, err)
					} else if err != nil && !testCheckBug63171(err) {
						bug63171 = err
					}
				},
			)

			if jErr := json.Unmarshal([]byte(output), &createResponse); jErr != nil {
				if allowFail {
					t.Logf("Error unmarshalling JSON: %v", jErr)
				} else {
					t.Errorf("failed to create a PAM provider: %v", jErr)
					t.FailNow()
				}
				return
			}

			if !allowFail {
				assert.NotEmpty(t, createResponse.(map[string]any)["Id"])
				assert.NotEmpty(t, createResponse.(map[string]any)["Name"])
				assert.Equal(t, createResponse.(map[string]any)["Name"], providerName)
				assert.NotEmpty(t, createResponse.(map[string]any)["ProviderType"])
			}
		},
	)

	if bug63171 != nil {
		return createResponse, bug63171
	}

	return createResponse, err
}

func testDeletePamProvider(t *testing.T, pID int, allowFail bool) error {
	var err error
	var output string
	t.Run(
		fmt.Sprintf("Deleting PAM provider %d", pID), func(t *testing.T) {
			testCmd := RootCmd

			testCmd.SetArgs([]string{"pam", "delete", "--id", strconv.Itoa(pID)})
			output = captureOutput(
				func() {
					err = testCmd.Execute()
					if !allowFail {
						assert.NoError(t, err)
					}
				},
			)
			if !allowFail {
				assert.Contains(t, output, fmt.Sprintf("Deleted PAM provider with ID %d", pID))
			}
		},
	)
	if err != nil && !allowFail {
		t.Log(output)
		return err
	}
	return nil
}

func testListPamProviderTypes(t *testing.T, name string, allowFail bool, allowEmpty bool) (any, error) {
	var err error
	var output string
	var pvTypes any

	testCmd := RootCmd
	testCmd.SetArgs([]string{"pam-types", "list"})
	output = captureOutput(
		func() {
			err = testCmd.Execute()
			if !allowFail {
				assert.NoError(t, err)
			}
		},
	)
	var pTypes []any
	if err = json.Unmarshal([]byte(output), &pTypes); err != nil && !allowFail {
		t.Errorf("Error unmarshalling JSON: %v", err)
		return nil, err
	}

	if !allowEmpty {
		assert.GreaterOrEqual(t, len(pTypes), 0)
	}

	if len(pTypes) > 0 {
		for _, p := range pTypes {
			providerConfig := p.(map[string]any)

			if !allowFail {
				assert.NotEmpty(t, providerConfig["Name"])
				assert.NotEmpty(t, providerConfig["Id"])

				if providerConfig["Name"] == name {
					pvTypes = p
				}
			}

			pTypeParams, ok := providerConfig["ProviderTypeParams"].([]any)
			if !ok {
				t.Logf("ProviderTypeParams is not a list of maps for %s", providerConfig["Name"])
				continue
			}

			if len(pTypeParams) > 0 {
				for _, param := range pTypeParams {
					if !allowFail {
						assert.NotEmpty(t, param.(map[string]any)["Id"])
						assert.NotEmpty(t, param.(map[string]any)["Name"])
						assert.NotEmpty(t, param.(map[string]any)["DataType"])
					}
				}
			} else {
				t.Logf("ProviderTypeParams is empty for %s (%s)", providerConfig["Name"], providerConfig["Id"])
			}
		}
	}
	return pvTypes, err
}

func testFormatPamCreateConfig(t *testing.T, inputFileName string, providerName string, isUpdate bool) (string, error) {
	pConfig, pErr := loadJSONFile(inputFileName)

	assert.NoError(t, pErr)
	if pErr != nil {
		t.Errorf("failed to load PAM provider config file '%s': %v", inputFileName, pErr)
		return "", pErr
	}

	cProviderType := pConfig["ProviderType"].(map[string]any)
	cProviderTypeName := cProviderType["Name"].(string)

	cProviderTypeParamValues := pConfig["ProviderTypeParamValues"].([]any)

	apiProviderType, pvtErr := testListPamProviderTypes(t, cProviderTypeName, false, false)

	if pvtErr != nil {
		t.Errorf("failed to find PAM provider type '%s' unable to create PAM provider: %v", cProviderTypeName, pvtErr)
		return "", pvtErr
	} else if apiProviderType == nil {
		t.Errorf("failed to find PAM provider type '%s' unable to create PAM provider: %v", cProviderTypeName, pvtErr)
		return "", pvtErr
	}

	switch apiProviderType.(type) {
	case map[string]any:
		aProviderType := apiProviderType.(map[string]any)
		cProviderType["Id"] = aProviderType["Id"]
		apiProviderTypeParams, ok := aProviderType["ProviderTypeParams"]
		if !ok || apiProviderTypeParams == nil {
			apiProviderTypeParams = aProviderType["Parameters"]
		}
		cProviderType["ProviderTypeParams"] = apiProviderTypeParams
		nameToIdMap := make(map[string]int)
		paramsFieldName := "ProviderTypeParams"
		_, ok = cProviderType[paramsFieldName]
		if ok && cProviderType[paramsFieldName] != nil {
			t.Logf("PAM definition is v10 or earlier")
			for _, cParam := range cProviderType[paramsFieldName].([]any) {
				paramId := cParam.(map[string]any)["Id"]
				paramName := cParam.(map[string]any)["Name"]
				nameToIdMap[paramName.(string)] = int(paramId.(float64))
			}
		}

		for idx, pValue := range cProviderTypeParamValues {
			pValueMap := pValue.(map[string]any)
			paramInfo := pValueMap["ProviderTypeParam"].(map[string]any)
			paramInfo["Id"] = nameToIdMap[paramInfo["Name"].(string)]
			pValueMap["ProviderTypeParam"] = paramInfo
			cProviderTypeParamValues[idx] = pValueMap
		}
	default:
		oErr := fmt.Errorf("failed to find PAM provider type '%s': unexpected type returned", cProviderTypeName)
		t.Error(oErr)
		return "", oErr
	}

	pConfig, pErr = loadJSONFile(inputFileName)
	assert.NoError(t, pErr)
	if pErr != nil {
		t.Errorf("failed to load PAM provider config file '%s': %v", inputFileName, pErr)
		return "", pErr
	}

	pConfig["ProviderType"] = cProviderType
	pConfig["ProviderTypeParamValues"] = cProviderTypeParamValues

	if providerName != "" {
		pConfig["Name"] = providerName
	}

	if isUpdate {
		t.Logf("listing PAM providers for update")
		providersList, err := testListPamProviders(t)
		assert.NoError(t, err)
		if err != nil {
			t.Fatalf("failed to list PAM providers: %v", err)
		}
		if len(providersList) > 0 {
			for _, p := range providersList {
				providerConfig := p.(map[string]any)
				if providerConfig["Name"] == providerName {
					idInt := int(providerConfig["Id"].(float64))
					pConfig["Id"] = idInt
					break
				}
			}
		} else {
			dErr := fmt.Errorf("0 PAM providers found, cannot test update")
			t.Error(dErr)
			return "", dErr
		}
	}

	updatedFileName := strings.Replace(inputFileName, "-template.json", ".json", 1)
	wErr := writeJSONFile(updatedFileName, pConfig)
	if wErr != nil {
		t.Errorf("failed to write updated PAM provider config file '%s': %v", inputFileName, wErr)
		return "", wErr
	}
	return updatedFileName, nil
}

func testCheckBug63171(err error) bool {
	if err != nil && strings.Contains(err.Error(), "not supported in Keyfactor Command version 12 and later") {
		return true
	}
	return false
}
