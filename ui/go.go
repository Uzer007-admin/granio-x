package refyne

import (
	"fmt"
	"go/format"
	"io"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/Uzer007-admin/granio-x/ui/internal/guidefs"
	"github.com/Uzer007-admin/granio-x/ui/internal/tools"
	"github.com/Uzer007-admin/granio/v2/container"

	"github.com/Uzer007-admin/granio/v2"
)

// ExportGo generates a full Go package for the given object and writes it to the provided file handle
func ExportGo(obj granio.CanvasObject, d Context, name string, w io.Writer) error {
	guidefs.InitOnce()

	tools.VarNames.Reset()

	packagesList := packagesRequired(obj, d)

	// Really this needs to be a full dependency analysis but for now a simple sort of widgets before containers may work
	varListWidgets, varListContainers := varsRequired(obj, d)
	sort.Strings(varListWidgets)
	sort.Strings(varListContainers)

	code := exportCode(packagesList, append(varListWidgets, varListContainers...), obj, d, name)

	_, err := w.Write([]byte(code))
	return err
}

// ExportGoPreview generates a preview version of the Go code with a `main()` method for the given object and writes it to the file handle
func ExportGoPreview(obj granio.CanvasObject, d Context, w io.Writer) error {
	guidefs.InitOnce()

	packagesList := packagesRequired(obj, d)
	packagesList = append(packagesList, "app")

	// Really this needs to be a full dependency analysis but for now a simple sort of widgets before containers may work
	varListWidgets, varListContainers := varsRequired(obj, d)
	sort.Strings(varListWidgets)
	sort.Strings(varListContainers)

	code := exportCode(packagesList, append(varListWidgets, varListContainers...), obj, d, "main")

	code += `
func main() {
	a := app.NewWithID("dummy-` + strconv.Itoa(rand.Int()) + `")
	w := a.NewWindow("Hello")
	gui := newGUI()
	gui.win = w
	w.SetContent(gui.makeUI())
	w.ShowAndRun()
}
`
	_, err := w.Write([]byte(code))

	return err
}

func countContainers(obj granio.CanvasObject) int {
	var children []granio.CanvasObject
	switch c := obj.(type) {
	case *granio.Container:
		children = c.Objects
	case *container.AppTabs:
		children = make([]granio.CanvasObject, len(c.Items))
		for i, c := range c.Items {
			children[i] = c.Content
		}
	case *container.DocTabs:
		children = make([]granio.CanvasObject, len(c.Items))
		for i, c := range c.Items {
			children[i] = c.Content
		}
	case *container.Scroll:
		children = []granio.CanvasObject{c.Content}
	default:
		return 0
	}

	r := 1
	for _, obj := range children {
		r += countContainers(obj)
	}
	return r
}

func stringMapToSlice(m map[string]string, fn func(a, b string) bool) []string {
	keys := make([]string, len(m))
	n := 0
	for key := range m {
		keys[n] = key
		n++
	}
	sort.Slice(keys, func(i, j int) bool {
		return fn(keys[i], keys[j])
	})
	r := make([]string, len(keys))
	for i, key := range keys {
		r[i] = m[key]
	}
	return r
}

