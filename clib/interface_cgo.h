#ifndef TAILSCALE_INTERFACE_CGO_H
#define TAILSCALE_INTERFACE_CGO_H

#include <stdint.h>
#include <stddef.h>

// AppContext C function declarations (for Go->C calls)
void app_log(const char *tag, const char *message);
int app_encrypt_to_pref(const char *key, const char *value);
char *app_decrypt_from_pref(const char *key);
char *app_get_os_version(void);
char *app_get_model_name(void);
char *app_get_install_source(void);
int app_should_use_google_dns_fallback(void);
int app_is_chrome_os(void);
char *app_get_interfaces_as_string(void);
char *app_get_platform_dns_config(void);
char *app_get_syspolicy_string_value(const char *key);
int app_get_syspolicy_boolean_value(const char *key);
char *app_get_syspolicy_string_array_json_value(const char *key);

// --- AppContext vtable (from app_context.h) ---
typedef struct app_context app_context;
typedef void (*log_fn)(app_context *self, const char *tag, const char *message);
typedef int (*encrypt_to_pref_fn)(app_context *self, const char *key, const char *value);
typedef char *(*decrypt_from_pref_fn)(app_context *self, const char *key);
typedef char *(*get_os_version_fn)(app_context *self);
typedef char *(*get_model_name_fn)(app_context *self);
typedef char *(*get_install_source_fn)(app_context *self);
typedef int (*should_use_google_dns_fallback_fn)(app_context *self);
typedef int (*is_chrome_os_fn)(app_context *self);
typedef char *(*get_interfaces_as_string_fn)(app_context *self);
typedef char *(*get_platform_dns_config_fn)(app_context *self);
typedef char *(*get_syspolicy_string_value_fn)(app_context *self, const char *key);
typedef int (*get_syspolicy_boolean_value_fn)(app_context *self, const char *key);
typedef char *(*get_syspolicy_string_array_json_value_fn)(app_context *self, const char *key);

typedef struct
{
	log_fn log;
	encrypt_to_pref_fn encrypt_to_pref;
	decrypt_from_pref_fn decrypt_from_pref;
	get_os_version_fn get_os_version;
	get_model_name_fn get_model_name;
	get_install_source_fn get_install_source;
	should_use_google_dns_fallback_fn should_use_google_dns_fallback;
	is_chrome_os_fn is_chrome_os;
	get_interfaces_as_string_fn get_interfaces_as_string;
	get_platform_dns_config_fn get_platform_dns_config;
	get_syspolicy_string_value_fn get_syspolicy_string_value;
	get_syspolicy_boolean_value_fn get_syspolicy_boolean_value;
	get_syspolicy_string_array_json_value_fn get_syspolicy_string_array_json_value;
} app_context_vtable;

typedef struct
{
	app_context_vtable *vtable;
	app_context *self;
} app_context_handle;

// --- Application vtable ---
typedef struct application application;
typedef void *(*call_local_api_fn)(application *self, int timeoutMillis, const char *method, const char *endpoint, void *body);
typedef void *(*call_local_api_multipart_fn)(application *self, int timeoutMillis, const char *method, const char *endpoint, void *parts);
typedef void (*notify_policy_changed_fn)(application *self);
typedef void *(*watch_notifications_fn)(application *self, int mask, void *cb);

typedef struct
{
	call_local_api_fn call_local_api;
	call_local_api_multipart_fn call_local_api_multipart;
	notify_policy_changed_fn notify_policy_changed;
	watch_notifications_fn watch_notifications;
} application_vtable;

typedef struct
{
	application_vtable *vtable;
	application *self;
} application_handle;

// --- IPNService vtable ---
typedef struct ipn_service ipn_service;
typedef int32_t (*protect_fn)(ipn_service *self, int32_t fd);
typedef void *(*new_builder_fn)(ipn_service *self);
typedef void (*ipn_close_fn)(ipn_service *self);
typedef void (*disconnect_vpn_fn)(ipn_service *self);
typedef void (*update_vpn_status_fn)(ipn_service *self, int status);

typedef struct
{
	protect_fn protect;
	new_builder_fn new_builder;
	ipn_close_fn close;
	disconnect_vpn_fn disconnect_vpn;
	update_vpn_status_fn update_vpn_status;
} ipn_service_vtable;

typedef struct
{
	ipn_service_vtable *vtable;
	ipn_service *self;
} ipn_service_handle;

// --- VPNServiceBuilder vtable ---
typedef struct vpn_service_builder vpn_service_builder;
typedef int (*set_mtu_fn)(vpn_service_builder *self, int32_t mtu);
typedef int (*add_dns_server_fn)(vpn_service_builder *self, const char *server);
// ... add other builder methods as needed

typedef struct
{
	set_mtu_fn set_mtu;
	add_dns_server_fn add_dns_server;
	// ... add other builder methods as needed
} vpn_service_builder_vtable;

typedef struct
{
	vpn_service_builder_vtable *vtable;
	vpn_service_builder *self;
} vpn_service_builder_handle;

// --- ParcelFileDescriptor vtable ---
typedef struct parcel_file_descriptor parcel_file_descriptor;
typedef int32_t (*detach_fn)(parcel_file_descriptor *self);

typedef struct
{
	detach_fn detach;
} parcel_file_descriptor_vtable;

typedef struct
{
	parcel_file_descriptor_vtable *vtable;
	parcel_file_descriptor *self;
} parcel_file_descriptor_handle;

// --- NotificationCallback vtable ---
typedef struct notification_callback notification_callback;
typedef int (*on_notify_fn)(notification_callback *self, const uint8_t *data, size_t len);

typedef struct
{
	on_notify_fn on_notify;
} notification_callback_vtable;

typedef struct
{
	notification_callback_vtable *vtable;
	notification_callback *self;
} notification_callback_handle;

// --- NotificationManager vtable ---
typedef struct notification_manager notification_manager;
typedef void (*stop_fn)(notification_manager *self);

typedef struct
{
	stop_fn stop;
} notification_manager_vtable;

typedef struct
{
	notification_manager_vtable *vtable;
	notification_manager *self;
} notification_manager_handle;

// --- InputStream vtable ---
typedef struct input_stream input_stream;
typedef ssize_t (*read_fn)(input_stream *self, uint8_t *buf, size_t len);
typedef void (*input_stream_close_fn)(input_stream *self);

typedef struct
{
	read_fn read;
	input_stream_close_fn close;
} input_stream_vtable;

typedef struct
{
	input_stream_vtable *vtable;
	input_stream *self;
} input_stream_handle;

// --- FilePart struct ---
typedef struct file_part file_part;
typedef struct
{
	int64_t content_length;
	const char *filename;
	input_stream_handle *body;
	const char *content_type;
} file_part_struct;

// --- LocalAPIResponse vtable ---
typedef struct local_api_response local_api_response;
typedef int (*status_code_fn)(local_api_response *self);
typedef void *(*body_bytes_fn)(local_api_response *self, size_t *out_len);
typedef input_stream_handle *(*body_input_stream_fn)(local_api_response *self);

typedef struct
{
	status_code_fn status_code;
	body_bytes_fn body_bytes;
	body_input_stream_fn body_input_stream;
} local_api_response_vtable;

typedef struct
{
	local_api_response_vtable *vtable;
	local_api_response *self;
} local_api_response_handle;

#endif // TAILSCALE_INTERFACE_CGO_H