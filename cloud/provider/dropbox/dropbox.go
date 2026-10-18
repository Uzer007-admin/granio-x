//go:generate granio bundle -o bundled.go -package dropbox Dropbox.svg

package dropbox

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/Uzer007-admin/granio-x/cloud/internal/settings"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/dialog"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/Uzer007-admin/granio/v2/theme"
)

type dropbox struct {
	store  granio.URI
	config string
	prefs  granio.Preferences
}

func (d *dropbox) SetConfig(str string) {
	d.config = str
}

func (d *dropbox) Cleanup(a granio.App) {
}

func (d *dropbox) Configure(a granio.App, w granio.Window) (data string, err error) {
	if granio.CurrentDevice().IsMobile() {
		data, err = d.mobileConfig(a)
	} else {
		data, err = d.desktopConfig(w)
	}

	if err == nil {
		d.config = data
	}
	return data, err
}

func (d *dropbox) Disconnect() {
	d.prefs.(interface{ Disconnect() }).Disconnect()
}

func (d *dropbox) ProviderDescription() string {
	return "Preferences and files are stored into the user's Dropbox storage.\n" +
		"You need to have Dropbox installed and running for this provider to work."
}

func (d *dropbox) ProviderIcon() granio.Resource {
	return theme.NewThemedResource(resourceDropboxSvg)
}

func (d *dropbox) ProviderName() string {
	return "Dropbox"
}

func (d *dropbox) Setup(a granio.App) error {
	if !granio.CurrentDevice().IsMobile() {
		return nil
	}
	if d.config != "" {
		d.store = storage.NewFileURI(d.config)
		// TODO validate
		return nil
	}
	data, err := d.mobileConfig(a)
	settings.SetProviderConfig(data)

	return err
}

func (d *dropbox) desktopConfig(w granio.Window) (string, error) {
	err := make(chan error)
	data := ""

	ask := dialog.NewConfirm("Set Dropbox location",
		"On the next screen please choose the the fynesync\nfolder, or whichever you would prefer",
		func(ok bool) {
			if !ok {
				err <- errors.New("user cancelled setup")
				return
			}

			dialog.ShowFolderOpen(func(list granio.ListableURI, e2 error) {
				if e2 != nil {
					err <- e2
					return
				}

				if list == nil {
					err <- errors.New("no folder was chosen")
					return
				}

				// TODO check content?...
				rootDir := list.Path()
				home, _ := os.UserHomeDir()
				if len(rootDir) > len(home) && strings.Index(rootDir, home) == 0 {
					rootDir = "~" + rootDir[len(home):]
				}

				data = rootDir
				err <- nil
			}, w)
		}, w)

	ask.SetConfirmText("OK")
	ask.SetDismissText("Cancel")
	ask.Show()
	return data, <-err
}

func (d *dropbox) mobileConfig(a granio.App) (string, error) {
	err := make(chan error)
	data := ""
	for len(a.Driver().AllWindows()) == 0 {
		time.Sleep(time.Millisecond * 100)
	}
	win := a.Driver().AllWindows()[0]

	ask := dialog.NewConfirm("Locate Dropbox files",
		"On the next screen please load the Dropbox file\nfynesync/"+a.UniqueID()+"/preferences.json",
		func(ok bool) {
			if !ok {
				err <- errors.New("user cancelled setup")
				return
			}

			dialog.ShowFileOpen(func(read granio.URIReadCloser, e2 error) {
				if e2 != nil {
					err <- e2
					return
				}

				if read == nil {
					err <- errors.New("no file was chosen")
					return
				}

				// TODO check content?...
				d.store = read.URI()
				data = d.store.Path()
				err <- nil
			}, win)
		}, win)

	ask.SetConfirmText("OK")
	ask.SetDismissText("Cancel")
	ask.Show()
	return data, <-err
}

func NewProvider() granio.CloudProvider {
	return &dropbox{}
}
