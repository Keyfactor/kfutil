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

package gui

import (
	"kfutil/pkg/gui/services"
	"kfutil/pkg/gui/views"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	NavHome       = "Home"
	NavSettings   = "Settings"
	NavStoreTypes = "Installed Store Types"
	NavCatalog    = "Store Type Catalog"
)

// App represents the main GUI application
type App struct {
	fyneApp      fyne.App
	mainWindow   fyne.Window
	authService  *services.AuthService
	storeService *services.StoreService
	content      *fyne.Container
	currentView  string
}

// NewApp creates a new GUI application instance
func NewApp() *App {
	fyneApp := app.New()
	fyneApp.Settings().SetTheme(NewKeyfactorTheme())

	guiApp := &App{
		fyneApp:      fyneApp,
		mainWindow:   fyneApp.NewWindow("Keyfactor Utility - Store Type Manager"),
		authService:  services.NewAuthService(),
		storeService: services.NewStoreService(),
	}

	return guiApp
}

// LaunchApp is the main entry point called from the CLI
func LaunchApp() error {
	guiApp := NewApp()
	guiApp.setupUI()
	guiApp.mainWindow.ShowAndRun()
	return nil
}

// setupUI initializes the main UI layout
func (a *App) setupUI() {
	a.mainWindow.Resize(fyne.NewSize(1200, 800))
	a.mainWindow.CenterOnScreen()

	// Create navigation sidebar
	nav := a.createNavigation()

	// Create content container (starts with home view)
	a.content = container.NewMax()
	a.showView(NavHome)

	// Create main layout with navigation on left, content on right
	split := container.NewHSplit(nav, a.content)
	split.SetOffset(0.15) // 15% for navigation

	a.mainWindow.SetContent(split)
}

// createNavigation creates the sidebar navigation
func (a *App) createNavigation() fyne.CanvasObject {
	homeBtn := widget.NewButton(
		NavHome, func() {
			a.showView(NavHome)
		},
	)
	settingsBtn := widget.NewButton(
		NavSettings, func() {
			a.showView(NavSettings)
		},
	)
	storeTypesBtn := widget.NewButton(
		NavStoreTypes, func() {
			a.showView(NavStoreTypes)
		},
	)
	catalogBtn := widget.NewButton(
		NavCatalog, func() {
			a.showView(NavCatalog)
		},
	)

	// Create navigation list
	navContainer := container.NewVBox(
		widget.NewLabelWithStyle("kfutil", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		homeBtn,
		settingsBtn,
		storeTypesBtn,
		catalogBtn,
	)

	return container.NewPadded(navContainer)
}

// showView switches to the specified view
func (a *App) showView(viewName string) {
	a.currentView = viewName

	var viewContent fyne.CanvasObject

	switch viewName {
	case NavHome:
		viewContent = views.NewHomeView(a.authService, a.navigateTo)
	case NavSettings:
		viewContent = views.NewSettingsView(a.authService, a.showNotification)
	case NavStoreTypes:
		viewContent = views.NewStoreManagerView(a.authService, a.storeService, a.showStoreDetail, a.navigateTo)
	case NavCatalog:
		viewContent = views.NewStoreCatalogView(a.authService, a.storeService, a.showCatalogDetail, a.showNotification)
	default:
		viewContent = widget.NewLabel("Unknown view")
	}

	a.content.Objects = []fyne.CanvasObject{viewContent}
	a.content.Refresh()
}

// navigateTo is a callback for views to navigate to other views
func (a *App) navigateTo(viewName string) {
	a.showView(viewName)
}

// showNotification displays a notification to the user
func (a *App) showNotification(title, message string) {
	fyne.CurrentApp().SendNotification(
		&fyne.Notification{
			Title:   title,
			Content: message,
		},
	)
}

// showStoreDetail shows the detail view for an installed store type
func (a *App) showStoreDetail(storeTypeID int) {
	detail := views.NewStoreDetailView(
		a.authService, a.storeService, storeTypeID, false, a.showNotification,
		func() {
			a.showView(NavStoreTypes) // Return to store manager after closing
		},
		func() {
			a.showView(NavStoreTypes) // Refresh store manager after successful create/deploy
		},
	)
	a.content.Objects = []fyne.CanvasObject{detail}
	a.content.Refresh()
}

// showCatalogDetail shows the detail view for a catalog store type
func (a *App) showCatalogDetail(shortName string) {
	detail := views.NewCatalogDetailView(
		a.authService, a.storeService, shortName, a.showNotification,
		func() {
			a.showView(NavCatalog) // Return to catalog after closing
		},
		func() {
			a.showView(NavStoreTypes) // Refresh store manager after successful create/deploy
		},
	)
	a.content.Objects = []fyne.CanvasObject{detail}
	a.content.Refresh()
}
