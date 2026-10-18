//go:build !ios && !android && !mobile
// +build !ios,!android,!mobile

// This file is copied from fyne.io/fyne/app/preferences_other.go to avoid an import loop

package internal

func (p *preferences) watch() {
	watchFile(p.file.Path(), func() {
		p.prefLock.RLock()
		shouldIgnoreChange := p.savedRecently
		p.prefLock.RUnlock()
		if shouldIgnoreChange {
			return
		}

		p.load()
	})
}
