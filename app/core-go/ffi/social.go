package main

/*
#include <stdint.h>
*/
import "C"

//export core_send_contact_request
func core_send_contact_request(handle C.uint64_t, contactCode *C.char) C.int {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return C.int(0)
	}

	if err := instance.SendContactRequest(goString(contactCode)); err != nil {
		lastError.set(err)
		return C.int(0)
	}

	lastError.set(nil)
	return C.int(1)
}

//export core_accept_contact_request
func core_accept_contact_request(handle C.uint64_t, fromUserID *C.char) C.int {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return C.int(0)
	}

	if err := instance.AcceptContactRequest(goString(fromUserID)); err != nil {
		lastError.set(err)
		return C.int(0)
	}

	lastError.set(nil)
	return C.int(1)
}

//export core_reject_contact_request
func core_reject_contact_request(handle C.uint64_t, fromUserID *C.char) C.int {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return C.int(0)
	}

	if err := instance.RejectContactRequest(goString(fromUserID)); err != nil {
		lastError.set(err)
		return C.int(0)
	}

	lastError.set(nil)
	return C.int(1)
}
