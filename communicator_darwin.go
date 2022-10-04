package bluetooth

import "C"
import (
	"unsafe"
)

//#cgo CFLAGS: -x objective-c
//#cgo LDFLAGS: -framework Foundation -framework CoreBluetooth
//#import <Foundation/Foundation.h>
//#import <CoreBluetooth/CoreBluetooth.h>
//#include "bluetooth.h"
/*
// Misc functions for C array creation
static char**makeCharArray(int size) {
	return calloc(sizeof(char*), size);
}

static void setArrayString(char **a, char *s, int n) {
	a[n] = s;
}

*/
import "C"
import (
	"errors"
	"time"
)

func CreateCoreBluetooth(verbose bool) coreBluetooth {
	if verbose {
		a := C.createAdapter(C.BOOL(True))
		return coreBluetooth{
			ptr: unsafe.Pointer(a),
		}
	} else {
		a := C.createAdapter(C.BOOL(False))
		return coreBluetooth{
			ptr: unsafe.Pointer(a),
		}
	}
}

type coreBluetooth struct {
	ptr unsafe.Pointer
}

func (cb *coreBluetooth) SetCharacteristicIDs(ids []string) {
	length := len(ids)
	arr := C.makeCharArray(C.int(length))
	for i, id := range ids {
		C.setArrayString(arr, C.CString(id), C.int(i))
	}
	C.setCharacteristicIDs(cb.ptr, arr, C.int(length))
}

func (cb *coreBluetooth) SetPeripheralID(id string) {
	C.setPeripheralID(cb.ptr, C.CString(id))
}

func (cb *coreBluetooth) GetPeripheralID() string {
	return C.GoString(C.getPeripheralID(cb.ptr))
}

func (cb *coreBluetooth) Running() bool {
	return (bool)(C.running(cb.ptr))
}

func (cb *coreBluetooth) Connected() bool {
	return (bool)(C.connected(cb.ptr))
}

func (cb *coreBluetooth) Connect() error {
	timeout := time.NewTimer(10 * time.Second)
	ch := make(chan error)
	go func(ch chan error) {
		defer func() {
			if err := recover(); err != nil {
				ch <- errors.New("connection failed")
			} else {
				ch <- nil
			}
			close(ch)
		}()
		C._connect(cb.ptr)
	}(ch)
	for err := range ch {
		if err != nil {
			return err
		}
		timeout.Stop()
	}
	return nil
}

func (cb *coreBluetooth) SendMessage(msg string) {
	C.sendMessage(cb.ptr, C.CString(msg))
}

func (cb *coreBluetooth) Read(dataLen int) (int, []byte, error) {
	var msg string = C.GoString(C.readMessage(cb.ptr))
	if dataLen < 0 {
		return len(msg), []byte(msg), nil
	}
	if len(msg) > dataLen {
		msg = msg[:dataLen]
	}
	return len(msg), []byte(msg), nil
}

func (cb *coreBluetooth) Write(d []byte) (int, error) {
	cb.SendMessage(string(d))
	return len(d), nil
}

func (cb *coreBluetooth) Close() error {
	//TODO implement me
	return nil
}

var _ Communicator = &coreBluetooth{}

func Connect(params Params) (Communicator, error) {
	bluetooth := CreateCoreBluetooth(params.Verbose)
	bluetooth.SetPeripheralID(params.Address)
	if params.CharacteristicIDs != nil {
		bluetooth.SetCharacteristicIDs(params.CharacteristicIDs)
	}
	if err := bluetooth.Connect(); err != nil {
		return nil, err
	}
	return &bluetooth, nil
}
