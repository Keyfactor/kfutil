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
	"encoding/json"
	"fmt"
	"sync"

	"kfutil/pkg/gui/services"
	"kfutil/pkg/gui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// NewStoreCatalogView creates the store type catalog view
func NewStoreCatalogView(
	authService *services.AuthService,
	storeService *services.StoreService,
	showDetail func(string),
	showNotification func(string, string),
) fyne.CanvasObject {
	// Search entry
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search by Name, ShortName, or Capability...")

	// Status label
	statusLabel := widget.NewLabel("Loading catalog...")

	// View mode state - use shared state for persistence
	// (SharedViewState.IsGridView is used directly)

	// Load catalog
	catalog, err := storeService.LoadCatalog()
	if err != nil {
		return container.NewCenter(widget.NewLabel("Failed to load catalog: " + err.Error()))
	}

	filteredCatalog := catalog

	// Multi-select: track selected indices using a map
	// Use mutex to protect concurrent access from Fyne's renderer
	var selMutex sync.Mutex
	selectedIndices := make(map[int]bool)

	// Helper to get selected count
	getSelectedCount := func() int {
		selMutex.Lock()
		defer selMutex.Unlock()
		return len(selectedIndices)
	}

	// Helper to get selected short names
	getSelectedShortNames := func() []string {
		selMutex.Lock()
		defer selMutex.Unlock()
		var names []string
		for idx := range selectedIndices {
			if idx >= 0 && idx < len(filteredCatalog) {
				if sn, ok := filteredCatalog[idx]["ShortName"].(string); ok {
					names = append(names, sn)
				}
			}
		}
		return names
	}

	// Container for the current view
	viewContainer := container.NewMax()

	// Selection label
	selectionLabel := widget.NewLabel("")
	updateSelectionLabel := func() {
		count := getSelectedCount()
		if count == 0 {
			selectionLabel.SetText("")
		} else if count == 1 {
			selectionLabel.SetText("1 item selected")
		} else {
			selectionLabel.SetText(fmt.Sprintf("%d items selected", count))
		}
	}

	// Grid view creator with checkboxes
	var createGridView func() fyne.CanvasObject
	var updateView func()

	createGridView = func() fyne.CanvasObject {
		if len(filteredCatalog) == 0 {
			return container.NewCenter(widget.NewLabel("No store types found"))
		}

		var cardObjects []fyne.CanvasObject
		for i, item := range filteredCatalog {
			idx := i
			name, _ := item["Name"].(string)
			shortName, _ := item["ShortName"].(string)
			description, _ := item["Description"].(string)
			capability, _ := item["Capability"].(string)

			card := widgets.NewStoreCard(0, name, shortName, description, capability)

			// Checkbox for selection
			check := widget.NewCheck(
				"", func(checked bool) {
					selMutex.Lock()
					if checked {
						selectedIndices[idx] = true
					} else {
						delete(selectedIndices, idx)
					}
					selMutex.Unlock()
					updateSelectionLabel()
				},
			)
			selMutex.Lock()
			isSelected := selectedIndices[idx]
			selMutex.Unlock()
			check.SetChecked(isSelected)

			card.OnDoubleTapped = func() {
				sn, _ := filteredCatalog[idx]["ShortName"].(string)
				showDetail(sn)
			}

			// Wrap card with checkbox
			cardWithCheck := container.NewBorder(nil, nil, check, nil, card)
			cardObjects = append(cardObjects, container.NewPadded(cardWithCheck))
		}

		grid := container.NewGridWithColumns(3, cardObjects...)
		return container.NewScroll(grid)
	}

	// Table view creator - creates a fresh table each time to avoid race conditions
	// with Fyne's internal table renderer state
	var createTableView func() fyne.CanvasObject
	createTableView = func() fyne.CanvasObject {
		if len(filteredCatalog) == 0 {
			return container.NewCenter(widget.NewLabel("No store types found"))
		}

		table := widget.NewTable(
			func() (int, int) {
				return len(filteredCatalog) + 1, 5 // +1 for header, 5 columns (checkbox + 4 data)
			},
			func() fyne.CanvasObject {
				return container.NewMax(widget.NewCheck("", nil))
			},
			func(id widget.TableCellID, cell fyne.CanvasObject) {
				cont := cell.(*fyne.Container)
				cont.Objects = nil

				if id.Row == 0 {
					// Header row
					headers := []string{"", "Name", "ShortName", "Capability", "Version"}
					label := widget.NewLabel(headers[id.Col])
					label.TextStyle = fyne.TextStyle{Bold: true}
					cont.Add(label)
				} else {
					rowIdx := id.Row - 1
					if id.Col == 0 {
						// Checkbox column
						check := widget.NewCheck(
							"", func(checked bool) {
								selMutex.Lock()
								if checked {
									selectedIndices[rowIdx] = true
								} else {
									delete(selectedIndices, rowIdx)
								}
								selMutex.Unlock()
								updateSelectionLabel()
							},
						)
						if rowIdx < len(filteredCatalog) {
							selMutex.Lock()
							isSelected := selectedIndices[rowIdx]
							selMutex.Unlock()
							check.SetChecked(isSelected)
						}
						cont.Add(check)
					} else {
						// Data columns
						label := widget.NewLabel("")
						if rowIdx < len(filteredCatalog) {
							item := filteredCatalog[rowIdx]
							switch id.Col {
							case 1:
								name, _ := item["Name"].(string)
								label.SetText(name)
							case 2:
								shortName, _ := item["ShortName"].(string)
								label.SetText(shortName)
							case 3:
								capability, _ := item["Capability"].(string)
								label.SetText(capability)
							case 4:
								label.SetText("N/A")
							}
						}
						cont.Add(label)
					}
				}
				cont.Refresh()
			},
		)

		table.SetColumnWidth(0, 40)  // Checkbox
		table.SetColumnWidth(1, 350) // Name
		table.SetColumnWidth(2, 150) // ShortName
		table.SetColumnWidth(3, 120) // Capability
		table.SetColumnWidth(4, 80)  // Version

		// Double-click on data columns to view details
		table.OnSelected = func(id widget.TableCellID) {
			if id.Row > 0 && id.Col > 0 {
				rowIdx := id.Row - 1
				if rowIdx < len(filteredCatalog) {
					shortName, _ := filteredCatalog[rowIdx]["ShortName"].(string)
					showDetail(shortName)
				}
			}
		}

		return table
	}

	// Function to update view
	updateView = func() {
		if SharedViewState.IsGridView {
			viewContainer.Objects = []fyne.CanvasObject{createGridView()}
		} else {
			viewContainer.Objects = []fyne.CanvasObject{createTableView()}
		}
		viewContainer.Refresh()
	}

	// View toggle button - set initial text based on current state
	viewToggleBtn := widget.NewButton("Table View", nil)
	if !SharedViewState.IsGridView {
		viewToggleBtn.SetText("Grid View")
	}
	viewToggleBtn.OnTapped = func() {
		SharedViewState.IsGridView = !SharedViewState.IsGridView
		if SharedViewState.IsGridView {
			viewToggleBtn.SetText("Table View")
		} else {
			viewToggleBtn.SetText("Grid View")
		}
		updateView()
	}

	statusLabel.SetText(fmt.Sprintf("Catalog contains %d store types", len(catalog)))

	// Search handler
	searchEntry.OnChanged = func(query string) {
		filteredCatalog = storeService.FilterCatalog(catalog, query)
		// Clear selections when search changes
		selMutex.Lock()
		selectedIndices = make(map[int]bool)
		selMutex.Unlock()
		updateSelectionLabel()
		updateView()
	}

	// Clear selection button
	clearSelectionBtn := widget.NewButton(
		"Clear Selection", func() {
			selMutex.Lock()
			selectedIndices = make(map[int]bool)
			selMutex.Unlock()
			updateSelectionLabel()
			updateView()
		},
	)

	// View details button
	viewBtn := widget.NewButton(
		"View Details", func() {
			if getSelectedCount() == 0 {
				dialog.ShowInformation(
					"No Selection",
					"Please select a store type first.",
					fyne.CurrentApp().Driver().AllWindows()[0],
				)
				return
			}
			if getSelectedCount() > 1 {
				dialog.ShowInformation(
					"Multiple Selection",
					"Please select only one store type to view details.",
					fyne.CurrentApp().Driver().AllWindows()[0],
				)
				return
			}
			shortNames := getSelectedShortNames()
			if len(shortNames) > 0 {
				showDetail(shortNames[0])
			}
		},
	)

	// Deploy button - supports multi-select
	deployBtn := widget.NewButton(
		"Deploy Selected", func() {
			if !authService.IsAuthenticated() {
				dialog.ShowInformation(
					"Authentication Required",
					"Please configure authentication in Settings before deploying store types.",
					fyne.CurrentApp().Driver().AllWindows()[0],
				)
				return
			}

			count := getSelectedCount()
			if count == 0 {
				dialog.ShowInformation(
					"No Selection",
					"Please select at least one store type.",
					fyne.CurrentApp().Driver().AllWindows()[0],
				)
				return
			}

			shortNames := getSelectedShortNames()

			// Build confirmation message
			message := fmt.Sprintf("Deploy %d store type(s) to Keyfactor Command?", count)
			if count <= 5 {
				message = fmt.Sprintf("Deploy the following %d store type(s)?\n\n", count)
				for _, sn := range shortNames {
					message += "- " + sn + "\n"
				}
			}

			dialog.ShowConfirm(
				"Deploy Store Types", message,
				func(confirmed bool) {
					if !confirmed {
						return
					}

					var errors []string
					var successCount int
					for _, shortName := range shortNames {
						// Check for duplicate
						isDup, _ := storeService.CheckDuplicateShortName(authService, shortName)
						if isDup {
							errors = append(errors, fmt.Sprintf("%s: already exists", shortName))
							continue
						}

						result, err := storeService.DeployFromCatalog(authService, shortName, shortName)
						if err != nil {
							errors = append(errors, fmt.Sprintf("%s: %v", shortName, err))
						} else {
							successCount++
							_ = result // silence unused warning
						}
					}

					if len(errors) > 0 {
						errMsg := fmt.Sprintf("Deployed %d of %d store types.\n\nErrors:\n", successCount, count)
						for _, e := range errors {
							errMsg += "- " + e + "\n"
						}
						dialog.ShowError(fmt.Errorf("%s", errMsg), fyne.CurrentApp().Driver().AllWindows()[0])
					} else {
						dialog.ShowInformation(
							"Deploy Successful",
							fmt.Sprintf("Successfully deployed %d store type(s).", successCount),
							fyne.CurrentApp().Driver().AllWindows()[0],
						)
					}
				}, fyne.CurrentApp().Driver().AllWindows()[0],
			)
		},
	)

	// Export button - supports multi-select
	exportBtn := widget.NewButton(
		"Export Selected", func() {
			count := getSelectedCount()
			if count == 0 {
				dialog.ShowInformation(
					"No Selection",
					"Please select at least one store type.",
					fyne.CurrentApp().Driver().AllWindows()[0],
				)
				return
			}

			// Collect selected items
			var exportData []interface{}
			var names []string
			selMutex.Lock()
			for idx := range selectedIndices {
				if idx >= 0 && idx < len(filteredCatalog) {
					exportData = append(exportData, filteredCatalog[idx])
					if name, ok := filteredCatalog[idx]["Name"].(string); ok {
						names = append(names, name)
					}
				}
			}
			selMutex.Unlock()

			var data []byte
			if len(exportData) == 1 {
				data, _ = json.MarshalIndent(exportData[0], "", "  ")
			} else {
				data, _ = json.MarshalIndent(exportData, "", "  ")
			}

			textWidget := widget.NewMultiLineEntry()
			textWidget.SetText(string(data))
			textWidget.Wrapping = fyne.TextWrapWord

			title := "Export Store Types"
			defaultFilename := "store_types.json"
			if count == 1 && len(names) > 0 {
				title = "Export Store Type - " + names[0]
				shortNames := getSelectedShortNames()
				if len(shortNames) > 0 {
					defaultFilename = shortNames[0] + ".json"
				}
			} else {
				title = fmt.Sprintf("Export %d Store Types", count)
			}

			// Copy to clipboard button
			copyBtn := widget.NewButton(
				"Copy to Clipboard", func() {
					fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(string(data))
					dialog.ShowInformation(
						"Copied",
						"JSON content copied to clipboard.",
						fyne.CurrentApp().Driver().AllWindows()[0],
					)
				},
			)

			// Save to file button
			saveBtn := widget.NewButton(
				"Save to File", func() {
					saveDialog := dialog.NewFileSave(
						func(writer fyne.URIWriteCloser, err error) {
							if err != nil || writer == nil {
								return
							}
							defer writer.Close()
							writer.Write(data)
							dialog.ShowInformation(
								"Saved",
								"File saved successfully.",
								fyne.CurrentApp().Driver().AllWindows()[0],
							)
						}, fyne.CurrentApp().Driver().AllWindows()[0],
					)
					saveDialog.SetFileName(defaultFilename)
					saveDialog.Show()
				},
			)

			buttonRow := container.NewHBox(copyBtn, saveBtn)
			content := container.NewBorder(nil, buttonRow, nil, nil, container.NewScroll(textWidget))

			d := dialog.NewCustom(title, "Close", content, fyne.CurrentApp().Driver().AllWindows()[0])
			d.Resize(fyne.NewSize(600, 500))
			d.Show()
		},
	)

	// Auth status indicator
	var authStatus string
	if authService.IsAuthenticated() {
		config := authService.GetConfig()
		authStatus = "Connected to: " + config.Hostname
	} else {
		authStatus = "Not connected (Deploy disabled)"
		deployBtn.Disable()
	}
	authLabel := widget.NewLabel(authStatus)

	// Toolbar
	toolbar := container.NewHBox(
		viewToggleBtn,
		viewBtn,
		deployBtn,
		exportBtn,
		clearSelectionBtn,
	)

	// Initial view update
	updateView()

	// Main layout
	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Store Type Catalog", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			authLabel,
			searchEntry,
			toolbar,
		),
		container.NewHBox(statusLabel, selectionLabel),
		nil, nil,
		viewContainer,
	)

	return container.NewPadded(content)
}
