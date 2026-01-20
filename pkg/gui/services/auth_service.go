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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Keyfactor/keyfactor-auth-client-go/auth_providers"
	"github.com/Keyfactor/keyfactor-go-client/v3/api"
)

const (
	AuthTypeBasic       = "basic"
	AuthTypeOAuthClient = "oauth_client"
	AuthTypeOAuthToken  = "oauth_token"
)

// AuthConfig holds the authentication configuration
type AuthConfig struct {
	AuthType      string   `json:"auth_type"`
	Hostname      string   `json:"hostname"`
	Port          int      `json:"port"`
	APIPath       string   `json:"api_path"`
	Username      string   `json:"username,omitempty"`
	Password      string   `json:"password,omitempty"`
	Domain        string   `json:"domain,omitempty"`
	ClientID      string   `json:"client_id,omitempty"`
	ClientSecret  string   `json:"client_secret,omitempty"`
	TokenURL      string   `json:"token_url,omitempty"`
	Audience      string   `json:"audience,omitempty"`
	Scopes        []string `json:"scopes,omitempty"`
	AccessToken   string   `json:"access_token,omitempty"`
	SkipTLSVerify bool     `json:"skip_tls_verify"`
}

// AuthService handles authentication operations
type AuthService struct {
	mu          sync.RWMutex
	config      *AuthConfig
	client      *api.Client
	isConnected bool
	lastError   error
}

// NewAuthService creates a new auth service instance
func NewAuthService() *AuthService {
	svc := &AuthService{
		config: &AuthConfig{
			Port:    443,
			APIPath: "KeyfactorAPI",
		},
	}
	// Try to load existing config
	svc.LoadConfigFromFile("")
	return svc
}

// GetConfig returns a copy of the current auth config
func (s *AuthService) GetConfig() AuthConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return AuthConfig{Port: 443, APIPath: "KeyfactorAPI"}
	}
	return *s.config
}

// SetConfig sets the auth configuration
func (s *AuthService) SetConfig(config AuthConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = &config
	s.client = nil
	s.isConnected = false
}

// IsAuthenticated returns true if we have a valid connection
func (s *AuthService) IsAuthenticated() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isConnected && s.client != nil
}

// GetClient returns the authenticated API client
func (s *AuthService) GetClient() (*api.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil && s.isConnected {
		return s.client, nil
	}

	if s.config == nil || s.config.Hostname == "" {
		return nil, fmt.Errorf("authentication not configured")
	}

	client, err := s.createClient()
	if err != nil {
		s.lastError = err
		return nil, err
	}

	s.client = client
	s.isConnected = true
	return client, nil
}

// TestConnection tests the current authentication configuration
func (s *AuthService) TestConnection() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config == nil || s.config.Hostname == "" {
		return fmt.Errorf("authentication not configured")
	}

	client, err := s.createClient()
	if err != nil {
		s.lastError = err
		s.isConnected = false
		return err
	}

	// Test by listing store types (minimal API call)
	_, err = client.ListCertificateStoreTypes()
	if err != nil {
		s.lastError = err
		s.isConnected = false
		return fmt.Errorf("connection test failed: %w", err)
	}

	s.client = client
	s.isConnected = true
	s.lastError = nil
	return nil
}

