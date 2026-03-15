package config

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	timeDurationType = reflect.TypeOf(time.Duration(0))
	timeTimeType     = reflect.TypeOf(time.Time{})
	urlType          = reflect.TypeOf(url.URL{})
)

// Load populates a struct pointer from environment variables described by env tags.
//
// Supported tag options:
//   - `env:"KEY"` reads KEY into the field
//   - `required` fails when the key is unset
//   - `default=value` uses value when the key is unset
//   - `sep=|` overrides []string separators (default `,`)
//   - `entrysep=;` and `kvsep=:` override map separators (defaults `,` and `=`)
//   - `layout=2006-01-02` defines the time.Time layout
//   - `oneof=a|b|c` constrains string values
//   - `format=bytes` enables byte-size parsing for integer fields
func Load(target any) error {
	if target == nil {
		return errors.New("config target must be a non-nil pointer to struct")
	}

	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("config target must be a non-nil pointer to struct")
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("config target must point to a struct")
	}

	errs, _ := loadStruct(elem, "")
	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

func loadStruct(target reflect.Value, parentPath string) ([]error, bool) {
	var (
		errs    []error
		changed bool
	)

	targetType := target.Type()
	for i := 0; i < target.NumField(); i++ {
		field := target.Field(i)
		structField := targetType.Field(i)

		if structField.PkgPath != "" {
			continue
		}

		fieldPath := structField.Name
		if parentPath != "" {
			fieldPath = parentPath + "." + structField.Name
		}

		opts, ok, err := parseFieldOptions(structField.Tag.Get("env"))
		if err != nil {
			errs = append(errs, fmt.Errorf("field %s: %w", fieldPath, err))
			continue
		}
		if !ok {
			childErrs, childChanged := loadNestedField(field, fieldPath)
			errs = append(errs, childErrs...)
			changed = changed || childChanged
			continue
		}

		fieldChanged, err := assignField(field, fieldPath, opts)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		changed = changed || fieldChanged
	}

	return errs, changed
}

type fieldOptions struct {
	key        string
	required   bool
	hasDefault bool
	defaultVal string
	sep        string
	entrySep   string
	kvSep      string
	layout     string
	oneOf      []string
	format     string
}

func parseFieldOptions(tag string) (fieldOptions, bool, error) {
	if tag == "" {
		return fieldOptions{}, false, nil
	}
	if tag == "-" {
		return fieldOptions{}, false, nil
	}

	parts := splitTag(tag)
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return fieldOptions{}, false, errors.New("env tag must start with a key")
	}

	opts := fieldOptions{
		key:      strings.TrimSpace(parts[0]),
		sep:      ",",
		entrySep: ",",
		kvSep:    "=",
	}

	for _, raw := range parts[1:] {
		part := strings.TrimSpace(raw)
		if part == "" {
			continue
		}

		switch {
		case part == "required":
			opts.required = true
		case strings.HasPrefix(part, "default="):
			opts.hasDefault = true
			opts.defaultVal = strings.TrimPrefix(part, "default=")
		case strings.HasPrefix(part, "sep="):
			opts.sep = strings.TrimPrefix(part, "sep=")
		case strings.HasPrefix(part, "entrysep="):
			opts.entrySep = strings.TrimPrefix(part, "entrysep=")
		case strings.HasPrefix(part, "kvsep="):
			opts.kvSep = strings.TrimPrefix(part, "kvsep=")
		case strings.HasPrefix(part, "layout="):
			opts.layout = strings.TrimPrefix(part, "layout=")
		case strings.HasPrefix(part, "oneof="):
			value := strings.TrimPrefix(part, "oneof=")
			if value != "" {
				opts.oneOf = strings.Split(value, "|")
			}
		case strings.HasPrefix(part, "format="):
			opts.format = strings.TrimPrefix(part, "format=")
		default:
			return fieldOptions{}, false, fmt.Errorf("unsupported env option %q", part)
		}
	}

	return opts, true, nil
}

func splitTag(tag string) []string {
	var (
		parts         []string
		current       strings.Builder
		lastWasOption bool
	)

	for i := 0; i < len(tag); i++ {
		ch := tag[i]
		if ch == ',' {
			part := current.String()
			if len(parts) > 0 && lastWasOption && !isTagOption(part) {
				parts[len(parts)-1] += "," + part
			} else {
				parts = append(parts, part)
			}
			lastWasOption = len(parts) > 1 && isTagOption(parts[len(parts)-1])
			current.Reset()
			continue
		}
		current.WriteByte(ch)
	}
	part := current.String()
	if len(parts) > 0 && lastWasOption && !isTagOption(part) {
		parts[len(parts)-1] += "," + part
	} else {
		parts = append(parts, part)
	}
	return parts
}

