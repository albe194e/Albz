package main

/*
#include <stdint.h>
*/
import "C"

import (
	"encoding/json"
	"fmt"
)

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

//export core_start_direct_conversation
func core_start_direct_conversation(handle C.uint64_t, contactUserID *C.char) *C.char {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return nil
	}

	conversationID, err := instance.StartDirectConversation(goString(contactUserID))
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
