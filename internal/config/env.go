package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// loadFromEnv loads configuration from environment variables using reflection
func loadFromEnv(config *Config) error {
	val := reflect.ValueOf(config).Elem()
	return loadStructFromEnv(val)
}

// loadStructFromEnv recursively loads environment variables into a struct
func loadStructFromEnv(val reflect.Value) error {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Get the env tag
		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			continue
		}

		// Handle nested structs
		if field.Kind() == reflect.Struct {
			if err := loadStructFromEnv(field); err != nil {
				return err
			}
			continue
		}

		// Get environment variable value
		envValue := os.Getenv(envTag)
		if envValue == "" {
			continue
		}

		// Set the field value based on its type
		if err := setFieldValue(field, envValue); err != nil {
			return fmt.Errorf("failed to set %s: %v", envTag, err)
		}
	}

	return nil
}

// setFieldValue sets a field's value from a string
func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)

	case reflect.Slice:
		// Handle string slices
		if field.Type().Elem().Kind() == reflect.String {
			values := strings.Split(value, ",")
			slice := reflect.MakeSlice(field.Type(), len(values), len(values))
			for i, v := range values {
				slice.Index(i).SetString(strings.TrimSpace(v))
			}
			field.Set(slice)
		} else {
			return fmt.Errorf("unsupported slice type: %v", field.Type().Elem().Kind())
		}

	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}

	return nil
}
