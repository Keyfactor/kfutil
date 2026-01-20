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

package services

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Keyfactor/keyfactor-go-client/v3/api"
)

//go:embed store_types.json
var embeddedStoreTypesJSON []byte

// StoreTypeInfo represents a store type for display
type StoreTypeInfo struct {
	ID          int    `json:"Id"`
	Name        string `json:"Name"`
	ShortName   string `json:"ShortName"`
	Description string `json:"Description,omitempty"`
	Capability  string `json:"Capability"`
	Version     string `json:"Version,omitempty"` // Not currently in API
}

// StoreService handles store type operations
type StoreService struct {
	catalogCache map[string]map[string]interface{}
}

// NewStoreService creates a new store service instance
func NewStoreService() *StoreService {
	return &StoreService{}
}

// ListInstalledStoreTypes retrieves all installed store types from the API
func (s *StoreService) ListInstalledStoreTypes(authService *AuthService) ([]StoreTypeInfo, error) {
	client, err := authService.GetClient()
	if err != nil {
		return nil, err
	}

	storeTypes, err := client.ListCertificateStoreTypes()
	if err != nil {
		return nil, fmt.Errorf("failed to list store types: %w", err)
	}

	var result []StoreTypeInfo
	if storeTypes != nil {
		for _, st := range *storeTypes {
			info := StoreTypeInfo{
				ID:         st.StoreType,
				Name:       st.Name,
				ShortName:  st.ShortName,
				Capability: st.Capability,
				Version:    "N/A",
			}
			result = append(result, info)
		}
	}

	return result, nil
}

// GetStoreType retrieves a specific store type by ID
func (s *StoreService) GetStoreType(authService *AuthService, id int) (*api.CertificateStoreType, error) {
	client, err := authService.GetClient()
	if err != nil {
		return nil, err
	}

	storeType, err := client.GetCertificateStoreType(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get store type: %w", err)
	}

	return storeType, nil
}

// GetStoreTypeByName retrieves a specific store type by name
func (s *StoreService) GetStoreTypeByName(authService *AuthService, name string) (*api.CertificateStoreType, error) {
	client, err := authService.GetClient()
	if err != nil {
		return nil, err
	}

	storeType, err := client.GetCertificateStoreType(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get store type: %w", err)
	}

	return storeType, nil
}

// CreateStoreType creates a new store type
func (s *StoreService) CreateStoreType(
	authService *AuthService,
	storeType *api.CertificateStoreType,
) (*api.CertificateStoreType, error) {
	client, err := authService.GetClient()
	if err != nil {
		return nil, err
	}

	result, err := client.CreateStoreType(storeType)
	if err != nil {
		return nil, fmt.Errorf("failed to create store type: %w", err)
	}

	return result, nil
}

// UpdateStoreType updates an existing store type
func (s *StoreService) UpdateStoreType(
	authService *AuthService,
	storeType *api.CertificateStoreType,
) (*api.CertificateStoreType, error) {
	client, err := authService.GetClient()
	if err != nil {
		return nil, err
	}

	result, err := client.UpdateStoreType(storeType)
	if err != nil {
		return nil, fmt.Errorf("failed to update store type: %w", err)
	}

	return result, nil
}

// DeleteStoreType deletes a store type by ID
func (s *StoreService) DeleteStoreType(authService *AuthService, id int) error {
	client, err := authService.GetClient()
	if err != nil {
		return err
	}

	_, err = client.DeleteCertificateStoreType(id)
	if err != nil {
		return fmt.Errorf("failed to delete store type: %w", err)
	}

	return nil
}

// LoadCatalog loads the store types catalog from embedded JSON
func (s *StoreService) LoadCatalog() ([]map[string]interface{}, error) {
	var catalog []map[string]interface{}
	if err := json.Unmarshal(embeddedStoreTypesJSON, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse catalog: %w", err)
	}

	// Cache by ShortName for quick lookup
	s.catalogCache = make(map[string]map[string]interface{})
	for _, item := range catalog {
		if shortName, ok := item["ShortName"].(string); ok {
			s.catalogCache[shortName] = item
		}
	}

	return catalog, nil
}

// GetCatalogItem returns a specific item from the catalog by ShortName
func (s *StoreService) GetCatalogItem(shortName string) (map[string]interface{}, error) {
	if s.catalogCache == nil {
		_, err := s.LoadCatalog()
		if err != nil {
			return nil, err
		}
	}

	item, ok := s.catalogCache[shortName]
	if !ok {
		return nil, fmt.Errorf("store type '%s' not found in catalog", shortName)
	}

	return item, nil
}

