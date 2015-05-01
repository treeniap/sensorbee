package tuple

import (
	"fmt"
	"math"
)

// Constants defining the largest/smallest float64 numbers that can
// be converted.
const (
	MaxConvFloat64 = float64(math.MaxInt64)
	MinConvFloat64 = float64(math.MinInt64)
)

// ToBool converts a given Value to a bool, if possible. The conversion
// rules are similar to those in Python:
//
//  * Null: false
//  * Bool: actual boolean value
//  * Int: true if non-zero
//  * Float: true if non-zero
//  * String: true if non-empty
//  * Blob: true if non-empty
//  * Timestamp: true if IsZero() is false
//  * Array: true if non-empty
//  * Map: true if non-empty
func ToBool(v Value) (bool, error) {
	defaultValue := false
	switch v.Type() {
	case TypeNull:
		return false, nil
	case TypeBool:
		val, e := v.AsBool()
		if e != nil {
			return defaultValue, e
		}
		return val, nil
	case TypeInt:
		val, e := v.AsInt()
		if e != nil {
			return defaultValue, e
		}
		return val != 0, nil
	case TypeFloat:
		val, e := v.AsFloat()
		if e != nil {
			return defaultValue, e
		}
		return val != 0.0, nil
	case TypeString:
		val, e := v.AsString()
		if e != nil {
			return defaultValue, e
		}
		return len(val) > 0, nil
	case TypeBlob:
		val, e := v.AsBlob()
		if e != nil {
			return defaultValue, e
		}
		return len(val) > 0, nil
	case TypeTimestamp:
		val, e := v.AsTimestamp()
		if e != nil {
			return defaultValue, e
		}
		return !val.IsZero(), nil
	case TypeArray:
		val, e := v.AsArray()
		if e != nil {
			return defaultValue, e
		}
		return len(val) > 0, nil
	case TypeMap:
		val, e := v.AsMap()
		if e != nil {
			return defaultValue, e
		}
		return len(val) > 0, nil
	default:
		return defaultValue,
			fmt.Errorf("cannot convert %T to bool", v)
	}
}

// ToInt converts a given Value to an int64, if possible. The conversion
// rules are as follows:
//
//  * Null: 0
//  * Bool: 0 if false, 1 if true
//  * Int: actual value
//  * Float: conversion as done by int64(value)
//    (values outside of valid int64 bounds will lead to an error)
//  * String: (error)
//  * Blob: (error)
//  * Timestamp: Unix time, see time.Time.Unix()
//  * Array: (error)
//  * Map: (error)
func ToInt(v Value) (int64, error) {
	defaultValue := int64(0)
	switch v.Type() {
	case TypeNull:
		return 0, nil
	case TypeBool:
		val, e := v.AsBool()
		if e != nil {
			return defaultValue, e
		}
		if val {
			return 1, nil
		}
		return 0, nil
	case TypeInt:
		val, e := v.AsInt()
		if e != nil {
			return defaultValue, e
		}
		return val, nil
	case TypeFloat:
		val, e := v.AsFloat()
		if e != nil {
			return defaultValue, e
		}
		if val >= MinConvFloat64 && val <= MaxConvFloat64 {
			return int64(val), nil
		} else {
			return defaultValue,
				fmt.Errorf("%v is out of bounds for int64 conversion", val)
		}
	case TypeTimestamp:
		val, e := v.AsTimestamp()
		if e != nil {
			return defaultValue, e
		}
		return val.Unix(), nil
	default:
		return defaultValue,
			fmt.Errorf("cannot convert %T to int64", v)
	}
}
