package main

import (
	"embed"
	"flag"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Uzer007-admin/granio-terminal"
	"github.com/Uzer007-admin/granio-terminal/cmd/fyneterm/data"
	fancyfs "github.com/Uzer007-admin/granio/v2/storage/foldermeta"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/lang"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/Uzer007-admin/granio/v2/theme"
)

//go:embed translation
var translations embed.FS

func termTitle() string {
	return lang.L("Title")
}

func guessCellSize(th *termTheme) granio.Size {
	cell := canvas.NewText("M", color.White)
	cell.TextStyle.Monospace = true

	min := cell.MinSize()
	scale := th.fontSize / theme.TextSize()
	return granio.NewSize(float32(math.Round(float64(min.Width*scale))), float32(math.Round(float64(min.Height*scale))))
}

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Show terminal debug messages")
	flag.Parse()

	_ = lang.AddTranslationsFS(translations, "translation")

	a := app.New()
	a.SetIcon(data.Icon)
	w := newTerminalWindow(a, debug)
	w.ShowAndRun()
}

func findTerminal(item *container.TabItem) *terminal.Terminal {
	override := item.Content.(*container.ThemeOverride)
	stack := override.Content.(*granio.Container)
	for _, obj := range stack.Objects {
		if t, ok := obj.(*terminal.Terminal); ok {
			return t
		}
	}
	return nil
}

func newTerminalWindow(a granio.App, debug bool) granio.Window {
	w, _, _ := buildTerminalWindow(a, debug, true)
	return w
}

// buildTerminalWindow constructs the terminal window and returns the window,
// its DocTabs and the updateView callback so the tab/resize behaviour can be
// exercised from tests.
func buildTerminalWindow(a granio.App, debug, runShell bool) (granio.Window, *container.DocTabs, func(bool)) {
	w := a.NewWindow(termTitle())
	w.SetPadded(false)
	th := newTermTheme()

	var wrapper *granio.Container
	tabs := container.NewDocTabs()
	// dirs tracks the working directory of each tab so new tabs can open in the same place
	dirs := make(map[*container.TabItem]string)

	// updateView swaps between showing DocTabs (2+ tabs) or the
	// single tab's content directly (1 tab, no tab bar visible).
	// It adjusts the window height to compensate for the tab bar.
	updateView := func(adding bool) {
		barHeight := tabs.MinSize().Height - tabs.Selected().Content.MinSize().Height
		size := w.Canvas().Size()
		if !adding && len(tabs.Items) == 1 {
			wrapper.Objects = []granio.CanvasObject{tabs.Items[0].Content}
			w.Resize(granio.NewSize(size.Width, size.Height-barHeight))
		} else if adding && len(tabs.Items) == 2 {
			wrapper.Objects = []granio.CanvasObject{tabs}
			w.Resize(granio.NewSize(size.Width, size.Height+barHeight))
		} else {
			wrapper.Objects = []granio.CanvasObject{tabs}
		}
		wrapper.Refresh()
	}

	firstTab := newTab(tabs, dirs, updateView, debug, th, w, a, runShell)
	tabs.Append(firstTab)
	tabs.CreateTab = func() *container.TabItem {
		tab := newTab(tabs, dirs, updateView, debug, th, w, a, true)
		w.Canvas().Focus(findTerminal(tab))
		return tab
	}

	tabs.OnClosed = func(item *container.TabItem) {
		findTerminal(item).Close()
		delete(dirs, item)
		if len(tabs.Items) == 0 {
			w.Close()
			return
		}
		updateView(false)
	}

	tabs.OnSelected = func(item *container.TabItem) {
		if item.Text == "" || item.Text == termTitle() {
			w.SetTitle(termTitle())
		} else {
			w.SetTitle(termTitle() + ": " + item.Text)
		}
		w.Canvas().Focus(findTerminal(item))
	}

	// Start in single-tab mode (no tab bar visible)
	wrapper = container.NewStack(firstTab.Content)
	w.SetContent(wrapper)

	cellSize := guessCellSize(th)
	w.Resize(granio.NewSize(cellSize.Width*80, cellSize.Height*24))

	w.Canvas().Focus(findTerminal(firstTab))

	return w, tabs, updateView
}

// currentDir returns the working directory of the selected tab,
// or an empty string if it is not known or no longer exists.
func currentDir(tabs *container.DocTabs, dirs map[*container.TabItem]string) string {
	pwd := dirs[tabs.Selected()]
	if pwd == "" {
		return ""
	}
	if info, err := os.Stat(pwd); err != nil || !info.IsDir() {
		return ""
	}
	return pwd
}

