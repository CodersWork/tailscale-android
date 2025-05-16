#include "interface_cgo.h"
#include <stdio.h>
#include <string.h>

#include "hilog/log.h"
#undef LOG_DOMAIN
#define LOG_DOMAIN 0x3200 // 全局domain宏，标识业务领域

// --- AppContext named C function implementations (Go will call these directly) ---
void app_log(const char *tag, const char *message)
{
	// printf("[LOG][%s] %s\n", tag, message);
	OH_LOG_Print(LOG_APP, LOG_DOMAIN, tag, "%{public}s", message);
}
int app_encrypt_to_pref(const char *key, const char *value) { return 0; }
char *app_decrypt_from_pref(const char *key) { return strdup("decrypted"); }
char *app_get_os_version(void) { return strdup("1.0.0"); }
char *app_get_model_name(void) { return strdup("Model"); }
char *app_get_install_source(void) { return strdup("store"); }
int app_should_use_google_dns_fallback(void) { return 1; }
int app_is_chrome_os(void) { return 0; }
char *app_get_interfaces_as_string(void) { return strdup("eth0"); }
char *app_get_platform_dns_config(void) { return strdup("8.8.8.8"); }
char *app_get_syspolicy_string_value(const char *key) { return strdup("policy"); }
int app_get_syspolicy_boolean_value(const char *key) { return 0; }
char *app_get_syspolicy_string_array_json_value(const char *key) { return strdup("[\"a\"]"); }

// --- Application vtable: fill with Go-exported functions (externs) ---
extern void *go_call_local_api(application *, int, const char *, const char *, void *);
extern void *go_call_local_api_multipart(application *, int, const char *, const char *, void *);
extern void go_notify_policy_changed(application *);
extern void *go_watch_notifications(application *, int, void *);

// --- IPNService example implementations ---
int32_t my_protect(ipn_service *self, int32_t fd)
{
	printf("Protect %d\n", fd);
	return 1;
}
void *my_new_builder(ipn_service *self)
{
	printf("NewBuilder\n");
	return NULL;
}
void my_ipn_close(ipn_service *self) { printf("Close\n"); }
void my_disconnect_vpn(ipn_service *self) { printf("DisconnectVPN\n"); }
void my_update_vpn_status(ipn_service *self, int status) { printf("UpdateVpnStatus %d\n", status); }

// --- VPNServiceBuilder example implementations ---
int my_set_mtu(vpn_service_builder *self, int32_t mtu)
{
	printf("SetMTU %d\n", mtu);
	return 0;
}
int my_add_dns_server(vpn_service_builder *self, const char *server)
{
	printf("AddDNSServer %s\n", server);
	return 0;
}

// --- ParcelFileDescriptor example implementations ---
int32_t my_detach(parcel_file_descriptor *self)
{
	printf("Detach\n");
	return 0;
}

// --- NotificationCallback example implementations ---
int my_on_notify(notification_callback *self, const uint8_t *data, size_t len)
{
	printf("OnNotify: %zu bytes\n", len);
	return 0;
}

// --- NotificationManager example implementations ---
void my_stop(notification_manager *self) { printf("Stop\n"); }

// --- InputStream example implementations ---
ssize_t my_read(input_stream *self, uint8_t *buf, size_t len)
{
	printf("Read %zu\n", len);
	return 0;
}
void my_input_stream_close(input_stream *self) { printf("InputStream Close\n"); }

// --- LocalAPIResponse example implementations ---
int my_status_code(local_api_response *self) { return 200; }
void *my_body_bytes(local_api_response *self, size_t *out_len)
{
	*out_len = 0;
	return NULL;
}
input_stream_handle *my_body_input_stream(local_api_response *self) { return NULL; }

int main()
{
	// AppContext: nothing to fill, just implement the named C functions above.

	// Application vtable: fill with Go-exported functions
	application_vtable app_vtable = {
		.call_local_api = go_call_local_api,
		.call_local_api_multipart = go_call_local_api_multipart,
		.notify_policy_changed = go_notify_policy_changed,
		.watch_notifications = go_watch_notifications};

	// Example: create a dummy Application handle (normally returned from Go)
	application_handle app_handle = {&app_vtable, NULL};

	// Example: call Application methods via vtable
	if (app_handle.vtable && app_handle.vtable->notify_policy_changed)
		app_handle.vtable->notify_policy_changed(app_handle.self);

	// IPNService vtable
	ipn_service_vtable ipn_vtable = {
		.protect = my_protect,
		.new_builder = my_new_builder,
		.close = my_ipn_close,
		.disconnect_vpn = my_disconnect_vpn,
		.update_vpn_status = my_update_vpn_status};

	// VPNServiceBuilder vtable
	vpn_service_builder_vtable vpn_vtable = {
		.set_mtu = my_set_mtu,
		.add_dns_server = my_add_dns_server};

	// ParcelFileDescriptor vtable
	parcel_file_descriptor_vtable parcel_vtable = {
		.detach = my_detach};

	// NotificationCallback vtable
	notification_callback_vtable notify_cb_vtable = {
		.on_notify = my_on_notify};

	// NotificationManager vtable
	notification_manager_vtable notify_mgr_vtable = {
		.stop = my_stop};

	// InputStream vtable
	input_stream_vtable input_vtable = {
		.read = my_read,
		.close = my_input_stream_close};

	// LocalAPIResponse vtable
	local_api_response_vtable resp_vtable = {
		.status_code = my_status_code,
		.body_bytes = my_body_bytes,
		.body_input_stream = my_body_input_stream};

	// Example: call IPNService methods
	ipn_service_handle ipn_handle = {&ipn_vtable, NULL};
	ipn_handle.vtable->protect(ipn_handle.self, 42);
	ipn_handle.vtable->new_builder(ipn_handle.self);
	ipn_handle.vtable->close(ipn_handle.self);
	ipn_handle.vtable->disconnect_vpn(ipn_handle.self);
	ipn_handle.vtable->update_vpn_status(ipn_handle.self, 1);
	// ... repeat for other interfaces as needed ...
	return 0;
}