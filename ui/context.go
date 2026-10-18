package refyne

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"

	"github.com/Uzer007-admin/granio-ui/internal/guidefs"
)

// Context defines a graphical builder context that contains metadata and theme information.
type Context = guidefs.Context

type context struct {
	meta map[granio.CanvasObject]map[string]string
	attr map[granio.CanvasObject][]string
	root granio.CanvasObject
}

// DefaultContext returns a simple context with an empty metadata map that will
// defer to the default theme for theming information.
func DefaultContext() Context {
	return &context{
		meta: make(map[granio.CanvasObject]map[string]string),
		attr: make(map[granio.CanvasObject][]string),
	}
}

func (c *context) Metadata() map[granio.CanvasObject]map[string]string {
	return c.meta
}

func (c *context) Attrs() map[granio.CanvasObject][]string {
	return c.attr
}

func (c *context) Root() granio.CanvasObject {
	return c.root
}

func (c *context) Theme() granio.Theme {
	return theme.DefaultTheme()
}

// ContainerOf returns the parent of the given CanvasObject, in the specified Context.
// The returned object will be in the tree descended from `c.Root()`, or nil.
func ContainerOf(obj granio.CanvasObject, c Context) granio.CanvasObject {
	return containerOf(obj, c.Root())
}

func containerOf(obj granio.CanvasObject, root granio.CanvasObject) granio.CanvasObject {
	switch c := root.(type) {
	case *granio.Container:
		for _, w := range c.Objects {
			if w == obj {
				return root
			}

			parent := containerOf(obj, w)
			if parent != nil {
				return parent
			}
		}

	case granio.Widget:
		drops := DropZonesForObject(root)

		for _, child := range drops {
			if child == obj {
				return root
			}

			parent := containerOf(obj, child)
			if parent != nil {
				return parent
			}
		}
	}

	return nil
}
