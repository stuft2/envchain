package envault

import (
	"fmt"
	"os"
	"testing"
)

func TestGetEnvOrDefault(t *testing.T) {
	t.Run("returns default when unset", func(t *testing.T) {
		const key = "TEST_GETENVORDEFAULT_UNSET"
		_ = os.Unsetenv(key) // ensure it's unset

		got := GetEnvOrDefault(key, "fallback")
		if got != "fallback" {
			t.Fatalf("expected default %q, got %q", "fallback", got)
		}
	})

	t.Run("returns set value", func(t *testing.T) {
		const key = "TEST_GETENVORDEFAULT_SET"
		t.Setenv(key, "9090")

		got := GetEnvOrDefault(key, "8080")
		if got != "9090" {
			t.Fatalf("expected %q, got %q", "9090", got)
		}
	})

	t.Run("treats empty string as set", func(t *testing.T) {
		const key = "TEST_GETENVORDEFAULT_EMPTY"
		t.Setenv(key, "")

		got := GetEnvOrDefault(key, "fallback")
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})

	t.Run("supports unicode values", func(t *testing.T) {
		const key = "TEST_GETENVORDEFAULT_UNICODE"
		const val = "áβ🙂"
		t.Setenv(key, val)

		got := GetEnvOrDefault(key, "fallback")
		if got != val {
			t.Fatalf("expected %q, got %q", val, got)
		}
	})

	t.Run("empty default is allowed", func(t *testing.T) {
		const key = "TEST_GETENVORDEFAULT_EMPTY_DEFAULT"
		_ = os.Unsetenv(key)

		got := GetEnvOrDefault(key, "")
		if got != "" {
			t.Fatalf("expected empty default, got %q", got)
		}
	})
}

// Example test shows typical usage.
// Keep environment hygiene so it doesn't leak to other tests.
func ExampleGetEnvOrDefault() {
	const key = "EXAMPLE_PORT"

	// Save original and restore after.
	orig, had := os.LookupEnv(key)
	if had {
		defer os.Setenv(key, orig)
	} else {
		defer os.Unsetenv(key)
	}

	_ = os.Setenv(key, "9090")
	fmt.Println(GetEnvOrDefault(key, "8080"))

	_ = os.Unsetenv(key)
	fmt.Println(GetEnvOrDefault(key, "8080"))

	// Output:
	// 9090
	// 8080
}
