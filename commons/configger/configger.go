package configger

import (
	"github.com/micro/go-micro/config"
	"github.com/micro/go-micro/config/reader"
)

func Load(filePath string) error {
	return config.LoadFile(filePath)
}

func GetValue(configPath ...string) reader.Value {
	return config.Get(configPath...)
}

func GetString(defaultValue string, configPath ...string) string {
	return config.Get(configPath...).String(defaultValue)
}

func GetInt(defaultValue int, configPath ...string) int {
	return config.Get(configPath...).Int(defaultValue)
}

func GetBool(defaultValue bool, configPath ...string) bool {
	return config.Get(configPath...).Bool(defaultValue)
}
