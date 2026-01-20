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
	"io"
	"sync"
	"sync/atomic"

	"kfutil/pkg/gui/services"
	"kfutil/pkg/gui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/Keyfactor/keyfactor-go-client/v3/api"
)

// tableDataSnapshot holds an immutable snapshot of table data for thread-safe access
type tableDataSnapshot struct {
	items []services.StoreTypeInfo
}

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

		// Store types data - use atomic pointer for thread-safe access from table callbacks
		// This avoids creating new widgets from goroutines which causes race conditions
		var tableSnapshot atomic.Pointer[tableDataSnapshot]
		tableSnapshot.Store(&tableDataSnapshot{items: nil})

		// Also keep the full unfiltered list for filtering operations (protected by mutex)
		var dataMutex sync.Mutex
		var storeTypes []services.StoreTypeInfo

		// Multi-select: track selected indices using a map for O(1) lookup
		// Use mutex to protect concurrent access from goroutines and Fyne's renderer
		var selMutex sync.Mutex
		selectedIndices := make(map[int]bool)

		// Helper to get current snapshot safely
		getSnapshot := func() []services.StoreTypeInfo {
			snap := tableSnapshot.Load()
			if snap == nil {
				return nil
			}
			return snap.items
		}

		// Helper to get selected count (must be called with lock held or in safe context)
		getSelectedCount := func() int {
			selMutex.Lock()
			defer selMutex.Unlock()
			return len(selectedIndices)
		}

		// Helper to get selected IDs (uses snapshot)
		getSelectedIDs := func() []int {
			selMutex.Lock()
			defer selMutex.Unlock()
			items := getSnapshot()
			var ids []int
			for idx := range selectedIndices {
				if idx >= 0 && idx < len(items) {
					ids = append(ids, items[idx].ID)
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

		// Create a PERSISTENT table that reads from the atomic snapshot
		// This avoids creating new widgets from goroutines which causes race conditions
		var persistentTable *widget.Table
		var emptyLabel *widget.Label
		var gridContainer *fyne.Container
		var gridScroll *container.Scroll

		emptyLabel = widget.NewLabel("No store types found")

		persistentTable = widget.NewTable(
			func() (int, int) {
				items := getSnapshot()
				if len(items) == 0 {
					return 0, 0
				}
				return len(items) + 1, 6 // +1 for header, 6 columns (checkbox + 5 data)
			},
			func() fyne.CanvasObject {
				// Create a container with both a label and checkbox - we'll show/hide as needed
				return container.NewMax(widget.NewLabel(""), widget.NewCheck("", nil))
			},
			func(id widget.TableCellID, cell fyne.CanvasObject) {
				cont := cell.(*fyne.Container)
				items := getSnapshot()

				// Get label and check from container
				var label *widget.Label
				var check *widget.Check
				for _, obj := range cont.Objects {
					if l, ok := obj.(*widget.Label); ok {
						label = l
					}
					if c, ok := obj.(*widget.Check); ok {
						check = c
					}
				}

				if id.Row == 0 {
					// Header row - show label, hide check
					if check != nil {
						check.Hide()
					}
					if label != nil {
						label.Show()
						headers := []string{"", "ID", "Name", "ShortName", "Capability", "Version"}
						label.TextStyle = fyne.TextStyle{Bold: true}
						label.SetText(headers[id.Col])
					}
				} else {
					rowIdx := id.Row - 1

					if id.Col == 0 {
						// Checkbox column - show check, hide label
						if label != nil {
							label.Hide()
						}
						if check != nil {
							check.Show()
							check.OnChanged = func(checked bool) {
								selMutex.Lock()
								if checked {
									selectedIndices[rowIdx] = true
								} else {
									delete(selectedIndices, rowIdx)
								}
								selMutex.Unlock()
								updateSelectionLabel()
							}
							if rowIdx < len(items) {
								selMutex.Lock()
								isSelected := selectedIndices[rowIdx]
								selMutex.Unlock()
								check.SetChecked(isSelected)
							}
						}
					} else {
						// Data columns - show label, hide check
						if check != nil {
							check.Hide()
						}
						if label != nil {
							label.Show()
							label.TextStyle = fyne.TextStyle{Bold: false}
							if rowIdx < len(items) {
								st := items[rowIdx]
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
							} else {
								label.SetText("")
							}
						}
					}
				}
			},
		)

		persistentTable.SetColumnWidth(0, 40)  // Checkbox
		persistentTable.SetColumnWidth(1, 60)  // ID
		persistentTable.SetColumnWidth(2, 350) // Name (wider to accommodate long names)
		persistentTable.SetColumnWidth(3, 150) // ShortName
		persistentTable.SetColumnWidth(4, 120) // Capability
		persistentTable.SetColumnWidth(5, 80)  // Version

		// Double-click to view details
		persistentTable.OnSelected = func(id widget.TableCellID) {
			if id.Row > 0 && id.Col > 0 {
				rowIdx := id.Row - 1
				items := getSnapshot()
				if rowIdx < len(items) {
					showDetail(items[rowIdx].ID)
				}
			}
		}

		// Grid container - we'll rebuild grid content only from main thread (button clicks)
		gridContainer = container.NewGridWithColumns(4)
		gridScroll = container.NewScroll(gridContainer)

		// Function to rebuild grid view - ONLY call from main thread (button handlers)
		rebuildGridView := func() {
			items := getSnapshot()
			gridContainer.Objects = nil

			if len(items) == 0 {
				return
			}

			for i, st := range items {
				idx := i
				stCopy := st
				card := widgets.NewStoreCard(
					stCopy.ID,
					stCopy.Name,
					stCopy.ShortName,
					stCopy.Description,
					stCopy.Capability,
				)

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
					showDetail(stCopy.ID)
				}

				cardWithCheck := container.NewBorder(nil, nil, check, nil, card)
				gridContainer.Objects = append(gridContainer.Objects, container.NewPadded(cardWithCheck))
			}
			gridContainer.Refresh()
		}

		// Function to update the view based on current mode
		// CRITICAL: When fromGoroutine=true, we ONLY refresh the table widget
		// We NEVER modify viewContainer.Objects from a goroutine as it causes race conditions
		var updateView func(fromGoroutine bool)
		updateView = func(fromGoroutine bool) {
			if fromGoroutine {
				// From goroutine: ONLY refresh the table if we're in table mode
				// Do NOT modify any container objects - this causes race conditions with Fyne's renderer
				if !SharedViewState.IsGridView {
					persistentTable.UnselectAll()
					persistentTable.ScrollToTop()
					persistentTable.Refresh()
				}
				// For grid mode, the user must click Refresh button (main thread) to see updates
				return
			}

			// From main thread: safe to modify UI structure
			items := getSnapshot()
			hasData := len(items) > 0

			if SharedViewState.IsGridView {
				// Grid mode
				rebuildGridView()
				if hasData {
					viewContainer.Objects = []fyne.CanvasObject{gridScroll}
				} else {
					viewContainer.Objects = []fyne.CanvasObject{container.NewCenter(emptyLabel)}
				}
				viewContainer.Refresh()
			} else {
				// Table mode
				if hasData {
					viewContainer.Objects = []fyne.CanvasObject{persistentTable}
				} else {
					viewContainer.Objects = []fyne.CanvasObject{container.NewCenter(emptyLabel)}
				}
				viewContainer.Refresh()

				// Table refresh - clear selection and scroll to top to force re-render
				persistentTable.UnselectAll()
				if hasData {
					persistentTable.ScrollToTop()
				}
				persistentTable.Refresh()
			}
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
			updateView(false) // Called from main thread (button handler)
		}

		// loadData fetches store types and updates the snapshot (can be called sync or async)
		loadData := func() error {
			types, err := storeService.ListInstalledStoreTypes(authService)
			if err != nil {
				return err
			}

			// Update data under lock, then store atomic snapshot
			dataMutex.Lock()
			storeTypes = types
			filtered := storeService.FilterStoreTypes(storeTypes, searchEntry.Text)
			dataMutex.Unlock()

			// Store new snapshot atomically
			tableSnapshot.Store(&tableDataSnapshot{items: filtered})

			// Clear selections on refresh - protected by mutex
			selMutex.Lock()
			selectedIndices = make(map[int]bool)
			selMutex.Unlock()

			return nil
		}

		// initialLoad does a synchronous load on first render so the view has data immediately
		initialLoad := func() {
			statusLabel.SetText("Loading...")
			err := loadData()
			if err != nil {
				statusLabel.SetText("Error: " + err.Error())
				return
			}
			items := getSnapshot()
			statusLabel.SetText(fmt.Sprintf("Loaded %d store types", len(items)))
		}

		// refreshList does an async refresh (for button clicks after initial load)
		var refreshList func()
		refreshList = func() {
			statusLabel.SetText("Loading...")

			go func() {
				err := loadData()
				if err != nil {
					statusLabel.SetText("Error: " + err.Error())
					return
				}

				updateSelectionLabel()
				items := getSnapshot()
				statusLabel.SetText(fmt.Sprintf("Loaded %d store types", len(items)))

				// For grid mode, we need to navigate to refresh the view since we can't
				// safely rebuild the grid from a goroutine (it creates widgets and modifies containers)
				// For table mode, we can just refresh the table widget
				if SharedViewState.IsGridView {
					// Navigate to trigger a fresh view creation on the main thread
					navigateTo("Installed Store Types")
				} else {
					updateView(true) // Called from goroutine - only refreshes table
				}
			}()
		}

		// Search handler
		searchEntry.OnChanged = func(query string) {
			dataMutex.Lock()
			filtered := storeService.FilterStoreTypes(storeTypes, query)
			dataMutex.Unlock()
			// Store new snapshot atomically
			tableSnapshot.Store(&tableDataSnapshot{items: filtered})
			// Clear selections when search changes
			selMutex.Lock()
			selectedIndices = make(map[int]bool)
			selMutex.Unlock()
			updateSelectionLabel()
			updateView(false) // Called from main thread (UI callback)
		}

		// Clear selection button
		clearSelectionBtn := widget.NewButton(
			"Clear Selection", func() {
				selMutex.Lock()
				selectedIndices = make(map[int]bool)
				selMutex.Unlock()
				updateSelectionLabel()
				updateView(false) // Called from main thread (button handler)
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
				items := getSnapshot()
				selMutex.Lock()
				for idx := range selectedIndices {
					if idx >= 0 && idx < len(items) {
						names = append(names, items[idx].Name)
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
					items := getSnapshot()
					selMutex.Lock()
					for idx := range selectedIndices {
						if idx >= 0 && idx < len(items) {
							title = "Export Store Type - " + items[idx].Name
							defaultFilename = items[idx].ShortName + ".json"
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
			"Open Store Type Catalog", func() {
				navigateTo("Store Type Catalog")
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

		// Do initial data load synchronously so the view has data immediately
		// This ensures grid view works correctly on first render
		initialLoad()

		// Now set up the view with the loaded data (runs on main thread)
		updateView(false)

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

	// Initial auth check - MUST be synchronous to avoid race conditions with Fyne's renderer
	// Modifying container.Objects and calling Refresh() from a goroutine causes concurrent map writes
	initialAuthCheck := func() {
		// Check if already authenticated - this is a quick in-memory check
		if authService.IsAuthenticated() {
			// Show main content directly (no goroutine needed)
			mainContainer.Objects = []fyne.CanvasObject{buildMainContent()}
			return
		}

		// Not authenticated - show auth check view with option to retry
		authStatusLabel.SetText("Not authenticated")
		authErrorLabel.SetText("Please configure authentication in Settings or click Retry to attempt connection.")
		retryBtn.Enable()
		mainContainer.Objects = []fyne.CanvasObject{authCheckView}
	}

	// Retry function - uses goroutine for network call but avoids modifying containers from goroutine
	retryAuth := func() {
		authStatusLabel.SetText("Checking authentication...")
		authErrorLabel.SetText("")
		retryBtn.Disable()

		go func() {
			// Try to test connection
			err := authService.TestConnection()
			if err != nil {
				// Update labels (safe from goroutine - Fyne widgets handle their own sync)
				authStatusLabel.SetText("Authentication failed")
				authErrorLabel.SetText(err.Error())
				retryBtn.Enable()
				return
			}

			// Authentication successful - navigate to trigger fresh view creation
			// This avoids modifying container objects from a goroutine which causes race conditions
			navigateTo("Installed Store Types")
		}()
	}

	// Set up retry button action
	retryBtn.OnTapped = retryAuth

	// Initial auth check (synchronous - safe)
	initialAuthCheck()

	return mainContainer
}
