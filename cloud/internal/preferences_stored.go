// This file is copied from fyne.io/fyne/app/preferences.go to avoid an import loop
package internal

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/storage"
)

type preferences struct {
	*InMemoryPreferences

	prefLock            sync.RWMutex
	disconnected        bool
	loadingInProgress   bool
	savedRecently       bool
	changedDuringSaving bool

	app                 granio.App
	needsSaveBeforeExit bool
	file                granio.URI
}

// Declare conformity with Preferences interface
var _ granio.Preferences = (*preferences)(nil)

func (p *preferences) Disconnect() {
	p.prefLock.Lock()
	p.disconnected = true
	p.prefLock.Unlock()
}

// forceImmediateSave writes preferences to file immediately, ignoring the debouncing
// logic in the change listener. Does nothing if preferences are not backed with a file.
func (p *preferences) forceImmediateSave() {
	if !p.needsSaveBeforeExit {
		return
	}
	err := p.save()
	if err != nil {
		granio.LogError("Failed on force saving preferences", err)
	}
}

func (p *preferences) resetSavedRecently() {
	go func() {
		time.Sleep(time.Millisecond * 100) // writes are not always atomic. 10ms worked, 100 is safer.
		p.prefLock.Lock()
		p.savedRecently = false
		changedDuringSaving := p.changedDuringSaving
		p.changedDuringSaving = false
		p.prefLock.Unlock()

		if changedDuringSaving {
			p.save()
		}
	}()
}

func (p *preferences) save() error {
	return p.saveToFile(p.file)
}

func (p *preferences) saveToFile(file granio.URI) error {
	p.prefLock.Lock()
	p.savedRecently = true
	p.prefLock.Unlock()
	defer p.resetSavedRecently()
	dir, _ := storage.Parent(file)
	if dir != nil { // may not be supported / needed if the file exists
		_ = storage.CreateListable(dir)
	}

	write, err := storage.Writer(file)
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
		return err
	}
	encode := json.NewEncoder(write)

	p.InMemoryPreferences.ReadValues(func(values map[string]interface{}) {
		err = encode.Encode(&values)
	})

	write.Close()
	return err
}

func (p *preferences) load() {
	err := p.loadFromFile(p.file)
	if err != nil {
		granio.LogError("Preferences load error:", err)
	}
}

func (p *preferences) loadFromFile(path granio.URI) (err error) {
	read, err := storage.Reader(path)
	if err != nil {
		return err
	}
	defer func() {
		if r := read.Close(); r != nil && err == nil {
			err = r
		}
	}()
	decode := json.NewDecoder(read)

	p.prefLock.Lock()
	p.loadingInProgress = true
	p.prefLock.Unlock()

	p.InMemoryPreferences.WriteValues(func(values map[string]interface{}) {
		err = decode.Decode(&values)
		if err != nil {
			return
		}
		convertLists(values)
	})

	p.prefLock.Lock()
	p.loadingInProgress = false
	p.prefLock.Unlock()

	return err
}

func NewPreferences(a granio.App, file granio.URI) *preferences {
	p := &preferences{app: a, file: file}
	p.InMemoryPreferences = NewInMemoryPreferences()

	// don't load or watch if not setup
	if a.UniqueID() == "" {
		return p
	}
	p.load()

	p.needsSaveBeforeExit = true
	p.AddChangeListener(func() {
		p.prefLock.Lock()
		shouldIgnoreChange := p.savedRecently || p.loadingInProgress || p.disconnected
		if p.savedRecently && !p.loadingInProgress {
			p.changedDuringSaving = true
		}
		p.prefLock.Unlock()

		if shouldIgnoreChange { // callback after loading file, or too many updates in a row
			return
		}

		err := p.save()
		if err != nil {
			granio.LogError("Failed on saving preferences", err)
		}
	})
	p.watch()
	return p
}

func convertLists(values map[string]interface{}) {
	for k, v := range values {
		if items, ok := v.([]interface{}); ok {
			if len(items) == 0 {
				continue
			}

			switch items[0].(type) {
			case bool:
				bools := make([]bool, len(items))
				for i, item := range items {
					bools[i] = item.(bool)
				}
				values[k] = bools
			case float64:
				floats := make([]float64, len(items))
				for i, item := range items {
					floats[i] = item.(float64)
				}
				values[k] = floats
			case int:
				ints := make([]int, len(items))
				for i, item := range items {
					ints[i] = item.(int)
				}
				values[k] = ints
			case string:
				strings := make([]string, len(items))
				for i, item := range items {
					strings[i] = item.(string)
				}
				values[k] = strings
			}
		}
	}
}
