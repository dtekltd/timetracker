package windows

import (
	"golang.org/x/sys/windows"
)

// SetDPIAwareness calls SetProcessDpiAwarenessContext with
// DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 so screenshots capture
// at native resolution on HiDPI and multi-monitor setups.
func SetDPIAwareness() {
	// DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = -4 (handle value)
	const dpiAwarenessContextPerMonitorAwareV2 = ^uintptr(3) // (HANDLE)(-4)

	user32 := windows.NewLazySystemDLL("user32.dll")
	proc := user32.NewProc("SetProcessDpiAwarenessContext")
	if proc.Find() != nil {
		// Windows 10 1703+; skip on older builds.
		return
	}
	_, _, _ = proc.Call(dpiAwarenessContextPerMonitorAwareV2)
}
