package envchain

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetEnv(key string) EnvContainer {
	if val, ok := os.LookupEnv(key); ok {
		return EnvContainer{key: key, value: val, ok: true}
	}
	return EnvContainer{key: key, value: "", ok: false}
}

type EnvContainer struct {
	key   string
	value string
	ok    bool
}

func (c EnvContainer) WithDefault(def string) EnvContainer {
	if c.ok {
		return c
	}
	c.value = def
	return c
}

func (c EnvContainer) Ok() bool {
	return c.ok
}

func (c EnvContainer) Required() EnvContainer {
	if !c.ok {
		panic(fmt.Sprintf("required environment variable %q is not set", c.key))
	}
	return c
}

func (c EnvContainer) asString() string {
	return c.value
}

func (c EnvContainer) asNumber() (float64, error) {
	n, err := strconv.ParseFloat(c.value, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (c EnvContainer) asBool() (bool, error) {
	b, err := strconv.ParseBool(c.value)
	if err != nil {
		return false, err
	}
	return b, nil
}

func (c EnvContainer) asInt() (int, error) {
	i, err := strconv.Atoi(c.value)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (c EnvContainer) asInt64() (int64, error) {
	i, err := strconv.ParseInt(c.value, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (c EnvContainer) asUint() (uint, error) {
	u, err := strconv.ParseUint(c.value, 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(u), nil
}

func (c EnvContainer) asUint64() (uint64, error) {
	u, err := strconv.ParseUint(c.value, 10, 64)
	if err != nil {
		return 0, err
	}
	return u, nil
}

func (c EnvContainer) asDuration() (time.Duration, error) {
	d, err := time.ParseDuration(c.value)
	if err != nil {
		return 0, err
	}
	return d, nil
}

func (c EnvContainer) asURL() (*url.URL, error) {
	u, err := url.Parse(c.value)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid URL %q: expected scheme and host", c.value)
	}
	return u, nil
}

func (c EnvContainer) asCSV() ([]string, error) {
	return c.asStringSlice(",")
}

func (c EnvContainer) asStringSlice(sep string) ([]string, error) {
	if sep == "" {
		return nil, errors.New("separator cannot be empty")
	}
	if strings.TrimSpace(c.value) == "" {
		return []string{}, nil
	}
	parts := strings.Split(c.value, sep)
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		v := strings.TrimSpace(part)
		if v == "" {
			continue
		}
		values = append(values, v)
	}
	return values, nil
}

func (c EnvContainer) asTime(layout string) (time.Time, error) {
	if layout == "" {
		return time.Time{}, errors.New("layout cannot be empty")
	}
	t, err := time.Parse(layout, c.value)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func (c EnvContainer) asBytes() (int64, error) {
	raw := strings.TrimSpace(c.value)
	if raw == "" {
		return 0, errors.New("byte size cannot be empty")
	}

	numPart := raw
	unitPart := ""
	for i := len(raw) - 1; i >= 0; i-- {
		ch := raw[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			continue
		}
		numPart = strings.TrimSpace(raw[:i+1])
		unitPart = strings.ToUpper(strings.TrimSpace(raw[i+1:]))
		break
	}
	if numPart == raw {
		unitPart = ""
	}

	n, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, errors.New("byte size cannot be negative")
	}

	mult, ok := map[string]float64{
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

	size := n * mult
	if size > math.MaxInt64 {
		return 0, fmt.Errorf("byte size %q overflows int64", c.value)
	}
	return int64(size), nil
}

func (c EnvContainer) asMap(sepKV, sepEntry string) (map[string]string, error) {
	if sepKV == "" || sepEntry == "" {
		return nil, errors.New("separators cannot be empty")
	}
	if strings.TrimSpace(c.value) == "" {
		return map[string]string{}, nil
	}

	entries := strings.Split(c.value, sepEntry)
	result := make(map[string]string, len(entries))
	for _, entry := range entries {
		pair := strings.SplitN(entry, sepKV, 2)
		if len(pair) != 2 {
			return nil, fmt.Errorf("invalid map entry %q", strings.TrimSpace(entry))
		}
		key := strings.TrimSpace(pair[0])
		val := strings.TrimSpace(pair[1])
		if key == "" {
			return nil, fmt.Errorf("invalid map entry %q: empty key", strings.TrimSpace(entry))
		}
		result[key] = val
	}
	return result, nil
}

func (c EnvContainer) asEnum(valid ...string) (string, error) {
	if len(valid) == 0 {
		return "", errors.New("enum options cannot be empty")
	}
	for _, candidate := range valid {
		if c.value == candidate {
			return c.value, nil
		}
	}
	return "", fmt.Errorf("value %q is not in enum %v", c.value, valid)
}
