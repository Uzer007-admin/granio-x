package terminal

import (
	"runtime"
	"time"
	"unicode/utf8"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
)

var lastKeyTime = time.Now()

// TypedRune is called when the user types a visible character
func (t *Terminal) TypedRune(r rune) {
	lastKeyTime = time.Now()
	b := make([]byte, utf8.UTFMax)
	size := utf8.EncodeRune(b, r)
	_, _ = t.in.Write(b[:size])
}

// TypedKey will be called if a non-printable keyboard event occurs
func (t *Terminal) TypedKey(e *granio.KeyEvent) {
	lastKeyTime = time.Now()
	if t.keyboardState.shiftPressed {
		t.keyTypedWithShift(e)
		return
	}

	switch e.Name {
	case granio.KeyReturn:
		_, _ = t.in.Write([]byte{'\r'})
	case granio.KeyEnter:
		if t.newLineMode {
			_, _ = t.in.Write([]byte{'\r'})
			return
		}
		_, _ = t.in.Write([]byte{'\n'})
	case granio.KeyTab:
		_, _ = t.in.Write([]byte{'\t'})
	case granio.KeyF1:
		_, _ = t.in.Write([]byte{asciiEscape, 'O', 'P'})
	case granio.KeyF2:
		_, _ = t.in.Write([]byte{asciiEscape, 'O', 'Q'})
	case granio.KeyF3:
		_, _ = t.in.Write([]byte{asciiEscape, 'O', 'R'})
	case granio.KeyF4:
		_, _ = t.in.Write([]byte{asciiEscape, 'O', 'S'})
	case granio.KeyF5:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '5', '~'})
	case granio.KeyF6:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '7', '~'})
	case granio.KeyF7:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '8', '~'})
	case granio.KeyF8:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '9', '~'})
	case granio.KeyF9:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '0', '~'})
	case granio.KeyF10:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '1', '~'})
	case granio.KeyF11:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '3', '~'})
	case granio.KeyF12:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '4', '~'})
	case granio.KeyEscape:
		_, _ = t.in.Write([]byte{asciiEscape})
	case granio.KeyBackspace:
		_, _ = t.in.Write([]byte{asciiBackspace})
	case granio.KeyDelete:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '3', '~'})
	case granio.KeyUp, granio.KeyDown, granio.KeyLeft, granio.KeyRight:
		t.typeCursorKey(e.Name)
	case granio.KeyPageUp:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '5', '~'})
	case granio.KeyPageDown:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '6', '~'})
	case granio.KeyHome:
		_, _ = t.in.Write([]byte{asciiEscape, 'O', 'H'})
	case granio.KeyInsert:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '~'})
	case granio.KeyEnd:
		_, _ = t.in.Write([]byte{asciiEscape, 'O', 'F'})
	}
}

func (t *Terminal) keyTypedWithShift(e *granio.KeyEvent) {
	switch e.Name {
	case granio.KeyF1:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '5', '~'})
	case granio.KeyF2:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '6', '~'})
	case granio.KeyF3:
		_, _ = t.in.Write([]byte{asciiEscape, 'O', 'R', ';', '2', '~'})
	case granio.KeyF4:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', ';', '2', 'S'})
	case granio.KeyF5:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '5', ';', '2', '~'})
	case granio.KeyF6:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '7', ';', '2', '~'})
	case granio.KeyF7:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '8', ';', '2', '~'})
	case granio.KeyF8:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', '9', ';', '2', '~'})
	case granio.KeyF9:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '0', ';', '2', '~'})
	case granio.KeyF10:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '1', ';', '2', '~'})
	case granio.KeyF11:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '3', ';', '2', '~'})
	case granio.KeyF12:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', '4', ';', '2', '~'})
	case granio.KeyPageUp:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '5', ';', '2', '~'})
	case granio.KeyPageDown:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '6', ';', '2', '~'})
	case granio.KeyHome:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', ';', '2', 'H'})
	case granio.KeyInsert:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '2', ';', '2', '~'})
	case granio.KeyDelete:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '3', ';', '2', '~'})
	case granio.KeyEnd:
		_, _ = t.in.Write([]byte{asciiEscape, '[', '1', ';', '2', 'F'})
	case granio.KeyUp:
		_, _ = t.in.Write([]byte{asciiEscape, '[', 'A', ';', '2'})
	case granio.KeyDown:
		_, _ = t.in.Write([]byte{asciiEscape, '[', 'B', ';', '2'})
	case granio.KeyLeft:
		_, _ = t.in.Write([]byte{asciiEscape, '[', 'D', ';', '2'})
	case granio.KeyRight:
		_, _ = t.in.Write([]byte{asciiEscape, '[', 'C', ';', '2'})
	}
}

