package util

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
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
		logger.Warn("Failed to set default endpoint, exiting: ", err)
		return false
	}
	return true
}

func GetCurrentAudioDeviceID() (string, error) {
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return "", fmt.Errorf("failed to initialize COM library, continuing anyway")
	}
	defer ole.CoUninitialize()

	var mmDeviceEnumerator *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(
		wca.CLSID_MMDeviceEnumerator,
		0,
		wca.CLSCTX_ALL,
		wca.IID_IMMDeviceEnumerator,
		&mmDeviceEnumerator,
	); err != nil {
		return "", fmt.Errorf("failed to create device enumerator: %w", err)
	}
	defer mmDeviceEnumerator.Release()

	var defaultDevice *wca.IMMDevice
	if err := mmDeviceEnumerator.GetDefaultAudioEndpoint(wca.EConsole, wca.DEVICE_STATE_ACTIVE, &defaultDevice); err != nil {
		return "", fmt.Errorf("failed to get default audio endpoint: %w", err)
	}
	defer defaultDevice.Release()

	var deviceID string
	if err := defaultDevice.GetId(&deviceID); err != nil {
		return "", fmt.Errorf("failed to get device ID: %w", err)
	}

	return deviceID, nil
}

// Finds the friendly name of a device by its ID using the Windows API.
func GetDeviceFriendlyNameByIdWinApi(wantDeviceID string) (string, error) {
	if wantDeviceID == "" {
		return "", errors.New("deviceID cannot be empty")
	}

	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return "", fmt.Errorf("failed to initialize COM library: %w", err)
	}
	defer ole.CoUninitialize()

	var mmDeviceEnumerator *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(
		wca.CLSID_MMDeviceEnumerator,
		0,
		wca.CLSCTX_ALL,
		wca.IID_IMMDeviceEnumerator,
		&mmDeviceEnumerator,
	); err != nil {
		return "", fmt.Errorf("failed to create device enumerator: %w", err)
	}
	defer mmDeviceEnumerator.Release()

	var deviceCollection *wca.IMMDeviceCollection
	if err := mmDeviceEnumerator.EnumAudioEndpoints(wca.EAll, wca.DEVICE_STATE_ACTIVE, &deviceCollection); err != nil {
		return "", fmt.Errorf("failed to enumerate audio endpoints: %w", err)
	}
	defer deviceCollection.Release()

	var deviceCount uint32
	if err := deviceCollection.GetCount(&deviceCount); err != nil {
		return "", fmt.Errorf("failed to get device count: %w", err)
	}

	for deviceIdx := uint32(0); deviceIdx < deviceCount; deviceIdx++ {
		var endpoint *wca.IMMDevice
		if err := deviceCollection.Item(deviceIdx, &endpoint); err != nil {
			return "", fmt.Errorf("failed to get device at index %d: %w", deviceIdx, err)
		}
		defer endpoint.Release()
		var currentDeviceID string
		if err := endpoint.GetId(&currentDeviceID); err != nil {
			return "", fmt.Errorf("failed to get device ID for device at index %d: %w", deviceIdx, err)
		}
		if currentDeviceID == wantDeviceID {
			var propertyStore *wca.IPropertyStore
			if err := endpoint.OpenPropertyStore(wca.STGM_READ, &propertyStore); err != nil {
				return "", fmt.Errorf("failed to open property store for device at index %d: %w", deviceIdx, err)
			}
			defer propertyStore.Release()

			value := &wca.PROPVARIANT{}
			if err := propertyStore.GetValue(&wca.PKEY_Device_FriendlyName, value); err != nil {
				return "", fmt.Errorf("failed to get friendly name for device at index %d: %w", deviceIdx, err)
			}
			friendlyName := value.String()
			return friendlyName, nil
		}
	}
	return "", fmt.Errorf("no device found with name: %s", wantDeviceID)
}

func GetDeviceFriendlyNameByIdExec(deviceID string) (string, error) {
	if deviceID == "" {
		return "", errors.New("deviceID cannot be empty")
	}

	// Escape single quotes in the device ID for PowerShell command
	psDeviceID := strings.ReplaceAll(deviceID, "'", "''")

	// Construct the PowerShell command
	psCommand := fmt.Sprintf(
		`Get-PnpDevice -InstanceId '*%s*' | Select-Object -ExpandProperty FriendlyName`,
		psDeviceID,
	)

	// Execute the PowerShell command
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCommand)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run() // Use Run instead of Output to capture stderr separately

	// Check for errors during execution or in stderr
	if err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return "", fmt.Errorf("powershell command failed: %v - stderr: %s", err, stderrStr)
		}
		return "", fmt.Errorf("powershell command failed: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return "", nil // No friendly name found
	}

	return output, nil
}

