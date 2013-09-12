// +build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	kernel32    = syscall.MustLoadDLL("kernel32.dll")
	CreateFile  = kernel32.MustFindProc("CreateFileW")
	CloseHandle = kernel32.MustFindProc("CloseHandle")
)

// Force a file update on ntfs, based on:
// http://blogs.msdn.com/b/oldnewthing/archive/2011/12/26/10251026.aspx
func update(path string) error {

	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	r, _, err := CreateFile.Call(
		uintptr(unsafe.Pointer(p)),
		0, // don't require any access at all
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		0, // lpSecurityAttributes
		syscall.OPEN_EXISTING,
		0, // dwFlagsAttributes
		0, // hTemplateFile
	)
	h := syscall.Handle(r)
	if h == syscall.InvalidHandle {
		return err
	}

	r, _, err = CloseHandle.Call(uintptr(h))
	if r == 0 {
		return err
	}
	return nil
}
