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

package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"kfutil/pkg/gui/services"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// NewSettingsView creates the settings view for authentication configuration
func NewSettingsView(authService *services.AuthService, showNotification func(string, string)) fyne.CanvasObject {
	config := authService.GetConfig()

	// Track last used config file path
	var lastConfigPath string
	homeDir, _ := os.UserHomeDir()
	defaultConfigPath := filepath.Join(homeDir, ".keyfactor", "command_config.json")

	// Profile selector
	profileSelect := widget.NewSelect([]string{}, nil)
	currentProfileLabel := widget.NewLabel("Current Profile: " + authService.GetCurrentProfile())

	// Load available profiles
	refreshProfiles := func() {
		profiles, err := authService.ListProfiles("")
		if err != nil || len(profiles) == 0 {
			profileSelect.Options = []string{"default"}
		} else {
			profileSelect.Options = profiles
		}
		profileSelect.SetSelected(authService.GetCurrentProfile())
		currentProfileLabel.SetText("Current Profile: " + authService.GetCurrentProfile())
	}

	// Auth type selector
	authTypeSelect := widget.NewSelect([]string{"Basic Auth", "OAuth2 Client Credentials", "OAuth2 Access Token"}, nil)

	// Map config auth type to display value
	switch config.AuthType {
	case services.AuthTypeBasic:
		authTypeSelect.SetSelected("Basic Auth")
	case services.AuthTypeOAuthClient:
		authTypeSelect.SetSelected("OAuth2 Client Credentials")
	case services.AuthTypeOAuthToken:
		authTypeSelect.SetSelected("OAuth2 Access Token")
	default:
		authTypeSelect.SetSelected("Basic Auth")
	}

	// Common fields
	hostnameEntry := widget.NewEntry()
	hostnameEntry.SetPlaceHolder("keyfactor.example.com")
	hostnameEntry.SetText(config.Hostname)

	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("443")
	if config.Port > 0 {
		portEntry.SetText(strconv.Itoa(config.Port))
	} else {
		portEntry.SetText("443")
	}

	apiPathEntry := widget.NewEntry()
	apiPathEntry.SetPlaceHolder("KeyfactorAPI")
	if config.APIPath != "" {
		apiPathEntry.SetText(config.APIPath)
	} else {
		apiPathEntry.SetText("KeyfactorAPI")
	}

	skipTLSCheck := widget.NewCheck("Skip TLS Verification", nil)
	skipTLSCheck.SetChecked(config.SkipTLSVerify)

	// Basic auth fields
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("username")
	usernameEntry.SetText(config.Username)

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("password")
	passwordEntry.SetText(config.Password)

	domainEntry := widget.NewEntry()
	domainEntry.SetPlaceHolder("DOMAIN")
	domainEntry.SetText(config.Domain)

	// OAuth client credentials fields
	clientIDEntry := widget.NewEntry()
	clientIDEntry.SetPlaceHolder("client_id")
	clientIDEntry.SetText(config.ClientID)

	clientSecretEntry := widget.NewPasswordEntry()
	clientSecretEntry.SetPlaceHolder("client_secret")
	clientSecretEntry.SetText(config.ClientSecret)

	tokenURLEntry := widget.NewEntry()
	tokenURLEntry.SetPlaceHolder("https://auth.example.com/oauth2/token")
	tokenURLEntry.SetText(config.TokenURL)

	// OAuth audience and scopes (optional, shared by both OAuth types)
	audienceEntry := widget.NewEntry()
	audienceEntry.SetPlaceHolder("https://keyfactor.example.com/KeyfactorAPI (optional)")
	audienceEntry.SetText(config.Audience)

	scopesEntry := widget.NewEntry()
	scopesEntry.SetPlaceHolder("openid,profile (comma-separated, optional)")
	if len(config.Scopes) > 0 {
		scopesEntry.SetText(strings.Join(config.Scopes, ","))
	}

	// OAuth access token fields
	accessTokenEntry := widget.NewPasswordEntry()
	accessTokenEntry.SetPlaceHolder("access_token")
	accessTokenEntry.SetText(config.AccessToken)

	// Create field containers
	basicAuthFields := container.NewVBox(
		widget.NewLabel("Username"),
		usernameEntry,
		widget.NewLabel("Password"),
		passwordEntry,
		widget.NewLabel("Domain (optional)"),
		domainEntry,
	)

	oauthClientFields := container.NewVBox(
		widget.NewLabel("Client ID"),
		clientIDEntry,
		widget.NewLabel("Client Secret"),
		clientSecretEntry,
		widget.NewLabel("Token URL"),
		tokenURLEntry,
		widget.NewLabel("Audience (optional)"),
		audienceEntry,
		widget.NewLabel("Scopes (optional, comma-separated)"),
		scopesEntry,
	)

	oauthTokenFields := container.NewVBox(
		widget.NewLabel("Access Token"),
		accessTokenEntry,
		widget.NewLabel("Audience (optional)"),
		audienceEntry,
		widget.NewLabel("Scopes (optional, comma-separated)"),
		scopesEntry,
	)

	// Dynamic fields container
	dynamicFields := container.NewMax()

	// Function to clear auth-specific fields
	clearAuthFields := func() {
		usernameEntry.SetText("")
		passwordEntry.SetText("")
		domainEntry.SetText("")
		clientIDEntry.SetText("")
		clientSecretEntry.SetText("")
		tokenURLEntry.SetText("")
		audienceEntry.SetText("")
		scopesEntry.SetText("")
		accessTokenEntry.SetText("")
	}

	// Track current auth type to detect changes
	var currentAuthType string

	updateDynamicFields := func(authType string) {
		// Clear fields when auth type changes
		if currentAuthType != "" && currentAuthType != authType {
			clearAuthFields()
		}
		currentAuthType = authType

		dynamicFields.Objects = nil
		switch authType {
		case "Basic Auth":
			dynamicFields.Objects = []fyne.CanvasObject{basicAuthFields}
		case "OAuth2 Client Credentials":
			dynamicFields.Objects = []fyne.CanvasObject{oauthClientFields}
		case "OAuth2 Access Token":
			dynamicFields.Objects = []fyne.CanvasObject{oauthTokenFields}
		}
		dynamicFields.Refresh()
	}

	authTypeSelect.OnChanged = updateDynamicFields
	currentAuthType = authTypeSelect.Selected // Initialize without clearing
	updateDynamicFields(authTypeSelect.Selected)

	// Status label
	statusLabel := widget.NewLabel("")

	// Build config from form
	buildConfig := func() services.AuthConfig {
		port, _ := strconv.Atoi(portEntry.Text)
		if port == 0 {
			port = 443
		}

		cfg := services.AuthConfig{
			Hostname:      hostnameEntry.Text,
			Port:          port,
			APIPath:       apiPathEntry.Text,
			SkipTLSVerify: skipTLSCheck.Checked,
		}

		switch authTypeSelect.Selected {
		case "Basic Auth":
			cfg.AuthType = services.AuthTypeBasic
			cfg.Username = usernameEntry.Text
			cfg.Password = passwordEntry.Text
			cfg.Domain = domainEntry.Text
		case "OAuth2 Client Credentials":
			cfg.AuthType = services.AuthTypeOAuthClient
			cfg.ClientID = clientIDEntry.Text
			cfg.ClientSecret = clientSecretEntry.Text
			cfg.TokenURL = tokenURLEntry.Text
			cfg.Audience = audienceEntry.Text
			if scopesEntry.Text != "" {
				cfg.Scopes = strings.Split(scopesEntry.Text, ",")
				// Trim whitespace from each scope
				for i, scope := range cfg.Scopes {
					cfg.Scopes[i] = strings.TrimSpace(scope)
				}
			}
		case "OAuth2 Access Token":
			cfg.AuthType = services.AuthTypeOAuthToken
			cfg.AccessToken = accessTokenEntry.Text
			cfg.Audience = audienceEntry.Text
			if scopesEntry.Text != "" {
				cfg.Scopes = strings.Split(scopesEntry.Text, ",")
				// Trim whitespace from each scope
				for i, scope := range cfg.Scopes {
					cfg.Scopes[i] = strings.TrimSpace(scope)
				}
			}
		}

		return cfg
	}

	// Test connection button with cancel support
	var testMutex sync.Mutex
	var testCancelled bool

	testBtn := widget.NewButton("Test Connection", nil)
	cancelBtn := widget.NewButton("Cancel", nil)
	cancelBtn.Disable()

	testBtn.OnTapped = func() {
		cfg := buildConfig()
		authService.SetConfig(cfg)
		statusLabel.SetText("Testing connection...")

		testMutex.Lock()
		testCancelled = false
		testMutex.Unlock()

		testBtn.Disable()
		cancelBtn.Enable()

		go func() {
			err := authService.TestConnection()

			testMutex.Lock()
			cancelled := testCancelled
			testMutex.Unlock()

			if cancelled {
				statusLabel.SetText("Connection test cancelled")
				showNotification("Cancelled", "Connection test was cancelled")
			} else if err != nil {
				statusLabel.SetText("Connection failed: " + err.Error())
				showNotification("Connection Failed", err.Error())
			} else {
				statusLabel.SetText("Connection successful!")
				showNotification("Success", "Connection test successful")
			}

			testBtn.Enable()
			cancelBtn.Disable()
		}()
	}

	cancelBtn.OnTapped = func() {
		testMutex.Lock()
		testCancelled = true
		testMutex.Unlock()
		authService.Disconnect()
		cancelBtn.Disable()
	}

	// Save button
	saveBtn := widget.NewButton(
		"Save Configuration", func() {
			cfg := buildConfig()
			authService.SetConfig(cfg)

			err := authService.SaveConfigToFile("")
			if err != nil {
				statusLabel.SetText("Failed to save: " + err.Error())
				showNotification("Save Failed", err.Error())
			} else {
				statusLabel.SetText("Configuration saved successfully")
				showNotification("Success", "Configuration saved")
			}
		},
	)

	// Function to update form from loaded config
	updateFormFromConfig := func(cfg services.AuthConfig) {
		hostnameEntry.SetText(cfg.Hostname)
		if cfg.Port > 0 {
			portEntry.SetText(strconv.Itoa(cfg.Port))
		}
		apiPathEntry.SetText(cfg.APIPath)
		skipTLSCheck.SetChecked(cfg.SkipTLSVerify)

		// Set the auth type first (this triggers clearing)
		// So temporarily set currentAuthType to the same to prevent clearing
		switch cfg.AuthType {
		case services.AuthTypeBasic:
			currentAuthType = "Basic Auth"
			authTypeSelect.SetSelected("Basic Auth")
		case services.AuthTypeOAuthClient:
			currentAuthType = "OAuth2 Client Credentials"
			authTypeSelect.SetSelected("OAuth2 Client Credentials")
		case services.AuthTypeOAuthToken:
			currentAuthType = "OAuth2 Access Token"
			authTypeSelect.SetSelected("OAuth2 Access Token")
		}

		// Then set the field values
		usernameEntry.SetText(cfg.Username)
		passwordEntry.SetText(cfg.Password)
		domainEntry.SetText(cfg.Domain)
		clientIDEntry.SetText(cfg.ClientID)
		clientSecretEntry.SetText(cfg.ClientSecret)
		tokenURLEntry.SetText(cfg.TokenURL)
		audienceEntry.SetText(cfg.Audience)
		if len(cfg.Scopes) > 0 {
			scopesEntry.SetText(strings.Join(cfg.Scopes, ","))
		} else {
			scopesEntry.SetText("")
		}
		accessTokenEntry.SetText(cfg.AccessToken)
	}

	// Profile selection handler
	profileSelect.OnChanged = func(selected string) {
		if selected == "" {
			return
		}
		err := authService.LoadProfile("", selected)
		if err != nil {
			statusLabel.SetText("Failed to load profile: " + err.Error())
			return
		}
		cfg := authService.GetConfig()
		updateFormFromConfig(cfg)
		currentProfileLabel.SetText("Current Profile: " + selected)
		statusLabel.SetText("Loaded profile: " + selected)
	}

	// Initialize profiles list
	refreshProfiles()

	// New profile button
	newProfileBtn := widget.NewButton(
		"New Profile", func() {
			// Create entry for new profile name
			nameEntry := widget.NewEntry()
			nameEntry.SetPlaceHolder("Enter profile name")

			dialog.ShowForm(
				"Create New Profile", "Create", "Cancel",
				[]*widget.FormItem{
					widget.NewFormItem("Profile Name", nameEntry),
				},
				func(confirmed bool) {
					if !confirmed || nameEntry.Text == "" {
						return
					}
					profileName := strings.TrimSpace(nameEntry.Text)
					if profileName == "" {
						return
					}

					// Check if profile already exists
					profiles, _ := authService.ListProfiles("")
					for _, p := range profiles {
						if p == profileName {
							dialog.ShowError(
								fmt.Errorf("profile '%s' already exists", profileName),
								fyne.CurrentApp().Driver().AllWindows()[0],
							)
							return
						}
					}

					// Set as current profile and clear form for new config
					authService.SetCurrentProfile(profileName)
					authService.SetConfig(
						services.AuthConfig{
							Port:    443,
							APIPath: "KeyfactorAPI",
						},
					)

					// Clear form
					hostnameEntry.SetText("")
					portEntry.SetText("443")
					apiPathEntry.SetText("KeyfactorAPI")
					skipTLSCheck.SetChecked(false)
					clearAuthFields()
					authTypeSelect.SetSelected("Basic Auth")

					// Refresh profile list and select new profile
					refreshProfiles()
					profileSelect.SetSelected(profileName)
					currentProfileLabel.SetText("Current Profile: " + profileName)
					statusLabel.SetText("New profile created: " + profileName + " (not saved yet)")
				},
				fyne.CurrentApp().Driver().AllWindows()[0],
			)
		},
	)

	// Delete profile button
	deleteProfileBtn := widget.NewButton(
		"Delete Profile", func() {
			currentProfile := authService.GetCurrentProfile()
			if currentProfile == "" {
				return
			}

			dialog.ShowConfirm(
				"Delete Profile",
				fmt.Sprintf("Are you sure you want to delete the profile '%s'?", currentProfile),
				func(confirmed bool) {
					if !confirmed {
						return
					}

					err := authService.DeleteProfile("", currentProfile)
					if err != nil {
						dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
						return
					}

					// Reload config to get another profile
					authService.LoadConfigFromFile("")
					cfg := authService.GetConfig()
					updateFormFromConfig(cfg)
					refreshProfiles()
					statusLabel.SetText("Profile deleted: " + currentProfile)
					showNotification("Deleted", "Profile '"+currentProfile+"' deleted")
				},
				fyne.CurrentApp().Driver().AllWindows()[0],
			)
		},
	)

	// Load from file button with file picker
	loadBtn := widget.NewButton(
		"Load from File", func() {
			fileDialog := dialog.NewFileOpen(
				func(reader fyne.URIReadCloser, err error) {
					if err != nil || reader == nil {
						return
					}
					defer reader.Close()

					filePath := reader.URI().Path()
					lastConfigPath = filePath

					err = authService.LoadConfigFromFile(filePath)
					if err != nil {
						statusLabel.SetText("Failed to load: " + err.Error())
						return
					}

					cfg := authService.GetConfig()
					updateFormFromConfig(cfg)
					refreshProfiles()
					statusLabel.SetText("Configuration loaded from: " + filePath)
				}, fyne.CurrentApp().Driver().AllWindows()[0],
			)

			// Set default location
			if lastConfigPath != "" {
				dir := filepath.Dir(lastConfigPath)
				if uri, err := storage.ListerForURI(storage.NewFileURI(dir)); err == nil {
					fileDialog.SetLocation(uri)
				}
			} else if _, err := os.Stat(defaultConfigPath); err == nil {
				dir := filepath.Dir(defaultConfigPath)
				if uri, err := storage.ListerForURI(storage.NewFileURI(dir)); err == nil {
					fileDialog.SetLocation(uri)
				}
			}

			fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
			fileDialog.Show()
		},
	)

	// Load default config button (quick access)
	loadDefaultBtn := widget.NewButton(
		"Load Default Config", func() {
			err := authService.LoadConfigFromFile("")
			if err != nil {
				statusLabel.SetText("Failed to load default config: " + err.Error())
			} else {
				cfg := authService.GetConfig()
				updateFormFromConfig(cfg)
				refreshProfiles()
				statusLabel.SetText("Configuration loaded from default location")
			}
		},
	)

	// Generate example config button with save and copy options
	generateBtn := widget.NewButton(
		"Generate Example Config", func() {
			var authType string
			switch authTypeSelect.Selected {
			case "Basic Auth":
				authType = services.AuthTypeBasic
			case "OAuth2 Client Credentials":
				authType = services.AuthTypeOAuthClient
			case "OAuth2 Access Token":
				authType = services.AuthTypeOAuthToken
			}

			example := authService.GenerateExampleConfig(authType)

			// Show in a dialog with save and copy options
			textWidget := widget.NewMultiLineEntry()
			textWidget.SetText(example)
			textWidget.Wrapping = fyne.TextWrapWord

			// Copy to clipboard button
			copyBtn := widget.NewButton(
				"Copy to Clipboard", func() {
					fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(example)
					dialog.ShowInformation(
						"Copied",
						"Configuration copied to clipboard.",
						fyne.CurrentApp().Driver().AllWindows()[0],
					)
				},
			)

			// Save to file button
			saveExampleBtn := widget.NewButton(
				"Save to File", func() {
					saveDialog := dialog.NewFileSave(
						func(writer fyne.URIWriteCloser, err error) {
							if err != nil || writer == nil {
								return
							}
							defer writer.Close()
							writer.Write([]byte(example))
							dialog.ShowInformation(
								"Saved",
								"Configuration saved successfully.",
								fyne.CurrentApp().Driver().AllWindows()[0],
							)
						}, fyne.CurrentApp().Driver().AllWindows()[0],
					)
					saveDialog.SetFileName("command_config.json")
					saveDialog.Show()
				},
			)

			buttonRow := container.NewHBox(copyBtn, saveExampleBtn)
			content := container.NewBorder(nil, buttonRow, nil, nil, container.NewScroll(textWidget))

			d := dialog.NewCustom("Example Configuration", "Close", content, fyne.CurrentApp().Driver().AllWindows()[0])
			d.Resize(fyne.NewSize(500, 400))
			d.Show()
		},
	)

	// Show environment variables button - secrets are masked by auth_service
	envBtn := widget.NewButton(
		"Show Environment Variables", func() {
			envVars := authService.GetEnvironmentVariables()

			var content string
			for key, value := range envVars {
				if value == "" {
					value = "(not set)"
				} else if key == "KEYFACTOR_PASSWORD" || key == "KEYFACTOR_AUTH_CLIENT_SECRET" || key == "KEYFACTOR_AUTH_ACCESS_TOKEN" {
					// Show that it's set but masked (the auth service already masks these)
					if value != "" && value != "(not set)" {
						value = value + " (from env var)"
					}
				}
				content += key + ": " + value + "\n"
			}

			textWidget := widget.NewMultiLineEntry()
			textWidget.SetText(content)
			textWidget.Wrapping = fyne.TextWrapWord
			textWidget.Disable() // Read-only

			d := dialog.NewCustom(
				"Environment Variables", "Close",
				container.NewScroll(textWidget), fyne.CurrentApp().Driver().AllWindows()[0],
			)
			d.Resize(fyne.NewSize(500, 400))
			d.Show()
		},
	)

	// Main form layout
	form := container.NewVBox(
		widget.NewLabelWithStyle("Authentication Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),

		// Profile section
		widget.NewLabelWithStyle("Profile", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		currentProfileLabel,
		container.NewBorder(nil, nil, widget.NewLabel("Select Profile:"), nil, profileSelect),
		container.NewGridWithColumns(2, newProfileBtn, deleteProfileBtn),

		widget.NewSeparator(),

		widget.NewLabel("Authentication Type"),
		authTypeSelect,

		widget.NewSeparator(),

		widget.NewLabel("Hostname"),
		hostnameEntry,

		container.NewGridWithColumns(
			2,
			container.NewVBox(widget.NewLabel("Port"), portEntry),
			container.NewVBox(widget.NewLabel("API Path"), apiPathEntry),
		),

		skipTLSCheck,

		widget.NewSeparator(),

		dynamicFields,

		widget.NewSeparator(),

		container.NewGridWithColumns(
			3,
			testBtn,
			cancelBtn,
			saveBtn,
		),

		container.NewGridWithColumns(
			2,
			loadBtn,
			loadDefaultBtn,
		),

		container.NewGridWithColumns(
			2,
			generateBtn,
			envBtn,
		),

		widget.NewSeparator(),

		statusLabel,
	)

	return container.NewPadded(container.NewScroll(form))
}
