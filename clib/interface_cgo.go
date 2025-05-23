// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

package main

/*
#include <stdlib.h>
#include <string.h>
// AppContext C function declarations
extern void app_log(const char* tag, const char* message);
extern int app_encrypt_to_pref(const char* key, const char* value);
extern char* app_decrypt_from_pref(const char* key);
extern char* app_get_os_version(void);
extern char* app_get_model_name(void);
extern char* app_get_install_source(void);
extern int app_should_use_google_dns_fallback(void);
extern int app_is_chrome_os(void);
extern char* app_get_interfaces_as_string(void);
extern char* app_get_platform_dns_config(void);
extern char* app_get_syspolicy_string_value(const char* key);
extern int app_get_syspolicy_boolean_value(const char* key);
extern char* app_get_syspolicy_string_array_json_value(const char* key);
#include "interface_cgo.h"
*/
import "C"
import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"unsafe"

	"github.com/tailscale/tailscale-android/libtailscale"
)

const prefFilePath = "/data/storage/el2/base/files/pref.db"

var prefFileLock sync.Mutex

// --- AppContext Go implementation using named C functions ---
type cgoAppContext struct{}

func (ctx *cgoAppContext) Log(tag, logLine string) {
	cTag := C.CString(tag)
	cLogLine := C.CString(logLine)
	defer C.free(unsafe.Pointer(cTag))
	defer C.free(unsafe.Pointer(cLogLine))
	C.app_log(cTag, cLogLine)
}
func (ctx *cgoAppContext) EncryptToPref(key, value string) error {
	prefFileLock.Lock()
	defer prefFileLock.Unlock()

	prefs := make(map[string]string)

	// Read existing preferences file.
	data, err := os.ReadFile(prefFilePath)
	if err != nil {
		if !os.IsNotExist(err) { // If error is not "file does not exist", return it.
			return errors.New("failed to read pref file: " + err.Error())
		}
		// If file does not exist, prefs will remain empty, and a new file will be created.
	} else {
		// If file exists and is not empty, try to unmarshal its JSON content.
		if len(data) > 0 {
			if errUnmarshal := json.Unmarshal(data, &prefs); errUnmarshal != nil {
				// File might be corrupted or not valid JSON. Return an error.
				return errors.New("failed to unmarshal pref file (it might be corrupted or not valid JSON): " + errUnmarshal.Error())
			}
		}
		// If file is empty, prefs remains an empty map, which is fine.
	}

	// Add or overwrite the key in the preferences map.
	prefs[key] = value

	// Marshal the updated preferences map back to JSON with indentation for readability.
	updatedData, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return errors.New("failed to marshal prefs to JSON: " + err.Error())
	}

	// Write the updated data back to the file.
	// os.WriteFile creates the file if it doesn't exist, and truncates it if it does.
	// 0600 permissions: owner can read/write.
	if err := os.WriteFile(prefFilePath, updatedData, 0600); err != nil {
		return errors.New("failed to write pref file: " + err.Error())
	}

	return nil
}
func (ctx *cgoAppContext) DecryptFromPref(key string) (string, error) {
	prefFileLock.Lock()
	defer prefFileLock.Unlock()

	prefs := make(map[string]string)

	data, err := os.ReadFile(prefFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("pref file does not exist or key not found") // Or a more specific error like "key not found"
		}
		return "", errors.New("failed to read pref file: " + err.Error())
	}

	// If file is empty, the key cannot be found.
	if len(data) == 0 {
		return "", errors.New("pref file is empty or key not found")
	}

	if errUnmarshal := json.Unmarshal(data, &prefs); errUnmarshal != nil {
		return "", errors.New("failed to unmarshal pref file (it might be corrupted or not valid JSON): " + errUnmarshal.Error())
	}

	value, ok := prefs[key]
	if !ok {
		return "", errors.New("key not found in prefs")
	}

	return value, nil
}
func (ctx *cgoAppContext) GetOSVersion() (string, error) {
	return "HarmonyOS 5.0.0", nil
}
func (ctx *cgoAppContext) GetModelName() (string, error) {
	return "Emulator", nil
}
func (ctx *cgoAppContext) GetInstallSource() string {
	return ""
}
func (ctx *cgoAppContext) ShouldUseGoogleDNSFallback() bool {
	return false
}
func (ctx *cgoAppContext) IsChromeOS() (bool, error) {
	return false, nil
}
func (ctx *cgoAppContext) GetInterfacesAsString() (string, error) {
	cValue := C.app_get_interfaces_as_string()
	if cValue == nil {
		return "", errors.New("failed to get interfaces")
	}
	defer C.free(unsafe.Pointer(cValue))
	return C.GoString(cValue), nil
}
func (ctx *cgoAppContext) GetPlatformDNSConfig() string {
	cValue := C.app_get_platform_dns_config()
	if cValue == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(cValue))
	return C.GoString(cValue)
}
func (ctx *cgoAppContext) GetSyspolicyStringValue(key string) (string, error) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	cValue := C.app_get_syspolicy_string_value(cKey)
	if cValue == nil {
		return "", errors.New("failed to get syspolicy string value")
	}
	defer C.free(unsafe.Pointer(cValue))
	return C.GoString(cValue), nil
}
func (ctx *cgoAppContext) GetSyspolicyBooleanValue(key string) (bool, error) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	result := C.app_get_syspolicy_boolean_value(cKey)
	return result != 0, nil
}
func (ctx *cgoAppContext) GetSyspolicyStringArrayJSONValue(key string) (string, error) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	cValue := C.app_get_syspolicy_string_array_json_value(cKey)
	if cValue == nil {
		return "", errors.New("failed to get syspolicy string array value")
	}
	defer C.free(unsafe.Pointer(cValue))
	return C.GoString(cValue), nil
}

