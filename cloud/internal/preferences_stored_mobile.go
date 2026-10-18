//go:build android || ios || mobile || nacl
// +build android ios mobile nacl

// This file is copied from fyne.io/fyne/app/preferences_mobile.go to avoid an import loop

package internal

func (p *preferences) watch() {
	// no-op as we are in mobile simulation mode
}
