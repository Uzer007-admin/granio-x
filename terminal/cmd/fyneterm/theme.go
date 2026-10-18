//go:generate fyne bundle -package main -o fonts.go font

package main

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"
)

type termTheme struct {
	granio.Theme

	fontSize float32
}

func newTermTheme() *termTheme {
	return &termTheme{Theme: theme.DefaultTheme(), fontSize: 12}
}

// Color fixes a bug < 2.1 where theme.DarkTheme() would not override user preference.
func (t *termTheme) Color(n granio.ThemeColorName, v granio.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground, theme.ColorNameForeground:
		return t.Theme.Color(n, v)
	}
	return t.Theme.Color(n, theme.VariantDark)
}

func (t *termTheme) Size(n granio.ThemeSizeName) float32 {
	if n == theme.SizeNameText {
		return t.fontSize
	}

	return t.Theme.Size(n)
}

func (t *termTheme) Font(s granio.TextStyle) granio.Resource {
	if !s.Monospace {
		return t.Theme.Font(s)
	}

	if s.Bold {
		if s.Italic {
			return resourceHackBoldItalicTtf
		}
		return resourceHackBoldTtf
	} else if s.Italic {
		return resourceHackItalicTtf
	}

	return resourceHackRegularTtf
}
