package refyne

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	_ "github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/widget"

	"github.com/Uzer007-admin/granio-x/ui/internal/guidefs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const labelJSON = `{
  "Type": "*widget.Label",%s
  "Struct": {
    "Hidden": false,
    "Text": "Hi",
    "Alignment": 1,
    "Wrapping": 0,
    "TextStyle": {
      "Bold": true,
      "Italic": false,
      "Monospace": false,
      "Symbol": false,
      "TabWidth": 0,
      "Underline": false,
      "Strikethrough": false
    },
    "Truncation": 0,
    "Importance": 0,
    "SizeName": "",
    "TextRole": 0,
    "Selectable": false
  }
}`

func labelJSONWith(indent string) string {
	in := fmt.Sprintf(labelJSON, "")
	out := ""

	rows := strings.Split(in, "\n")
	for i, row := range rows {
		in := indent
		if i == 0 {
			in = ""
		}
		end := "\n"
		if i >= len(rows)-1 {
			end = ""
		}
		out = out + in + row + end
	}

	return out
}

var splitJSON = `{
  "Type": "*container.Split",
  "Name": "mySplit",
  "Struct": {
    "Horizontal": true,
    "Leading": ` + labelJSONWith("    ") + `,
    "Offset": 0.75,
    "Trailing": ` + labelJSONWith("    ") + `
  }
}
`

func TestDecodeObject(t *testing.T) {
	guidefs.InitOnce()

	buf := bytes.NewReader([]byte(fmt.Sprintf(labelJSON, "\n  \"Name\": \"myLabel\",")))
	meta := make(map[granio.CanvasObject]map[string]string)
	obj, err := DecodeObject(buf, &context{meta: meta})
	assert.Nil(t, err)

	l, ok := obj.(*widget.Label)
	require.True(t, ok)
	assert.Equal(t, "Hi", l.Text)
	assert.Equal(t, "myLabel", meta[l]["name"])
	assert.Equal(t, granio.TextAlignCenter, l.Alignment)
	assert.Equal(t, granio.TextStyle{Bold: true}, l.TextStyle)
}

func TestDecodeSplit(t *testing.T) {
	buf := bytes.NewReader([]byte(splitJSON))
	meta := make(map[granio.CanvasObject]map[string]string)
	obj, err := DecodeObject(buf, &context{meta: meta})
	assert.Nil(t, err)

	s, ok := obj.(*container.Split)
	require.True(t, ok)
	assert.True(t, s.Horizontal)
	assert.Equal(t, 0.75, s.Offset)
	assert.NotNil(t, s.Leading)
	assert.NotNil(t, s.Trailing)
	assert.Equal(t, "mySplit", meta[s]["name"])

	o1, ok := s.Leading.(*widget.Label)
	assert.True(t, ok)
	o2, ok := s.Trailing.(*widget.Label)
	assert.True(t, ok)
	assert.Equal(t, "Hi", o1.Text)
	assert.Equal(t, "Hi", o2.Text)
	assert.Equal(t, granio.TextStyle{Bold: true}, o1.TextStyle)
	assert.Equal(t, granio.TextStyle{Bold: true}, o2.TextStyle)
}

func TestEncodeObject(t *testing.T) {
	l := widget.NewLabelWithStyle("Hi", granio.TextAlignCenter, granio.TextStyle{Bold: true})

	props := map[string]string{"name": "myLabel"}
	meta := map[granio.CanvasObject]map[string]string{l: props}

	var buf bytes.Buffer
	err := EncodeObject(l, &context{meta: meta}, &buf)
	assert.Nil(t, err)
	assert.Equal(t, fmt.Sprintf(labelJSON, "\n  \"Name\": \"myLabel\",")+"\n", buf.String())
}

func TestEncodeSplit(t *testing.T) {
	l1 := widget.NewLabelWithStyle("Hi", granio.TextAlignCenter, granio.TextStyle{Bold: true})
	l2 := widget.NewLabelWithStyle("Hi", granio.TextAlignCenter, granio.TextStyle{Bold: true})
	s := container.NewVSplit(l1, l2)
	s.Horizontal = true
	s.Offset = 0.75

	props := map[string]string{"name": "mySplit"}
	meta := map[granio.CanvasObject]map[string]string{s: props}

	var buf bytes.Buffer
	err := EncodeObject(s, &context{meta: meta}, &buf)
	assert.Nil(t, err)
	assert.Equal(t, splitJSON, buf.String())
}
