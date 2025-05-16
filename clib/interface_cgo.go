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
	"errors"
	"sync"
	"unsafe"

	"github.com/tailscale/tailscale-android/libtailscale"
)

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
	cKey := C.CString(key)
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cKey))
	defer C.free(unsafe.Pointer(cValue))
	result := C.app_encrypt_to_pref(cKey, cValue)
	if result != 0 {
		return errors.New("failed to encrypt to pref")
	}
	return nil
}
func (ctx *cgoAppContext) DecryptFromPref(key string) (string, error) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	cValue := C.app_decrypt_from_pref(cKey)
	if cValue == nil {
		return "", errors.New("failed to decrypt from pref")
	}
	defer C.free(unsafe.Pointer(cValue))
	return C.GoString(cValue), nil
}
func (ctx *cgoAppContext) GetOSVersion() (string, error) {
	cValue := C.app_get_os_version()
	if cValue == nil {
		return "", errors.New("failed to get OS version")
	}
	defer C.free(unsafe.Pointer(cValue))
	return C.GoString(cValue), nil
}
func (ctx *cgoAppContext) GetModelName() (string, error) {
	cValue := C.app_get_model_name()
	if cValue == nil {
		return "", errors.New("failed to get model name")
	}
	defer C.free(unsafe.Pointer(cValue))
	return C.GoString(cValue), nil
}
func (ctx *cgoAppContext) GetInstallSource() string {
	cValue := C.app_get_install_source()
	if cValue == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(cValue))
	return C.GoString(cValue)
}
func (ctx *cgoAppContext) ShouldUseGoogleDNSFallback() bool {
	return C.app_should_use_google_dns_fallback() != 0
}
func (ctx *cgoAppContext) IsChromeOS() (bool, error) {
	result := C.app_is_chrome_os()
	return result != 0, nil
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
