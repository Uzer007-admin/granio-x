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
	ioLock              sync.Mutex
	saveWorkers         sync.WaitGroup
	saveTimer           *time.Timer
	stopWatching        func()
	watchDone           <-chan struct{}
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
	if p.saveTimer != nil && p.saveTimer.Stop() {
		p.saveWorkers.Done()
	}
	p.saveTimer = nil
	stop := p.stopWatching
	p.stopWatching = nil
	p.prefLock.Unlock()
	if stop != nil {
		stop()
	}
	p.saveWorkers.Wait()
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
	p.prefLock.Lock()
	defer p.prefLock.Unlock()
	if p.disconnected {
		return
	}
	if p.saveTimer != nil && p.saveTimer.Stop() {
		p.saveWorkers.Done()
	}
	p.saveWorkers.Add(1)
	var timer *time.Timer
	timer = time.AfterFunc(100*time.Millisecond, func() {
		defer p.saveWorkers.Done()
		p.prefLock.Lock()
		if p.disconnected || p.saveTimer != timer {
			p.prefLock.Unlock()
			return
		}
		p.saveTimer = nil
		p.savedRecently = false
		changedDuringSaving := p.changedDuringSaving
		p.changedDuringSaving = false
		p.prefLock.Unlock()

		if changedDuringSaving {
			p.save()
		}
	})
	p.saveTimer = timer
}

func (p *preferences) save() error {
	return p.saveToFile(p.file)
}

func (p *preferences) saveToFile(file granio.URI) error {
	p.ioLock.Lock()
	defer p.ioLock.Unlock()
	p.prefLock.Lock()
	if p.disconnected {
		p.prefLock.Unlock()
		return nil
	}
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

	closeErr := write.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func (p *preferences) load() {
	err := p.loadFromFile(p.file)
	if err != nil {
		granio.LogError("Preferences load error:", err)
	}
}

func (p *preferences) loadFromFile(path granio.URI) (err error) {
	p.ioLock.Lock()
	p.prefLock.RLock()
	disconnected := p.disconnected
	p.prefLock.RUnlock()
	if disconnected {
		p.ioLock.Unlock()
		return nil
	}
	read, err := storage.Reader(path)
	if err != nil {
		p.ioLock.Unlock()
		return err
	}
	decode := json.NewDecoder(read)
	values := make(map[string]interface{})
	err = decode.Decode(&values)
	closeErr := read.Close()
	p.ioLock.Unlock()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	convertLists(values)

	p.prefLock.Lock()
	p.loadingInProgress = true
	p.prefLock.Unlock()

	p.InMemoryPreferences.lock.Lock()
	p.InMemoryPreferences.values = values
	p.InMemoryPreferences.lock.Unlock()
	p.InMemoryPreferences.fireChange()

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
