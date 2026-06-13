package main

/*
#include <stdint.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"sync"

	"github.com/albe194e/albz/app/core-go/bridge"
)

type bridgeRegistry struct {
	mu        sync.RWMutex
	nextID    uint64
	instances map[uint64]*bridgeInstance
}

func newBridgeRegistry() *bridgeRegistry {
	return &bridgeRegistry{
		nextID:    1,
		instances: make(map[uint64]*bridgeInstance),
	}
}

type bridgeInstance struct {
	bridge *bridge.Bridge

	eventsMu sync.Mutex
	events   []string
}

func (i *bridgeInstance) OnCoreEvent(eventJSON string) {
	i.eventsMu.Lock()
	defer i.eventsMu.Unlock()

	i.events = append(i.events, eventJSON)
}

func (i *bridgeInstance) pollEvent() string {
	i.eventsMu.Lock()
	defer i.eventsMu.Unlock()

	if len(i.events) == 0 {
		return ""
	}

	event := i.events[0]
	i.events = i.events[1:]
	return event
}

func (r *bridgeRegistry) add(instance *bridgeInstance) uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.nextID
	r.nextID++
	r.instances[id] = instance
	return id
}

func (r *bridgeRegistry) get(id uint64) (*bridgeInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instance, ok := r.instances[id]
	if !ok {
		return nil, fmt.Errorf("bridge handle %d is not initialized", id)
	}

	return instance, nil
}

func (r *bridgeRegistry) remove(id uint64) (*bridgeInstance, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance, ok := r.instances[id]
	if !ok {
		return nil, false
	}

	delete(r.instances, id)
	return instance, true
}

type lastErrorStore struct {
	mu    sync.Mutex
	value string
}

func (s *lastErrorStore) set(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err == nil {
		s.value = ""
		return
	}

	s.value = err.Error()
}

func (s *lastErrorStore) take() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	value := s.value
	s.value = ""
	return value
}

var (
	registry  = newBridgeRegistry()
	lastError lastErrorStore
)

func goString(value *C.char) string {
	if value == nil {
		return ""
	}

	return C.GoString(value)
}

func cString(value string) *C.char {
	return C.CString(value)
}

func loadInstance(handle C.uint64_t) (*bridgeInstance, error) {
	return registry.get(uint64(handle))
}

func loadBridge(handle C.uint64_t) (*bridge.Bridge, error) {
	instance, err := loadInstance(handle)
	if err != nil {
		return nil, err
	}

	return instance.bridge, nil
}

func main() {}