func isTagOption(part string) bool {
	part = strings.TrimSpace(part)
	if part == "required" {
		return true
	}
	for _, prefix := range []string{
		"default=",
		"sep=",
		"entrysep=",
		"kvsep=",
		"layout=",
		"oneof=",
		"format=",
	} {
		if strings.HasPrefix(part, prefix) {
			return true
		}
	}
	return false
}

func loadNestedField(field reflect.Value, fieldPath string) ([]error, bool) {
	fieldType := field.Type()

	switch {
	case fieldType.Kind() == reflect.Struct && shouldRecurseIntoStruct(fieldType):
		return loadStruct(field, fieldPath)
	case fieldType.Kind() == reflect.Pointer && fieldType.Elem().Kind() == reflect.Struct && shouldRecurseIntoStruct(fieldType.Elem()):
		if field.IsNil() {
			child := reflect.New(fieldType.Elem())
			errs, changed := loadStruct(child.Elem(), fieldPath)
			if changed {
				field.Set(child)
			}
			return errs, changed
		}
		return loadStruct(field.Elem(), fieldPath)
	default:
		return nil, false
	}
}

func shouldRecurseIntoStruct(fieldType reflect.Type) bool {
	return fieldType != timeTimeType && fieldType != urlType
}

func assignField(field reflect.Value, fieldName string, opts fieldOptions) (bool, error) {
	raw, ok := os.LookupEnv(opts.key)
	if !ok {
		if opts.hasDefault {
			raw = opts.defaultVal
			ok = true
		} else if opts.required {
			return false, fmt.Errorf("field %s: environment variable %q is required", fieldName, opts.key)
		} else {
			return false, nil
		}
	}

	if !field.CanSet() {
		return false, fmt.Errorf("field %s: cannot set value", fieldName)
	}

	if err := setValue(field, raw, opts); err != nil {
		return false, fmt.Errorf("field %s: env %q value %q: %w", fieldName, opts.key, raw, err)
	}

	return true, nil
}

func setValue(field reflect.Value, raw string, opts fieldOptions) error {
	fieldType := field.Type()

	switch {
	case fieldType == timeDurationType:
		value, err := time.ParseDuration(raw)
		if err != nil {
			return err
		}
		field.SetInt(int64(value))
		return nil
	case fieldType == timeTimeType:
		if opts.layout == "" {
			return errors.New("time.Time fields require layout")
		}
		value, err := time.Parse(opts.layout, raw)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(value))
		return nil
	case fieldType == urlType:
		value, err := parseURL(raw)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(*value))
		return nil
	case fieldType.Kind() == reflect.Pointer && fieldType.Elem() == urlType:
		value, err := parseURL(raw)
		if err != nil {
			return err
		}
		ptr := reflect.New(urlType)
		ptr.Elem().Set(reflect.ValueOf(*value))
		field.Set(ptr)
		return nil
	case fieldType.Kind() == reflect.Struct:
		return errors.New("nested structs are not supported")
	case fieldType.Kind() == reflect.String:
		if len(opts.oneOf) > 0 && !slices.Contains(opts.oneOf, raw) {
			return fmt.Errorf("value is not in enum %v", opts.oneOf)
		}
		field.SetString(raw)
		return nil
	case fieldType.Kind() == reflect.Bool:
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		field.SetBool(value)
		return nil
	case fieldType.Kind() == reflect.Int, fieldType.Kind() == reflect.Int8, fieldType.Kind() == reflect.Int16, fieldType.Kind() == reflect.Int32, fieldType.Kind() == reflect.Int64:
		value, err := parseSignedInteger(raw, fieldType.Bits(), opts.format)
		if err != nil {
			return err
		}
		field.SetInt(value)
		return nil
	case fieldType.Kind() == reflect.Uint, fieldType.Kind() == reflect.Uint8, fieldType.Kind() == reflect.Uint16, fieldType.Kind() == reflect.Uint32, fieldType.Kind() == reflect.Uint64:
		value, err := parseUnsignedInteger(raw, fieldType.Bits())
		if err != nil {
			return err
		}
		field.SetUint(value)
		return nil
	case fieldType.Kind() == reflect.Float32, fieldType.Kind() == reflect.Float64:
		value, err := strconv.ParseFloat(raw, fieldType.Bits())
		if err != nil {
			return err
		}
		field.SetFloat(value)
		return nil
	case fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.String:
		value, err := parseStringSlice(raw, opts.sep)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(value))
		return nil
	case fieldType.Kind() == reflect.Map && fieldType.Key().Kind() == reflect.String && fieldType.Elem().Kind() == reflect.String:
		value, err := parseMap(raw, opts.kvSep, opts.entrySep)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(value))
		return nil
	default:
		return fmt.Errorf("unsupported field type %s", fieldType)
	}
}

