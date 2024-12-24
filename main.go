package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"flag"
)

// Function to show help information
func showHelp() {
	fmt.Println("Usage: upload [OPTIONS]")
	fmt.Println("")
	fmt.Println("Upload a file to a remote server.")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -h, --help          Show this help message")
	fmt.Println("  -v, --version       Show version information")
	fmt.Println("  -d, --device        Specify the device to upload to: tcl, redmi, or server")
	fmt.Println("  -r, --remote        Specify the remote directory to upload to (default for each device is prefilled)")
	fmt.Println("  -f, --file          Specify the file or folder to upload")
	fmt.Println("  -l, --list          List available devices for upload")
	fmt.Println("")
	fmt.Println("Example: upload --device redmi --file ./file.txt --remote /sdcard/")
	fmt.Println("Example: upload -d server -f ./folder --remote /ffp/home/root/")
}

// Function to show version information
func showVersion() {
	fmt.Println("upload v1.3")
	fmt.Println("by PhateValleyman")
	fmt.Println("Jonas.Ned@outlook.com")
}

// Function to list available devices
func listDevices() {
	fmt.Println("Default devices:")
	fmt.Printf("%-15s %-10s %-10s\n", "IP", "PORT", "DEVICE")
	fmt.Printf("%-15s %-10s %-10s\n", "192.168.1.15", "22", "redmi")
	fmt.Printf("%-15s %-10s %-10s\n", "192.168.1.20", "22", "server")
	fmt.Printf("%-15s %-10s %-10s\n", "192.168.1.12", "8022", "tcl")
	fmt.Println("")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error: Unable to get home directory")
		return
	}

	sshConfigPath := homeDir + "/.ssh/config"
	if _, err := os.Stat(sshConfigPath); os.IsNotExist(err) {
		fmt.Println("Error: ~/.ssh/config not found.")
		return
	}

	file, err := os.Open(sshConfigPath)
	if err != nil {
		fmt.Println("Error: Unable to open ~/.ssh/config")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentDevice, ip, port string
	fmt.Println("Devices from ~/.ssh/config:")
	fmt.Printf("%-15s %-10s %-15s\n", "IP", "PORT", "DEVICE")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Host ") {
			if currentDevice != "" && ip != "" {
				fmt.Printf("%-15s %-10s %-15s\n", ip, port, currentDevice)
			}
			currentDevice = strings.TrimSpace(strings.TrimPrefix(line, "Host"))
			ip, port = "", ""
		} else if strings.HasPrefix(line, "HostName ") {
			ip = strings.TrimSpace(strings.TrimPrefix(line, "HostName"))
		} else if strings.HasPrefix(line, "Port ") {
			port = strings.TrimSpace(strings.TrimPrefix(line, "Port"))
		}
	}

	if currentDevice != "" && ip != "" {
		fmt.Printf("%-15s %-10s %-15s\n", ip, port, currentDevice)
	}
}

// Function to get device information from ~/.ssh/config or default values
func getDeviceInfo(device string) (string, string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error: Unable to get home directory")
		return "", ""
	}

	sshConfigPath := homeDir + "/.ssh/config"
	if _, err := os.Stat(sshConfigPath); os.IsNotExist(err) {
		// Default devices
		switch device {
		case "tcl":
			return "192.168.1.12", "8022"
		case "redmi":
			return "192.168.1.15", "22"
		case "server":
			return "192.168.1.20", "22"
		default:
			fmt.Println("Error: Invalid device '" + device + "'. Use tcl, redmi, or server.")
			os.Exit(1)
		}
	}

	file, err := os.Open(sshConfigPath)
	if err != nil {
		fmt.Println("Error: Unable to open ~/.ssh/config")
		return "", ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var ip, port string
	isTargetHost := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Host ") {
			isTargetHost = strings.TrimSpace(strings.TrimPrefix(line, "Host")) == device
		} else if isTargetHost {
			if strings.HasPrefix(line, "HostName ") {
				ip = strings.TrimSpace(strings.TrimPrefix(line, "HostName"))
			} else if strings.HasPrefix(line, "Port ") {
				port = strings.TrimSpace(strings.TrimPrefix(line, "Port"))
			}
			if ip != "" && port != "" {
				break
			}
		}
	}

	if ip == "" {
		fmt.Println("Error: Device '" + device + "' not found in ~/.ssh/config.")
		os.Exit(1)
	}
	if port == "" {
		port = "22"
	}
	return ip, port
}
// Function to handle file upload using SCP
func upload(file, remoteDir, device string) {
	ip, port := getDeviceInfo(device)

	if file == "" {
		fmt.Println("Error: File or directory not specified.")
		return
	}

	// Check if file exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Println("Error: File or directory does not exist.")
		return
	}

	if remoteDir == "" {
		remoteDir = "/"
	}

	// Prepare SCP command
	cmd := exec.Command("scp", "-P", port, "-i", os.Getenv("HOME")+"/.ssh/server", file, fmt.Sprintf("root@%s:%s", ip, remoteDir))
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error: File upload failed:", err)
		return
	}

	fmt.Println("File uploaded successfully!")
}

func main() {
	// Command-line flags
	helpFlag := flag.Bool("h", false, "Show help message")
	versionFlag := flag.Bool("v", false, "Show version information")
	deviceFlag := flag.String("d", "", "Specify the device to upload to: tcl, redmi, or server")
	remoteFlag := flag.String("r", "", "Specify the remote directory to upload to")
	fileFlag := flag.String("f", "", "Specify the file or folder to upload")
	listFlag := flag.Bool("l", false, "List available devices for upload")

	flag.Parse()

	// Show help or version information if requested
	if *helpFlag {
		showHelp()
		return
	}
	if *versionFlag {
		showVersion()
		return
	}
	if *listFlag {
		listDevices()
		return
	}

	if *deviceFlag == "" || *fileFlag == "" {
		showHelp()
		return
	}

	upload(*fileFlag, *remoteFlag, *deviceFlag)
}
