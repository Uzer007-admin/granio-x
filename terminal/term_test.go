package terminal

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	_ "github.com/Uzer007-admin/granio/v2/test"

	"github.com/stretchr/testify/assert"
)

func TestNewTerminal(t *testing.T) {
	term := New()
	assert.NotNil(t, term)

	term.Refresh() // ensure visuals set up
	assert.NotNil(t, term.content)
}

func TestTerminal_CreateRenderer_KeepsContent(t *testing.T) {
	term := New()
	term.Resize(granio.NewSize(200, 100))
	term.handleOutput([]byte("Hello"))
	assert.Equal(t, "Hello", term.Text())

	grid, cursor := term.content, term.cursor
	// Fyne destroys renderers of widgets that have not been drawn for a while
	// (e.g. a background tab) and then calls CreateRenderer again.
	term.CreateRenderer()

	assert.Same(t, grid, term.content)
	assert.Same(t, cursor, term.cursor)
	assert.Equal(t, "Hello", term.Text())
}

func TestExitCode(t *testing.T) {
	term := New()
	assert.Equal(t, term.ExitCode(), int(-1))
}

func testExitCodeN(t *testing.T, n int) {
	t.Helper()
	term := New()
	term.Resize(granio.NewSize(45, 45))
	t.Cleanup(term.Close)
	done := make(chan error, 1)
	go func() { done <- term.RunLocalShell() }()
	newline := "\n"
	if runtime.GOOS == "windows" {
		newline = "\r"
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := fmt.Fprintf(term, "exit %d%s", n, newline); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("оболочка не открыла канал ввода")
		}
		time.Sleep(10 * time.Millisecond)
	}
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("оболочка не завершилась")
	}
	assert.Eventually(t, func() bool { return term.ExitCode() == n }, time.Second, 10*time.Millisecond)
}

func TestExitCode01(t *testing.T) {
	testExitCodeN(t, 0)
	testExitCodeN(t, 1)
}

func TestTerminal_Resize(t *testing.T) {
	term := New()
	term.Resize(granio.NewSize(45, 45))

	assert.Equal(t, uint(5), term.config.Columns)
	assert.Equal(t, uint(2), term.config.Rows)
}

func TestTerminal_Resize_Narrower(t *testing.T) {
	term := New()
	term.Resize(granio.NewSize(85, 45))
	assert.Equal(t, uint(10), term.config.Columns)

	term.handleOutput([]byte("12345678"))
	assert.Equal(t, 8, term.cursorCol)

	term.Resize(granio.NewSize(45, 45))
	assert.Equal(t, uint(5), term.config.Columns)
	assert.Equal(t, 4, term.cursorCol)

	term.handleOutput([]byte("abcdefgh"))
	assert.LessOrEqual(t, term.cursorCol, 5)
	for i, row := range term.content.Rows {
		if i == 0 {
			continue // written before we got narrower
		}
		assert.LessOrEqual(t, len(row.Cells), 5)
	}
}

func TestTerminal_AddListener(t *testing.T) {
	term := New()
	listen := make(chan Config, 1)
	term.AddListener(listen)
	assert.Equal(t, 1, len(term.listeners))

	go term.onConfigure()
	select {
	case <-listen: // passed
	case <-time.After(time.Millisecond * 100):
		t.Error("Failed waiting for configure callback")
	}
	term.RemoveListener(listen)
	assert.Equal(t, 0, len(term.listeners))
}

func TestTerminal_SanitizePosition(t *testing.T) {
	tests := []struct {
		name   string
		pos    granio.Position
		width  int
		height int
		want   granio.Position
	}{
		{
			name:   "WithinBounds",
			pos:    granio.NewPos(2, 3),
			width:  45,
			height: 45,
			want:   granio.NewPos(2, 3),
		},
		{
			name:   "NegativeX",
			pos:    granio.NewPos(-1, 2),
			width:  45,
			height: 45,
			want:   granio.NewPos(0, 2),
		},
		{
			name:   "ExceedsWidth",
			pos:    granio.NewPos(46, 2),
			width:  45,
			height: 45,
			want:   granio.NewPos(45, 2),
		},
		{
			name:   "NegativeY",
			pos:    granio.NewPos(3, -1),
			width:  45,
			height: 45,
			want:   granio.NewPos(3, 0),
		},
		{
			name:   "ExceedsHeight",
			pos:    granio.NewPos(3, 46),
			width:  45,
			height: 45,
			want:   granio.NewPos(3, 45),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			term := New()
			term.Resize(granio.NewSize(float32(tt.width), float32(tt.height)))

			got := term.sanitizePosition(tt.pos)

			if *got != tt.want {
				t.Errorf("got %v, want %v", *got, tt.want)
			}
		})
	}
}