// GetDeviceIDByNameExec finds the PNPDeviceID (InstanceId) of devices matching the given name
// by executing a PowerShell command.
// Returns the Device ID and an error if the command fails, or if there are multiple devices matching the name (since we're using wildcards).
func GetDeviceIDByNameExec(deviceName string) (string, error) {
	if deviceName == "" {
		return "", errors.New("deviceName cannot be empty")
	}

	// Escape single quotes in the device name for PowerShell command
	psDeviceName := strings.ReplaceAll(deviceName, "'", "''")

	// Construct the PowerShell command
	// We search both FriendlyName and Name properties for a match using wildcards.
	// Select-Object -ExpandProperty InstanceId outputs only the ID strings, one per line.
	psCommand := fmt.Sprintf(
		`Get-PnpDevice | Where-Object { $_.FriendlyName -like '*%s*' -or $_.Name -like '*%s*' } | Select-Object -ExpandProperty InstanceId | ForEach-Object { $_.Split('\\')[2] }`,
		psDeviceName, psDeviceName,
	)

	// Execute the PowerShell command
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCommand)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run() // Use Run instead of Output to capture stderr separately

	// Check for errors during execution or in stderr
	if err != nil {
		// Attempt to provide more context if stderr has content
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return "", fmt.Errorf("powershell command failed: %v - stderr: %s", err, stderrStr)
		}
		return "", fmt.Errorf("powershell command failed: %v", err)
	}

	// Process the output
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		// No devices found matching the name
		return "", nil
	}

	// Split the output by newline characters (handling Windows \r\n and Unix \n)
	ids := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")

	// Filter out any potential empty strings resulting from splitting
	var validIDs []string
	for _, id := range ids {
		trimmedID := strings.TrimSpace(id)
		if trimmedID != "" {
			validIDs = append(validIDs, trimmedID)
		}
	}

	if len(validIDs) == 0 {
		return "", fmt.Errorf("no valid device IDs found for device name: %s", deviceName)
	} else if len(validIDs) != 1 {
		return "", fmt.Errorf("multiple device IDs found for device name: %s", deviceName)
	}

	return validIDs[0], nil
}

// For future reference - this is the Windows API version of the function
func GetDeviceIDByNameWinAPI(deviceName string) (string, error) {
	if deviceName == "" {
		return "", errors.New("deviceName cannot be empty")
	}

	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return "", fmt.Errorf("failed to initialize COM library: %w", err)
	}
	defer ole.CoUninitialize()

	var mmDeviceEnumerator *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(
		wca.CLSID_MMDeviceEnumerator,
		0,
		wca.CLSCTX_ALL,
		wca.IID_IMMDeviceEnumerator,
		&mmDeviceEnumerator,
	); err != nil {
		return "", fmt.Errorf("failed to create device enumerator: %w", err)
	}
	defer mmDeviceEnumerator.Release()

	var deviceCollection *wca.IMMDeviceCollection
	if err := mmDeviceEnumerator.EnumAudioEndpoints(wca.EAll, wca.DEVICE_STATE_ACTIVE, &deviceCollection); err != nil {
		return "", fmt.Errorf("failed to enumerate audio endpoints: %w", err)
	}
	defer deviceCollection.Release()

	var deviceCount uint32
	if err := deviceCollection.GetCount(&deviceCount); err != nil {
		return "", fmt.Errorf("failed to get device count: %w", err)
	}

	for deviceIdx := uint32(0); deviceIdx < deviceCount; deviceIdx++ {
		var endpoint *wca.IMMDevice
		if err := deviceCollection.Item(deviceIdx, &endpoint); err != nil {
			return "", fmt.Errorf("failed to get device at index %d: %w", deviceIdx, err)
		}
		defer endpoint.Release()

		var propertyStore *wca.IPropertyStore
		if err := endpoint.OpenPropertyStore(wca.STGM_READ, &propertyStore); err != nil {
			return "", fmt.Errorf("failed to open property store for device at index %d: %w", deviceIdx, err)
		}
		defer propertyStore.Release()

		value := &wca.PROPVARIANT{}
		if err := propertyStore.GetValue(&wca.PKEY_Device_FriendlyName, value); err != nil {
			return "", fmt.Errorf("failed to get friendly name for device at index %d: %w", deviceIdx, err)
		}

		friendlyName := value.String()
		if strings.EqualFold(friendlyName, deviceName) {
			var deviceID string
			if err := endpoint.GetId(&deviceID); err != nil {
				return "", fmt.Errorf("failed to get device ID for device at index %d: %w", deviceIdx, err)
			}
			return deviceID, nil
		}
	}

	return "", fmt.Errorf("no device found with name: %s", deviceName)
}
