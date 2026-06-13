package main

/*
#include <stdint.h>
*/
import "C"

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
