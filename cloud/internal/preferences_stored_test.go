// This file is copied from fyne.io/fyne/app/preferences_test.go to avoid an import loop
package internal

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/stretchr/testify/assert"
)

func loadPreferences(t *testing.T) *preferences {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("testdata", "dummy.json"))
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(t.TempDir(), "preferences.json")
	if err = os.WriteFile(filename, content, 0600); err != nil {
		t.Fatal(err)
	}
	return openTestPreferences(t, storage.NewFileURI(filename))
}

func openTestPreferences(t *testing.T, uri granio.URI) *preferences {
	t.Helper()
	p := NewPreferences(test.NewApp(), uri)
	t.Cleanup(func() {
		p.Disconnect()
		if p.watchDone != nil {
			<-p.watchDone
		}
	})
	return p
}

func TestPreferencesNullFileAllowsWrite(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "preferences.json")
	if err := os.WriteFile(filename, []byte("null"), 0600); err != nil {
		t.Fatal(err)
	}
	p := openTestPreferences(t, storage.NewFileURI(filename))
	assert.NoError(t, p.loadFromFile(p.file))
	assert.NotPanics(t, func() { p.SetString("key", "value") })
	assert.Equal(t, "value", p.String("key"))
}

func TestPreferences_Remove(t *testing.T) {
	p := loadPreferences(t)
	p.WriteValues(func(val map[string]interface{}) {
		val["keyString"] = "value"
		val["keyInt"] = 4
	})

	p.RemoveValue("keyString")
	err := p.saveToFile(p.file)
	assert.Nil(t, err)

	// check it doesn't write values that were removed
	p.Disconnect()
	if p.watchDone != nil {
		<-p.watchDone
	}
	p = openTestPreferences(t, p.file)
	assert.Equal(t, 4, p.Int("keyInt"))
	assert.Equal(t, "missing", p.StringWithFallback("keyString", "missing"))
}

func TestPreferences_Save(t *testing.T) {
	p := loadPreferences(t)
	p.WriteValues(func(val map[string]interface{}) {
		val["keyString"] = "value"
		val["keyStringList"] = []string{"1", "2", "3"}
		val["keyInt"] = 4
		val["keyIntList"] = []int{1, 2, 3}
		val["keyFloat"] = 3.5
		val["keyFloatList"] = []float64{1.1, 2.2, 3.3}
		val["keyBool"] = true
		val["keyBoolList"] = []bool{true, false, true}
	})

	path := filepath.Join(t.TempDir(), "preferences.json")
	defer os.Remove(path)
	err := p.saveToFile(storage.NewFileURI(path))
	assert.Nil(t, err)

	expected, err := os.ReadFile(filepath.Join("testdata", "dummy.json"))
	if err != nil {
		assert.Fail(t, "Failed to load, %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		assert.Fail(t, "Failed to load, %v", err)
	}
	assert.JSONEq(t, string(expected), string(content))

	// check it reads the saved output
	p.Disconnect()
	if p.watchDone != nil {
		<-p.watchDone
	}
	p = openTestPreferences(t, p.file)
	assert.Equal(t, "value", p.String("keyString"))
	assert.Equal(t, 3, len(p.StringList("keyStringList")))
}

func TestPreferences_Save_OverwriteFast(t *testing.T) {
	p := loadPreferences(t)
	p.WriteValues(func(val map[string]interface{}) {
		val["key"] = "value"
	})

	path := filepath.Join(t.TempDir(), "preferences.json")
	defer os.Remove(path)
	p.saveToFile(storage.NewFileURI(path))

	p.WriteValues(func(val map[string]interface{}) {
		val["key2"] = "value2"
	})
	p.saveToFile(storage.NewFileURI(path))

	p2 := loadPreferences(t)
	p2.loadFromFile(storage.NewFileURI(path))
	assert.Equal(t, "value2", p2.String("key2"))
}

func TestPreferences_Load(t *testing.T) {
	p := loadPreferences(t)
	path := filepath.Join("testdata", "dummy.json")
	p.loadFromFile(storage.NewFileURI(path))

	assert.Equal(t, "value", p.String("keyString"))
	assert.Equal(t, []string{"1", "2", "3"}, p.StringList("keyStringList"))
	assert.Equal(t, 4, p.Int("keyInt"))
	assert.Equal(t, []int{1, 2, 3}, p.IntList("keyIntList"))
	assert.Equal(t, 3.5, p.Float("keyFloat"))
	assert.Equal(t, []float64{1.1, 2.2, 3.3}, p.FloatList("keyFloatList"))
	assert.Equal(t, true, p.Bool("keyBool"))
	assert.Equal(t, []bool{true, false, true}, p.BoolList("keyBoolList"))
}

func TestPreferencesDisconnectStopsPendingSave(t *testing.T) {
	p := loadPreferences(t)
	p.SetString("first", "saved")
	p.SetString("second", "pending")
	p.Disconnect()
	if p.watchDone != nil {
		<-p.watchDone
	}
	before, err := os.ReadFile(p.file.Path())
	if err != nil {
		t.Fatal(err)
	}
	p.SetString("third", "ignored")
	time.Sleep(150 * time.Millisecond)
	after, err := os.ReadFile(p.file.Path())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, before, after)
	p.prefLock.RLock()
	assert.Nil(t, p.saveTimer)
	p.prefLock.RUnlock()
}
