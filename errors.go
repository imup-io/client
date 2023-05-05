package main

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/honeybadger-io/honeybadger-go"
)

var setupHoneybadger sync.Once
var HoneybadgerAPIKey string

type ErrMap struct {
	sync.RWMutex
	internal map[string]error
}

func NewErrMap(id string) *ErrMap {
	setupHoneybadger.Do(func() {
		hostContext := fmt.Sprintf("%s os: %s version: %s", id, runtime.GOOS, ClientVersion)

		honeybadger.Configure(honeybadger.Configuration{
			APIKey:   HoneybadgerAPIKey,
			Env:      ClientVersion,
			Hostname: hostContext,
		})
	})

	return &ErrMap{internal: make(map[string]error)}
}

func (m *ErrMap) read(key string) (bool, error) {
	m.RLock()
	defer m.RUnlock()
	value, ok := m.internal[key]

	return ok, value
}

func (m *ErrMap) write(key string, value error) {
	m.Lock()
	defer m.Unlock()

	m.internal[key] = value
}

func (m *ErrMap) delete(key string) {
	m.Lock()
	defer m.Unlock()

	delete(m.internal, key)
}

func (m *ErrMap) reportErrors(key string) {
	if ok, value := m.read(key); ok {
		honeybadger.Notify(value.Error(), honeybadger.ErrorClass{Name: key})

		m.delete(key)
	}
}
