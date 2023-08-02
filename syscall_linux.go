package bluetooth

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

func ioR(t, nr, size uint) uint {
	return (2 << 30) | (t << 8) | nr | (size << 16)
}

func writev(fd int, iovs []unix.Iovec) (n int, err error) {
	var _p0 unsafe.Pointer
	n = 0
	if len(iovs) > 0 {
		_p0 = unsafe.Pointer(&iovs[0])
	} else {
		return
	}
	r0, _, e1 := unix.Syscall(unix.SYS_WRITEV, uintptr(fd), uintptr(_p0), uintptr(len(iovs)))
	n = int(r0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

func getsockopt(s int, level int, name int, val unsafe.Pointer, vallen *socklen) (err error) {
	_, _, e1 := unix.Syscall6(unix.SYS_GETSOCKOPT, uintptr(s), uintptr(level), uintptr(name), uintptr(val), uintptr(unsafe.Pointer(vallen)), 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

func setsockopt(s int, level int, name int, val unsafe.Pointer, vallen uintptr) (err error) {
	_, _, e1 := unix.Syscall6(unix.SYS_SETSOCKOPT, uintptr(s), uintptr(level), uintptr(name), uintptr(val), uintptr(vallen), 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

func ioctl(fd int, req uint, arg unsafe.Pointer) (err error) {
	_, _, e1 := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(req), uintptr(arg))
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var (
	errEAGAIN error = syscall.EAGAIN
	errEINVAL error = syscall.EINVAL
	errENOENT error = syscall.ENOENT
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case syscall.EAGAIN:
		return errEAGAIN
	case syscall.EINVAL:
		return errEINVAL
	case syscall.ENOENT:
		return errENOENT
	}
	return e
}
