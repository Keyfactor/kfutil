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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Keyfactor/keyfactor-auth-client-go/auth_providers"
)

func TestMain(m *testing.M) {
	cleanups := ensureTestAuthConfigs()
	code := m.Run()
	for i := len(cleanups) - 1; i >= 0; i-- {
		cleanups[i]()
	}
	os.Exit(code)
}

func ensureTestAuthConfigs() []func() {
	if os.Getenv(auth_providers.EnvKeyfactorHostName) == "" {
		return nil
	}

	config := buildTestAuthConfig()
	if len(config.Servers) == 0 {
		return nil
	}

	type authConfigPath struct {
		path      string
		overwrite bool
	}

	var paths []authConfigPath
	homeDir, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths, authConfigPath{path: filepath.Join(homeDir, auth_providers.DefaultConfigFilePath)})
	}
	paths = append(paths, authConfigPath{
		path:      filepath.Join("$HOME", ".keyfactor", "extra_config.json"),
		overwrite: true,
	})

	var cleanups []func()
	for _, configPath := range paths {
		if configPath.path == "" {
			continue
		}
		previousConfig, readErr := auth_providers.ReadConfigFromJSON(configPath.path)
		existed := readErr == nil
		if existed && !configPath.overwrite {
			if _, ok := previousConfig.Servers[auth_providers.DefaultConfigProfile]; ok {
				continue
			}
			merged := &auth_providers.Config{Servers: map[string]auth_providers.Server{}}
			for name, server := range previousConfig.Servers {
				merged.Servers[name] = server
			}
			for name, server := range config.Servers {
				if _, ok := merged.Servers[name]; !ok || name == auth_providers.DefaultConfigProfile {
					merged.Servers[name] = server
				}
			}
			if err := auth_providers.WriteConfigToJSON(configPath.path, merged); err == nil {
				pathToRestore := configPath.path
				configToRestore := previousConfig
				cleanups = append(cleanups, func() {
					_ = auth_providers.WriteConfigToJSON(pathToRestore, configToRestore)
				})
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(configPath.path), 0700); err != nil {
			continue
		}
		if err := auth_providers.WriteConfigToJSON(configPath.path, config); err == nil {
			pathToCleanup := configPath.path
			if existed {
				configToRestore := previousConfig
				cleanups = append(cleanups, func() {
					_ = auth_providers.WriteConfigToJSON(pathToCleanup, configToRestore)
				})
			} else {
				cleanups = append(cleanups, func() {
					_ = os.Remove(pathToCleanup)
				})
			}
		}
	}
	return cleanups
}

func buildTestAuthConfig() *auth_providers.Config {
	config := &auth_providers.Config{
		Servers: map[string]auth_providers.Server{},
	}

	host := os.Getenv(auth_providers.EnvKeyfactorHostName)
	apiPath := os.Getenv(auth_providers.EnvKeyfactorAPIPath)
	if apiPath == "" {
		apiPath = auth_providers.DefaultCommandAPIPath
	}

	username := os.Getenv(auth_providers.EnvKeyfactorUsername)
	password := os.Getenv(auth_providers.EnvKeyfactorPassword)
	domain := os.Getenv(auth_providers.EnvKeyfactorDomain)
	if username != "" && password != "" {
		config.Servers[auth_providers.DefaultConfigProfile] = auth_providers.Server{
			Host:          host,
			APIPath:       apiPath,
			Username:      username,
			Password:      password,
			Domain:        domain,
			SkipTLSVerify: true,
			AuthType:      "basic",
		}
	}

	clientID := os.Getenv(auth_providers.EnvKeyfactorClientID)
	clientSecret := os.Getenv(auth_providers.EnvKeyfactorClientSecret)
	tokenURL := os.Getenv(auth_providers.EnvKeyfactorAuthTokenURL)
	if clientID != "" && clientSecret != "" && tokenURL != "" {
		oauthServer := auth_providers.Server{
			Host:          host,
			APIPath:       apiPath,
			ClientID:      clientID,
			ClientSecret:  clientSecret,
			OAuthTokenUrl: tokenURL,
			Scopes:        testAuthScopes(),
			Audience:      os.Getenv(auth_providers.EnvKeyfactorAuthAudience),
			SkipTLSVerify: true,
			AuthType:      "oauth",
		}
		config.Servers["oauth"] = oauthServer
		if _, ok := config.Servers[auth_providers.DefaultConfigProfile]; !ok {
			config.Servers[auth_providers.DefaultConfigProfile] = oauthServer
		}
	}

	return config
}

func testAuthScopes() []string {
	scopesCSV := os.Getenv(auth_providers.EnvKeyfactorAuthScopes)
	if scopesCSV == "" {
		return []string{"openid"}
	}
	var scopes []string
	for _, scope := range strings.Split(scopesCSV, ",") {
		scope = strings.TrimSpace(scope)
		if scope != "" {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}
