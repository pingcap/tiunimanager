package config

import "errors"

var LocalConfig map[Key]Instance

type Instance struct {
	Key     Key
	Value   interface{}
	Version int
}

type Key string

const (
	KEY_REGISTRY_ADDRESS Key = "config_registry_address"

	KEY_API_LOG     = "config_key_api_log"
	KEY_CLUSTER_LOG = "config_key_cluster_log"
	KEY_METADB_LOG  = "config_key_metadb_log"
	KEY_TIUPLIB_LOG = "config_key_tiupmgr_log"
	KEY_DEFAULT_LOG = "config_key_default_log"

	KEY_API_PORT     = "config_key_api_port"
	KEY_CLUSTER_PORT = "config_key_cluster_port"
	KEY_MANAGER_PORT = "config_key_manager_port"
	KEY_METADB_PORT  = "config_key_metadb_port"

	KEY_CERTIFICATES     = "config_key_Certificates"
	KEY_SQLITE_FILE_PATH = "config_key_sqlite_file_path"

	KEY_SERVER_ID           = "config_server_id"
	KEY_PREFIX_SERVICE_PORT = "config_service_port"
)

func CreateInstance(key Key, value interface{}) Instance {
	return Instance{
		key, value, 0,
	}
}

func Get(key Key) (interface{}, error) {
	instance, ok := LocalConfig[key]
	if !ok {
		return nil, errors.New("undefined config")
	}

	return instance.Value, nil
}

func GetWithDefault(key Key, value interface{}) interface{} {
	instance, ok := LocalConfig[key]
	if !ok {
		return value
	}

	return instance.Value
}

func GetStringWithDefault(key Key, value string) string {
	result := GetWithDefault(key, value)
	return result.(string)
}

func GetIntegerWithDefault(key Key, value int) int {
	result := GetWithDefault(key, value)
	return result.(int)
}

func UpdateLocal(key Key, value interface{}, newVersion int) bool {
	return true
}

func ModifyLocalServiceConfig(key Key, value interface{}) bool {
	return true
}

func ModifyGlobalConfig(key Key, value interface{}) bool {
	return true
}
