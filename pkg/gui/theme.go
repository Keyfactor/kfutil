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
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// KeyfactorTheme is a custom theme for the kfutil GUI
type KeyfactorTheme struct{}

// NewKeyfactorTheme creates a new Keyfactor-branded theme
func NewKeyfactorTheme() fyne.Theme {
	return &KeyfactorTheme{}
}

// Color returns the color for the specified theme color name
func (t *KeyfactorTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0, G: 122, B: 204, A: 255} // Keyfactor blue
	case theme.ColorNameButton:
		return color.NRGBA{R: 0, G: 122, B: 204, A: 255}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0, G: 150, B: 230, A: 255}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0, G: 122, B: 204, A: 100}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font returns the font resource for the specified text style
func (t *KeyfactorTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns the icon resource for the specified icon name
func (t *KeyfactorTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the size for the specified size name
func (t *KeyfactorTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 4
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 20
	case theme.SizeNameSubHeadingText:
		return 16
	default:
		return theme.DefaultTheme().Size(name)
	}
}
