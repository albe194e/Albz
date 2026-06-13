package main

/*
#include <stdint.h>
*/
import "C"

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
