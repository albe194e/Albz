package main

/*
#include <stdint.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

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
