package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef struct {
	unsigned char* data;
	int len;
} ByteBuffer;
*/
import "C"

import "unsafe"

//export core_get_qr_contact
func core_get_qr_contact(handle C.uint64_t) C.ByteBuffer {
	instance, err := loadBridge(handle)
	if err != nil {
		lastError.set(err)
		return C.ByteBuffer{}
	}

	bytes, err := instance.GetContactQRCode()
	if err != nil {
		lastError.set(err)
		return C.ByteBuffer{}
	}

	lastError.set(nil)
	return cByteBuffer(bytes)
}

//export core_bytes_free
func core_bytes_free(data *C.uchar) {
	if data == nil {
		return
	}

	C.free(unsafe.Pointer(data))
}

func cByteBuffer(value []byte) C.ByteBuffer {
	if len(value) == 0 {
		return C.ByteBuffer{}
	}

	ptr := C.malloc(C.size_t(len(value)))
	if ptr == nil {
		return C.ByteBuffer{}
	}

	target := unsafe.Slice((*byte)(ptr), len(value))
	copy(target, value)

	return C.ByteBuffer{
		data: (*C.uchar)(ptr),
		len:  C.int(len(value)),
	}
}
