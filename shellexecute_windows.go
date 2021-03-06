package su

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

type _ShellExecuteInfo struct {
	size          uint32
	mask          uint32
	hwnd          uintptr
	verb          *uint16
	file          *uint16
	parameter     *uint16
	directory     *uint16
	show          int
	instApp       uintptr
	idList        uintptr
	class         *uint16
	keyClass      uintptr
	hotkey        uint32
	iconOrMonitor uintptr
	hProcess      windows.Handle
}

var kernel32 = windows.NewLazySystemDLL("kernel32.dll")
var shell32 = windows.NewLazySystemDLL("shell32.dll")
var procShellExecute = shell32.NewProc("ShellExecuteExW")
var procGetProcessId = kernel32.NewProc("GetProcessId")

const (
	_SEE_MASK_NOCLOSEPROCESS = 0x40
	_SEE_MASK_UNICODE        = 0x4000
)

// ShellExecute calls ShellExecute-API: edit,explore,open and so on.
func (i Param) shellExecute() (int, error) {
	var p _ShellExecuteInfo
	var pid uintptr
	var err error

	p.size = uint32(unsafe.Sizeof(p))

	p.mask = _SEE_MASK_UNICODE | _SEE_MASK_NOCLOSEPROCESS

	p.verb, err = windows.UTF16PtrFromString(i.Action)
	if err != nil {
		return 0, err
	}
	p.file, err = windows.UTF16PtrFromString(i.Path)
	if err != nil {
		return 0, err
	}
	p.parameter, err = windows.UTF16PtrFromString(i.Param)
	if err != nil {
		return 0, err
	}
	p.directory, err = windows.UTF16PtrFromString(i.Directory)
	if err != nil {
		return 0, err
	}

	p.show = i.Show
	status, _, err := procShellExecute.Call(uintptr(unsafe.Pointer(&p)))

	if p.hProcess != 0 {
		pid, _, _ = procGetProcessId.Call(uintptr(p.hProcess))
		if err := windows.CloseHandle(p.hProcess); err != nil {
			println("windows.Closehandle()=", err.Error())
		}
	}

	if status == 0 {
		// ShellExecute and ShellExecuteExA's error is lower than 32
		// But, ShellExecuteExW's error is FALSE.

		if err != nil {
			return int(pid), err
		} else if err = windows.GetLastError(); err != nil {
			return int(pid), err
		} else {
			return int(pid), fmt.Errorf("Error(%d) in ShellExecuteExW()", status)
		}
	}
	return int(pid), nil
}
