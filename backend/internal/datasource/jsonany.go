package datasource

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func anyToString(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}

func anyToFloat64(value interface{}) (float64, bool) {
	switch typed := value.(type) {
	case nil:
		return 0, false
	case float64:
		return typed, true
	case int64:
		return float64(typed), true
	case int:
		return float64(typed), true
	case json.Number:
		floatVal, err := typed.Float64()
		if err != nil {
			return 0, false
		}
		return floatVal, true
	case string:
		if typed == "" {
			return 0, false
		}
		floatVal, err := strconv.ParseFloat(typed, 64)
		if err != nil {
			return 0, false
		}
		return floatVal, true
	}

	return 0, false
}
