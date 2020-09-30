package pager

import (
	"strconv"
	"strings"
)

// StringToInt convert string to int
//  In fact, if the string type data is not passed in, then the int type data will not be returned here, but the original data will be returned.
func StringToInt(val interface{}) interface{} {
	v, ok := val.(string)
	if !ok {
		return val
	}
	rs, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return rs
}

// StringToFloat32 convert string to float32
//  In fact, if the string type data is not passed in, then the int32 type data will not be returned here, but the original data will be returned.
func StringToFloat32(val interface{}) interface{} {
	v, ok := val.(string)
	if !ok {
		return val
	}
	rs, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return 0
	}
	return rs
}

// StringToFloat64 convert string to float64
//  In fact, if the string type data is not passed in, then the float64 type data will not be returned here, but the original data will be returned.
func StringToFloat64(val interface{}) interface{} {
	v, ok := val.(string)
	if !ok {
		return val
	}
	rs, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return rs
}

// StringToBool convert string to bool
func StringToBool(val interface{}) interface{} {
	if strings.ToLower(val.(string)) == "true" || val == "1" || (val != "0" && len(val.(string)) > 0) {
		return true
	}
	return false
}
