package common

import (
	"fmt"
	"strconv"
)

func ToString(i interface{}) string {
	switch i.(type) {
	case string:
		return i.(string)
	case int:
		return fmt.Sprintf("%d", i.(int))
	case int64:
		return fmt.Sprintf("%d", i.(int64))
	case float64:
		return strconv.FormatFloat(i.(float64), 'f', 6, 64)

	}
	return ""
}

func ToInt64(i interface{}) int64 {
	switch i.(type) {
	case int64:
		return i.(int64)
	case int:
		return int64(i.(int))
	case string:
		result, err := strconv.ParseInt(i.(string), 10, 0)
		if err == nil {
			return result
		}
	case float64:
		return int64(i.(float64))
	}

	return 0
}

func ToInt(i interface{}) int {
	switch i.(type) {
	case int64:
		return int(i.(int64))
	case int:
		return i.(int)
	case string:
		result, err := strconv.Atoi(i.(string))
		if err == nil {
			return result
		}
	case float64:
		return int(i.(float64))
	}

	return 0
}
