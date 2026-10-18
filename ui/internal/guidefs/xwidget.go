package guidefs

import (
	"fmt"
	"strconv"

	xWidget "github.com/Uzer007-admin/granio-x/widget"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func initMapXWidget() WidgetInfo {
	return WidgetInfo{
		Name: "Map",
		Create: func(Context) granio.CanvasObject {
			return xWidget.NewMapWithOptions(
				xWidget.AtLatLon(55.9486, -3.1999),
				xWidget.AtZoomLevel(9))
		},
		Edit: func(obj granio.CanvasObject, c Context, _ func([]*widget.FormItem), onchanged func()) []*widget.FormItem {
			m := obj.(*xWidget.Map)
			props := c.Metadata()[obj]
			if props["lat"] == "" {
				props["lat"] = "55.9486"
			}
			if props["lon"] == "" {
				props["lon"] = "-3.1999"
			}

			repan := func() {
				lon, _ := strconv.ParseFloat(props["lon"], 64)
				lat, _ := strconv.ParseFloat(props["lat"], 64)
				m.PanToLatLon(lat, lon)
			}
			repan()

			lat := widget.NewEntry()
			lat.SetText(props["lat"])
			lat.OnChanged = func(s string) {
				props["lat"] = s
				repan()
				onchanged()
			}
			lon := widget.NewEntry()
			lon.SetText(props["lon"])
			lon.OnChanged = func(s string) {
				props["lon"] = s
				repan()
				onchanged()
			}

			if props["zoom"] == "" {
				props["zoom"] = "9"
			}
			rezoom := func(s string) {
				if i, err := strconv.Atoi(s); err == nil {
					m.Zoom(i)
				}
			}
			rezoom(props["zoom"])
			zoom := widget.NewEntry()
			zoom.SetText(props["zoom"])
			zoom.OnChanged = func(s string) {
				props["zoom"] = s
				rezoom(s)
				onchanged()
			}

			return []*widget.FormItem{
				widget.NewFormItem("Latitude", lat),
				widget.NewFormItem("Longitude", lon),
				widget.NewFormItem("Zoom", zoom),
			}
		},
		Gostring: func(obj granio.CanvasObject, c Context, defs map[string]string) string {
			props := c.Metadata()[obj]

			lat, _ := strconv.ParseFloat(props["lat"], 64)
			lon, _ := strconv.ParseFloat(props["lon"], 64)
			zoom, _ := strconv.Atoi(props["zoom"])
			return widgetRef(obj, c, defs,
				fmt.Sprintf("xWidget.NewMapWithOptions(xWidget.AtLatLon(%f, %f), xWidget.AtZoomLevel(%d))", lat, lon, zoom))
		},
		Packages: func(obj granio.CanvasObject, _ Context) []string {
			return []string{"xWidget"}
		},
	}
}
