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
	"kfutil/pkg/gui/services"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NewHomeView creates the home view
func NewHomeView(authService *services.AuthService, navigateTo func(string)) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Welcome to kfutil", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabelWithStyle(
		"Keyfactor Command Utility - Store Type Manager",
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)

	// Connection status
	var statusText string
	var statusStyle fyne.TextStyle
	if authService.IsAuthenticated() {
		config := authService.GetConfig()
		statusText = "Connected to: " + config.Hostname
		statusStyle = fyne.TextStyle{}
	} else {
		statusText = "Not connected - Please configure authentication in Settings"
		statusStyle = fyne.TextStyle{Italic: true}
	}
	status := widget.NewLabelWithStyle(statusText, fyne.TextAlignCenter, statusStyle)

	// Quick action buttons
	settingsBtn := widget.NewButton(
		"Configure Authentication", func() {
			navigateTo("Settings")
		},
	)

	storeTypesBtn := widget.NewButton(
		"Manage Store Types", func() {
			navigateTo("Store Types")
		},
	)

	catalogBtn := widget.NewButton(
		"Browse Catalog", func() {
			navigateTo("Catalog")
		},
	)

	// Feature descriptions
	features := widget.NewRichTextFromMarkdown(
		`
## Features

- **Settings**: Configure authentication to connect to Keyfactor Command
- **Store Types**: View and manage installed certificate store types
- **Catalog**: Browse and deploy store types from the internal catalog

## Getting Started

1. Go to **Settings** to configure your authentication
2. Test your connection to Keyfactor Command
3. Use **Store Types** to view installed store types
4. Use **Catalog** to deploy new store types
`,
	)

	content := container.NewVBox(
		widget.NewSeparator(),
		title,
		subtitle,
		widget.NewSeparator(),
		status,
		widget.NewSeparator(),
		container.NewHBox(
			settingsBtn,
			storeTypesBtn,
			catalogBtn,
		),
		widget.NewSeparator(),
		features,
	)

	return container.NewPadded(container.NewScroll(content))
}
