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

package widgets

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// StoreCard represents a card widget for displaying store type information
type StoreCard struct {
	widget.BaseWidget

	ID          int
	Name        string
	ShortName   string
	Description string
	Capability  string
	Version     string

	OnTapped       func()
	OnDoubleTapped func()
}

// NewStoreCard creates a new store card widget
func NewStoreCard(id int, name, shortName, description, capability string) *StoreCard {
	card := &StoreCard{
		ID:          id,
		Name:        name,
		ShortName:   shortName,
		Description: description,
		Capability:  capability,
		Version:     "N/A",
	}
	card.ExtendBaseWidget(card)
	return card
}

// CreateRenderer implements fyne.Widget
func (c *StoreCard) CreateRenderer() fyne.WidgetRenderer {
	// Try to load icon
	icon := c.loadIcon()

	// Create labels
	nameLabel := widget.NewLabelWithStyle(c.Name, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	nameLabel.Wrapping = fyne.TextWrapWord
	shortNameLabel := widget.NewLabel(fmt.Sprintf("ShortName: %s", c.ShortName))
	shortNameLabel.Alignment = fyne.TextAlignCenter
	capabilityLabel := widget.NewLabel(fmt.Sprintf("Capability: %s", c.Capability))
	capabilityLabel.Alignment = fyne.TextAlignCenter

	var idLabel *widget.Label
	if c.ID > 0 {
		idLabel = widget.NewLabel(fmt.Sprintf("ID: %d", c.ID))
		idLabel.Alignment = fyne.TextAlignCenter
	} else {
		idLabel = widget.NewLabel("")
	}

	// Layout: Name at top, then icon centered below, then other details
	content := container.NewVBox(
		nameLabel,
		container.NewCenter(icon),
		shortNameLabel,
		capabilityLabel,
		idLabel,
	)

	// Add border/background
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.CornerRadius = 8

	return &storeCardRenderer{
		card:    c,
		bg:      bg,
		content: content,
	}
}

// loadIcon attempts to load the store type icon
func (c *StoreCard) loadIcon() fyne.CanvasObject {
	// Try multiple icon paths for the specific store type
	iconPaths := []string{
		filepath.Join("icons", fmt.Sprintf("storetype_%s.png", c.ShortName)),
		filepath.Join("icons", fmt.Sprintf("%s.png", c.ShortName)),
		filepath.Join("pkg", "gui", "assets", "icons", fmt.Sprintf("storetype_%s.png", c.ShortName)),
	}

	for _, path := range iconPaths {
		if _, err := os.Stat(path); err == nil {
			img := canvas.NewImageFromFile(path)
			img.SetMinSize(fyne.NewSize(64, 64))
			img.FillMode = canvas.ImageFillContain
			return img
		}
	}

	// Try default icon paths
	defaultPaths := []string{
		filepath.Join("icons", "storetype_default.png"),
		filepath.Join("pkg", "gui", "assets", "icons", "storetype_default.png"),
	}

	for _, path := range defaultPaths {
		if _, err := os.Stat(path); err == nil {
			img := canvas.NewImageFromFile(path)
			img.SetMinSize(fyne.NewSize(64, 64))
			img.FillMode = canvas.ImageFillContain
			return img
		}
	}

	// Fallback to theme icon if no default image found
	icon := widget.NewIcon(theme.DocumentIcon())
	icon.Resize(fyne.NewSize(64, 64))
	return icon
}

// Tapped implements fyne.Tappable
func (c *StoreCard) Tapped(*fyne.PointEvent) {
	if c.OnTapped != nil {
		c.OnTapped()
	}
}

// DoubleTapped implements fyne.DoubleTappable
func (c *StoreCard) DoubleTapped(*fyne.PointEvent) {
	if c.OnDoubleTapped != nil {
		c.OnDoubleTapped()
	}
}

// storeCardRenderer implements fyne.WidgetRenderer
type storeCardRenderer struct {
	card    *StoreCard
	bg      *canvas.Rectangle
	content *fyne.Container
}

func (r *storeCardRenderer) Destroy() {}

func (r *storeCardRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.content.Resize(size)
}

func (r *storeCardRenderer) MinSize() fyne.Size {
	return r.content.MinSize()
}

func (r *storeCardRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.content}
}

func (r *storeCardRenderer) Refresh() {
	r.bg.Refresh()
	r.content.Refresh()
}

// StoreCardList creates a scrollable list of store cards
func StoreCardList(cards []*StoreCard) fyne.CanvasObject {
	list := container.NewVBox()
	for _, card := range cards {
		list.Add(container.NewPadded(card))
	}
	return container.NewScroll(list)
}