func (t *Terminal) trackKeyboardState(down bool, e *granio.KeyEvent) {
	switch e.Name {
	case desktop.KeyShiftLeft:
		t.keyboardState.shiftPressed = down
	case desktop.KeyAltLeft:
		t.keyboardState.altPressed = down
	case desktop.KeyControlLeft:
		t.keyboardState.ctrlPressed = down
	case desktop.KeyShiftRight:
		t.keyboardState.shiftPressed = down
	case desktop.KeyAltRight:
		t.keyboardState.altPressed = down
	case desktop.KeyControlRight:
		t.keyboardState.ctrlPressed = down
	}
}

// KeyDown is called when we get a down key event
func (t *Terminal) KeyDown(e *granio.KeyEvent) {
	t.trackKeyboardState(true, e)
}

// KeyUp is called when we get an up key event
func (t *Terminal) KeyUp(e *granio.KeyEvent) {
	t.trackKeyboardState(false, e)
}

// FocusGained notifies the terminal that it has focus
func (t *Terminal) FocusGained() {
	t.focused = true
	t.Refresh()
}

// TypedShortcut handles key combinations, we pass them on to the tty.
func (t *Terminal) TypedShortcut(s granio.Shortcut) {
	if ds, ok := s.(*desktop.CustomShortcut); ok {
		t.ShortcutHandler.TypedShortcut(s) // it's not clear how we can check if this consumed the event

		// handle CTRL+A to CTRL+_ and everything in-between
		if ds.Modifier == granio.KeyModifierControl && len(ds.KeyName) > 0 {
			char := ds.KeyName[0]
			off := char - 'A' + 1
			switch {
			case ds.Key() == granio.KeySpace:
				fallthrough
			case ds.Key() == "@":
				off = 0
				fallthrough
			case char >= 'A' && char <= '_':
				_, _ = t.in.Write([]byte{off})
			}
		}
		return
	}

	if runtime.GOOS == "darwin" {
		// do the default thing for macOS as they separate ctrl/cmd
		t.ShortcutHandler.TypedShortcut(s)
	} else {
		// we need to override the default ctrl-X/C/V/A for non-mac and do it ourselves
		switch sh := s.(type) {
		case *granio.ShortcutCut:
			if sh.Secondary {
				// shift+del - cut is no-op
			} else {
				_, _ = t.in.Write([]byte{0x18})
			}
		case *granio.ShortcutCopy:
			if sh.Secondary {
				t.copySelectedText(t.selectClipboard()) // ctrl+ins
			} else {
				_, _ = t.in.Write([]byte{0x3})
			}
		case *granio.ShortcutPaste:
			if sh.Secondary {
				t.pasteText(t.selectClipboard()) // shift+ins
			} else {
				_, _ = t.in.Write([]byte{0x16})
			}
		case *granio.ShortcutUndo:
			_, _ = t.in.Write([]byte{0x1a})
		case *granio.ShortcutRedo:
			_, _ = t.in.Write([]byte{0x19})
		case *granio.ShortcutSelectAll:
			_, _ = t.in.Write([]byte{0x1})
		}
	}
}

// FocusLost tells the terminal it no longer has focus
func (t *Terminal) FocusLost() {
	t.focused = false
	// the key up for anything held down will go to whatever is focused next
	t.keyboardState.shiftPressed = false
	t.keyboardState.ctrlPressed = false
	t.keyboardState.altPressed = false
	t.Refresh()
}

// Focused is used to determine if this terminal currently has focus
func (t *Terminal) Focused() bool {
	return t.focused
}

func (t *Terminal) typeCursorKey(key granio.KeyName) {
	cursorPrefix := byte('[')
	if t.bufferMode {
		cursorPrefix = 'O'
	}

	switch key {
	case granio.KeyUp:
		_, _ = t.in.Write([]byte{asciiEscape, cursorPrefix, 'A'})
	case granio.KeyDown:
		_, _ = t.in.Write([]byte{asciiEscape, cursorPrefix, 'B'})
	case granio.KeyLeft:
		_, _ = t.in.Write([]byte{asciiEscape, cursorPrefix, 'D'})
	case granio.KeyRight:
		_, _ = t.in.Write([]byte{asciiEscape, cursorPrefix, 'C'})
	}
}

type discardWriter struct{}

func (d discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (d discardWriter) Close() error {
	return nil
}
