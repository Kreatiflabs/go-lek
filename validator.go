package golek

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Validator defines the interface for struct validation.
type Validator interface {
	Validate(v any) error
}

// DefaultValidator validates structs using the `validate` struct tag.
type DefaultValidator struct{}

// Validate checks fields based on the `validate` tag.
func (v *DefaultValidator) Validate(s any) error {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	var errs ValidationErrors

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct) {
				if err := v.Validate(field.Interface()); err != nil {
					if ve, ok := err.(ValidationErrors); ok {
						errs = append(errs, ve...)
					}
				}
			}
			continue
		}

		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			if err := validateRule(rule, field, fieldType.Name); err != nil {
				errs = append(errs, *err)
			}
		}

		if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct) {
			if err := v.Validate(field.Interface()); err != nil {
				if ve, ok := err.(ValidationErrors); ok {
					errs = append(errs, ve...)
				}
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func validateRule(rule string, val reflect.Value, fieldName string) *ValidationError {
	if val.Kind() == reflect.Ptr {
		if val.IsNil() && rule == "required" {
			return &ValidationError{Field: fieldName, Tag: "required", Message: "field is required"}
		}
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	parts := strings.SplitN(rule, "=", 2)
	rName := parts[0]
	rVal := ""
	if len(parts) > 1 {
		rVal = parts[1]
	}

	isZero := val.IsZero()

	switch rName {
	case "required":
		if isZero {
			return &ValidationError{Field: fieldName, Tag: "required", Message: "field is required"}
		}
	case "email":
		if !isZero && val.Kind() == reflect.String {
			matched, _ := regexp.MatchString(`^[^@\s]+@[^@\s]+\.[^@\s]+$`, val.String())
			if !matched {
				return &ValidationError{Field: fieldName, Tag: "email", Value: val.String(), Message: "invalid email format"}
			}
		}
	case "min":
		if !isZero {
			min, _ := strconv.ParseFloat(rVal, 64)
			if val.Kind() == reflect.String && float64(val.Len()) < min {
				return &ValidationError{Field: fieldName, Tag: "min", Value: val.String(), Message: "length must be at least " + rVal}
			} else if isNumeric(val.Kind()) && getNumeric(val) < min {
				return &ValidationError{Field: fieldName, Tag: "min", Value: getNumeric(val), Message: "value must be at least " + rVal}
			}
		}
	case "max":
		if !isZero {
			max, _ := strconv.ParseFloat(rVal, 64)
			if val.Kind() == reflect.String && float64(val.Len()) > max {
				return &ValidationError{Field: fieldName, Tag: "max", Value: val.String(), Message: "length must be at most " + rVal}
			} else if isNumeric(val.Kind()) && getNumeric(val) > max {
				return &ValidationError{Field: fieldName, Tag: "max", Value: getNumeric(val), Message: "value must be at most " + rVal}
			}
		}
	case "gte":
		if !isZero && isNumeric(val.Kind()) {
			min, _ := strconv.ParseFloat(rVal, 64)
			if getNumeric(val) < min {
				return &ValidationError{Field: fieldName, Tag: "gte", Value: getNumeric(val), Message: "value must be >= " + rVal}
			}
		}
	case "lte":
		if !isZero && isNumeric(val.Kind()) {
			max, _ := strconv.ParseFloat(rVal, 64)
			if getNumeric(val) > max {
				return &ValidationError{Field: fieldName, Tag: "lte", Value: getNumeric(val), Message: "value must be <= " + rVal}
			}
		}
	case "len":
		if !isZero && val.Kind() == reflect.String {
			l, _ := strconv.Atoi(rVal)
			if val.Len() != l {
				return &ValidationError{Field: fieldName, Tag: "len", Value: val.String(), Message: "length must be exactly " + rVal}
			}
		}
	case "oneof":
		if !isZero {
			opts := strings.Split(rVal, " ")
			strVal := ""
			if val.Kind() == reflect.String {
				strVal = val.String()
			} else if isNumeric(val.Kind()) {
				strVal = strconv.FormatFloat(getNumeric(val), 'f', -1, 64)
			}
			found := false
			for _, o := range opts {
				if o == strVal {
					found = true
					break
				}
			}
			if !found {
				return &ValidationError{Field: fieldName, Tag: "oneof", Value: strVal, Message: "must be one of: " + rVal}
			}
		}
	// Simplified url, uuid, alpha, alphanum, numeric for brevity, can be expanded with real regex
	}
	return nil
}

func isNumeric(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func getNumeric(v reflect.Value) float64 {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		return v.Float()
	}
	return 0
}
