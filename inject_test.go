package envault

import (
	"context"
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

type stubContextProvider struct {
	err    error
	called *int
	ctx    *context.Context
}

var _ internal.Provider = stubContextProvider{}
var _ internal.ContextProvider = stubContextProvider{}

func (s stubContextProvider) Inject() error {
	if s.called != nil {
		*s.called++
	}
	return nil
}

func (s stubContextProvider) InjectContext(ctx context.Context) error {
	if s.called != nil {
		*s.called++
	}
	if s.ctx != nil {
		*s.ctx = ctx
	}
	if s.err != nil {
		return s.err
	}
	return ctx.Err()
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

func TestInjectWithContextUsesContextProvider(t *testing.T) {
	var contextCalls, legacyCalls int
	var got context.Context
	type ctxKey string
	ctx := context.WithValue(context.Background(), ctxKey("k"), "v")

	err := InjectWithContext(
		ctx,
		stubContextProvider{called: &contextCalls, ctx: &got, err: nil},
		stubProvider{called: &legacyCalls},
	)
	if err != nil {
		t.Fatalf("InjectWithContext: %v", err)
	}
	if contextCalls != 1 || legacyCalls != 1 {
		t.Fatalf("providers called counts: context=%d legacy=%d", contextCalls, legacyCalls)
	}
	if got != ctx {
		t.Fatal("expected context provider to receive the shared context")
	}
}

func TestInjectWithContextCancelledContext(t *testing.T) {
	var contextCalls int
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := InjectWithContext(ctx, stubContextProvider{called: &contextCalls})
	if err == nil {
		t.Fatal("expected canceled context error, got nil")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("expected context canceled, got %q", err.Error())
	}
	if contextCalls != 1 {
		t.Fatalf("expected context provider to be called once, got %d", contextCalls)
	}
}

func TestInjectWithContextJoinsErrors(t *testing.T) {
	err := InjectWithContext(
		context.Background(),
		stubContextProvider{err: errors.New("ctx provider")},
		stubProvider{err: errors.New("legacy provider")},
	)
	if err == nil {
		t.Fatal("expected joined error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "ctx provider") || !strings.Contains(msg, "legacy provider") {
		t.Fatalf("expected both errors in message, got %q", msg)
	}
}
