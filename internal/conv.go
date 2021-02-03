package internal

import (
	"fmt"
	"reflect"
	"strconv"
)

// InterfaceToString converts a value having interface type to string
func InterfaceToString(i interface{}) (string, error) {
	if i == nil {
		return "", ErrInvalidParams
	}

	switch i.(type) {
	case int:
		return strconv.FormatInt(int64(i.(int)), 10), nil
	case int64:
		return strconv.FormatInt(i.(int64), 10), nil
	case int32:
		return strconv.FormatInt(int64(i.(int32)), 10), nil
	case float32:
		return strconv.FormatFloat(float64(i.(float32)), 'f', -1, 64), nil
	case float64:
		return strconv.FormatFloat(i.(float64), 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(i.(bool)), nil
	case string:
		return i.(string), nil
	default:
		return "", ErrUnknownValue
	}
}

// InterfaceToMap converts interface{} type to map[string]string
func InterfaceToMap(m interface{}) (map[string]string, error) {
	result := make(map[string]string)
	if m == nil {
		return result, nil
	}

	rv := reflect.ValueOf(m)
	switch rv.Type().Kind() {
	case reflect.Map:
		keys := rv.MapKeys()
		for _, k := range keys {
			name, err := InterfaceToString(k.Interface())
			if err != nil {
				return nil, err
			}
			vvof := reflect.ValueOf(rv.MapIndex(k).Interface())
			switch vvof.Type().Kind() {
			case reflect.Map:
				innerMap, err := InterfaceToMap(rv.MapIndex(k).Interface())
				if err != nil {
					return nil, err
				}
				for kk, vv := range innerMap {
					result[name+"."+kk] = vv
				}
			case reflect.Slice:
				result[name] = fmt.Sprintf("%v", rv.MapIndex(k).Interface())
			default:
				vv, err := InterfaceToString(rv.MapIndex(k).Interface())
				if err != nil {
					return nil, err
				}
				result[name] = vv
			}
		}
	default:
		return nil, ErrInvalidParams
	}

	return result, nil
}
