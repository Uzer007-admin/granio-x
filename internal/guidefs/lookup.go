package guidefs

import (
	"reflect"

	"github.com/Uzer007-admin/granio/v2"
)

// Lookup returns the [WidgetInfo] for the given widget type.
func Lookup(clazz string) *WidgetInfo {
	if match, ok := Widgets[clazz]; ok {
		return &match
	}
	if match, ok := Collections[clazz]; ok {
		return &match
	}
	if match, ok := Containers[clazz]; ok {
		return &match
	}
	if match, ok := Graphics[clazz]; ok {
		return &match
	}

	return nil
}

// TypeName returns the unique name for this object, i.e. "canvas.Line" or "xWidget.Map"
func TypeName(o granio.CanvasObject) string {
	class := reflect.TypeOf(o).String()
	info := Lookup(class)
	if info == nil {
		class = "*xW" + class[2:] // xWidget imports
	}

	return class
}
