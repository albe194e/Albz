package main

/*
#include <stdint.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/albe194e/albz/app/core-go/bridge"
)

//export core_create
func core_create(profileName *C.char, dataDir *C.char, serverURL *C.char) C.uint64_t {
	bridgeInstanceValue, err := bridge.New(bridge.Options{
		ProfileName: goString(profileName),
		DataDir:     goString(dataDir),
		ServerURL:   goString(serverURL),
	})
	if err != nil {
		lastError.set(err)
		return C.uint64_t(0)
	}

	instance := &bridgeInstance{
		bridge: bridgeInstanceValue,
	}
	if err := bridgeInstanceValue.SetEventSink(instance); err != nil {
		_ = bridgeInstanceValue.Close()
		lastError.set(err)
		return C.uint64_t(0)
	}

	lastError.set(nil)
	return C.uint64_t(registry.add(instance))
}

//export core_close
func core_close(handle C.uint64_t) {
	if handle == 0 {
		lastError.set(nil)
		return
	}

	instance, ok := registry.remove(uint64(handle))
	if !ok {
		lastError.set(fmt.Errorf("bridge handle %d is not initialized", uint64(handle)))
		return
	}

	lastError.set(instance.bridge.Close())
}

//export core_config_json
func core_config_json(handle C.uint64_t) *C.char {
	instance, err := registry.get(uint64(handle))
	if err != nil {
		lastError.set(err)
		return nil
	}

	configJSON, err := instance.bridge.ConfigJSON()
	if err != nil {
		lastError.set(err)
		return nil
	}

	lastError.set(nil)
	return cString(configJSON)
}

//export core_snapshot_json
func core_snapshot_json(handle C.uint64_t) *C.char {
	instance, err := registry.get(uint64(handle))
	if err != nil {
		lastError.set(err)
		return nil
	}

	snapshotJSON, err := instance.bridge.SnapshotJSON()
	if err != nil {
		lastError.set(err)
		return nil
	}

	lastError.set(nil)
	return cString(snapshotJSON)
}

//export core_poll_event_json
func core_poll_event_json(handle C.uint64_t) *C.char {
	instance, err := loadInstance(handle)
	if err != nil {
		lastError.set(err)
		return nil
	}

	event := instance.pollEvent()
	lastError.set(nil)
	if event == "" {
		return nil
	}

	return cString(event)
}

//export core_take_last_error
func core_take_last_error() *C.char {
	value := lastError.take()
	if value == "" {
		return nil
	}

	return cString(value)
}

//export core_string_free
func core_string_free(value *C.char) {
	if value == nil {
		return
	}

	C.free(unsafe.Pointer(value))
}
