package terminal

import (
	"image/color"

	widget2 "github.com/Uzer007-admin/granio-terminal/internal/widget"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/theme"
)

const cursorWidth = 2

type render struct {
	term *Terminal
}

func (r *render) Layout(s granio.Size) {
	r.term.content.Resize(s)
}

func (r *render) MinSize() granio.Size {
	return granio.NewSize(0, 0) // don't get propped open by the text cells
}

func (r *render) Refresh() {
	granio.Do(func() { // TODO fix root cause in refresh on wrong thread
		r.moveCursor()
		r.term.refreshCursor()
	})

	r.term.content.Refresh()
}

func (r *render) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *render) Objects() []granio.CanvasObject {
	return []granio.CanvasObject{r.term.content, r.term.cursor}
}

func (r *render) Destroy() {
}

func (r *render) moveCursor() {
	cell := r.term.guessCellSize()
	r.term.cursor.Move(granio.NewPos(cell.Width*float32(r.term.cursorCol), cell.Height*float32(r.term.cursorRow)))
}

func (t *Terminal) refreshCursor() {
	t.cursor.Hidden = !t.focused || t.cursorHidden
	if t.bell {
		t.cursor.FillColor = theme.Color(theme.ColorNameError)
	} else {
		t.cursor.FillColor = theme.Color(theme.ColorNamePrimary)
	}
	t.cursor.Resize(granio.NewSize(cursorWidth, t.guessCellSize().Height))
	t.cursor.Refresh()
}

// CreateRenderer requests a new renderer for this terminal (just a wrapper around the TextGrid)
func (t *Terminal) CreateRenderer() granio.WidgetRenderer {
	t.ExtendBaseWidget(t)

	// Retain any existing content when a new renderer is created.
	if t.content == nil {
		t.content = widget2.NewTermGrid()
		t.setupShortcuts()
	}

	if t.cursor == nil {
		t.cursor = canvas.NewRectangle(theme.Color(theme.ColorNamePrimary))
		t.cursor.Hidden = true
		t.cursor.Resize(granio.NewSize(cursorWidth, t.guessCellSize().Height))
	}

	r := &render{term: t}
	t.cursorMoved = r.moveCursor
	return r
}
