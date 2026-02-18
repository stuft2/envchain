package envchain

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestGetEnv(t *testing.T) {
	t.Run("returns default when unset", func(t *testing.T) {
		const key = "TEST_GETENVORDEFAULT_UNSET"
		_ = os.Unsetenv(key) // ensure it's unset

		got := GetEnv(key).WithDefault("fallback")
		if got.asString() != "fallback" {
			t.Fatalf("expected default %q, got %q", "fallback", got.asString())
		}
		if got.Ok() {
			t.Fatalf("expected lookup ok=false for unset key")
		}
	})

	t.Run("returns set value", func(t *testing.T) {
		const key = "TEST_GETENVORDEFAULT_SET"
		t.Setenv(key, "9090")

		got := GetEnv(key).WithDefault("8080")
		if got.asString() != "9090" {
			t.Fatalf("expected %q, got %q", "9090", got.asString())
		}
		if !got.Ok() {
			t.Fatalf("expected lookup ok=true for set key")
		}
	})

	t.Run("treats empty string as set", func(t *testing.T) {
		const key = "TEST_GETENVORDEFAULT_EMPTY"
		t.Setenv(key, "")

		got := GetEnv(key).WithDefault("fallback")
		if got.asString() != "" {
			t.Fatalf("expected empty string, got %q", got.asString())
		}
		if !got.Ok() {
			t.Fatalf("expected lookup ok=true for empty-but-set key")
		}
	})

	t.Run("supports unicode values", func(t *testing.T) {
		const key = "TEST_GETENVORDEFAULT_UNICODE"
		const val = "áβ🙂"
		t.Setenv(key, val)

		got := GetEnv(key).WithDefault("fallback")
		if got.asString() != val {
			t.Fatalf("expected %q, got %q", val, got.asString())
		}
		if !got.Ok() {
			t.Fatalf("expected lookup ok=true for set key")
		}
	})

	t.Run("empty default is allowed", func(t *testing.T) {
		const key = "TEST_GETENVORDEFAULT_EMPTY_DEFAULT"
		_ = os.Unsetenv(key)

		got := GetEnv(key).WithDefault("")
		if got.asString() != "" {
			t.Fatalf("expected empty default, got %q", got.asString())
		}
		if got.Ok() {
			t.Fatalf("expected lookup ok=false for unset key")
		}
	})
}

func TestEnvContainerRequired(t *testing.T) {
	t.Run("returns self when set", func(t *testing.T) {
		const key = "TEST_GETENV_REQUIRED_SET"
		t.Setenv(key, "value")

		got := GetEnv(key).Required()
		if !got.Ok() {
			t.Fatalf("expected Ok() true")
		}
		if got.asString() != "value" {
			t.Fatalf("expected %q, got %q", "value", got.asString())
		}
	})

	t.Run("panics when unset", func(t *testing.T) {
		const key = "TEST_GETENV_REQUIRED_UNSET"
		_ = os.Unsetenv(key)

		defer func() {
			r := recover()
			if r == nil {
				t.Fatalf("expected panic for unset required env")
			}
			want := `required environment variable "TEST_GETENV_REQUIRED_UNSET" is not set`
			if r != want {
				t.Fatalf("expected panic %q, got %v", want, r)
			}
		}()
		_ = GetEnv(key).Required()
	})
}

// Example test shows typical usage.
// Keep environment hygiene so it doesn't leak to other tests.
func ExampleGetEnv() {
	const key = "EXAMPLE_PORT"

	// Save original and restore after.
	orig, had := os.LookupEnv(key)
	if had {
		defer os.Setenv(key, orig)
	} else {
		defer os.Unsetenv(key)
	}

	_ = os.Setenv(key, "9090")
	fmt.Println(GetEnv(key).WithDefault("8080").asString())

	_ = os.Unsetenv(key)
	fmt.Println(GetEnv(key).WithDefault("8080").asString())

	// Output:
	// 9090
	// 8080
}