// CatalogItemToStoreType converts a catalog item to a CertificateStoreType
func (s *StoreService) CatalogItemToStoreType(item map[string]interface{}) (*api.CertificateStoreType, error) {
	data, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	var storeType api.CertificateStoreType
	if err := json.Unmarshal(data, &storeType); err != nil {
		return nil, fmt.Errorf("failed to convert catalog item: %w", err)
	}

	return &storeType, nil
}

// CheckDuplicateShortName checks if a ShortName already exists in installed store types
func (s *StoreService) CheckDuplicateShortName(authService *AuthService, shortName string) (bool, error) {
	if !authService.IsAuthenticated() {
		return false, nil // Can't check, assume no duplicate
	}

	installed, err := s.ListInstalledStoreTypes(authService)
	if err != nil {
		return false, nil // Can't check, assume no duplicate
	}

	for _, st := range installed {
		if strings.EqualFold(st.ShortName, shortName) {
			return true, nil
		}
	}

	return false, nil
}

// DeployFromCatalog deploys a store type from the catalog
func (s *StoreService) DeployFromCatalog(
	authService *AuthService,
	shortName string,
	overrideShortName string,
) (*api.CertificateStoreType, error) {
	item, err := s.GetCatalogItem(shortName)
	if err != nil {
		return nil, err
	}

	// Override ShortName if provided
	if overrideShortName != "" {
		item["ShortName"] = overrideShortName
	}

	storeType, err := s.CatalogItemToStoreType(item)
	if err != nil {
		return nil, err
	}

	// Check for duplicate
	finalShortName := storeType.ShortName
	if overrideShortName != "" {
		finalShortName = overrideShortName
	}

	isDup, _ := s.CheckDuplicateShortName(authService, finalShortName)
	if isDup {
		return nil, fmt.Errorf("store type with ShortName '%s' already exists", finalShortName)
	}

	return s.CreateStoreType(authService, storeType)
}

// ExportStoreTypesToJSON exports store types to a JSON file
func (s *StoreService) ExportStoreTypesToJSON(storeTypes []*api.CertificateStoreType, filePath string) error {
	data, err := json.MarshalIndent(storeTypes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal store types: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// ImportStoreTypesFromJSON imports store types from a JSON file
func (s *StoreService) ImportStoreTypesFromJSON(filePath string) ([]*api.CertificateStoreType, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var storeTypes []*api.CertificateStoreType
	if err := json.Unmarshal(data, &storeTypes); err != nil {
		// Try single object
		var single api.CertificateStoreType
		if err2 := json.Unmarshal(data, &single); err2 != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
		storeTypes = []*api.CertificateStoreType{&single}
	}

	return storeTypes, nil
}

// GetIconPath returns the path to the icon for a store type
func (s *StoreService) GetIconPath(shortName string) string {
	// Check in icons directory
	iconPaths := []string{
		filepath.Join("icons", fmt.Sprintf("storetype_%s.png", shortName)),
		filepath.Join("icons", fmt.Sprintf("%s.png", shortName)),
	}

	for _, path := range iconPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "" // No icon found
}

// FilterStoreTypes filters store types by search query
func (s *StoreService) FilterStoreTypes(storeTypes []StoreTypeInfo, query string) []StoreTypeInfo {
	if query == "" {
		return storeTypes
	}

	query = strings.ToLower(query)
	var filtered []StoreTypeInfo

	for _, st := range storeTypes {
		if strings.Contains(strings.ToLower(st.Name), query) ||
			strings.Contains(strings.ToLower(st.ShortName), query) ||
			strings.Contains(strings.ToLower(st.Capability), query) ||
			strings.Contains(fmt.Sprintf("%d", st.ID), query) {
			filtered = append(filtered, st)
		}
	}

	return filtered
}

// FilterCatalog filters catalog items by search query
func (s *StoreService) FilterCatalog(catalog []map[string]interface{}, query string) []map[string]interface{} {
	if query == "" {
		return catalog
	}

	query = strings.ToLower(query)
	var filtered []map[string]interface{}

	for _, item := range catalog {
		name, _ := item["Name"].(string)
		shortName, _ := item["ShortName"].(string)
		capability, _ := item["Capability"].(string)

		if strings.Contains(strings.ToLower(name), query) ||
			strings.Contains(strings.ToLower(shortName), query) ||
			strings.Contains(strings.ToLower(capability), query) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}
