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
	"strconv"
	"strings"

	"kfutil/pkg/gui/services"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/Keyfactor/keyfactor-go-client/v3/api"
)

// NewStoreDetailView creates a detail view for an installed store type
func NewStoreDetailView(
	authService *services.AuthService,
	storeService *services.StoreService,
	storeTypeID int,
	editMode bool,
	showNotification func(string, string),
	onClose func(),
	onSuccess func(), // Called after successful create/deploy to refresh list
) fyne.CanvasObject {
	// Load store type
	storeType, err := storeService.GetStoreType(authService, storeTypeID)
	if err != nil {
		return container.NewCenter(
			container.NewVBox(
				widget.NewLabel("Error loading store type: "+err.Error()),
				widget.NewButton("Close", onClose),
			),
		)
	}

	return createDetailView(authService, storeService, storeType, true, showNotification, onClose, onSuccess)
}

// NewCatalogDetailView creates a detail view for a catalog store type
func NewCatalogDetailView(
	authService *services.AuthService,
	storeService *services.StoreService,
	shortName string,
	showNotification func(string, string),
	onClose func(),
	onSuccess func(), // Called after successful create/deploy to refresh list
) fyne.CanvasObject {
	// Load from catalog
	item, err := storeService.GetCatalogItem(shortName)
	if err != nil {
		return container.NewCenter(
			container.NewVBox(
				widget.NewLabel("Error loading store type: "+err.Error()),
				widget.NewButton("Close", onClose),
			),
		)
	}

	storeType, err := storeService.CatalogItemToStoreType(item)
	if err != nil {
		return container.NewCenter(
			container.NewVBox(
				widget.NewLabel("Error converting store type: "+err.Error()),
				widget.NewButton("Close", onClose),
			),
		)
	}

	return createDetailView(authService, storeService, storeType, false, showNotification, onClose, onSuccess)
}

