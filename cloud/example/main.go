package main

import (
	"fmt"

	"github.com/Uzer007-admin/granio-cloud"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/data/binding"
	"github.com/Uzer007-admin/granio/v2/layout"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func main() {
	a := app.NewWithID("io.fyne.cloud.example")
	w := a.NewWindow("Cloud")

	current := widget.NewLabel("No configured provider")
	ch := make(chan granio.Settings)
	a.Settings().AddChangeListener(ch)
	cloud.Enable(a)

	choose := widget.NewButton("Cloud settings", func() {
		cloud.ShowSettings(a, w)
	})
	testEntry := widget.NewEntryWithData(binding.BindPreferenceString("test", a.Preferences()))
	testEntry.Validator = nil

	go func() {
		for range ch {
			if a.CloudProvider() == nil {
				return
			}
			current.SetText(fmt.Sprint("Current cloud: ", a.CloudProvider().ProviderName()))
		}
	}()

	w.SetContent(container.NewVBox(layout.NewSpacer(), current,
		container.NewBorder(nil, nil, widget.NewLabel("Test store value"), nil, testEntry),
		layout.NewSpacer(), choose))
	w.ShowAndRun()
}
