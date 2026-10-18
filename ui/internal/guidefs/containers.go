package guidefs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func initContainers() {
	Containers = map[string]WidgetInfo{
		"*granio.Container": {
			Name: "Container",
			Create: func(Context) granio.CanvasObject {
				return container.NewVBox()
			},
			Edit: func(obj granio.CanvasObject, ctx Context, refresh func([]*widget.FormItem), onchanged func()) []*widget.FormItem {
				props := ctx.Metadata()[obj]
				c := obj.(*granio.Container)

				choose := widget.NewFormItem("Layout", widget.NewSelect(layoutNames, nil))
				items := []*widget.FormItem{choose}
				ready := false
				choose.Widget.(*widget.Select).OnChanged = func(l string) {
					lay := Layouts[l]
					props["layout"] = l
					c.Layout = lay.Create(c, ctx)
					c.Refresh()

					edit := lay.Edit
					items = []*widget.FormItem{choose}
					if edit != nil {
						items = append(items, edit(c, ctx)...)
					}

					if ready {
						refresh(items)
						onchanged()
					}
				}
				choose.Widget.(*widget.Select).SetSelected(props["layout"])
				ready = true
				return items
			},
			Gostring: func(obj granio.CanvasObject, ctx Context, defs map[string]string) string {
				props := ctx.Metadata()
				c := obj.(*granio.Container)
				l := props[c]["layout"]
				if l == "" {
					l = "VBox"
				}
				lay := Layouts[l]
				if lay.goText != nil {
					code := lay.goText(c, ctx, defs)
					return widgetRef(obj, ctx, defs, code)
				}

				str := &strings.Builder{}
				if l == "Form" {
					str.WriteString("container.New(layout.NewFormLayout(),\n")
				}
				if l == "GridWrap" {
					width := props[c]["width"]
					if width == "" {
						width = "100"
					}
					height := props[c]["height"]
					if height == "" {
						height = "100"
					}
					w, err := strconv.ParseInt(width, 0, 0)
					if err != nil {
						w = 100
					}
					h, err := strconv.ParseInt(height, 0, 0)
					if err != nil {
						h = 100
					}

					str.WriteString(fmt.Sprintf("container.New(layout.NewGridWrapLayout(granio.NewSize(float32(%d), float32(%d))),\n", w, h))
				} else {
					str.WriteString(fmt.Sprintf("container.New%s(", l))
				}
				writeGoStringExcluding(str, nil, ctx, defs, c.Objects...)
				str.WriteString(")")
				return widgetRef(obj, ctx, defs, str.String())
			},
		},
		"*container.AppTabs": {
			Name: "App Tabs",
			Children: func(o granio.CanvasObject) []granio.CanvasObject {
				tabs := o.(*container.AppTabs)

				children := make([]granio.CanvasObject, len(tabs.Items))
				for i, c := range tabs.Items {
					children[i] = c.Content
				}
				return children
			},
			AddChild: func(parent, o granio.CanvasObject) {
				tabs := o.(*container.AppTabs)

				item := container.NewTabItem("Untitled", o)
				tabs.Append(item)
			},
			Create: func(Context) granio.CanvasObject {
				return container.NewAppTabs(container.NewTabItem("Untitled", container.NewStack()))
			},
			Edit: func(obj granio.CanvasObject, c Context, setItems func([]*widget.FormItem), onchanged func()) []*widget.FormItem {
				tabs := obj.(*container.AppTabs)
				props := c.Metadata()[tabs]
				items := make([]*widget.FormItem, len(tabs.Items)+3)
				itemNames := make([]string, len(tabs.Items))

				update := func(s string, tabs *container.AppTabs) {
					switch s {
					case "Bottom":
						tabs.SetTabLocation(container.TabLocationBottom)
					case "Leading":
						tabs.SetTabLocation(container.TabLocationLeading)
					case "Trailing":
						tabs.SetTabLocation(container.TabLocationTrailing)
					default:
						tabs.SetTabLocation(container.TabLocationTop)
					}
				}
				loc := props["location"]
				update(loc, tabs)

				locations := widget.NewSelect([]string{"Top", "Bottom", "Leading", "Trailing"}, func(s string) {
					update(s, tabs)

					props["location"] = s
				})
				locations.SetSelected(props["location"])
				items[0] = widget.NewFormItem("Location", locations)

				newRow := func(item *container.TabItem, i int) *widget.FormItem {
					icon := newIconSelectorButton(item.Icon, func(i granio.Resource) {
						item.Icon = i
						tabs.Refresh()
						onchanged()
					}, false)
					edit := widget.NewEntry()
					edit.SetText(item.Text)
					edit.OnChanged = func(s string) {
						item.Text = s
						tabs.Refresh()
						onchanged()
					}
					del := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
						if i == len(tabs.Items)-1 {
							tabs.Items = tabs.Items[:i]
							items = items[:i+1]
							itemNames = itemNames[:i]
						} else {
							tabs.Items = append(tabs.Items[:i], tabs.Items[i+1:]...)
							items = append(items[:i+1], items[i+2:]...)
							itemNames = append(itemNames[:i], itemNames[i+1:]...)
						}
						tabs.Refresh()
						setItems(items)
						onchanged()
					})
					del.Importance = widget.DangerImportance

					tools := container.NewBorder(nil, nil, icon, del, edit)
					return widget.NewFormItem(fmt.Sprintf("Tab %d", i+1), tools)
				}
				for i, c := range tabs.Items {
					items[i+1] = newRow(c, i)
					itemNames[i] = fmt.Sprintf("%s (%d)", c.Text, i+1)
				}

				items[len(items)-2] = widget.NewFormItem("",
					widget.NewButton("Add Tab", func() {
						title := fmt.Sprintf("Tab %d", len(tabs.Items)+1)
						item := container.NewTabItem(title, container.NewStack())

						add := items[len(items)-2]
						sel := items[len(items)-1]
						newItem := newRow(item, len(tabs.Items))
						items = append(items[:len(items)-2], newItem, add, sel)
						itemNames = append(itemNames, title)

						tabs.Append(item)
						setItems(items)
						onchanged()
					}))
				ready := false
				selected := widget.NewSelect(itemNames, nil)
				selected.OnChanged = func(_ string) {
					tabs.SelectIndex(selected.SelectedIndex())
					if ready {
						onchanged()
					}
				}
				selected.SetSelectedIndex(tabs.SelectedIndex())
				ready = true
				items[len(items)-1] = widget.NewFormItem("Selected", selected)
				return items
			},
			Gostring: func(obj granio.CanvasObject, ctx Context, defs map[string]string) string {
				props := ctx.Metadata()[obj]
				tabs := obj.(*container.AppTabs)
				str := &strings.Builder{}

				loc := props["location"]
				addLoc := loc != "" && loc != "Top"

				if addLoc {
					str.WriteString("func() *container.AppTabs {\ntabs := ")
				}

				str.WriteString("container.NewAppTabs(")

				for i, c := range tabs.Items {
					if i > 0 {
						str.WriteString(",\n")
					}

					hasIcon := c.Icon != nil
					constr := "NewTabItem"
					if hasIcon {
						constr = "NewTabItemWithIcon"
					}
					str.WriteString(fmt.Sprintf("container.%s(\"%s\", ", constr, c.Text))
					if hasIcon {
						str.WriteString("theme." + IconName(c.Icon) + "(), ")
					}
					writeGoStringExcluding(str, nil, ctx, defs, c.Content)
					str.WriteString(")")
				}
				str.WriteString(")")

				if addLoc {
					locName := "container.TabLocationTop"
					switch loc {
					case "Bottom":
						locName = "container.TabLocationBottom"
					case "Leading":
						locName = "container.TabLocationLeading"
					case "Trailing":
						locName = "container.TabLocationTrailing"
					}

					str.WriteString(fmt.Sprintf("\n\ttabs.SetTabLocation(%s)\n", locName))
					str.WriteString("return tabs\n}()")
				}

				return widgetRef(obj, ctx, defs, str.String())
			},
			Packages: func(obj granio.CanvasObject, _ Context) []string {
				tabs := obj.(*container.AppTabs)
				for _, c := range tabs.Items {
					if c.Icon != nil {
						return []string{"container", "theme"}
					}
				}
				return []string{"container"}
			},
		},
		"*container.Clip": {
			Name: "Clip",
			Children: func(o granio.CanvasObject) []granio.CanvasObject {
				c := o.(*container.Clip)
				return []granio.CanvasObject{c.Content}
			},
			AddChild: func(parent, o granio.CanvasObject) {
				c := parent.(*container.Clip)
				c.Content = o
				c.Refresh()
			},
			Create: func(Context) granio.CanvasObject {
				return container.NewClip(container.NewStack())
			},
			Edit: func(obj granio.CanvasObject, c Context, _ func([]*widget.FormItem), _ func()) []*widget.FormItem {
				return []*widget.FormItem{}
			},
			Gostring: func(obj granio.CanvasObject, c Context, defs map[string]string) string {
				clip := obj.(*container.Clip)
				str := &strings.Builder{}

				str.WriteString("container.NewClip(")
				writeGoStringExcluding(str, nil, c, defs, clip.Content)
				str.WriteString(")")

				return widgetRef(obj, c, defs, str.String())
			},
			Packages: func(_ granio.CanvasObject, _ Context) []string {
				return []string{"container"}
			},
		},
		"*container.Navigation": {
			Name: "Navigation",
			Children: func(o granio.CanvasObject) []granio.CanvasObject {
				nav := o.(*container.Navigation)
				return []granio.CanvasObject{nav.Root}
			},
			AddChild: func(parent, o granio.CanvasObject) {
				scr := parent.(*container.Navigation)
				scr.Root = o
				scr.Refresh()
			},
			Create: func(Context) granio.CanvasObject {
				return container.NewNavigation(container.NewStack())
			},
			Edit: func(obj granio.CanvasObject, c Context, _ func([]*widget.FormItem), _ func()) []*widget.FormItem {
				nav := obj.(*container.Navigation)
				title := widget.NewEntry()
				title.SetText(nav.Title)
				title.OnChanged = func(s string) {
					nav.SetTitle(s)
				}
				return []*widget.FormItem{
					widget.NewFormItem("Title", title),
				}
			},
			Gostring: func(obj granio.CanvasObject, c Context, defs map[string]string) string {
				n := obj.(*container.Navigation)
				str := &strings.Builder{}

				if n.Title != "" {
					str.WriteString("container.NewNavigationWithTitle(")
					writeGoStringExcluding(str, nil, c, defs, n.Root)
					str.WriteString(fmt.Sprintf(", \"%s\")", n.Title))
				} else {
					str.WriteString("container.NewNavigation(")
					writeGoStringExcluding(str, nil, c, defs, n.Root)
					str.WriteString(")")
				}

				return widgetRef(obj, c, defs, str.String())
			},
			Packages: func(_ granio.CanvasObject, _ Context) []string {
				return []string{"container"}
			},
		},
		"*container.Scroll": {
			Name: "Scroll",
			Children: func(o granio.CanvasObject) []granio.CanvasObject {
				scr := o.(*container.Scroll)
				return []granio.CanvasObject{scr.Content}
			},
			AddChild: func(parent, o granio.CanvasObject) {
				scr := parent.(*container.Scroll)
				scr.Content = o
				scr.Refresh()
			},
			Create: func(Context) granio.CanvasObject {
				return container.NewScroll(container.NewStack())
			},
			Edit: func(obj granio.CanvasObject, c Context, _ func([]*widget.FormItem), _ func()) []*widget.FormItem {
				scroll := obj.(*container.Scroll)
				dirVert := widget.NewCheck("", func(on bool) {
					isHorizontal := scroll.Direction == container.ScrollBoth || scroll.Direction == container.ScrollHorizontalOnly
					if on {
						if isHorizontal {
							scroll.Direction = container.ScrollBoth
						} else {
							scroll.Direction = container.ScrollVerticalOnly
						}
					} else {
						if isHorizontal {
							scroll.Direction = container.ScrollHorizontalOnly
						} else {
							scroll.Direction = container.ScrollNone
						}
					}

					scroll.Refresh()
				})
				dirVert.Checked = scroll.Direction == container.ScrollBoth || scroll.Direction == container.ScrollVerticalOnly
				dirHoriz := widget.NewCheck("", func(on bool) {
					isVertical := scroll.Direction == container.ScrollBoth || scroll.Direction == container.ScrollVerticalOnly
					if on {
						if isVertical {
							scroll.Direction = container.ScrollBoth
						} else {
							scroll.Direction = container.ScrollHorizontalOnly
						}
					} else {
						if isVertical {
							scroll.Direction = container.ScrollVerticalOnly
						} else {
							scroll.Direction = container.ScrollNone
						}
					}

					scroll.Refresh()
				})
				dirHoriz.Checked = scroll.Direction == container.ScrollBoth || scroll.Direction == container.ScrollHorizontalOnly
				return []*widget.FormItem{
					widget.NewFormItem("Scroll Vertical", dirVert),
					widget.NewFormItem("Scroll Horizontal", dirHoriz),
				}
			},
			Gostring: func(obj granio.CanvasObject, c Context, defs map[string]string) string {
				s := obj.(*container.Scroll)
				str := &strings.Builder{}

				if s.Direction != container.ScrollBoth {
					str.WriteString(fmt.Sprintf("&container.Scroll{Direction: %s,\nContent: ", directionName(s.Direction)))
					writeGoStringExcluding(str, nil, c, defs, s.Content)
					str.WriteString("}")
				} else {
					str.WriteString("container.NewScroll(")
					writeGoStringExcluding(str, nil, c, defs, s.Content)
					str.WriteString(")")
				}

				return widgetRef(obj, c, defs, str.String())
			},
			Packages: func(_ granio.CanvasObject, _ Context) []string {
				return []string{"container"}
			},
		},
		"*container.Split": {
			Name: "Split",
			Children: func(o granio.CanvasObject) []granio.CanvasObject {
				split := o.(*container.Split)
				return []granio.CanvasObject{split.Leading, split.Trailing}
			},
			AddChild: func(parent, o granio.CanvasObject) {
				split := parent.(*container.Split)
				if split.Leading == nil {
					split.Leading = o
				} else {
					split.Trailing = o
				}
				split.Refresh()
			},
			Create: func(Context) granio.CanvasObject {
				return container.NewHSplit(container.NewStack(), container.NewStack())
			},
			Edit: func(obj granio.CanvasObject, c Context, _ func([]*widget.FormItem), onchanged func()) []*widget.FormItem {
				split := obj.(*container.Split)
				offset := widget.NewEntry()
				offset.SetText(fmt.Sprintf("%f", split.Offset))
				offset.OnChanged = func(s string) {
					if f, err := strconv.ParseFloat(s, 64); err == nil {
						split.SetOffset(f)
					}
					onchanged()
				}
				// TODO - add Fyne split.OnChanged
				vert := widget.NewCheck("", func(on bool) {
					split.Horizontal = !on
					split.Refresh()
					onchanged()
				})
				vert.Checked = !split.Horizontal
				return []*widget.FormItem{
					widget.NewFormItem("Offset", offset),
					widget.NewFormItem("Vertical", vert),
				}
			},
			Gostring: func(obj granio.CanvasObject, c Context, defs map[string]string) string {
				s := obj.(*container.Split)
				str := &strings.Builder{}
				str.WriteString(fmt.Sprintf("&container.Split{Horizontal: %t, Offset: %f, Leading: ", s.Horizontal, s.Offset))
				writeGoStringExcluding(str, nil, c, defs, s.Leading)
				str.WriteString(", Trailing: ")
				writeGoStringExcluding(str, nil, c, defs, s.Trailing)
				str.WriteString("}")
				return widgetRef(obj, c, defs, str.String())
			},
			Packages: func(_ granio.CanvasObject, _ Context) []string {
				return []string{"container"}
			},
		},
		"*container.ThemeOverride": {
			Name: "ThemeOverride",
			Children: func(o granio.CanvasObject) []granio.CanvasObject {
				over := o.(*container.ThemeOverride)
				return []granio.CanvasObject{over.Content}
			},
			AddChild: func(parent, o granio.CanvasObject) {
				over := o.(*container.ThemeOverride)
				over.Content = o
				over.Refresh()
			},
			Create: func(c Context) granio.CanvasObject {
				return container.NewThemeOverride(container.NewStack(), c.Theme())
			},
			Edit: func(obj granio.CanvasObject, c Context, _ func([]*widget.FormItem), onchanged func()) []*widget.FormItem {
				props := c.Metadata()[obj]
				over := obj.(*container.ThemeOverride)
				custom := widget.NewMultiLineEntry()
				custom.SetText(props["data"])
				custom.OnChanged = func(s string) {
					th, err := theme.FromJSONWithFallback(s, c.Theme())
					if err != nil {
						return
					}

					props["data"] = custom.Text
					over.Theme = th
					over.Refresh()
					onchanged()
				}

				return []*widget.FormItem{
					widget.NewFormItem("Theme Data", custom),
				}
			},
			Gostring: func(obj granio.CanvasObject, c Context, defs map[string]string) string {
				props := c.Metadata()[obj]
				over := obj.(*container.ThemeOverride)
				str := &strings.Builder{}
				str.WriteString("container.NewThemeOverride(")
				writeGoStringExcluding(str, nil, c, defs, over.Content)
				str.WriteString(", ")
				str.WriteString("func() granio.Theme { th, _ := theme.FromJSONWithFallback(`")
				str.WriteString(props["data"])
				str.WriteString("`, granio.CurrentApp().Settings().Theme()); return th}())")
				return widgetRef(obj, c, defs, str.String())
			},
			Packages: func(_ granio.CanvasObject, _ Context) []string {
				return []string{"container", "theme"}
			},
		},
	}

	Containers["*widget.Scroll"] = Containers["*container.Scroll"] // internal widget name may be used

	ContainerNames = extractNames(Containers)
}
