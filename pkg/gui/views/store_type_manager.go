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

package views

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"kfutil/pkg/gui/services"
	"kfutil/pkg/gui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/Keyfactor/keyfactor-go-client/v3/api"
)

// NewStoreManagerView creates the store type manager view for installed store types
func NewStoreManagerView(
	authService *services.AuthService,
	storeService *services.StoreService,
	showDetail func(int),
	navigateTo func(string),
) fyne.CanvasObject {
	// Create a container that will hold either the auth check view or the main content
	mainContainer := container.NewMax()

	// Auth status widgets
	authStatusLabel := widget.NewLabel("Checking authentication...")
	authErrorLabel := widget.NewLabel("")
	authErrorLabel.Wrapping = fyne.TextWrapWord

	retryBtn := widget.NewButton("Retry Connection", nil)
	settingsBtn := widget.NewButton(
		"Go to Settings", func() {
			navigateTo("Settings")
		},
	)

	authCheckView := container.NewCenter(
		container.NewVBox(
			widget.NewLabelWithStyle("Authentication Required", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			authStatusLabel,
			authErrorLabel,
			container.NewHBox(retryBtn, settingsBtn),
		),
	)

	// Function to build the main store manager content
	buildMainContent := func() fyne.CanvasObject {
		// Search entry
		searchEntry := widget.NewEntry()
		searchEntry.SetPlaceHolder("Search by Name, ShortName, ID, or Capability...")

		// Status label
		statusLabel := widget.NewLabel("Loading...")

		// View mode state - use shared state for persistence
		// (SharedViewState.IsGridView is used directly)

		// Store types data
		var storeTypes []services.StoreTypeInfo
		var filteredTypes []services.StoreTypeInfo

		// Multi-select: track selected indices using a map for O(1) lookup
		// Use mutex to protect concurrent access from goroutines and Fyne's renderer
		var selMutex sync.Mutex
		selectedIndices := make(map[int]bool)

		// Helper to get selected count (must be called with lock held or in safe context)
		getSelectedCount := func() int {
			selMutex.Lock()
			defer selMutex.Unlock()
			return len(selectedIndices)
		}

		// Helper to get selected IDs (must be called with lock held or in safe context)
		getSelectedIDs := func() []int {
			selMutex.Lock()
			defer selMutex.Unlock()
			var ids []int
			for idx := range selectedIndices {
				if idx >= 0 && idx < len(filteredTypes) {
					ids = append(ids, filteredTypes[idx].ID)
				}
			}
			return ids
		}

		// Container that will hold the current view (grid or table)
		viewContainer := container.NewMax()

		// Selection label to show count
		selectionLabel := widget.NewLabel("")
		var updateSelectionLabel func()
		updateSelectionLabel = func() {
			count := getSelectedCount()
			if count == 0 {
				selectionLabel.SetText("")
			} else if count == 1 {
				selectionLabel.SetText("1 item selected")
			} else {
				selectionLabel.SetText(fmt.Sprintf("%d items selected", count))
			}
		}

		// Grid view - uses a scrollable grid of cards with checkboxes
		var createGridView func() fyne.CanvasObject
		createGridView = func() fyne.CanvasObject {
			if len(filteredTypes) == 0 {
				return container.NewCenter(widget.NewLabel("No store types found"))
			}

			// Create cards for each store type with checkbox
			var cardObjects []fyne.CanvasObject
			for i, st := range filteredTypes {
				idx := i // capture for closure
				card := widgets.NewStoreCard(st.ID, st.Name, st.ShortName, st.Description, st.Capability)

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
					showDetail(filteredTypes[idx].ID)
				}

				// Wrap card with checkbox
				cardWithCheck := container.NewBorder(nil, nil, check, nil, card)
				cardObjects = append(cardObjects, container.NewPadded(cardWithCheck))
			}

			// Create a grid with 4 columns
			grid := container.NewGridWithColumns(4, cardObjects...)
			return container.NewScroll(grid)
		}

		// Table view creator - creates a fresh table each time to avoid race conditions
		// with Fyne's internal table renderer state
		var createTableView func() fyne.CanvasObject
		createTableView = func() fyne.CanvasObject {
			if len(filteredTypes) == 0 {
				return container.NewCenter(widget.NewLabel("No store types found"))
			}

			table := widget.NewTable(
				func() (int, int) {
					return len(filteredTypes) + 1, 6 // +1 for header, 6 columns (checkbox + 5 data)
				},
				func() fyne.CanvasObject {
					// Create a container that can hold either a label or checkbox
					return container.NewMax(widget.NewCheck("", nil))
				},
				func(id widget.TableCellID, cell fyne.CanvasObject) {
					cont := cell.(*fyne.Container)

					if id.Row == 0 {
						// Header row
						cont.Objects = []fyne.CanvasObject{}
						headers := []string{"", "ID", "Name", "ShortName", "Capability", "Version"}
						label := widget.NewLabel(headers[id.Col])
						label.TextStyle = fyne.TextStyle{Bold: true}
						cont.Add(label)
					} else {
						cont.Objects = []fyne.CanvasObject{}
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
							if rowIdx < len(filteredTypes) {
								selMutex.Lock()
								isSelected := selectedIndices[rowIdx]
								selMutex.Unlock()
								check.SetChecked(isSelected)
							}
							cont.Add(check)
						} else {
							// Data columns
							label := widget.NewLabel("")
							if rowIdx < len(filteredTypes) {
								st := filteredTypes[rowIdx]
								switch id.Col {
								case 1:
									label.SetText(fmt.Sprintf("%d", st.ID))
								case 2:
									label.SetText(st.Name)
								case 3:
									label.SetText(st.ShortName)
								case 4:
									label.SetText(st.Capability)
								case 5:
									label.SetText(st.Version)
								}
							}
							cont.Add(label)
						}
					}
					cont.Refresh()
				},
			)

			table.SetColumnWidth(0, 40)  // Checkbox
			table.SetColumnWidth(1, 60)  // ID
			table.SetColumnWidth(2, 350) // Name (wider to accommodate long names)
			table.SetColumnWidth(3, 150) // ShortName
			table.SetColumnWidth(4, 120) // Capability
			table.SetColumnWidth(5, 80)  // Version

			// Double-click to view details
			table.OnSelected = func(id widget.TableCellID) {
				if id.Row > 0 && id.Col > 0 {
					// Double-click on data cells opens detail
					rowIdx := id.Row - 1
					if rowIdx < len(filteredTypes) {
						showDetail(filteredTypes[rowIdx].ID)
					}
				}
			}

			return table
		}

		// Function to update the view based on current mode
		var updateView func()
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

		// Refresh function
		var refreshList func()
		refreshList = func() {
			statusLabel.SetText("Loading...")

			go func() {
				types, err := storeService.ListInstalledStoreTypes(authService)
				if err != nil {
					statusLabel.SetText("Error: " + err.Error())
					return
				}

				storeTypes = types
				filteredTypes = storeService.FilterStoreTypes(storeTypes, searchEntry.Text)

				// Clear selections on refresh - protected by mutex
				selMutex.Lock()
				selectedIndices = make(map[int]bool)
				selMutex.Unlock()
				updateSelectionLabel()

				statusLabel.SetText(fmt.Sprintf("Loaded %d store types", len(storeTypes)))
				updateView()
			}()
		}

		// Search handler
		searchEntry.OnChanged = func(query string) {
			filteredTypes = storeService.FilterStoreTypes(storeTypes, query)
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

		// View details button (for single selection)
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
				// Get the single selected ID
				ids := getSelectedIDs()
				if len(ids) > 0 {
					showDetail(ids[0])
				}
			},
		)

		// Refresh button
		refreshBtn := widget.NewButton("Refresh", refreshList)

		// Delete button - supports multi-select
		deleteBtn := widget.NewButton(
			"Delete Selected", func() {
				count := getSelectedCount()
				if count == 0 {
					dialog.ShowInformation(
						"No Selection",
						"Please select at least one store type.",
						fyne.CurrentApp().Driver().AllWindows()[0],
					)
					return
				}

				ids := getSelectedIDs()
				var names []string
				selMutex.Lock()
				for idx := range selectedIndices {
					if idx >= 0 && idx < len(filteredTypes) {
						names = append(names, filteredTypes[idx].Name)
					}
				}
				selMutex.Unlock()

				message := fmt.Sprintf("Are you sure you want to delete %d store type(s)?", count)
				if count <= 5 {
					message = fmt.Sprintf("Are you sure you want to delete the following %d store type(s)?\n\n", count)
					for _, name := range names {
						message += "- " + name + "\n"
					}
				}

				dialog.ShowConfirm(
					"Confirm Delete", message,
					func(confirmed bool) {
						if !confirmed {
							return
						}

						// Delete all selected
						var errors []string
						for _, id := range ids {
							err := storeService.DeleteStoreType(authService, id)
							if err != nil {
								errors = append(errors, fmt.Sprintf("ID %d: %v", id, err))
							}
						}

						if len(errors) > 0 {
							dialog.ShowError(
								fmt.Errorf("some deletions failed:\n%s", errors),
								fyne.CurrentApp().Driver().AllWindows()[0],
							)
						}
						refreshList()
					}, fyne.CurrentApp().Driver().AllWindows()[0],
				)
			},
		)

		// Export button - supports multi-select with save and copy options
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

				ids := getSelectedIDs()

				// Collect all selected store types
				var exportData []interface{}
				for _, id := range ids {
					fullType, err := storeService.GetStoreType(authService, id)
					if err != nil {
						dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
						return
					}
					exportData = append(exportData, fullType)
				}

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
				if count == 1 {
					selMutex.Lock()
					for idx := range selectedIndices {
						if idx >= 0 && idx < len(filteredTypes) {
							title = "Export Store Type - " + filteredTypes[idx].Name
							defaultFilename = filteredTypes[idx].ShortName + ".json"
							break
						}
					}
					selMutex.Unlock()
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

		// Import button with file picker and paste options
		importBtn := widget.NewButton(
			"Import from JSON", func() {
				textWidget := widget.NewMultiLineEntry()
				textWidget.SetPlaceHolder("Paste JSON here, or use 'Load from File' button...")
				textWidget.Wrapping = fyne.TextWrapWord

				// Load from file button
				loadFileBtn := widget.NewButton(
					"Load from File", func() {
						fileDialog := dialog.NewFileOpen(
							func(reader fyne.URIReadCloser, err error) {
								if err != nil || reader == nil {
									return
								}
								defer reader.Close()

								data, readErr := io.ReadAll(reader)
								if readErr != nil {
									dialog.ShowError(
										fmt.Errorf("failed to read file: %w", readErr),
										fyne.CurrentApp().Driver().AllWindows()[0],
									)
									return
								}
								textWidget.SetText(string(data))
							}, fyne.CurrentApp().Driver().AllWindows()[0],
						)
						fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
						fileDialog.Show()
					},
				)

				content := container.NewBorder(loadFileBtn, nil, nil, nil, container.NewScroll(textWidget))

				d := dialog.NewCustomConfirm(
					"Import Store Type", "Import", "Cancel",
					content,
					func(confirmed bool) {
						if !confirmed || textWidget.Text == "" {
							return
						}

						// Try to parse as a single store type or an array
						var storeTypes []api.CertificateStoreType

						// First try as array
						if err := json.Unmarshal([]byte(textWidget.Text), &storeTypes); err != nil {
							// Try as single object
							var single api.CertificateStoreType
							if err := json.Unmarshal([]byte(textWidget.Text), &single); err != nil {
								dialog.ShowError(
									fmt.Errorf("invalid JSON: %w", err),
									fyne.CurrentApp().Driver().AllWindows()[0],
								)
								return
							}
							storeTypes = []api.CertificateStoreType{single}
						}

						if len(storeTypes) == 0 {
							dialog.ShowError(
								fmt.Errorf("no store types found in JSON"),
								fyne.CurrentApp().Driver().AllWindows()[0],
							)
							return
						}

						// Import each store type
						var errors []string
						var successCount int
						for _, st := range storeTypes {
							// Clear ID for new creation
							st.StoreType = 0
							// Clear deprecated JobProperties
							st.JobProperties = nil

							result, createErr := storeService.CreateStoreType(authService, &st)
							if createErr != nil {
								errors = append(errors, fmt.Sprintf("%s: %v", st.ShortName, createErr))
							} else {
								successCount++
								_ = result
							}
						}

						if len(errors) > 0 {
							errMsg := fmt.Sprintf(
								"Imported %d of %d store type(s).\n\nErrors:\n",
								successCount,
								len(storeTypes),
							)
							for _, e := range errors {
								errMsg += "- " + e + "\n"
							}
							dialog.ShowError(fmt.Errorf("%s", errMsg), fyne.CurrentApp().Driver().AllWindows()[0])
						} else {
							dialog.ShowInformation(
								"Import Successful",
								fmt.Sprintf("Successfully imported %d store type(s).", successCount),
								fyne.CurrentApp().Driver().AllWindows()[0],
							)
						}
						refreshList()
					}, fyne.CurrentApp().Driver().AllWindows()[0],
				)
				d.Resize(fyne.NewSize(600, 500))
				d.Show()
			},
		)

		// Open catalog button
		catalogBtn := widget.NewButton(
			"Open Catalog", func() {
				navigateTo("Catalog")
			},
		)

		// Toolbar
		toolbar := container.NewHBox(
			refreshBtn,
			viewToggleBtn,
			viewBtn,
			deleteBtn,
			exportBtn,
			importBtn,
			catalogBtn,
			clearSelectionBtn,
		)

		// Initial load
		refreshList()

		// Main layout with selection info
		content := container.NewBorder(
			container.NewVBox(
				widget.NewLabelWithStyle("Installed Store Types", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				searchEntry,
				toolbar,
			),
			container.NewHBox(statusLabel, selectionLabel),
			nil, nil,
			viewContainer,
		)

		return container.NewPadded(content)
	}

	// Function to check authentication and show appropriate view
	checkAuthAndLoad := func() {
		authStatusLabel.SetText("Checking authentication...")
		authErrorLabel.SetText("")
		retryBtn.Disable()

		mainContainer.Objects = []fyne.CanvasObject{authCheckView}
		mainContainer.Refresh()

		go func() {
			// Check if already authenticated
			if authService.IsAuthenticated() {
				// Show main content
				mainContainer.Objects = []fyne.CanvasObject{buildMainContent()}
				mainContainer.Refresh()
				return
			}

			// Try to test connection
			err := authService.TestConnection()
			if err != nil {
				authStatusLabel.SetText("Authentication failed")
				authErrorLabel.SetText(err.Error())
				retryBtn.Enable()
				mainContainer.Refresh()
				return
			}

			// Authentication successful, show main content
			mainContainer.Objects = []fyne.CanvasObject{buildMainContent()}
			mainContainer.Refresh()
		}()
	}

	// Set up retry button action
	retryBtn.OnTapped = checkAuthAndLoad

	// Initial auth check
	checkAuthAndLoad()

	return mainContainer
}
