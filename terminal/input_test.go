package terminal

import (
	"bytes"
	"io"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/test"
)

// NopCloser returns a WriteCloser with a no-op Close method wrapping
// the provided writer w.
func NopCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}

type nopCloser struct {
	io.Writer
}

func (n nopCloser) Close() error {
	return nil
}

func TestTerminal_TypedKey(t *testing.T) {
	tests := map[string]struct {
		key          granio.KeyName
		bufferMode   bool
		shiftPressed bool
		want         []byte
	}{
		"F1":             {granio.KeyF1, false, false, []byte{asciiEscape, 'O', 'P'}},
		"F2":             {granio.KeyF2, false, false, []byte{asciiEscape, 'O', 'Q'}},
		"F3":             {granio.KeyF3, false, false, []byte{asciiEscape, 'O', 'R'}},
		"F4":             {granio.KeyF4, false, false, []byte{asciiEscape, 'O', 'S'}},
		"F5":             {granio.KeyF5, false, false, []byte{asciiEscape, '[', '1', '5', '~'}},
		"F6":             {granio.KeyF6, false, false, []byte{asciiEscape, '[', '1', '7', '~'}},
		"F7":             {granio.KeyF7, false, false, []byte{asciiEscape, '[', '1', '8', '~'}},
		"F8":             {granio.KeyF8, false, false, []byte{asciiEscape, '[', '1', '9', '~'}},
		"F9":             {granio.KeyF9, false, false, []byte{asciiEscape, '[', '2', '0', '~'}},
		"F10":            {granio.KeyF10, false, false, []byte{asciiEscape, '[', '2', '1', '~'}},
		"F11":            {granio.KeyF11, false, false, []byte{asciiEscape, '[', '2', '3', '~'}},
		"F12":            {granio.KeyF12, false, false, []byte{asciiEscape, '[', '2', '4', '~'}},
		"Shift+F1":       {granio.KeyF1, false, true, []byte{asciiEscape, '[', '2', '5', '~'}},
		"Shift+F2":       {granio.KeyF2, false, true, []byte{asciiEscape, '[', '2', '6', '~'}},
		"Shift+F3":       {granio.KeyF3, false, true, []byte{asciiEscape, 'O', 'R', ';', '2', '~'}},
		"Shift+F4":       {granio.KeyF4, false, true, []byte{asciiEscape, '[', '1', ';', '2', 'S'}},
		"Shift+F5":       {granio.KeyF5, false, true, []byte{asciiEscape, '[', '1', '5', ';', '2', '~'}},
		"Shift+F6":       {granio.KeyF6, false, true, []byte{asciiEscape, '[', '1', '7', ';', '2', '~'}},
		"Shift+F7":       {granio.KeyF7, false, true, []byte{asciiEscape, '[', '1', '8', ';', '2', '~'}},
		"Shift+F8":       {granio.KeyF8, false, true, []byte{asciiEscape, '[', '1', '9', ';', '2', '~'}},
		"Shift+F9":       {granio.KeyF9, false, true, []byte{asciiEscape, '[', '2', '0', ';', '2', '~'}},
		"Shift+F10":      {granio.KeyF10, false, true, []byte{asciiEscape, '[', '2', '1', ';', '2', '~'}},
		"Shift+F11":      {granio.KeyF11, false, true, []byte{asciiEscape, '[', '2', '3', ';', '2', '~'}},
		"Shift+F12":      {granio.KeyF12, false, true, []byte{asciiEscape, '[', '2', '4', ';', '2', '~'}},
		"Shift+PageUp":   {granio.KeyPageUp, false, true, []byte{asciiEscape, '[', '5', ';', '2', '~'}},
		"Shift+PageDown": {granio.KeyPageDown, false, true, []byte{asciiEscape, '[', '6', ';', '2', '~'}},
		"Shift+Home":     {granio.KeyHome, false, true, []byte{asciiEscape, '[', '1', ';', '2', 'H'}},
		"Shift+Insert":   {granio.KeyInsert, false, true, []byte{asciiEscape, '[', '2', ';', '2', '~'}},
		"Shift+Delete":   {granio.KeyDelete, false, true, []byte{asciiEscape, '[', '3', ';', '2', '~'}},
		"Shift+End":      {granio.KeyEnd, false, true, []byte{asciiEscape, '[', '1', ';', '2', 'F'}},
		"Shift+Up":       {granio.KeyUp, false, true, []byte{asciiEscape, '[', 'A', ';', '2'}},
		"Shift+Down":     {granio.KeyDown, false, true, []byte{asciiEscape, '[', 'B', ';', '2'}},
		"Shift+Left":     {granio.KeyLeft, false, true, []byte{asciiEscape, '[', 'D', ';', '2'}},
		"Shift+Right":    {granio.KeyRight, false, true, []byte{asciiEscape, '[', 'C', ';', '2'}},

		"PageUp":    {granio.KeyPageUp, false, false, []byte{asciiEscape, '[', '5', '~'}},
		"PageDown":  {granio.KeyPageDown, false, false, []byte{asciiEscape, '[', '6', '~'}},
		"Home":      {granio.KeyHome, false, false, []byte{asciiEscape, 'O', 'H'}},
		"Insert":    {granio.KeyInsert, false, false, []byte{asciiEscape, '[', '2', '~'}},
		"Delete":    {granio.KeyDelete, false, false, []byte{asciiEscape, '[', '3', '~'}},
		"End":       {granio.KeyEnd, false, false, []byte{asciiEscape, 'O', 'F'}},
		"Enter":     {granio.KeyEnter, false, false, []byte{'\n'}}, // Modify as needed for Windows or bufferMode
		"Tab":       {granio.KeyTab, false, false, []byte{'\t'}},
		"Escape":    {granio.KeyEscape, false, false, []byte{asciiEscape}},
		"Backspace": {granio.KeyBackspace, false, false, []byte{asciiBackspace}},
		"Up":        {granio.KeyUp, false, false, []byte{asciiEscape, '[', 'A'}},
		"Down":      {granio.KeyDown, false, false, []byte{asciiEscape, '[', 'B'}},
		"Left":      {granio.KeyLeft, false, false, []byte{asciiEscape, '[', 'D'}},
		"Right":     {granio.KeyRight, false, false, []byte{asciiEscape, '[', 'C'}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Creating a mock terminal
			inBuffer := bytes.NewBuffer([]byte{})
			term := &Terminal{in: NopCloser(inBuffer), bufferMode: tt.bufferMode}
			term.keyboardState.shiftPressed = tt.shiftPressed
			keyEvent := &granio.KeyEvent{Name: tt.key}

			term.TypedKey(keyEvent)

			got := inBuffer.Bytes()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("TypedKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTerminal_TypedKey_LineMode(t *testing.T) {
	tests := map[string]struct {
		key         granio.KeyName
		newLineMode bool
		want        []byte
	}{
		"Enter":                 {granio.KeyEnter, false, []byte{'\n'}},
		"Enter with line mode":  {granio.KeyEnter, true, []byte{'\r'}},
		"Return":                {granio.KeyReturn, false, []byte{'\r'}},
		"Return with line mode": {granio.KeyReturn, true, []byte{'\r'}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Creating a mock terminal
			inBuffer := bytes.NewBuffer([]byte{})
			term := &Terminal{in: NopCloser(inBuffer), newLineMode: tt.newLineMode}
			keyEvent := &granio.KeyEvent{Name: tt.key}

			term.TypedKey(keyEvent)

			got := inBuffer.Bytes()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("TypedKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTerminal_TypedShortcut(t *testing.T) {
	tests := map[string]struct {
		shortcut granio.Shortcut
		want     []byte
	}{
		"LeftOption+U": {
			shortcut: &desktop.CustomShortcut{
				Modifier: granio.KeyModifierAlt,
				KeyName:  granio.KeyU,
			},
			want: []byte{},
		},
		"Control+@": {
			shortcut: &desktop.CustomShortcut{
				Modifier: granio.KeyModifierControl,
				KeyName:  "@",
			},
			want: []byte{0},
		},
		"Control+Space": {
			shortcut: &desktop.CustomShortcut{
				Modifier: granio.KeyModifierControl,
				KeyName:  granio.KeySpace,
			},
			want: []byte{0},
		},
		"Control+C": {
			shortcut: &desktop.CustomShortcut{
				Modifier: granio.KeyModifierControl,
				KeyName:  granio.KeyC,
			},
			want: []byte{3},
		},
		"Control+_": {
			shortcut: &desktop.CustomShortcut{
				Modifier: granio.KeyModifierControl,
				KeyName:  "_",
			},
			want: []byte{31},
		},
		"Control+X": {
			shortcut: &desktop.CustomShortcut{
				Modifier: granio.KeyModifierControl,
				KeyName:  granio.KeyX,
			},
			want: []byte{24},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Creating a mock terminal
			inBuffer := bytes.NewBuffer([]byte{})
			term := &Terminal{in: NopCloser(inBuffer)}

			term.TypedShortcut(tt.shortcut)

			got := inBuffer.Bytes()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("TypedShortcut() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTerminal_TypedKey_ModifierHeldOnFocusChange(t *testing.T) {
	buf1 := bytes.NewBuffer([]byte{})
	term1 := New()
	term1.in = NopCloser(buf1)
	term2 := New()

	w := test.NewTempWindow(t, container.NewGridWithColumns(2, term1, term2))
	w.Canvas().Focus(term1)

	// a shortcut like ctrl+shift+T moves focus whilst the modifier is still down
	term1.KeyDown(&granio.KeyEvent{Name: desktop.KeyShiftLeft})
	w.Canvas().Focus(term2)
	term2.KeyUp(&granio.KeyEvent{Name: desktop.KeyShiftLeft})

	w.Canvas().Focus(term1)
	term1.TypedKey(&granio.KeyEvent{Name: granio.KeyReturn})

	got := buf1.Bytes()
	want := []byte{'\r'}
	if !bytes.Equal(got, want) {
		t.Errorf("TypedKey() = %v, want %v", got, want)
	}
}