// --- Application registry and CGO glue ---
var (
	appRegistry      = make(map[uintptr]libtailscale.Application)
	appRegistryMutex sync.Mutex
	nextAppHandle    uintptr = 1
)

func registerApp(app libtailscale.Application) uintptr {
	appRegistryMutex.Lock()
	defer appRegistryMutex.Unlock()
	h := nextAppHandle
	nextAppHandle++
	appRegistry[h] = app
	return h
}

func getApp(h uintptr) libtailscale.Application {
	appRegistryMutex.Lock()
	defer appRegistryMutex.Unlock()
	return appRegistry[h]
}

// --- Application CGO wrappers ---
//
//export go_call_local_api
func go_call_local_api(self *C.application, timeoutMillis C.int, method, endpoint *C.char, body unsafe.Pointer) unsafe.Pointer {
	h := uintptr(unsafe.Pointer(self))
	app := getApp(h)
	if app == nil {
		return nil
	}
	goMethod := C.GoString(method)
	goEndpoint := C.GoString(endpoint)
	// TODO: Convert body to InputStream
	resp, err := app.CallLocalAPI(int(timeoutMillis), goMethod, goEndpoint, nil)
	if err != nil {
		return nil
	}
	_ = resp // TODO: Convert resp to C struct
	return nil
}

//export go_notify_policy_changed
func go_notify_policy_changed(self *C.application) {
	h := uintptr(unsafe.Pointer(self))
	app := getApp(h)
	if app == nil {
		return
	}
	app.NotifyPolicyChanged()
}

// --- IPNService CGO wrappers (example) ---
// You would have a registry and wrappers for each interface as above.
// For brevity, only one method is shown for each interface.

//export go_ipn_protect
func go_ipn_protect(self *C.ipn_service, fd C.int32_t) C.int32_t {
	// TODO: Lookup Go object and call Protect
	return 1 // Example: always succeed
}

// --- VPNServiceBuilder CGO wrappers (example) ---
//
//export go_vpn_set_mtu
func go_vpn_set_mtu(self *C.vpn_service_builder, mtu C.int32_t) C.int {
	// TODO: Lookup Go object and call SetMTU
	return 0 // Example: always succeed
}

// --- ParcelFileDescriptor CGO wrappers (example) ---
//
//export go_parcel_detach
func go_parcel_detach(self *C.parcel_file_descriptor) C.int32_t {
	// TODO: Lookup Go object and call Detach
	return 0 // Example: always succeed
}

// --- NotificationCallback CGO wrappers (example) ---
//
//export go_notify_on_notify
func go_notify_on_notify(self *C.notification_callback, data *C.uint8_t, len C.size_t) C.int {
	// TODO: Lookup Go object and call OnNotify
	return 0 // Example: always succeed
}

// --- NotificationManager CGO wrappers (example) ---
//
//export go_notification_manager_stop
func go_notification_manager_stop(self *C.notification_manager) {
	// TODO: Lookup Go object and call Stop
}

// --- InputStream CGO wrappers (example) ---
//
//export go_input_stream_read
func go_input_stream_read(self *C.input_stream, buf *C.uint8_t, len C.size_t) C.ssize_t {
	// TODO: Lookup Go object and call Read
	return 0 // Example: always EOF
}

//export go_input_stream_close
func go_input_stream_close(self *C.input_stream) {
	// TODO: Lookup Go object and call Close
}

// --- LocalAPIResponse CGO wrappers (example) ---
//
//export go_local_api_status_code
func go_local_api_status_code(self *C.local_api_response) C.int {
	// TODO: Lookup Go object and call StatusCode
	return 200 // Example: always 200
}

// --- Factory function ---
//
//export start_with_context
func start_with_context(dataDir, directFileRoot *C.char, vtable *C.app_context_vtable) *C.application_handle {
	ctx := &cgoAppContext{}
	app := libtailscale.Start(C.GoString(dataDir), C.GoString(directFileRoot), ctx)
	h := registerApp(app)

	handle := (*C.application_handle)(C.malloc(C.size_t(unsafe.Sizeof(C.application_handle{}))))
	// vtable must be filled in C, not Go!
	handle.vtable = nil
	handle.self = (*C.application)(unsafe.Pointer(h))
	return handle
}
func main() {}
