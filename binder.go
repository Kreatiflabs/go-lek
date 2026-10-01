package golek

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

// Binder defines the interface for request data binding.
type Binder interface {
	Bind(c *Context, v any) error
}

// DefaultBinder implements Binder with support for JSON body, path params, query params, and headers.
type DefaultBinder struct{}

// Bind binds data from the request into the provided struct pointer.
func (b *DefaultBinder) Bind(c *Context, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return errors.New("binder: target must be a non-nil pointer to a struct")
	}
	rt := rv.Elem().Type()
	re := rv.Elem()

	// Bind path params, query params, and headers
	for i := 0; i < rt.NumField(); i++ {
		fieldInfo := rt.Field(i)
		fieldVal := re.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		if paramTag := fieldInfo.Tag.Get("param"); paramTag != "" {
			if val := c.Param(paramTag); val != "" {
				if err := setField(fieldVal, val); err != nil {
					return NewHTTPError(400, "Invalid param value for "+paramTag)
				}
			}
		}

		if queryTag := fieldInfo.Tag.Get("query"); queryTag != "" {
			if val := c.Query(queryTag); val != "" {
				if err := setField(fieldVal, val); err != nil {
					return NewHTTPError(400, "Invalid query value for "+queryTag)
				}
			}
		}

		if headerTag := fieldInfo.Tag.Get("header"); headerTag != "" {
			if val := c.Header(headerTag); val != "" {
				if err := setField(fieldVal, val); err != nil {
					return NewHTTPError(400, "Invalid header value for "+headerTag)
				}
			}
		}
	}

	// Bind JSON body if applicable
	if req := c.Request; req != nil && req.Body != nil {
		contentType := req.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/json") {
			// We might need to ensure body isn't empty, but json.NewDecoder handles EOF gracefully or we can catch it
			err := json.NewDecoder(req.Body).Decode(v)
			if err != nil && err.Error() != "EOF" {
				return NewHTTPError(400, "Invalid JSON body: "+err.Error())
			}
		}
	}

	return nil
}

func setField(field reflect.Value, val string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			// Handle comma-separated list or just append
			vals := strings.Split(val, ",")
			slice := reflect.MakeSlice(field.Type(), len(vals), len(vals))
			for i, v := range vals {
				slice.Index(i).SetString(strings.TrimSpace(v))
			}
			field.Set(slice)
		}
	}
	return nil
}
