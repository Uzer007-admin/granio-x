package terminal

import (
	"github.com/Uzer007-admin/granio/v2"
)

func (t *Terminal) handleMouseDownV200(btn int, mods granio.KeyModifier, pos granio.Position) {
	_, _ = t.Write(t.encodeMouse(btn, mods, pos))
}

func (t *Terminal) handleMouseDownX10(btn int, _ granio.KeyModifier, pos granio.Position) {
	_, _ = t.Write(t.encodeMouse(btn, 0, pos))
}

func (t *Terminal) handleMouseUpV200(btn int, mods granio.KeyModifier, pos granio.Position) {
	_, _ = t.Write(t.encodeMouse(0, mods, pos))
}

func (t *Terminal) handleMouseUpX10(_ int, _ granio.KeyModifier, _ granio.Position) {
	// no-op for X10 mode
}

func (t *Terminal) encodeMouse(button int, mods granio.KeyModifier, pos granio.Position) []byte {
	p := t.getTermPosition(pos)
	var btn byte
	if button == 0 {
		btn = 3
	} else {
		btn = byte(button) - 1
	}

	if mods&granio.KeyModifierShift != 0 {
		btn += 4
	}
	if mods&granio.KeyModifierAlt != 0 {
		btn += 8
	}
	if mods&granio.KeyModifierControl != 0 {
		btn += 16
	}

	return []byte{asciiEscape, '[', 'M', 32 + btn, 32 + byte(p.Col), 32 + byte(p.Row)}
}
