package util

func GetAsByte(arguments []interface{}, idx int, defaultValue int8) int8 {
	if len(arguments) < (idx + 1) {
		return defaultValue
	}
	return arguments[idx].(int8)
}

func GetAsShort(arguments []interface{}, idx int, defaultValue int16) int16 {
	if len(arguments) < (idx + 1) {
		return defaultValue
	}
	return arguments[idx].(int16)
}

func GetAsInt(arguments []interface{}, idx int, defaultValue int32) int32 {
	if len(arguments) < (idx + 1) {
		return defaultValue
	}
	return arguments[idx].(int32)
}

func GetAsLong(arguments []interface{}, idx int, defaultValue int64) int64 {
	if len(arguments) < (idx + 1) {
		return defaultValue
	}
	return arguments[idx].(int64)
}

func GetAsFloat(arguments []interface{}, idx int, defaultValue float32) float32 {
	if len(arguments) < (idx + 1) {
		return defaultValue
	}
	return arguments[idx].(float32)
}

func GetAsDouble(arguments []interface{}, idx int, defaultValue float64) float64 {
	if len(arguments) < (idx + 1) {
		return defaultValue
	}
	return arguments[idx].(float64)
}

func GetAsString(arguments []interface{}, idx int, defaultValue string) string {
	if len(arguments) < (idx + 1) {
		return defaultValue
	}
	return arguments[idx].(string)
}

func GetAsObject(arguments []interface{}, idx int, defaultValue interface{}) interface{} {
	if len(arguments) < (idx + 1) {
		return defaultValue
	}
	return arguments[idx]
}
