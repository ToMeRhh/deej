package util

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	ole "github.com/go-ole/go-ole"
	"github.com/lxn/win"
	"github.com/mitchellh/go-ps"
	wca "github.com/moutend/go-wca/pkg/wca"
	"go.uber.org/zap"
)

const (
	getCurrentWindowInternalCooldown = time.Millisecond * 350
)

var (
	lastGetCurrentWindowResult []string
	lastGetCurrentWindowCall   = time.Now()
)

func getCurrentWindowProcessNames() ([]string, error) {

	// apply an internal cooldown on this function to avoid calling windows API functions too frequently.
	// return a cached value during that cooldown
	now := time.Now()
	if lastGetCurrentWindowCall.Add(getCurrentWindowInternalCooldown).After(now) {
		return lastGetCurrentWindowResult, nil
	}

	lastGetCurrentWindowCall = now

	// the logic of this implementation is a bit convoluted because of the way UWP apps
	// (also known as "modern win 10 apps" or "microsoft store apps") work.
	// these are rendered in a parent container by the name of ApplicationFrameHost.exe.
	// when windows's GetForegroundWindow is called, it returns the window owned by that parent process.
	// so whenever we get that, we need to go and look through its child windows until we find one with a different PID.
	// this behavior is most common with UWP, but it actually applies to any "container" process:
	// an acceptable approach is to return a slice of possible process names that could be the "right" one, looking
	// them up is fairly cheap and covers the most bases for apps that hide their audio-playing inside another process
	// (like steam, and the league client, and any UWP app)

	result := []string{}

	// a callback that will be called for each child window of the foreground window, if it has any
	enumChildWindowsCallback := func(childHWND *uintptr, lParam *uintptr) uintptr {

		// cast the outer lp into something we can work with (maybe closures are good enough?)
		ownerPID := (*uint32)(unsafe.Pointer(lParam))

		// get the child window's real PID
		var childPID uint32
		win.GetWindowThreadProcessId((win.HWND)(unsafe.Pointer(childHWND)), &childPID)

		// compare it to the parent's - if they're different, add the child window's process to our list of process names
		if childPID != *ownerPID {

			// warning: this can silently fail, needs to be tested more thoroughly and possibly reverted in the future
			actualProcess, err := ps.FindProcess(int(childPID))
			if err == nil {
				result = append(result, actualProcess.Executable())
			}
		}

		// indicates to the system to keep iterating
		return 1
	}

	// get the current foreground window
	hwnd := win.GetForegroundWindow()
	var ownerPID uint32

	// get its PID and put it in our window info struct
	win.GetWindowThreadProcessId(hwnd, &ownerPID)

	// check for system PID (0)
	if ownerPID == 0 {
		return nil, nil
	}

	// find the process name corresponding to the parent PID
	process, err := ps.FindProcess(int(ownerPID))
	if err != nil {
		return nil, fmt.Errorf("get parent process for pid %d: %w", ownerPID, err)
	}

	// add it to our result slice
	result = append(result, process.Executable())

	// iterate its child windows, adding their names too
	win.EnumChildWindows(hwnd, syscall.NewCallback(enumChildWindowsCallback), (uintptr)(unsafe.Pointer(&ownerPID)))

	// cache & return whichever executable names we ended up with
	lastGetCurrentWindowResult = result
	return result, nil
}

type IPolicyConfigVista struct {
	ole.IUnknown
}

type IPolicyConfigVistaVtbl struct {
	ole.IUnknownVtbl
	GetMixFormat          uintptr
	GetDeviceFormat       uintptr
	SetDeviceFormat       uintptr
	GetProcessingPeriod   uintptr
	SetProcessingPeriod   uintptr
	GetShareMode          uintptr
	SetShareMode          uintptr
	GetPropertyValue      uintptr
	SetPropertyValue      uintptr
	SetDefaultEndpoint    uintptr
	SetEndpointVisibility uintptr
}

func (v *IPolicyConfigVista) VTable() *IPolicyConfigVistaVtbl {
	return (*IPolicyConfigVistaVtbl)(unsafe.Pointer(v.RawVTable))
}

func (v *IPolicyConfigVista) SetDefaultEndpoint(deviceID string, eRole uint32) (err error) {
	err = pcvSetDefaultEndpoint(v, deviceID, eRole)
	return
}

func pcvSetDefaultEndpoint(pcv *IPolicyConfigVista, deviceID string, eRole uint32) (err error) {
	var ptr *uint16
	if ptr, err = syscall.UTF16PtrFromString(deviceID); err != nil {
		return
	}
	hr, _, _ := syscall.Syscall(
		pcv.VTable().SetDefaultEndpoint,
		3,
		uintptr(unsafe.Pointer(pcv)),
		uintptr(unsafe.Pointer(ptr)),
		uintptr(uint32(eRole)))
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}

func SetAudioDeviceByID(deviceID string, logger *zap.SugaredLogger) bool {
	GUID_IPolicyConfigVista := ole.NewGUID("{568b9108-44bf-40b4-9006-86afe5b5a620}")
	GUID_CPolicyConfigVistaClient := ole.NewGUID("{294935CE-F637-4E7C-A41B-AB255460B862}")
	var policyConfig *IPolicyConfigVista

	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		logger.Warn("Failed to initialize COM library, continuing anyway")
	}
	defer ole.CoUninitialize()

	if err := wca.CoCreateInstance(GUID_CPolicyConfigVistaClient, 0, wca.CLSCTX_ALL, GUID_IPolicyConfigVista, &policyConfig); err != nil {
		logger.Warn("Failed to create policy config library, exiting")
		return false
	}
	defer policyConfig.Release()

	if err := policyConfig.SetDefaultEndpoint(deviceID, wca.EConsole); err != nil {
		logger.Warn("Failed to set default endpoint, exiting")
		return false
	}
	return true
}