func exportCode(pkgs, vars []string, obj granio.CanvasObject, d Context, name string) string {
	for i := 0; i < len(pkgs); i++ {
		if pkgs[i] == "xWidget" {
			pkgs[i] = `xWidget	"github.com/Uzer007-admin/granio-x/widget"`

			continue
		}

		if pkgs[i] != "fmt" && !strings.Contains(pkgs[i], "/") {
			pkgs[i] = "github.com/Uzer007-admin/granio/v2/" + pkgs[i]
		}

		pkgs[i] = fmt.Sprintf(`	"%s"`, pkgs[i])
	}

	battrs := make(map[granio.CanvasObject][]string)
	for obj, attrs := range d.Attrs() {
		battrs[obj] = attrs
	}

	deps := make(map[string]int)
	for obj, props := range d.Metadata() {
		name := props["name"]

		deps[name] = countContainers(obj)
	}

	defs := make(map[string]string)

	_, clazz := getTypeOf(obj)
	main := guidefs.GoString(clazz, obj, d, defs)
	setupBeforeMap := make(map[string]string)
	setupAfterMap := make(map[string]string)

	genids := make(map[string]bool)
	for _, p := range d.Metadata() {
		if g, ok := p["name-is-generated"]; ok && g == "1" {
			name := p["name"]
			genids[name] = true
			if deps[name] > 0 {
				setupAfterMap[name] = name + " := " + defs[name]
			} else {
				setupBeforeMap[name] = name + " := " + defs[name]
			}
		}
	}

	for _, key := range vars {
		name := strings.Split(key, " ")[0]
		if genids[name] {
			continue
		}
		if deps[name] > 0 {
			setupAfterMap[name] = "g." + name + " = " + defs[name]
		} else {
			setupBeforeMap[name] = "g." + name + " = " + defs[name]
		}
	}

	setupBefore := stringMapToSlice(setupBeforeMap, func(a, b string) bool {
		return a < b
	})

	setupAfter := stringMapToSlice(setupAfterMap, func(a, b string) bool {
		if deps[a] == deps[b] {
			return a < b
		}
		return deps[a] < deps[b]
	})

	attrs := []string{}
	for obj, props := range d.Metadata() {
		id := "g." + props["name"]
		if g, ok := props["name-is-generated"]; ok && g == "1" {
			id = props["name"]
		}

		for _, attr := range d.Attrs()[obj] {
			attrs = append(attrs, id+"."+attr)
		}
	}
	sort.Strings(attrs)

	for obj := range d.Attrs() {
		if _, exists := battrs[obj]; !exists {
			delete(d.Attrs(), obj)
		}
	}

	for obj, attrs := range battrs {
		d.Attrs()[obj] = attrs
	}

	for obj, props := range d.Metadata() {
		if props["name-is-generated"] != "1" {
			continue
		}
		delete(props, "name-is-generated")
		delete(props, "name")

		d.Metadata()[obj] = props
	}

	guiName := "gui"
	guiNameUpper := ""
	layoutHelper := ""
	if name != "main" {
		guiName = name + "Gui"
		guiNameUpper = strings.ToUpper(string([]byte{name[0]})) + name[1:]
	} else {
		layoutHelper = `type wrappedLayout struct {
	layout  func([]granio.CanvasObject, granio.Size)
	minSize func([]granio.CanvasObject) granio.Size
}

func (w wrappedLayout) Layout(objs []granio.CanvasObject, s granio.Size) {
	w.layout(objs, s)
}

func (w wrappedLayout) MinSize(objs []granio.CanvasObject) granio.Size {
	return w.minSize(objs)
}

func wrapLayout(l func([]granio.CanvasObject, granio.Size), m func([]granio.CanvasObject) granio.Size) granio.Layout {
	return wrappedLayout{layout: l, minSize: m}
}`
	}

	data := struct {
		Pkgs         []string
		LayoutHelper string
		GuiName      string
		GuiNameUpper string
		Vars         []string
		Attrs        []string
		SetupBefore  []string
		SetupAfter   []string
		Main         string
	}{
		Pkgs:         pkgs,
		LayoutHelper: layoutHelper,
		GuiName:      guiName,
		GuiNameUpper: guiNameUpper,
		Vars:         vars,
		Attrs:        attrs,
		SetupBefore:  setupBefore,
		SetupAfter:   setupAfter,
		Main:         main,
	}
	code, err := tools.RenderCode(`// auto-generated
// Code generated by GUI builder.

package main

import (
	"github.com/Uzer007-admin/granio/v2"
{{- range .Pkgs }}
	{{.}}
{{- end }}
)

{{.LayoutHelper}}

type {{.GuiName}} struct {
	win granio.Window

{{ range .Vars }}
	{{.}}
{{- end }}
}

func new{{.GuiNameUpper}}GUI() *{{.GuiName}} {
	return &{{.GuiName}}{}
}

func (g *{{.GuiName}}) makeUI() granio.CanvasObject {
	{{ range .SetupBefore -}}
	{{.}}
	{{ end -}}
	{{ range .SetupAfter -}}
	{{.}}
	{{ end -}}

	{{- range .Attrs}}
		{{.}}
	{{- end}}

	return {{.Main}}
}`, data)
	if err != nil {
		granio.LogError("Failed to generate GUI code", err)
		return code
	}

	formatted, err := format.Source([]byte(code))
	if err != nil {
		granio.LogError("Failed to format GUI code", err)
		return code
	}
	return string(formatted)
}

func packagesRequired(obj granio.CanvasObject, d Context) []string {
	ret := []string{"container"}
	var objs []granio.CanvasObject
	if c, ok := obj.(*granio.Container); ok {
		objs = c.Objects
		layout, ok := d.Metadata()[c]["layout"]
		if ok && (layout == "Form" || layout == "CustomPadded" || layout == "GridWrap") {
			ret = append(ret, "layout")
		}
	} else {
		class := reflect.TypeOf(obj).String()
		info := guidefs.Lookup(class)

		if info != nil && info.IsContainer() {
			ret = packagesRequiredForWidget(obj, d)
			objs = info.Children(obj)
		} else {
			return packagesRequiredForWidget(obj, d)
		}
	}

	for _, w := range objs {
		for _, p := range packagesRequired(w, d) {
			added := false
			for _, exists := range ret {
				if p == exists {
					added = true
					break
				}
			}
			if !added {
				ret = append(ret, p)
			}
		}
	}
	return ret
}

func packagesRequiredForWidget(w granio.CanvasObject, d Context) []string {
	_, name := getTypeOf(w)
	if pkgs := guidefs.Lookup(name).Packages; pkgs != nil {
		return pkgs(w, d)
	}

	if _, ok := w.(granio.Widget); ok {
		return []string{"widget"}
	}

	return []string{}
}

func varsRequired(obj granio.CanvasObject, d Context) (widgets, containers []string) {
	name := d.Metadata()[obj]["name"]

	if c, ok := obj.(*granio.Container); ok {
		if name != "" {
			containers = append(containers, name+" *granio.Container")
		}

		for _, w := range c.Objects {
			w2, c2 := varsRequired(w, d)
			if len(w2) > 0 {
				widgets = append(widgets, w2...)
			}
			if len(c2) > 0 {
				containers = append(containers, c2...)
			}
		}
	} else {
		class := reflect.TypeOf(obj).String()
		info := guidefs.Lookup(class)

		if info != nil && info.IsContainer() {
			for _, child := range info.Children(obj) {
				w2, c2 := varsRequired(child, d)

				if len(w2) > 0 {
					widgets = append(widgets, w2...)
				}
				if len(c2) > 0 {
					containers = append(containers, c2...)
				}
			}
		}

		if name != "" {
			_, class := getTypeOf(obj)
			widgets = append(widgets, name+" "+class)
		}
	}

	return
}
