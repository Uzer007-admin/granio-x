// Package cloud provides various cloud provider implementations and utilities to add the cloud services
// into your Fyne app. Developers can choose to load a specific provider, or they can present a configuration
// user interface allowing the end-user to choose the provider they wish to use.
//
// A simple usage where an app uses AWS for cloud provision may look like this:
//
//	package main
//
//	import (
//		"github.com/Uzer007-admin/granio-x/cloud/provider/aws"
//		"github.com/Uzer007-admin/granio/v2/app"
//		"github.com/Uzer007-admin/granio/v2/widget"
//	)
//
//	func main() {
//		a := app.New()
//		a.SetCloudProvider(aws.NewProvider()) // if aws provider existed ;)
//		w := a.NewWindow("Cloud")
//
//		w.SetContent(widget.NewLabel("Add content here"))
//		w.ShowAndRun()
//	}
//
// Alternatively to allow the user to choose a cloud provider for their storage etc use:
//
//	package main
//
//	import (
//		"github.com/Uzer007-admin/granio-x/cloud"
//		"github.com/Uzer007-admin/granio/v2/app"
//		"github.com/Uzer007-admin/granio/v2/widget"
//	)
//
//	func main() {
//		a := app.New()
//		cloud.Enable(a)
//		w := a.NewWindow("Cloud")
//
//		w.SetContent(widget.NewButton("Choose cloud provider", func() {
//			cloud.ShowSettings(a, w)
//		}))
//		w.ShowAndRun()
//	}
package cloud // import "github.com/Uzer007-admin/granio-x/cloud"

import (
	"github.com/Uzer007-admin/granio-x/cloud/internal/settings"
	"github.com/Uzer007-admin/granio-x/cloud/provider/dropbox"
	"github.com/Uzer007-admin/granio/v2"
)

// Disconnectable interface describes a cloud provider that can respond to being disconnected.
// This is typically used before a replacement provider is loaded.
type Disconnectable interface {
	// Disconnect the cloud provider from application and ignore future events.
	Disconnect()
}

var providers []granio.CloudProvider

func Enable(a granio.App) {
	setCloud(lookupConfiguredProvider(), a)
}

func Register(p granio.CloudProvider) {
	providers = append(providers, p)
}

func init() {
	Register(dropbox.NewProvider())
}

func lookupConfiguredProvider() granio.CloudProvider {
	name := currentProviderName()

	for _, p := range providers {
		if p.ProviderName() == name {
			return p
		}
	}
	return nil
}

func setCloud(p granio.CloudProvider, a granio.App) {
	if dis, ok := a.CloudProvider().(Disconnectable); ok {
		dis.Disconnect()
	}
	if config, ok := p.(Configurable); ok {
		schema := settings.Load()
		config.SetConfig(schema.CloudConfig)
	}

	a.SetCloudProvider(p)
}