// createClient creates a new API client from the current config
func (s *AuthService) createClient() (*api.Client, error) {
	if s.config == nil {
		return nil, fmt.Errorf("no configuration available")
	}

	// Build server config
	serverConfig := &auth_providers.Server{
		Host:          s.config.Hostname,
		Port:          s.config.Port,
		APIPath:       s.config.APIPath,
		SkipTLSVerify: s.config.SkipTLSVerify,
	}

	switch s.config.AuthType {
	case AuthTypeBasic:
		serverConfig.Username = s.config.Username
		serverConfig.Password = s.config.Password
		serverConfig.Domain = s.config.Domain

	case AuthTypeOAuthClient:
		serverConfig.ClientID = s.config.ClientID
		serverConfig.ClientSecret = s.config.ClientSecret
		serverConfig.OAuthTokenUrl = s.config.TokenURL
		serverConfig.Audience = s.config.Audience
		serverConfig.Scopes = s.config.Scopes

	case AuthTypeOAuthToken:
		serverConfig.AccessToken = s.config.AccessToken
		serverConfig.Audience = s.config.Audience
		serverConfig.Scopes = s.config.Scopes

	default:
		return nil, fmt.Errorf("unknown auth type: %s", s.config.AuthType)
	}

	// Create the client
	client, err := api.NewKeyfactorClient(serverConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client, nil
}

// LoadConfigFromFile loads configuration from a JSON file
func (s *AuthService) LoadConfigFromFile(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(homeDir, ".keyfactor", "command_config.json")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No config file is OK
		}
		return err
	}

	// Try to parse as our AuthConfig format first
	var config AuthConfig
	if err := json.Unmarshal(data, &config); err == nil && config.Hostname != "" {
		s.config = &config
		s.client = nil
		s.isConnected = false
		return nil
	}

	// Try to parse as Keyfactor command_config.json format
	var cmdConfig struct {
		Servers map[string]struct {
			Host          string   `json:"host"`
			Port          int      `json:"port"`
			Username      string   `json:"username"`
			Password      string   `json:"password"`
			Domain        string   `json:"domain"`
			APIPath       string   `json:"api_path"`
			ClientID      string   `json:"client_id"`
			ClientSecret  string   `json:"client_secret"`
			TokenURL      string   `json:"token_url"`
			Audience      string   `json:"audience"`
			Scopes        []string `json:"scopes"`
			AccessToken   string   `json:"access_token"`
			SkipTLSVerify bool     `json:"skip_tls_verify"`
		} `json:"servers"`
	}

	if err := json.Unmarshal(data, &cmdConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Use default profile
	if server, ok := cmdConfig.Servers["default"]; ok {
		s.config = &AuthConfig{
			Hostname:      server.Host,
			Port:          server.Port,
			APIPath:       server.APIPath,
			Username:      server.Username,
			Password:      server.Password,
			Domain:        server.Domain,
			ClientID:      server.ClientID,
			ClientSecret:  server.ClientSecret,
			TokenURL:      server.TokenURL,
			Audience:      server.Audience,
			Scopes:        server.Scopes,
			AccessToken:   server.AccessToken,
			SkipTLSVerify: server.SkipTLSVerify,
		}

		// Determine auth type
		if s.config.AccessToken != "" {
			s.config.AuthType = AuthTypeOAuthToken
		} else if s.config.ClientID != "" && s.config.ClientSecret != "" {
			s.config.AuthType = AuthTypeOAuthClient
		} else if s.config.Username != "" {
			s.config.AuthType = AuthTypeBasic
		}

		if s.config.Port == 0 {
			s.config.Port = 443
		}
		if s.config.APIPath == "" {
			s.config.APIPath = "KeyfactorAPI"
		}

		s.client = nil
		s.isConnected = false
	}

	return nil
}