// createDetailView creates the actual detail view UI
func createDetailView(
	authService *services.AuthService,
	storeService *services.StoreService,
	storeType *api.CertificateStoreType,
	isInstalled bool,
	showNotification func(string, string),
	onClose func(),
	onSuccess func(), // Called after successful create/deploy to refresh list
) fyne.CanvasObject {
	// Title
	title := widget.NewLabelWithStyle("Store Type Details", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	if isInstalled {
		title.SetText("Store Type Details (Installed)")
	} else {
		title.SetText("Store Type Details (Catalog)")
	}

	// Read-only ID field (only for installed, not shown for catalog)
	idLabel := widget.NewLabel("")
	if isInstalled && storeType.StoreType > 0 {
		idLabel.SetText(fmt.Sprintf("ID: %d", storeType.StoreType))
	}

	// Editable fields
	nameEntry := widget.NewEntry()
	nameEntry.SetText(storeType.Name)

	shortNameEntry := widget.NewEntry()
	shortNameEntry.SetText(storeType.ShortName)
	if isInstalled {
		shortNameEntry.Disable() // ShortName can't be changed for installed types
	}

	capabilityEntry := widget.NewEntry()
	capabilityEntry.SetText(storeType.Capability)
	if isInstalled {
		capabilityEntry.Disable()
	}

	// Boolean fields
	localStoreCheck := widget.NewCheck("Local Store", nil)
	localStoreCheck.SetChecked(storeType.LocalStore)

	serverRequiredCheck := widget.NewCheck("Server Required", nil)
	serverRequiredCheck.SetChecked(storeType.ServerRequired)

	blueprintAllowedCheck := widget.NewCheck("Blueprint Allowed", nil)
	blueprintAllowedCheck.SetChecked(storeType.BlueprintAllowed)

	powerShellCheck := widget.NewCheck("PowerShell", nil)
	powerShellCheck.SetChecked(storeType.PowerShell)

	// Select fields
	privateKeySelect := widget.NewSelect([]string{"Optional", "Required", "Forbidden"}, nil)
	if storeType.PrivateKeyAllowed != "" {
		privateKeySelect.SetSelected(storeType.PrivateKeyAllowed)
	}

	customAliasSelect := widget.NewSelect([]string{"Optional", "Required", "Forbidden"}, nil)
	customAliasSelect.SetSelected(storeType.CustomAliasAllowed)

	// Store path fields
	storePathTypeSelect := widget.NewSelect([]string{"", "Freeform", "Fixed", "MultipleChoice"}, nil)

	// Infer StorePathType if not set but StorePathValue looks like an array
	// (workaround for Keyfactor Command API bug where StorePathType is not returned)
	inferredStorePathType := storeType.StorePathType
	if inferredStorePathType == "" && storeType.StorePathValue != "" {
		val := storeType.StorePathValue
		// Check if it looks like a JSON array or escaped JSON array
		if (strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]")) ||
			(strings.HasPrefix(val, `"[`) || strings.Contains(val, `\",\"`)) {
			inferredStorePathType = "MultipleChoice"
		}
	}
	storePathTypeSelect.SetSelected(inferredStorePathType)

	storePathValueEntry := widget.NewEntry()
	storePathValueEntry.SetText(storeType.StorePathValue)

	// Supported Operations
	var addOp, createOp, discoveryOp, enrollmentOp, removeOp bool
	if storeType.SupportedOperations != nil {
		addOp = storeType.SupportedOperations.Add
		createOp = storeType.SupportedOperations.Create
		discoveryOp = storeType.SupportedOperations.Discovery
		enrollmentOp = storeType.SupportedOperations.Enrollment
		removeOp = storeType.SupportedOperations.Remove
	}

	addCheck := widget.NewCheck("Add", nil)
	addCheck.SetChecked(addOp)
	createCheck := widget.NewCheck("Create", nil)
	createCheck.SetChecked(createOp)
	discoveryCheck := widget.NewCheck("Discovery", nil)
	discoveryCheck.SetChecked(discoveryOp)
	enrollmentCheck := widget.NewCheck("Enrollment", nil)
	enrollmentCheck.SetChecked(enrollmentOp)
	removeCheck := widget.NewCheck("Remove", nil)
	removeCheck.SetChecked(removeOp)

	// Properties - editable list
	var properties []api.StoreTypePropertyDefinition
	if storeType.Properties != nil {
		properties = *storeType.Properties
	}

	// Entry Parameters - editable list
	var entryParams []api.EntryParameter
	if storeType.EntryParameters != nil {
		entryParams = *storeType.EntryParameters
	}

	// Track selected properties for removal
	selectedProperties := make(map[int]bool)

	// Create properties table view
	propertiesContainer := container.NewMax()
	var refreshPropertiesView func()

	refreshPropertiesView = func() {
		// Clear selections when refreshing
		selectedProperties = make(map[int]bool)

		if len(properties) == 0 {
			propertiesContainer.Objects = []fyne.CanvasObject{
				container.NewCenter(widget.NewLabel("No Properties available")),
			}
		} else {
			// Create table for properties with checkbox column
			table := widget.NewTable(
				func() (int, int) { return len(properties) + 1, 5 }, // 5 columns: checkbox + 4 data
				func() fyne.CanvasObject {
					return container.NewMax(widget.NewCheck("", nil))
				},
				func(id widget.TableCellID, cell fyne.CanvasObject) {
					cont := cell.(*fyne.Container)
					cont.Objects = nil

					if id.Row == 0 {
						// Header row
						headers := []string{"", "DisplayName", "Type", "Required", "DefaultValue"}
						label := widget.NewLabel(headers[id.Col])
						label.TextStyle = fyne.TextStyle{Bold: true}
						cont.Add(label)
					} else {
						rowIdx := id.Row - 1
						if id.Col == 0 {
							// Checkbox column
							check := widget.NewCheck(
								"", func(checked bool) {
									if checked {
										selectedProperties[rowIdx] = true
									} else {
										delete(selectedProperties, rowIdx)
									}
								},
							)
							check.SetChecked(selectedProperties[rowIdx])
							cont.Add(check)
						} else {
							// Data columns
							label := widget.NewLabel("")
							if rowIdx < len(properties) {
								prop := properties[rowIdx]
								switch id.Col {
								case 1:
									label.SetText(prop.DisplayName)
								case 2:
									label.SetText(prop.Type)
								case 3:
									if prop.Required {
										label.SetText("Yes")
									} else {
										label.SetText("No")
									}
								case 4:
									if prop.DefaultValue != nil {
										label.SetText(fmt.Sprintf("%v", prop.DefaultValue))
									}
								}
							}
							cont.Add(label)
						}
					}
					cont.Refresh()
				},
			)
			table.SetColumnWidth(0, 40)  // Checkbox
			table.SetColumnWidth(1, 300) // DisplayName (doubled for longer values)
			table.SetColumnWidth(2, 100) // Type
			table.SetColumnWidth(3, 80)  // Required
			table.SetColumnWidth(4, 150) // DefaultValue

			// Click on data columns to edit
			table.OnSelected = func(id widget.TableCellID) {
				if id.Row > 0 && id.Col > 0 {
					rowIdx := id.Row - 1
					if rowIdx < len(properties) {
						showPropertyDetailDialog(
							properties[rowIdx], isInstalled, func(updated api.StoreTypePropertyDefinition) {
								properties[rowIdx] = updated
								refreshPropertiesView()
							},
						)
					}
				}
			}

			// Set minimum height for 5 rows - use full width
			tableWithHeight := container.NewStack(
				container.NewGridWrap(fyne.NewSize(100, 180), widget.NewLabel("")), // height spacer
				table,
			)
			propertiesContainer.Objects = []fyne.CanvasObject{tableWithHeight}
		}
		propertiesContainer.Refresh()
	}

	// Add property button
	addPropertyBtn := widget.NewButton(
		"Add Property", func() {
			newProp := api.StoreTypePropertyDefinition{
				Name:        "NewProperty",
				DisplayName: "New Property",
				Type:        "String",
				Required:    false,
			}
			showPropertyDetailDialog(
				newProp, isInstalled, func(updated api.StoreTypePropertyDefinition) {
					properties = append(properties, updated)
					refreshPropertiesView()
				},
			)
		},
	)

	// Remove property button
	removePropertyBtn := widget.NewButton(
		"Remove Selected", func() {
			if len(selectedProperties) == 0 {
				dialog.ShowInformation(
					"No Selection",
					"Please select properties to remove using the checkboxes.",
					fyne.CurrentApp().Driver().AllWindows()[0],
				)
				return
			}

			count := len(selectedProperties)
			message := fmt.Sprintf("Remove %d selected property(ies)?", count)

			dialog.ShowConfirm(
				"Remove Properties", message, func(confirmed bool) {
					if !confirmed {
						return
					}
					// Remove selected properties in reverse order to maintain indices
					var newProperties []api.StoreTypePropertyDefinition
					for i, prop := range properties {
						if !selectedProperties[i] {
							newProperties = append(newProperties, prop)
						}
					}
					properties = newProperties
					refreshPropertiesView()
				}, fyne.CurrentApp().Driver().AllWindows()[0],
			)
		},
	)

	refreshPropertiesView()

	// Track selected entry params for removal
	selectedEntryParams := make(map[int]bool)

	// Create entry parameters table view
	entryParamsContainer := container.NewMax()
	var refreshEntryParamsView func()

	refreshEntryParamsView = func() {
		// Clear selections when refreshing
		selectedEntryParams = make(map[int]bool)

		if len(entryParams) == 0 {
			entryParamsContainer.Objects = []fyne.CanvasObject{
				container.NewCenter(widget.NewLabel("No Entry Parameters available")),
			}
		} else {
			// Create table for entry parameters with checkbox column
			table := widget.NewTable(
				func() (int, int) { return len(entryParams) + 1, 5 }, // 5 columns: checkbox + 4 data
				func() fyne.CanvasObject {
					return container.NewMax(widget.NewCheck("", nil))
				},
				func(id widget.TableCellID, cell fyne.CanvasObject) {
					cont := cell.(*fyne.Container)
					cont.Objects = nil

					if id.Row == 0 {
						// Header row
						headers := []string{"", "DisplayName", "Type", "Required", "DefaultValue"}
						label := widget.NewLabel(headers[id.Col])
						label.TextStyle = fyne.TextStyle{Bold: true}
						cont.Add(label)
					} else {
						rowIdx := id.Row - 1
						if id.Col == 0 {
							// Checkbox column
							check := widget.NewCheck(
								"", func(checked bool) {
									if checked {
										selectedEntryParams[rowIdx] = true
									} else {
										delete(selectedEntryParams, rowIdx)
									}
								},
							)
							check.SetChecked(selectedEntryParams[rowIdx])
							cont.Add(check)
						} else {
							// Data columns
							label := widget.NewLabel("")
							if rowIdx < len(entryParams) {
								param := entryParams[rowIdx]
								switch id.Col {
								case 1:
									label.SetText(param.DisplayName)
								case 2:
									label.SetText(param.Type)
								case 3:
									// Check RequiredWhen
									required := param.RequiredWhen.OnAdd || param.RequiredWhen.OnRemove || param.RequiredWhen.OnReenrollment
									if required {
										label.SetText("Yes")
									} else {
										label.SetText("No")
									}
								case 4:
									label.SetText(param.DefaultValue)
								}
							}
							cont.Add(label)
						}
					}
					cont.Refresh()
				},
			)
			table.SetColumnWidth(0, 40)  // Checkbox
			table.SetColumnWidth(1, 300) // DisplayName (doubled for longer values)
			table.SetColumnWidth(2, 100) // Type
			table.SetColumnWidth(3, 80)  // Required
			table.SetColumnWidth(4, 150) // DefaultValue

			// Click on data columns to edit
			table.OnSelected = func(id widget.TableCellID) {
				if id.Row > 0 && id.Col > 0 {
					rowIdx := id.Row - 1
					if rowIdx < len(entryParams) {
						showEntryParamDetailDialog(
							entryParams[rowIdx], isInstalled, func(updated api.EntryParameter) {
								entryParams[rowIdx] = updated
								refreshEntryParamsView()
							},
						)
					}
				}
			}

			// Set minimum height for 5 rows - use full width
			tableWithHeight := container.NewStack(
				container.NewGridWrap(fyne.NewSize(100, 180), widget.NewLabel("")), // height spacer
				table,
			)
			entryParamsContainer.Objects = []fyne.CanvasObject{tableWithHeight}
		}
		entryParamsContainer.Refresh()
	}

	// Add entry parameter button
	addEntryParamBtn := widget.NewButton(
		"Add Entry Parameter", func() {
			newParam := api.EntryParameter{
				Name:        "NewParam",
				DisplayName: "New Parameter",
				Type:        "String",
			}
			showEntryParamDetailDialog(
				newParam, isInstalled, func(updated api.EntryParameter) {
					entryParams = append(entryParams, updated)
					refreshEntryParamsView()
				},
			)
		},
	)

	// Remove entry parameter button
	removeEntryParamBtn := widget.NewButton(
		"Remove Selected", func() {
			if len(selectedEntryParams) == 0 {
				dialog.ShowInformation(
					"No Selection",
					"Please select entry parameters to remove using the checkboxes.",
					fyne.CurrentApp().Driver().AllWindows()[0],
				)
				return
			}

			count := len(selectedEntryParams)
			message := fmt.Sprintf("Remove %d selected entry parameter(s)?", count)

			dialog.ShowConfirm(
				"Remove Entry Parameters", message, func(confirmed bool) {
					if !confirmed {
						return
					}
					// Remove selected entry params
					var newEntryParams []api.EntryParameter
					for i, param := range entryParams {
						if !selectedEntryParams[i] {
							newEntryParams = append(newEntryParams, param)
						}
					}
					entryParams = newEntryParams
					refreshEntryParamsView()
				}, fyne.CurrentApp().Driver().AllWindows()[0],
			)
		},
	)

	refreshEntryParamsView()

	// Build updated store type from form
	buildStoreType := func() *api.CertificateStoreType {
		updated := *storeType // Copy

		updated.Name = nameEntry.Text
		updated.ShortName = shortNameEntry.Text
		updated.Capability = capabilityEntry.Text
		updated.LocalStore = localStoreCheck.Checked
		updated.ServerRequired = serverRequiredCheck.Checked
		updated.BlueprintAllowed = blueprintAllowedCheck.Checked
		updated.PowerShell = powerShellCheck.Checked
		updated.PrivateKeyAllowed = privateKeySelect.Selected
		updated.CustomAliasAllowed = customAliasSelect.Selected
		updated.StorePathType = storePathTypeSelect.Selected
		updated.StorePathValue = storePathValueEntry.Text

		// Use the edited properties and entry parameters
		if len(properties) > 0 {
			updated.Properties = &properties
		} else {
			updated.Properties = nil
		}

		if len(entryParams) > 0 {
			updated.EntryParameters = &entryParams
		} else {
			updated.EntryParameters = nil
		}

		// Supported operations
		updated.SupportedOperations = &api.StoreTypeSupportedOperations{
			Add:        addCheck.Checked,
			Create:     createCheck.Checked,
			Discovery:  discoveryCheck.Checked,
			Enrollment: enrollmentCheck.Checked,
			Remove:     removeCheck.Checked,
		}

		// JobProperties is deprecated - always set to nil for POST/PUT operations
		updated.JobProperties = nil

		return &updated
	}

	// Close button
	closeBtn := widget.NewButton("Close", onClose)

	// Save button (only for installed)
	saveBtn := widget.NewButton(
		"Save Changes", func() {
			updated := buildStoreType()
			result, err := storeService.UpdateStoreType(authService, updated)
			if err != nil {
				dialog.ShowError(
					fmt.Errorf("failed to save store type: %w", err),
					fyne.CurrentApp().Driver().AllWindows()[0],
				)
			} else {
				dialog.ShowInformation(
					"Save Successful",
					fmt.Sprintf("Store type '%s' saved successfully (ID: %d)", updated.ShortName, result.StoreType),
					fyne.CurrentApp().Driver().AllWindows()[0],
				)
				if onSuccess != nil {
					onSuccess()
				}
				onClose()
			}
		},
	)
	if !isInstalled {
		saveBtn.Disable()
	}

	// Save as new button - requires unique Name, ShortName, and Capability
	saveAsNewBtn := widget.NewButton(
		"Save as New", func() {
			newNameEntry := widget.NewEntry()
			newNameEntry.SetText(nameEntry.Text + " (Copy)")

			newShortNameEntry := widget.NewEntry()
			newShortNameEntry.SetText(shortNameEntry.Text + "_copy")

			newCapabilityEntry := widget.NewEntry()
			newCapabilityEntry.SetText(capabilityEntry.Text + "_copy")

			formContent := container.NewVBox(
				widget.NewLabel("All three fields must be unique:"),
				widget.NewSeparator(),
				widget.NewLabel("Name"),
				newNameEntry,
				widget.NewLabel("ShortName"),
				newShortNameEntry,
				widget.NewLabel("Capability"),
				newCapabilityEntry,
			)

			d := dialog.NewCustomConfirm(
				"Save as New Store Type",
				"Create", "Cancel",
				formContent,
				func(confirmed bool) {
					if !confirmed {
						return
					}

					// Validate all fields are filled
					if newNameEntry.Text == "" || newShortNameEntry.Text == "" || newCapabilityEntry.Text == "" {
						dialog.ShowError(
							fmt.Errorf("all fields (Name, ShortName, Capability) are required"),
							fyne.CurrentApp().Driver().AllWindows()[0],
						)
						return
					}

					// Check for duplicate ShortName
					isDup, _ := storeService.CheckDuplicateShortName(authService, newShortNameEntry.Text)
					if isDup {
						dialog.ShowError(
							fmt.Errorf("store type with ShortName '%s' already exists", newShortNameEntry.Text),
							fyne.CurrentApp().Driver().AllWindows()[0],
						)
						return
					}

					updated := buildStoreType()
					updated.Name = newNameEntry.Text
					updated.ShortName = newShortNameEntry.Text
					updated.Capability = newCapabilityEntry.Text
					updated.StoreType = 0

					result, err := storeService.CreateStoreType(authService, updated)
					if err != nil {
						dialog.ShowError(
							fmt.Errorf("failed to create store type: %w", err),
							fyne.CurrentApp().Driver().AllWindows()[0],
						)
					} else {
						dialog.ShowInformation(
							"Create Successful",
							fmt.Sprintf(
								"Store type '%s' created successfully with ID: %d",
								newShortNameEntry.Text,
								result.StoreType,
							),
							fyne.CurrentApp().Driver().AllWindows()[0],
						)
						if onSuccess != nil {
							onSuccess()
						}
					}
				}, fyne.CurrentApp().Driver().AllWindows()[0],
			)
			d.Resize(fyne.NewSize(400, 300))
			d.Show()
		},
	)
	if !authService.IsAuthenticated() {
		saveAsNewBtn.Disable()
	}

	// Deploy button (only for catalog)
	deployBtn := widget.NewButton(
		"Deploy", func() {
			updated := buildStoreType()

			isDup, _ := storeService.CheckDuplicateShortName(authService, updated.ShortName)
			if isDup {
				dialog.ShowError(
					fmt.Errorf("store type with ShortName '%s' already exists", updated.ShortName),
					fyne.CurrentApp().Driver().AllWindows()[0],
				)
				return
			}

			result, err := storeService.CreateStoreType(authService, updated)
			if err != nil {
				dialog.ShowError(
					fmt.Errorf("failed to deploy store type: %w", err),
					fyne.CurrentApp().Driver().AllWindows()[0],
				)
			} else {
				dialog.ShowInformation(
					"Deploy Successful",
					fmt.Sprintf(
						"Store type '%s' deployed successfully with ID: %d",
						updated.ShortName,
						result.StoreType,
					),
					fyne.CurrentApp().Driver().AllWindows()[0],
				)
				if onSuccess != nil {
					onSuccess()
				}
			}
		},
	)
	if isInstalled || !authService.IsAuthenticated() {
		deployBtn.Disable()
	}

	// Export button with save and copy options
	exportBtn := widget.NewButton(
		"Export to JSON", func() {
			updated := buildStoreType()
			data, _ := json.MarshalIndent(updated, "", "  ")

			textWidget := widget.NewMultiLineEntry()
			textWidget.SetText(string(data))
			textWidget.Wrapping = fyne.TextWrapWord

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

			fileSaveBtn := widget.NewButton(
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
					saveDialog.SetFileName(updated.ShortName + ".json")
					saveDialog.Show()
				},
			)

			buttonRow := container.NewHBox(copyBtn, fileSaveBtn)
			content := container.NewBorder(nil, buttonRow, nil, nil, container.NewScroll(textWidget))

			d := dialog.NewCustom("Export Store Type", "Close", content, fyne.CurrentApp().Driver().AllWindows()[0])
			d.Resize(fyne.NewSize(600, 500))
			d.Show()
		},
	)

	// Button row
	buttons := container.NewHBox(
		closeBtn,
		saveBtn,
		saveAsNewBtn,
		deployBtn,
		exportBtn,
	)

	// Properties section with add/remove buttons
	propertiesSection := container.NewVBox(
		widget.NewLabelWithStyle("Properties", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(addPropertyBtn, removePropertyBtn),
		propertiesContainer,
	)

	// Entry parameters section with add/remove buttons
	entryParamsSection := container.NewVBox(
		widget.NewLabelWithStyle("Entry Parameters", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(addEntryParamBtn, removeEntryParamBtn),
		entryParamsContainer,
	)

	// Form layout - only show non-null fields
	formItems := []fyne.CanvasObject{
		title,
	}

	// Only show ID for installed types
	if isInstalled && storeType.StoreType > 0 {
		formItems = append(formItems, idLabel)
	}

	formItems = append(
		formItems,
		widget.NewSeparator(),
		widget.NewLabel("Name"),
		nameEntry,
		container.NewGridWithColumns(
			2,
			container.NewVBox(widget.NewLabel("ShortName"), shortNameEntry),
			container.NewVBox(widget.NewLabel("Capability"), capabilityEntry),
		),
		widget.NewSeparator(),
		widget.NewLabel("Options"),
		container.NewGridWithColumns(
			2,
			localStoreCheck,
			serverRequiredCheck,
		),
		container.NewGridWithColumns(
			2,
			blueprintAllowedCheck,
			powerShellCheck,
		),
		container.NewGridWithColumns(
			2,
			container.NewVBox(widget.NewLabel("Private Key Allowed"), privateKeySelect),
			container.NewVBox(widget.NewLabel("Custom Alias Allowed"), customAliasSelect),
		),
		widget.NewSeparator(),
		widget.NewLabel("Store Path"),
		container.NewGridWithColumns(
			2,
			container.NewVBox(widget.NewLabel("Store Path Type"), storePathTypeSelect),
			container.NewVBox(widget.NewLabel("Store Path Value"), storePathValueEntry),
		),
		widget.NewSeparator(),
		widget.NewLabel("Supported Operations"),
		container.NewHBox(addCheck, createCheck, discoveryCheck, enrollmentCheck, removeCheck),
		widget.NewSeparator(),
		propertiesSection,
		widget.NewSeparator(),
		entryParamsSection,
		widget.NewSeparator(),
		buttons,
	)

	form := container.NewVBox(formItems...)

	return container.NewPadded(container.NewScroll(form))
}

// showPropertyDetailDialog shows a dialog to edit a property
func showPropertyDetailDialog(
	prop api.StoreTypePropertyDefinition,
	isInstalled bool,
	onSave func(api.StoreTypePropertyDefinition),
) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(prop.Name)

	displayNameEntry := widget.NewEntry()
	displayNameEntry.SetText(prop.DisplayName)

	typeSelect := widget.NewSelect([]string{"String", "Bool", "Secret", "MultipleChoice", "PamProviderParameter"}, nil)
	if prop.Type != "" {
		typeSelect.SetSelected(prop.Type)
	} else {
		typeSelect.SetSelected("String")
	}

	requiredCheck := widget.NewCheck("Required", nil)
	requiredCheck.SetChecked(prop.Required)

	defaultValueEntry := widget.NewEntry()
	if prop.DefaultValue != nil {
		defaultValueEntry.SetText(fmt.Sprintf("%v", prop.DefaultValue))
	}

	dependsOnEntry := widget.NewEntry()
	if prop.DependsOn != nil {
		dependsOnEntry.SetText(fmt.Sprintf("%v", prop.DependsOn))
	}

	// Show StoreTypeId only if it's non-zero and isInstalled
	storeTypeIdLabel := widget.NewLabel("")
	if isInstalled && prop.StoreTypeID > 0 {
		storeTypeIdLabel.SetText(fmt.Sprintf("StoreTypeId: %d (read-only)", prop.StoreTypeID))
	}

	content := container.NewVBox(
		storeTypeIdLabel,
		widget.NewLabel("Name"),
		nameEntry,
		widget.NewLabel("Display Name"),
		displayNameEntry,
		widget.NewLabel("Type"),
		typeSelect,
		requiredCheck,
		widget.NewLabel("Default Value"),
		defaultValueEntry,
		widget.NewLabel("Depends On"),
		dependsOnEntry,
	)

	d := dialog.NewCustomConfirm(
		"Edit Property", "Save", "Cancel",
		container.NewScroll(content),
		func(confirmed bool) {
			if !confirmed {
				return
			}

			updated := prop
			updated.Name = nameEntry.Text
			updated.DisplayName = displayNameEntry.Text
			updated.Type = typeSelect.Selected
			updated.Required = requiredCheck.Checked
			if defaultValueEntry.Text != "" {
				updated.DefaultValue = defaultValueEntry.Text
			} else {
				updated.DefaultValue = nil
			}
			if dependsOnEntry.Text != "" {
				updated.DependsOn = dependsOnEntry.Text
			} else {
				updated.DependsOn = nil
			}

			onSave(updated)
		}, fyne.CurrentApp().Driver().AllWindows()[0],
	)
	d.Resize(fyne.NewSize(450, 400))
	d.Show()
}

// showEntryParamDetailDialog shows a dialog to edit an entry parameter
func showEntryParamDetailDialog(param api.EntryParameter, isInstalled bool, onSave func(api.EntryParameter)) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(param.Name)

	displayNameEntry := widget.NewEntry()
	displayNameEntry.SetText(param.DisplayName)

	typeSelect := widget.NewSelect([]string{"String", "Bool", "Secret", "MultipleChoice"}, nil)
	if param.Type != "" {
		typeSelect.SetSelected(param.Type)
	} else {
		typeSelect.SetSelected("String")
	}

	defaultValueEntry := widget.NewEntry()
	defaultValueEntry.SetText(param.DefaultValue)

	dependsOnEntry := widget.NewEntry()
	dependsOnEntry.SetText(param.DependsOn)

	optionsEntry := widget.NewEntry()
	optionsEntry.SetText(param.Options)

	// RequiredWhen checkboxes
	hasPrivateKeyCheck := widget.NewCheck("Has Private Key", nil)
	hasPrivateKeyCheck.SetChecked(param.RequiredWhen.HasPrivateKey)
	onAddCheck := widget.NewCheck("On Add", nil)
	onAddCheck.SetChecked(param.RequiredWhen.OnAdd)
	onRemoveCheck := widget.NewCheck("On Remove", nil)
	onRemoveCheck.SetChecked(param.RequiredWhen.OnRemove)
	onReenrollmentCheck := widget.NewCheck("On Reenrollment", nil)
	onReenrollmentCheck.SetChecked(param.RequiredWhen.OnReenrollment)

	// Show StoreTypeId only if it's non-zero and isInstalled
	storeTypeIdLabel := widget.NewLabel("")
	if isInstalled && param.StoreTypeId > 0 {
		storeTypeIdLabel.SetText(fmt.Sprintf("StoreTypeId: %d (read-only)", param.StoreTypeId))
	}

	content := container.NewVBox(
		storeTypeIdLabel,
		widget.NewLabel("Name"),
		nameEntry,
		widget.NewLabel("Display Name"),
		displayNameEntry,
		widget.NewLabel("Type"),
		typeSelect,
		widget.NewLabel("Default Value"),
		defaultValueEntry,
		widget.NewLabel("Depends On"),
		dependsOnEntry,
		widget.NewLabel("Options"),
		optionsEntry,
		widget.NewSeparator(),
		widget.NewLabel("Required When:"),
		container.NewGridWithColumns(2, hasPrivateKeyCheck, onAddCheck),
		container.NewGridWithColumns(2, onRemoveCheck, onReenrollmentCheck),
	)

	d := dialog.NewCustomConfirm(
		"Edit Entry Parameter", "Save", "Cancel",
		container.NewScroll(content),
		func(confirmed bool) {
			if !confirmed {
				return
			}

			updated := param
			updated.Name = nameEntry.Text
			updated.DisplayName = displayNameEntry.Text
			updated.Type = typeSelect.Selected
			updated.DefaultValue = defaultValueEntry.Text
			updated.DependsOn = dependsOnEntry.Text
			updated.Options = optionsEntry.Text
			updated.RequiredWhen.HasPrivateKey = hasPrivateKeyCheck.Checked
			updated.RequiredWhen.OnAdd = onAddCheck.Checked
			updated.RequiredWhen.OnRemove = onRemoveCheck.Checked
			updated.RequiredWhen.OnReenrollment = onReenrollmentCheck.Checked

			onSave(updated)
		}, fyne.CurrentApp().Driver().AllWindows()[0],
	)
	d.Resize(fyne.NewSize(450, 500))
	d.Show()
}

// Helper to get int from interface
func getIntFromInterface(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}
