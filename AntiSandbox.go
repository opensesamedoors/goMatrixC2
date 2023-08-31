package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// Credits: https://github.com/Arvanaghi/CheckPlease/blob/master/Go/checkplease.go

type MEMORYSTATUSEX struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

type PROCESSENTRY32 struct {
	dwSize              uint32
	cntUsage            uint32
	th32ProcessID       uint32
	th32DefaultHeapID   uintptr
	th32ModuleID        uint32
	cntThreads          uint32
	th32ParentProcessID uint32
	pcPriClassBase      int32
	dwFlags             uint32
	szExeFile           [260]uint16
}

func ProcessNames() bool {
	EvidenceOfSandbox := make([]string, 0)
	sandboxProcesses := [...]string{`vmsrvc`, `tcpview`, `wireshark`, `visual basic`, `fiddler`, `vmware`, `vbox`, `process explorer`, `autoit`, `vboxtray`, `vmtools`, `vmrawdsk`, `vmusbmouse`, `vmvss`, `vmscsi`, `vmxnet`, `vmx_svga`, `vmmemctl`, `df5serv`, `vboxservice`, `vmhgfs`}
	hProcessSnap, _, _ := CreateToolhelp32Snapshot.Call(2, 0)
	if hProcessSnap < 0 {
		fmt.Println("[---] Unable to create Snapshot, exiting.")
		os.Exit(-1)
	}
	defer CloseHandle.Call(hProcessSnap)
	exeNames := make([]string, 0, 100)
	var pe32 PROCESSENTRY32
	pe32.dwSize = uint32(unsafe.Sizeof(pe32))
	Process32First.Call(hProcessSnap, uintptr(unsafe.Pointer(&pe32)))
	for {
		exeNames = append(exeNames, syscall.UTF16ToString(pe32.szExeFile[:260]))
		retVal, _, _ := Process32Next.Call(hProcessSnap, uintptr(unsafe.Pointer(&pe32)))
		if retVal == 0 {
			break
		}
	}
	for _, exe := range exeNames {
		for _, sandboxProc := range sandboxProcesses {
			if strings.Contains(strings.ToLower(exe), strings.ToLower(sandboxProc)) {
				EvidenceOfSandbox = append(EvidenceOfSandbox, exe)
			}
		}
	}
	if len(EvidenceOfSandbox) == 0 {
		return false
	}
	return true
}

func DetectDebugging() bool {
	var IsDebuggerPresent = kernel32.NewProc("IsDebuggerPresent").Addr()
	var nargs uintptr = 0
	if debuggerPresent, _, err := syscall.Syscall(IsDebuggerPresent, nargs, 0, 0, 0); err != 0 {
		return true
	} else {
		if debuggerPresent != 0 {
			return true
		}
	}
	return false
}

func DiskSize(sizer float32) bool {
	minDiskSizeGB := sizer
	var getDiskFreeSpaceEx = kernel32.NewProc("GetDiskFreeSpaceExW")
	lpFreeBytesAvailable := int64(0)
	lpTotalNumberOfBytes := int64(0)
	lpTotalNumberOfFreeBytes := int64(0)
	getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("C:"))),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)))
	diskSizeGB := float32(lpTotalNumberOfBytes) / 1073741824
	if diskSizeGB > minDiskSizeGB {
		return true
	}
	return false
}

func FilePathChecker() bool {
	EvidenceOfSandbox := make([]string, 0)
	FilePathsToCheck := [...]string{`C:\windows\System32\Drivers\Vmmouse.sys`,
		`C:\windows\System32\Drivers\vm3dgl.dll`, `C:\windows\System32\Drivers\vmdum.dll`,
		`C:\windows\System32\Drivers\vm3dver.dll`, `C:\windows\System32\Drivers\vmtray.dll`,
		`C:\windows\System32\Drivers\vmci.sys`, `C:\windows\System32\Drivers\vmusbmouse.sys`,
		`C:\windows\System32\Drivers\vmx_svga.sys`, `C:\windows\System32\Drivers\vmxnet.sys`,
		`C:\windows\System32\Drivers\VMToolsHook.dll`, `C:\windows\System32\Drivers\vmhgfs.dll`,
		`C:\windows\System32\Drivers\vmmousever.dll`, `C:\windows\System32\Drivers\vmGuestLib.dll`,
		`C:\windows\System32\Drivers\VmGuestLibJava.dll`, `C:\windows\System32\Drivers\vmscsi.sys`,
		`C:\windows\System32\Drivers\VBoxMouse.sys`, `C:\windows\System32\Drivers\VBoxGuest.sys`,
		`C:\windows\System32\Drivers\VBoxSF.sys`, `C:\windows\System32\Drivers\VBoxVideo.sys`,
		`C:\windows\System32\vboxdisp.dll`, `C:\windows\System32\vboxhook.dll`,
		`C:\windows\System32\vboxmrxnp.dll`, `C:\windows\System32\vboxogl.dll`,
		`C:\windows\System32\vboxoglarrayspu.dll`, `C:\windows\System32\vboxoglcrutil.dll`,
		`C:\windows\System32\vboxoglerrorspu.dll`, `C:\windows\System32\vboxoglfeedbackspu.dll`,
		`C:\windows\System32\vboxoglpackspu.dll`, `C:\windows\System32\vboxoglpassthroughspu.dll`,
		`C:\windows\System32\vboxservice.exe`, `C:\windows\System32\vboxtray.exe`,
		`C:\windows\System32\VBoxControl.exe`}

	for _, FilePath := range FilePathsToCheck {
		if _, err := os.Stat(FilePath); err == nil {
			EvidenceOfSandbox = append(EvidenceOfSandbox, FilePath)
		}
	}
	if len(EvidenceOfSandbox) == 0 {
		return false
	}
	return true

}

func RAMCheck() bool {
	var memInfo MEMORYSTATUSEX
	memInfo.dwLength = uint32(unsafe.Sizeof(memInfo))
	globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memInfo)))
	if memInfo.ullTotalPhys/1073741824 > 4 { //check if its more then 4 GB
		return true
	}
	return false
}

func AntiSandBox() {
	ProcessNames()
	DetectDebugging()
	DiskSize(60.0) //check if disk C has more then 60 GB
	FilePathChecker()
	RAMCheck()
}
