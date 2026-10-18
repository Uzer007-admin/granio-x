package cloud

import (
	"fmt"

	"github.com/Uzer007-admin/granio-x/cloud/internal/settings"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/dialog"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func ShowSettings(a granio.App, w granio.Window) {
	prov := a.CloudProvider()
	if prov == nil {
		showChoice(a, w)
		return
	}

	config := widget.NewButton("Configure", func() {
		go func() {
			str, err := prov.(Configurable).Configure(a, w)
			if err != nil {
				dialog.ShowError(err, w)
			} else {
				if dis, ok := a.CloudProvider().(Disconnectable); ok {
					dis.Disconnect()
				}

				settings.SetProviderConfig(str)
				prov.(Configurable).SetConfig(str)
				a.SetCloudProvider(prov) // we don't use setCloud here as we have a new config
			}
		}()
	})
	if _, ok := prov.(Configurable); !ok {
		config.Hide()
	}
	current := widget.NewLabel(cloudText(prov))
	ch := make(chan granio.Settings)
	a.Settings().AddChangeListener(ch)
	go func() {
		for range ch {
			current.SetText(cloudText(prov))
		}
	}()

	var d dialog.Dialog
	d = dialog.NewCustomConfirm("Cloud configuration", "Change Provider", "Cancel",
		container.NewBorder(nil, nil, nil, config, current),
		func(change bool) {
			if !change {
				return
			}

			d.Hide()
			showChoice(a, w)
		}, w)
	d.Show()
}

func showChoice(a granio.App, w granio.Window) {
	var selected granio.CloudProvider
	ui := widget.NewList(func() int {
		return len(providers)
	},
		func() granio.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(nil),
				widget.NewLabel(""))
		},
		func(id widget.ListItemID, o granio.CanvasObject) {
			p := providers[id]
			o.(*granio.Container).Objects[0].(*widget.Icon).SetResource(p.ProviderIcon())
			o.(*granio.Container).Objects[1].(*widget.Label).SetText(p.ProviderName())
		})
	ui.OnSelected = func(id widget.ListItemID) {
		selected = providers[id]
	}

	dialog.ShowCustomConfirm("Choose Cloud provider", "Enable", "Cancel",
		container.NewVBox(widget.NewLabel("Choose one of the providers below to sync\nyour preferences to the cloud."), ui),
		func(ok bool) {
			if !ok || selected == nil {
				return
			}

			chooseProvider(a, selected)
		}, w)
}

func chooseProvider(a granio.App, p granio.CloudProvider) {
	settings.SetCurrentProviderName(p.ProviderName())
	setCloud(p, a)
}

func cloudText(p granio.CloudProvider) string {
	if p == nil {
		return "No Cloud configured"
	}

	return fmt.Sprintf("Using %s cloud provider", p.ProviderName())
}
