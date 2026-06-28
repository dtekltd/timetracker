package windows

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ActiveWindowInfo holds information about the currently focused window.
type ActiveWindowInfo struct {
	ProcessName  string
	ProcessPath  string
	WindowTitle  string
	WindowHandle string
}

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	kernel32                = windows.NewLazySystemDLL("kernel32.dll")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW      = user32.NewProc("GetWindowTextW")
	procGetWindowThreadPID  = user32.NewProc("GetWindowThreadProcessId")
	procOpenProcess         = kernel32.NewProc("OpenProcess")
	procQueryFullProcName   = kernel32.NewProc("QueryFullProcessImageNameW")
)

const (
	processQueryLimitedInformation = 0x1000
)

// GetActiveWindowInfo returns info about the foreground window.
func GetActiveWindowInfo() ActiveWindowInfo {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return ActiveWindowInfo{}
	}

	// Window title.
	titleBuf := make([]uint16, 512)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&titleBuf[0])), uintptr(len(titleBuf)))
	title := syscall.UTF16ToString(titleBuf)

	// Process ID from window handle.
	var pid uint32
	procGetWindowThreadPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return ActiveWindowInfo{
			WindowTitle:  title,
			WindowHandle: fmt.Sprintf("%x", hwnd),
		}
	}

	// Open process to query path.
	hProc, _, _ := procOpenProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	defer func() {
		if hProc != 0 {
			_, _, _ = kernel32.NewProc("CloseHandle").Call(hProc)
		}
	}()

	var exePath string
	var exeName string
	if hProc != 0 {
		pathBuf := make([]uint16, windows.MAX_PATH)
		size := uint32(len(pathBuf))
		ret, _, _ := procQueryFullProcName.Call(hProc, 0, uintptr(unsafe.Pointer(&pathBuf[0])), uintptr(unsafe.Pointer(&size)))
		if ret != 0 {
			exePath = syscall.UTF16ToString(pathBuf[:size])
			// Extract just the filename from the full path.
			for i := len(exePath) - 1; i >= 0; i-- {
				if exePath[i] == '\\' || exePath[i] == '/' {
					exeName = exePath[i+1:]
					break
				}
			}
			if exeName == "" {
				exeName = exePath
			}
		}
	}

	return ActiveWindowInfo{
		ProcessName:  exeName,
		ProcessPath:  exePath,
		WindowTitle:  title,
		WindowHandle: fmt.Sprintf("%x", hwnd),
	}
}
