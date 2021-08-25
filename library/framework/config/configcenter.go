package config

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

type Configuration map[Key]Instance

var LocalConfig map[Key]Instance

type Instance struct {
	Key     Key
	Value   interface{}
	Version int
}

type Key string

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

func GetInstance(key Key) (Instance, error) {
	instance, ok := LocalConfig[key]
	if !ok {
		return instance, errors.New("undefined config")
	}

	return instance, nil
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

func UpdateLocalConfig(key Key, value interface{}, newVersion int) (bool, int) {
	instance, err := GetInstance(key)
	if err != nil {
		log.Fatal(err)
		return false, -1
	}
	if newVersion < instance.Version {
		return false, instance.Version
	}
	LocalConfig[key] = Instance{key, value, newVersion}
	return true, newVersion
}

func ModifyLocalServiceConfig(key Key, value interface{}) bool {
	return true
}

func ModifyGlobalConfig(key Key, value interface{}) bool {
	return true
}