func newTab(tabs *container.DocTabs, dirs map[*container.TabItem]string, refresh func(bool), debug bool, th *termTheme, w granio.Window, a granio.App, runShell bool) *container.TabItem {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	img := canvas.NewImageFromResource(data.FyneLogo)
	img.FillMode = canvas.ImageFillContain
	img.Translucency = 0.8

	setDir := func(pwd string) {
		ff, err := fancyfs.DetailsForFolder(storage.NewFileURI(pwd))
		if ff == nil || err != nil {
			if err != nil && err != fancyfs.ErrNoMetadata {
				granio.LogError("Could not read dir metadata", err)
			}

			img.File = ""
			img.Resource = data.FyneLogo
			img.Image = nil
			img.FillMode = canvas.ImageFillContain
			img.Refresh()
			return
		}

		if ff.BackgroundURI != nil {
			img.File = ff.BackgroundURI.Path()
		} else {
			img.File = ""
		}

		homeStr, _ := os.UserHomeDir()
		if pwd == "" || pwd == homeStr {
			img.Resource = data.FyneLogo
		} else {
			img.Resource = ff.BackgroundResource
		}
		img.FillMode = ff.BackgroundFill
		img.Refresh()
	}

	t := terminal.New()
	t.SetDebug(debug)
	startDir := currentDir(tabs, dirs)
	if startDir != "" {
		t.SetStartDir(startDir)
		setDir(startDir)
	} else if len(os.Args) >= 2 {
		s, err := filepath.Abs(os.Args[1])
		if err == nil {
			startDir = s
			t.SetStartDir(s)
			setDir(s)
		}
	} else {
		wd, err := os.Getwd()
		if err == nil {
			setDir(wd)
		}
	}

	a.Settings().AddListener(func(s granio.Settings) {
		bg.FillColor = theme.Color(theme.ColorNameBackground)
		bg.Refresh()
	})

	sizeOverride := container.NewThemeOverride(container.NewStack(bg, img, t), th)
	tabItem := container.NewTabItem(termTitle(), sizeOverride)
	dirs[tabItem] = startDir

	var listen chan terminal.Config
	if runShell {
		listen = make(chan terminal.Config)
		go func() {
			for config := range listen {
				granio.Do(func() {
					title := config.Title
					if title == "" {
						tabItem.Text = termTitle()
					} else {
						tabItem.Text = title
					}
					if len(tabs.Items) > 1 {
						tabs.Refresh()
					}

					if tabs.Selected() == tabItem {
						if title == "" {
							w.SetTitle(termTitle())
						} else {
							w.SetTitle(termTitle() + ": " + title)
						}
					}

					setDir(config.PWD)
					if _, ok := dirs[tabItem]; ok && config.PWD != "" {
						dirs[tabItem] = config.PWD
					}
				})
			}
		}()
		t.AddListener(listen)
	}

	// New window shortcut
	newWin := func(_ granio.Shortcut) {
		w := newTerminalWindow(a, debug)
		w.Show()
	}
	t.AddShortcut(&desktop.CustomShortcut{KeyName: granio.KeyN, Modifier: granio.KeyModifierControl | granio.KeyModifierShift}, newWin)
	if runtime.GOOS == "darwin" {
		t.AddShortcut(&desktop.CustomShortcut{KeyName: granio.KeyN, Modifier: granio.KeyModifierSuper}, newWin)
	}

	// New tab shortcut
	newTabShortcut := func(_ granio.Shortcut) {
		item := newTab(tabs, dirs, refresh, debug, th, w, a, true)
		tabs.Append(item)
		tabs.Select(item)
		refresh(true)
		w.Canvas().Focus(findTerminal(item))
	}
	t.AddShortcut(&desktop.CustomShortcut{KeyName: granio.KeyT, Modifier: granio.KeyModifierControl | granio.KeyModifierShift}, newTabShortcut)
	if runtime.GOOS == "darwin" {
		t.AddShortcut(&desktop.CustomShortcut{KeyName: granio.KeyT, Modifier: granio.KeyModifierSuper}, newTabShortcut)
	}

	// Font size shortcuts
	refreshAllTabs := func() {
		for _, item := range tabs.Items {
			item.Content.Refresh()
		}
	}
	t.AddShortcut(&desktop.CustomShortcut{KeyName: granio.KeyEqual, Modifier: granio.KeyModifierShortcutDefault | granio.KeyModifierShift},
		func(_ granio.Shortcut) {
			th.fontSize++
			refreshAllTabs()
		})
	t.AddShortcut(&desktop.CustomShortcut{KeyName: granio.KeyMinus, Modifier: granio.KeyModifierShortcutDefault},
		func(_ granio.Shortcut) {
			th.fontSize--
			refreshAllTabs()
		})

	if runShell {
		go func() {
			err := t.RunLocalShell()
			if err != nil {
				granio.LogError("Failure in terminal", err)
			}
			t.RemoveListener(listen)
			granio.Do(func() {
				if _, ok := dirs[tabItem]; !ok {
					return
				}
				delete(dirs, tabItem)
				tabs.Remove(tabItem)
				if len(tabs.Items) == 0 {
					w.Close()
					return
				}
				refresh(false)
				w.Canvas().Focus(findTerminal(tabs.Selected()))
			})
		}()
	}

	return tabItem
}
