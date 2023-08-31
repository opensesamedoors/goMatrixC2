package main

import (
	"golang.org/x/sys/windows"
	"os"
	"syscall"
	"unsafe"
)

// Credits: https://github.com/timwhitez/Doge-SelfDelete/blob/main/selfdel.go
func openHndl(pwPath *uint16) uintptr {
	hndl, e := syscall.CreateFile(pwPath, windows.DELETE, 0, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if e != nil {
		return 0
	}
	return uintptr(hndl)
}

func renameHndl(hndl uintptr) error {
	DsStreamRename := ":" + GetRandomString(6)

	type FileRenameInfo struct {
		Flags          uint32
		RootDirectory  syscall.Handle
		FileNameLength uint32
		FileName       [1]uint16
	}

	var fRename FileRenameInfo
	memset(uintptr(unsafe.Pointer(&fRename)), 0, unsafe.Sizeof(fRename))
	lpwStream, _ := syscall.UTF16PtrFromString(DsStreamRename)
	fRename.FileNameLength = uint32(unsafe.Sizeof(lpwStream))
	rcmem := syscall.NewLazyDLL(string([]byte{'k', 'e', 'r', 'n', 'e', 'l', '3', '2'})).NewProc(string([]byte{'R', 't', 'l', 'C', 'o', 'p', 'y', 'M', 'e', 'm', 'o', 'r', 'y'}))
	rcmem.Call(uintptr(unsafe.Pointer(&fRename.FileName)), uintptr(unsafe.Pointer(lpwStream)), unsafe.Sizeof(lpwStream))
	SFIByHandle := syscall.NewLazyDLL(string([]byte{'k', 'e', 'r', 'n', 'e', 'l', '3', '2'})).NewProc(string([]byte{'S', 'e', 't', 'F', 'i', 'l', 'e', 'I', 'n', 'f', 'o', 'r', 'm', 'a', 't', 'i', 'o', 'n', 'B', 'y', 'H', 'a', 'n', 'd', 'l', 'e'}))
	r1, _, e := SFIByHandle.Call(hndl, 3, uintptr(unsafe.Pointer(&fRename)), unsafe.Sizeof(fRename)+unsafe.Sizeof(lpwStream))
	if r1 == 0 {
		return e
	}
	return nil
}

func depositeHndl(hndl uintptr) error {
	type FileDispositionInfo struct {
		DeleteFile uint32
	}

	fDelete := FileDispositionInfo{}
	memset(uintptr(unsafe.Pointer(&fDelete)), 0, unsafe.Sizeof(fDelete))
	fDelete.DeleteFile = 1

	SFIByHandle := syscall.NewLazyDLL(string([]byte{'k', 'e', 'r', 'n', 'e', 'l', '3', '2'})).NewProc(string([]byte{'S', 'e', 't', 'F', 'i', 'l', 'e', 'I', 'n', 'f', 'o', 'r', 'm', 'a', 't', 'i', 'o', 'n', 'B', 'y', 'H', 'a', 'n', 'd', 'l', 'e'}))
	r1, _, e := SFIByHandle.Call(hndl, 4, uintptr(unsafe.Pointer(&fDelete)), unsafe.Sizeof(fDelete))
	if r1 == 0 {
		return e
	}
	return nil
}

func memset(ptr uintptr, c byte, n uintptr) {
	var i uintptr
	for i = 0; i < n; i++ {
		pByte := (*byte)(unsafe.Pointer(ptr + 1))
		*pByte = c
	}
}

func uninstall() {
	wcPath := make([]uint16, syscall.MAX_PATH)
	memset(uintptr(unsafe.Pointer(&wcPath[0])), 0, syscall.MAX_PATH)

	modkernel32 := syscall.NewLazyDLL("kernel32.dll")
	gmfn := modkernel32.NewProc("GetModuleFileNameW")
	syscall.Syscall(gmfn.Addr(), 3, 0, uintptr(unsafe.Pointer(&wcPath[0])), uintptr(syscall.MAX_PATH))

	hCurrent := openHndl(&wcPath[0])
	if hCurrent == ^uintptr(0) || hCurrent == 0 {
		os.Exit(0)
	}

	if renameHndl(hCurrent) != nil {
		windows.CloseHandle(windows.Handle(hCurrent))
		os.Exit(0)
	}

	windows.CloseHandle(windows.Handle(hCurrent))

	memset(uintptr(unsafe.Pointer(&wcPath)), 0, syscall.MAX_PATH)
	syscall.Syscall(gmfn.Addr(), 3, 0, uintptr(unsafe.Pointer(&wcPath[0])), uintptr(syscall.MAX_PATH))

	hCurrent = openHndl(&wcPath[0])

	if hCurrent == ^uintptr(0) || hCurrent == 0 {
		os.Exit(0)
	}

	e := depositeHndl(hCurrent)
	if e != nil {
		windows.CloseHandle(windows.Handle(hCurrent))
		os.Exit(0)
	}

	windows.CloseHandle(windows.Handle(hCurrent))

	os.Exit(0)
}
