package fault_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/quixiq/polyglot/pkg/fault"
)

func TestNewCarriesKind(t *testing.T) {
	err := fault.New(fault.KindNotFound, "billing: invoice not found")
	if got := fault.KindOf(err); got != fault.KindNotFound {
		t.Fatalf("KindOf = %v, want %v", got, fault.KindNotFound)
	}
	if err.Error() != "billing: invoice not found" {
		t.Fatalf("Error() = %q", err.Error())
	}
}

func TestKindOfTraversesWrapChain(t *testing.T) {
	sentinel := fault.New(fault.KindPermissionDenied, "device: unauthorized access")
	wrapped := fmt.Errorf("get device %s: %w", "d-1", sentinel)
	wrappedTwice := fmt.Errorf("handle request: %w", wrapped)
	if got := fault.KindOf(wrappedTwice); got != fault.KindPermissionDenied {
		t.Fatalf("KindOf = %v, want %v", got, fault.KindPermissionDenied)
	}
	if !errors.Is(wrappedTwice, sentinel) {
		t.Fatal("errors.Is must still match the sentinel through the chain")
	}
}

func TestWrapAttachesKind(t *testing.T) {
	cause := errors.New("connection refused")
	err := fault.Wrap(fault.KindUnavailable, cause)
	if got := fault.KindOf(err); got != fault.KindUnavailable {
		t.Fatalf("KindOf = %v, want %v", got, fault.KindUnavailable)
	}
	if !errors.Is(err, cause) {
		t.Fatal("wrapped error must unwrap to its cause")
	}
	if err.Error() != "connection refused" {
		t.Fatalf("Error() = %q, want cause message", err.Error())
	}
	if fault.Wrap(fault.KindUnknown, nil) != nil {
		t.Fatal("Wrap(nil) must return nil")
	}
}

func TestKindOfUnknownForPlainError(t *testing.T) {
	if got := fault.KindOf(errors.New("boom")); got != fault.KindUnknown {
		t.Fatalf("KindOf = %v, want %v", got, fault.KindUnknown)
	}
	if got := fault.KindOf(nil); got != fault.KindUnknown {
		t.Fatalf("KindOf(nil) = %v, want %v", got, fault.KindUnknown)
	}
}

func TestKindStringStable(t *testing.T) {
	kinds := map[fault.Kind]string{
		fault.KindUnknown:            "unknown",
		fault.KindNotFound:           "not_found",
		fault.KindInvalidInput:       "invalid_input",
		fault.KindAlreadyExists:      "already_exists",
		fault.KindPermissionDenied:   "permission_denied",
		fault.KindUnauthenticated:    "unauthenticated",
		fault.KindFailedPrecondition: "failed_precondition",
		fault.KindConflict:           "conflict",
		fault.KindUnavailable:        "unavailable",
		fault.KindResourceExhausted:  "resource_exhausted",
	}
	for k, want := range kinds {
		if k.String() != want {
			t.Errorf("Kind(%d).String() = %q, want %q", int(k), k.String(), want)
		}
	}
	if got := fault.Kind(9999).String(); got != "unknown" {
		t.Errorf("Kind(9999).String() = %q, want unknown", got)
	}
}

func TestErrorsAsExtraction(t *testing.T) {
	sentinel := fault.New(fault.KindConflict, "billing: conflict detected")
	wrapped := fmt.Errorf("level 1: %w", sentinel)
	wrapped2 := fmt.Errorf("level 2: %w", wrapped)

	var ferr *fault.Error
	if !errors.As(wrapped2, &ferr) {
		t.Fatal("errors.As failed to extract *fault.Error from wrap chain")
	}
	if ferr.Kind() != fault.KindConflict {
		t.Errorf("extracted Kind = %v, want %v", ferr.Kind(), fault.KindConflict)
	}
	if ferr.Error() != "billing: conflict detected" {
		t.Errorf("extracted Error = %q, want 'billing: conflict detected'", ferr.Error())
	}
	if ferr.Unwrap() != nil {
		t.Errorf("sentinel root Unwrap() should be nil, got: %v", ferr.Unwrap())
	}
}

func TestErrorsJoinWithFault(t *testing.T) {
	err1 := errors.New("something background failed")
	err2 := fault.New(fault.KindNotFound, "customer: customer not found")
	joined := errors.Join(err1, err2)

	if got := fault.KindOf(joined); got != fault.KindNotFound {
		t.Errorf("KindOf(joined) = %v, want %v", got, fault.KindNotFound)
	}
	if !errors.Is(joined, err2) {
		t.Fatal("errors.Is must find err2 in joined error")
	}
}
