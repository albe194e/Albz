package main

/*
#include <stdint.h>
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"sync"
	"unsafe"

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

//export core_try_load_session
func core_try_load_session(handle C.uint64_t) C.int {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return C.int(-1)
	}

	loaded, err := instance.TryLoadSession()
	if err != nil {
		lastError.set(err)
		return C.int(-1)
	}

	lastError.set(nil)
	if loaded {
		return C.int(1)
	}

	return C.int(0)
}

//export core_login
func core_login(handle C.uint64_t, username *C.char, password *C.char) C.int {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return C.int(0)
	}

	if err := instance.Login(goString(username), goString(password)); err != nil {
		lastError.set(err)
		return C.int(0)
	}

	lastError.set(nil)
	return C.int(1)
}

//export core_logout
func core_logout(handle C.uint64_t) C.int {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return C.int(0)
	}

	if err := instance.Logout(); err != nil {
		lastError.set(err)
		return C.int(0)
	}

	lastError.set(nil)
	return C.int(1)
}

//export core_register
func core_register(handle C.uint64_t, name *C.char, username *C.char, password *C.char, profilePicturePath *C.char) C.int {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return C.int(0)
	}

	if err := instance.Register(
		goString(name),
		goString(username),
		goString(password),
		goString(profilePicturePath),
	); err != nil {
		lastError.set(err)
		return C.int(0)
	}

	lastError.set(nil)
	return C.int(1)
}

//export core_open_conversation
func core_open_conversation(handle C.uint64_t, conversationID *C.char) C.int {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return C.int(0)
	}

	if err := instance.OpenConversation(goString(conversationID)); err != nil {
		lastError.set(err)
		return C.int(0)
	}

	lastError.set(nil)
	return C.int(1)
}

//export core_send_message
func core_send_message(handle C.uint64_t, conversationID *C.char, body *C.char) C.int {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return C.int(0)
	}

	if err := instance.SendMessage(goString(conversationID), goString(body)); err != nil {
		lastError.set(err)
		return C.int(0)
	}

	lastError.set(nil)
	return C.int(1)
}

//export core_start_direct_conversation
func core_start_direct_conversation(handle C.uint64_t, friendUserID *C.char) *C.char {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return nil
	}

	conversationID, err := instance.StartDirectConversation(goString(friendUserID))
	if err != nil {
		lastError.set(err)
		return nil
	}

	lastError.set(nil)
	return cString(conversationID)
}

//export core_create_conversation
func core_create_conversation(handle C.uint64_t, name *C.char, participantUserIDsJSON *C.char) *C.char {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return nil
	}

	participantUserIDs := make([]string, 0)
	if jsonPayload := goString(participantUserIDsJSON); jsonPayload != "" {
		if err := json.Unmarshal([]byte(jsonPayload), &participantUserIDs); err != nil {
			lastError.set(fmt.Errorf("decode participant IDs: %w", err))
			return nil
		}
	}

	conversationID, err := instance.CreateConversation(goString(name), participantUserIDs)
	if err != nil {
		lastError.set(err)
		return nil
	}

	lastError.set(nil)
	return cString(conversationID)
}

//export core_send_friend_request
func core_send_friend_request(handle C.uint64_t, friendCode *C.char) C.int {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return C.int(0)
	}

	if err := instance.SendFriendRequest(goString(friendCode)); err != nil {
		lastError.set(err)
		return C.int(0)
	}

	lastError.set(nil)
	return C.int(1)
}

//export core_accept_friend_request
func core_accept_friend_request(handle C.uint64_t, fromUserID *C.char) C.int {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return C.int(0)
	}

	if err := instance.AcceptFriendRequest(goString(fromUserID)); err != nil {
		lastError.set(err)
		return C.int(0)
	}

	lastError.set(nil)
	return C.int(1)
}

//export core_reject_friend_request
func core_reject_friend_request(handle C.uint64_t, fromUserID *C.char) C.int {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return C.int(0)
	}

	if err := instance.RejectFriendRequest(goString(fromUserID)); err != nil {
		lastError.set(err)
		return C.int(0)
	}

	lastError.set(nil)
	return C.int(1)
}

//export core_save_profile_image
func core_save_profile_image(handle C.uint64_t, data *C.uchar, length C.int, filename *C.char) *C.char {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return nil
	}

	if data == nil || length <= 0 {
		lastError.set(fmt.Errorf("image data is required"))
		return nil
	}

	savedPath, err := instance.SaveProfileImage(
		C.GoBytes(unsafe.Pointer(data), length),
		goString(filename),
	)
	if err != nil {
		lastError.set(err)
		return nil
	}

	lastError.set(nil)
	return cString(savedPath)
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

func main() {}
