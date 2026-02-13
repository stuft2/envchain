package envault

import (
	"errors"
	"strings"
	"testing"

	"github.com/stuft2/envault/internal"
)

type stubProvider struct {
	err    error
	called *int
}

// Ensure the provider is implemented
var _ internal.Provider = stubProvider{}

func (s stubProvider) Inject() error {
	if s.called != nil {
		*s.called++
	}
	return s.err
}

func TestInjectReturnsNilWhenProvidersSucceed(t *testing.T) {
	var count int
	if err := Inject(stubProvider{called: &count}); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected provider to be called once, got %d", count)
	}
}

func TestInjectContinuesAfterErrors(t *testing.T) {
	var first, second int
	err := Inject(
		stubProvider{called: &first, err: errors.New("first")},
		stubProvider{called: &second},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if first != 1 || second != 1 {
		t.Fatalf("providers called counts: first=%d second=%d", first, second)
	}
}

func TestInjectJoinsMultipleErrors(t *testing.T) {
	err := Inject(
		stubProvider{err: errors.New("first")},
		stubProvider{err: errors.New("second")},
	)
	if err == nil {
		t.Fatal("expected joined error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "first") || !strings.Contains(msg, "second") {
		t.Fatalf("expected both errors in message, got %q", msg)
	}
}
