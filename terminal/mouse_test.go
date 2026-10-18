package terminal

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/stretchr/testify/assert"
)

func TestEncodeMouse(t *testing.T) {
	term := New()
	assert.Equal(t, "\x1b[M !!", string(term.encodeMouse(1, 0, granio.NewPos(4, 4))))
	assert.Equal(t, "\x1b[M!$#", string(term.encodeMouse(2, 0, granio.NewPos(30, 36))))
	assert.Equal(t, "\x1b[M#!!", string(term.encodeMouse(0, 0, granio.NewPos(4, 4))))
}

func TestEncodeMouse_Mods(t *testing.T) {
	term := New()
	assert.Equal(t, "\x1b[M$!!", string(term.encodeMouse(1,
		granio.KeyModifierShift, granio.NewPos(4, 4))))
	assert.Equal(t, "\x1b[M4!!", string(term.encodeMouse(1,
		granio.KeyModifierShift|granio.KeyModifierControl, granio.NewPos(4, 4))))
	assert.Equal(t, "\x1b[M%!!", string(term.encodeMouse(2,
		granio.KeyModifierShift, granio.NewPos(4, 4))))
	assert.Equal(t, "\x1b[M5!!", string(term.encodeMouse(2,
		granio.KeyModifierShift|granio.KeyModifierControl, granio.NewPos(4, 4))))
}
