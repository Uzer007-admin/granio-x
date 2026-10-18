package cloud

import (
	"github.com/Uzer007-admin/granio-x/cloud/internal/settings"
	"github.com/Uzer007-admin/granio/v2"
)

// Configurable interface describes the functions required for a cloud provider to store configuration.
type Configurable interface {
	// Configure requests that the cloud provider show some configuration options as a dialog on the specified window.
	// It returns a serialised configuration or an error.
	Configure(granio.App, granio.Window) (string, error)
	// SetConfig is used to apply a previous configuration to this provider.
	SetConfig(string)
}

func currentProviderName() string {
	store := settings.Load()
	if store == nil {
		return ""
	}
	return store.CloudName
}