// SaveConfigToFile saves the current configuration to a JSON file
func (s *AuthService) SaveConfigToFile(path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir := filepath.Join(homeDir, ".keyfactor")
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
		path = filepath.Join(dir, "command_config.json")
	}

	// Save in command_config.json format for CLI compatibility
	cmdConfig := map[string]interface{}{
		"servers": map[string]interface{}{
			"default": map[string]interface{}{
				"host":            s.config.Hostname,
				"port":            s.config.Port,
				"api_path":        s.config.APIPath,
				"username":        s.config.Username,
				"password":        s.config.Password,
				"domain":          s.config.Domain,
				"client_id":       s.config.ClientID,
				"client_secret":   s.config.ClientSecret,
				"token_url":       s.config.TokenURL,
				"audience":        s.config.Audience,
				"scopes":          s.config.Scopes,
				"access_token":    s.config.AccessToken,
				"skip_tls_verify": s.config.SkipTLSVerify,
			},
		},
	}

	data, err := json.MarshalIndent(cmdConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// GenerateExampleConfig generates an example config file content
func (s *AuthService) GenerateExampleConfig(authType string) string {
	var example map[string]interface{}

	switch authType {
	case AuthTypeBasic:
		example = map[string]interface{}{
			"servers": map[string]interface{}{
				"default": map[string]interface{}{
					"host":            "keyfactor.example.com",
					"port":            443,
					"api_path":        "KeyfactorAPI",
					"username":        "your_username",
					"password":        "your_password",
					"domain":          "your_domain",
					"skip_tls_verify": false,
				},
			},
		}
	case AuthTypeOAuthClient:
		example = map[string]interface{}{
			"servers": map[string]interface{}{
				"default": map[string]interface{}{
					"host":            "keyfactor.example.com",
					"port":            443,
					"api_path":        "KeyfactorAPI",
					"client_id":       "your_client_id",
					"client_secret":   "your_client_secret",
					"token_url":       "https://auth.example.com/oauth2/token",
					"audience":        "https://keyfactor.example.com/KeyfactorAPI",
					"scopes":          []string{"openid"},
					"skip_tls_verify": false,
				},
			},
		}
	case AuthTypeOAuthToken:
		example = map[string]interface{}{
			"servers": map[string]interface{}{
				"default": map[string]interface{}{
					"host":            "keyfactor.example.com",
					"port":            443,
					"api_path":        "KeyfactorAPI",
					"access_token":    "your_access_token",
					"audience":        "https://keyfactor.example.com/KeyfactorAPI",
					"scopes":          []string{"openid"},
					"skip_tls_verify": false,
				},
			},
		}
	default:
		return "{}"
	}

	data, _ := json.MarshalIndent(example, "", "  ")
	return string(data)
}

// GetEnvironmentVariables returns the relevant environment variable values
func (s *AuthService) GetEnvironmentVariables() map[string]string {
	vars := map[string]string{
		"KEYFACTOR_HOSTNAME":           os.Getenv("KEYFACTOR_HOSTNAME"),
		"KEYFACTOR_PORT":               os.Getenv("KEYFACTOR_PORT"),
		"KEYFACTOR_API_PATH":           os.Getenv("KEYFACTOR_API_PATH"),
		"KEYFACTOR_USERNAME":           os.Getenv("KEYFACTOR_USERNAME"),
		"KEYFACTOR_PASSWORD":           maskValue(os.Getenv("KEYFACTOR_PASSWORD")),
		"KEYFACTOR_DOMAIN":             os.Getenv("KEYFACTOR_DOMAIN"),
		"KEYFACTOR_AUTH_CLIENT_ID":     os.Getenv("KEYFACTOR_AUTH_CLIENT_ID"),
		"KEYFACTOR_AUTH_CLIENT_SECRET": maskValue(os.Getenv("KEYFACTOR_AUTH_CLIENT_SECRET")),
		"KEYFACTOR_AUTH_TOKEN_URL":     os.Getenv("KEYFACTOR_AUTH_TOKEN_URL"),
		"KEYFACTOR_AUTH_AUDIENCE":      os.Getenv("KEYFACTOR_AUTH_AUDIENCE"),
		"KEYFACTOR_AUTH_SCOPES":        os.Getenv("KEYFACTOR_AUTH_SCOPES"),
		"KEYFACTOR_AUTH_ACCESS_TOKEN":  maskValue(os.Getenv("KEYFACTOR_AUTH_ACCESS_TOKEN")),
		"KEYFACTOR_SKIP_VERIFY":        os.Getenv("KEYFACTOR_SKIP_VERIFY"),
	}
	return vars
}

// maskValue masks sensitive values for display
func maskValue(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + "****" + value[len(value)-2:]
}

// Disconnect clears the current connection
func (s *AuthService) Disconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client = nil
	s.isConnected = false
}