func parseSignedInteger(raw string, bits int, format string) (int64, error) {
	if format == "bytes" {
		value, err := parseBytes(raw)
		if err != nil {
			return 0, err
		}
		maxValue, minValue := signedRange(bits)
		if value > maxValue || value < minValue {
			return 0, fmt.Errorf("value %d overflows %d-bit signed integer", value, bits)
		}
		return value, nil
	}
	return strconv.ParseInt(raw, 10, bits)
}

func parseUnsignedInteger(raw string, bits int) (uint64, error) {
	return strconv.ParseUint(raw, 10, bits)
}

func signedRange(bits int) (int64, int64) {
	if bits <= 0 || bits >= 64 {
		return math.MaxInt64, math.MinInt64
	}
	maxValue := int64(1)<<(bits-1) - 1
	minValue := -int64(1) << (bits - 1)
	return maxValue, minValue
}

func parseURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid URL %q: expected scheme and host", raw)
	}
	return u, nil
}

func parseStringSlice(raw, sep string) ([]string, error) {
	if sep == "" {
		return nil, errors.New("separator cannot be empty")
	}
	if strings.TrimSpace(raw) == "" {
		return []string{}, nil
	}

	parts := strings.Split(raw, sep)
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		values = append(values, value)
	}
	return values, nil
}

func parseMap(raw, sepKV, sepEntry string) (map[string]string, error) {
	if sepKV == "" || sepEntry == "" {
		return nil, errors.New("separators cannot be empty")
	}
	if strings.TrimSpace(raw) == "" {
		return map[string]string{}, nil
	}

	entries := strings.Split(raw, sepEntry)
	result := make(map[string]string, len(entries))
	for _, entry := range entries {
		pair := strings.SplitN(entry, sepKV, 2)
		if len(pair) != 2 {
			return nil, fmt.Errorf("invalid map entry %q", strings.TrimSpace(entry))
		}
		key := strings.TrimSpace(pair[0])
		value := strings.TrimSpace(pair[1])
		if key == "" {
			return nil, fmt.Errorf("invalid map entry %q: empty key", strings.TrimSpace(entry))
		}
		result[key] = value
	}
	return result, nil
}

func parseBytes(raw string) (int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, errors.New("byte size cannot be empty")
	}

	numberPart := value
	unitPart := ""
	for i := len(value) - 1; i >= 0; i-- {
		ch := value[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			continue
		}
		numberPart = strings.TrimSpace(value[:i+1])
		unitPart = strings.ToUpper(strings.TrimSpace(value[i+1:]))
		break
	}
	if numberPart == value {
		unitPart = ""
	}

	number, err := strconv.ParseFloat(numberPart, 64)
	if err != nil {
		return 0, err
	}
	if number < 0 {
		return 0, errors.New("byte size cannot be negative")
	}

	multiplier, ok := map[string]float64{
		"":    1,
		"B":   1,
		"K":   1024,
		"KB":  1024,
		"KIB": 1024,
		"M":   1024 * 1024,
		"MB":  1024 * 1024,
		"MIB": 1024 * 1024,
		"G":   1024 * 1024 * 1024,
		"GB":  1024 * 1024 * 1024,
		"GIB": 1024 * 1024 * 1024,
		"T":   1024 * 1024 * 1024 * 1024,
		"TB":  1024 * 1024 * 1024 * 1024,
		"TIB": 1024 * 1024 * 1024 * 1024,
	}[unitPart]
	if !ok {
		return 0, fmt.Errorf("unsupported byte unit %q", unitPart)
	}

	size := number * multiplier
	if size > math.MaxInt64 {
		return 0, fmt.Errorf("byte size %q overflows int64", raw)
	}
	return int64(size), nil
}
