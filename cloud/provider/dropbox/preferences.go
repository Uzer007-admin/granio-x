package dropbox

import (
	"os"
	"path/filepath"

	"github.com/Uzer007-admin/granio-x/cloud/internal"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/storage"
)

func (d *dropbox) CloudPreferences(a granio.App) granio.Preferences {
	var p granio.Preferences
	if d.store != nil {
		p = internal.NewPreferences(a, d.store)
	} else {
		filePath := filepath.Join(d.dropboxLocation(), a.UniqueID(), "preferences.json")
		p = internal.NewPreferences(a, storage.NewFileURI(filePath))
	}

	d.prefs = p
	return p
}

func (d *dropbox) dropboxLocation() string {
	home, _ := os.UserHomeDir()
	dir := d.config
	if dir == "" {
		return filepath.Join(home, "Dropbox", "fynesync")
	}
	if dir[0] == '~' {
		dir = home + dir[1:]
	}
	return dir
}