func TestEnvContainerAsNumber(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    float64
		wantErr bool
	}{
		{name: "integer", value: "123", want: 123, wantErr: false},
		{name: "float", value: "3.14", want: 3.14, wantErr: false},
		{name: "scientific notation", value: "1e3", want: 1000, wantErr: false},
		{name: "invalid value returns error", value: "abc", want: 0, wantErr: true},
		{name: "empty value returns error", value: "", want: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asNumber()
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestEnvContainerAsBool(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    bool
		wantErr bool
	}{
		{name: "true", value: "true", want: true, wantErr: false},
		{name: "one", value: "1", want: true, wantErr: false},
		{name: "false", value: "false", want: false, wantErr: false},
		{name: "zero", value: "0", want: false, wantErr: false},
		{name: "invalid value returns error", value: "not-bool", want: false, wantErr: true},
		{name: "empty value returns error", value: "", want: false, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asBool()
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestEnvContainerAsInt(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{name: "valid int", value: "42", want: 42},
		{name: "invalid int", value: "4.2", wantErr: true},
		{name: "empty", value: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asInt()
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestEnvContainerAsInt64(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    int64
		wantErr bool
	}{
		{name: "valid int64", value: "922337203685477580", want: 922337203685477580},
		{name: "invalid int64", value: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asInt64()
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestEnvContainerAsUint(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    uint
		wantErr bool
	}{
		{name: "valid uint", value: "42", want: 42},
		{name: "negative uint", value: "-1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asUint()
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestEnvContainerAsUint64(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    uint64
		wantErr bool
	}{
		{name: "valid uint64", value: "1844674407370955161", want: 1844674407370955161},
		{name: "invalid uint64", value: "1.1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asUint64()
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestEnvContainerAsDuration(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    time.Duration
		wantErr bool
	}{
		{name: "valid duration", value: "1h30m", want: time.Hour + 30*time.Minute},
		{name: "invalid duration", value: "90minutes", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asDuration()
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestEnvContainerAsURL(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    *url.URL
		wantErr bool
	}{
		{
			name:  "valid url",
			value: "https://example.com/path",
			want: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/path",
			},
		},
		{name: "invalid url", value: "://bad-url", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asURL()
			if tt.want != nil {
				if got == nil || got.Scheme != tt.want.Scheme || got.Host != tt.want.Host || got.Path != tt.want.Path {
					t.Fatalf("expected %v, got %v", tt.want, got)
				}
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestEnvContainerAsCSV(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{name: "csv values", value: "a, b ,c", want: []string{"a", "b", "c"}},
		{name: "empty string", value: "", want: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asCSV()
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestEnvContainerAsStringSlice(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		sep     string
		want    []string
		wantErr bool
	}{
		{name: "pipe separated", value: "a| b|c ", sep: "|", want: []string{"a", "b", "c"}},
		{name: "empty separator", value: "a,b", sep: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asStringSlice(tt.sep)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestEnvContainerAsTime(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		layout  string
		want    time.Time
		wantErr bool
	}{
		{
			name:   "rfc3339",
			value:  "2026-02-18T12:00:00Z",
			layout: time.RFC3339,
			want:   time.Date(2026, 2, 18, 12, 0, 0, 0, time.UTC),
		},
		{name: "bad time", value: "2026/02/18", layout: time.RFC3339, wantErr: true},
		{name: "empty layout", value: "2026-02-18T12:00:00Z", layout: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asTime(tt.layout)
			if !tt.want.IsZero() && !got.Equal(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestEnvContainerAsBytes(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    int64
		wantErr bool
	}{
		{name: "plain bytes", value: "1024", want: 1024},
		{name: "decimal unit", value: "1.5MB", want: 1572864},
		{name: "binary unit", value: "2GiB", want: 2147483648},
		{name: "bad unit", value: "10XB", wantErr: true},
		{name: "bad value", value: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asBytes()
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestEnvContainerAsMap(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		sepKV    string
		sepEntry string
		want     map[string]string
		wantErr  bool
	}{
		{
			name:     "valid map",
			value:    "a=1,b=2",
			sepKV:    "=",
			sepEntry: ",",
			want:     map[string]string{"a": "1", "b": "2"},
		},
		{name: "missing separator", value: "a=1", sepKV: "", sepEntry: ",", wantErr: true},
		{name: "malformed pair", value: "a=1,b", sepKV: "=", sepEntry: ",", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asMap(tt.sepKV, tt.sepEntry)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestEnvContainerAsEnum(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		valid   []string
		want    string
		wantErr bool
	}{
		{name: "valid enum", value: "prod", valid: []string{"dev", "prod"}, want: "prod"},
		{name: "invalid enum", value: "qa", valid: []string{"dev", "prod"}, wantErr: true},
		{name: "no options", value: "prod", valid: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asEnum(tt.valid...)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}
