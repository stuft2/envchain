package config

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
		_ = os.Unsetenv(key)

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

func ExampleGetEnv() {
	const key = "EXAMPLE_PORT"

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
		{name: "negative", value: "-1", wantErr: true},
		{name: "float", value: "4.2", wantErr: true},
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
		{name: "invalid uint64", value: "abc", wantErr: true},
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
		{name: "seconds", value: "5s", want: 5 * time.Second},
		{name: "minutes", value: "2m", want: 2 * time.Minute},
		{name: "invalid", value: "later", wantErr: true},
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
		{name: "valid", value: "https://example.com", want: &url.URL{Scheme: "https", Host: "example.com"}},
		{name: "missing host", value: "https://", wantErr: true},
		{name: "missing scheme", value: "example.com", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asURL()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Scheme != tt.want.Scheme || got.Host != tt.want.Host {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestEnvContainerAsCSV(t *testing.T) {
	got, err := EnvContainer{value: "a,b, c ,,d"}.asCSV()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"a", "b", "c", "d"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
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
		{name: "valid", value: "a|b| c ", sep: "|", want: []string{"a", "b", "c"}},
		{name: "empty value", value: " ", sep: "|", want: []string{}},
		{name: "empty separator", value: "a,b", sep: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvContainer{value: tt.value}.asStringSlice(tt.sep)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestEnvContainerAsTime(t *testing.T) {
	want := time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC)
	got, err := EnvContainer{value: "2026-03-10"}.asTime("2006-01-02")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Equal(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestEnvContainerAsBytes(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    int64
		wantErr bool
	}{
		{name: "plain bytes", value: "42", want: 42},
		{name: "kb", value: "2KB", want: 2048},
		{name: "mib", value: "1MiB", want: 1024 * 1024},
		{name: "invalid unit", value: "1XB", wantErr: true},
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
	got, err := EnvContainer{value: "a=1,b=2"}.asMap("=", ",")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]string{"a": "1", "b": "2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestEnvContainerAsEnum(t *testing.T) {
	got, err := EnvContainer{value: "dev"}.asEnum("dev", "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "dev" {
		t.Fatalf("expected %q, got %q", "dev", got)
	}
}
