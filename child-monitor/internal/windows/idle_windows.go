package windows

import (
	"time"
	"unsafe"
)

// LASTINPUTINFO matches the Windows LASTINPUTINFO struct.
type lastInputInfo struct {
	cbSize uint32
	dwTime uint32
}

var procGetLastInputInfo = user32.NewProc("GetLastInputInfo")
var procGetTickCount = kernel32.NewProc("GetTickCount")

// GetIdleSeconds returns how many seconds have passed since the last
// keyboard or mouse input event.
func GetIdleSeconds() int {
	var info lastInputInfo
	info.cbSize = uint32(unsafe.Sizeof(info))
	procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&info)))

	tick, _, _ := procGetTickCount.Call()
	idleMs := uint32(tick) - info.dwTime

	d := time.Duration(idleMs) * time.Millisecond
	return int(d.Seconds())
}
