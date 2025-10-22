package doterr_test

import (
	"errors"
	"os"
	"testing"

	. "github.com/xmlui-org/xmlui-test-server/xmluisvr/doterr"
)

var (
	ErrTest  = errors.New("test")
	ErrOther = errors.New("other")
)

func TestWith_NoBase_WithCause(t *testing.T) {
	cause := errors.New("cause")
	e := WithErr("k", 1, cause)
	if !errors.Is(e, cause) {
		t.Fatal("lost cause")
	}
}

func TestWith_BaseOnly_MiddleEmpty(t *testing.T) {
	base := NewErr(ErrTest, "a", 1)
	e := WithErr(base) // no middle, no cause
	// Should preserve metadata from base node
	kvs := ErrMeta(e)
	if len(kvs) != 1 || kvs[0].Key() != "a" || kvs[0].Value() != 1 {
		t.Fatal("expected single node with original metadata")
	}
}

func TestWith_BaseAndCause_Middle(t *testing.T) {
	cause := errors.New("cause")
	base := NewErr(ErrTest, "a", 1, cause)
	e := WithErr(base, "b", 2)
	// Should enrich rightmost node and keep original cause
	if !errors.Is(e, cause) {
		t.Fatal("lost cause")
	}
	kvs := ErrMeta(e)
	if len(kvs) < 2 {
		t.Fatal("expected enriched kvs")
	}
}

func TestWith_NoBase_NoCause_NodeOnly(t *testing.T) {
	e := WithErr("k", 1)
	// Should create a single node with metadata
	kvs := ErrMeta(e)
	if len(kvs) != 1 || kvs[0].Key() != "k" || kvs[0].Value() != 1 {
		t.Fatal("expected single node with metadata")
	}
}

type customError struct{ msg string }

// implement error
func (c *customError) Error() string { return c.msg }

func TestFind_FindsTypedError(t *testing.T) {

	cause := &customError{"boom"}
	err := errors.Join(NewErr(ErrTest, "k", 1), cause)

	got, ok := FindErr[*customError](err)
	if //goland:noinspection GoDirectComparisonOfErrors
	!ok || got != cause {
		t.Fatalf("expected to extract *custom; ok=%v got=%v", ok, got)
	}
}

func TestFind_NoneFound(t *testing.T) {
	_, ok := FindErr[*os.PathError](NewErr(ErrTest, "k", 1))
	if ok {
		t.Fatalf("did not expect to extract *os.PathError")
	}
}

// ============================================================================
// NewErr validation tests
// ============================================================================

func TestNewErr_Valid_OneSentinelOneKV(t *testing.T) {
	err := NewErr(ErrTest, "key", "value")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, ErrTest) {
		t.Error("expected error to contain ErrTest sentinel")
	}
}

func TestNewErr_Valid_TwoSentinelsMultipleKVs(t *testing.T) {
	err := NewErr(ErrTest, ErrOther, "key1", "value1", "key2", 42)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, ErrTest) {
		t.Error("expected error to contain ErrTest sentinel")
	}
	if !errors.Is(err, ErrOther) {
		t.Error("expected error to contain ErrOther sentinel")
	}
}

func TestNewErr_Valid_JustSentinels(t *testing.T) {
	err := NewErr(ErrTest)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, ErrTest) {
		t.Error("expected error to contain ErrTest sentinel")
	}
}

func TestNewErr_Invalid_NoSentinel(t *testing.T) {
	err := NewErr("key", "value")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrMissingSentinel) {
		t.Errorf("expected ErrMissingSentinel, got: %v", err)
	}
}

func TestNewErr_Invalid_TrailingKeyWithoutValue(t *testing.T) {
	err := NewErr(ErrTest, "key", "value", "orphan")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrTrailingKey) {
		t.Errorf("expected ErrTrailingKey, got: %v", err)
	}

	// Check structured metadata - ErrMeta() now unwraps automatically
	kvs := ErrMeta(err)
	var foundKey bool
	for _, kv := range kvs {
		if kv.Key() == "key" && kv.Value() == "orphan" {
			foundKey = true
			break
		}
	}
	if !foundKey {
		t.Error("expected metadata to contain key='orphan'")
	}
}

func TestNewErr_Valid_ErrorAfterKVPairs_IsTrailingCause(t *testing.T) {
	// This is now VALID - the error after KV pairs is treated as a trailing cause
	err := NewErr(ErrTest, "key", "value", ErrOther)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, ErrTest) {
		t.Error("expected error to contain ErrTest sentinel")
	}
	if !errors.Is(err, ErrOther) {
		t.Error("expected error to contain ErrOther as trailing cause")
	}
}

func TestNewErr_Invalid_NonStringNonErrorValue(t *testing.T) {
	err := NewErr(ErrTest, "key", "value", 123)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrInvalidArgumentType) {
		t.Errorf("expected ErrInvalidArgumentType, got: %v", err)
	}

	// Check structured metadata - ErrMeta() now unwraps automatically
	kvs := ErrMeta(err)
	var foundType bool
	for _, kv := range kvs {
		if kv.Key() == "type" && kv.Value() == "int" {
			foundType = true
			break
		}
	}
	if !foundType {
		t.Error("expected metadata to contain type='int'")
	}
}

func TestNewErr_Invalid_EmptyParts(t *testing.T) {
	err := NewErr()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrMissingSentinel) {
		t.Errorf("expected ErrMissingSentinel, got: %v", err)
	}
}

// Tests for trailing cause functionality

func TestNewErr_Valid_WithTrailingCause(t *testing.T) {
	cause := errors.New("underlying cause")
	err := NewErr(ErrTest, "key", "value", cause)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, ErrTest) {
		t.Error("expected error to contain ErrTest sentinel")
	}
	if !errors.Is(err, cause) {
		t.Error("expected error to contain cause")
	}
}

func TestNewErr_Valid_ErrorAsValue(t *testing.T) {
	causeValue := errors.New("this is a value")
	trailingCause := errors.New("this is the real cause")
	err := NewErr(ErrTest, "cause", causeValue, trailingCause)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, ErrTest) {
		t.Error("expected error to contain ErrTest sentinel")
	}
	if !errors.Is(err, trailingCause) {
		t.Error("expected error to contain trailing cause")
	}

	// The causeValue should be in metadata - ErrMeta() unwraps automatically
	kvs := ErrMeta(err)
	found := false
	for _, kv := range kvs {
		if kv.Key() == "cause" {
			if errVal, ok := kv.Value().(error); ok && errVal.Error() == causeValue.Error() {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected causeValue to be in metadata with key 'cause'")
	}
}

func TestNewErr_Valid_MultipleSentinelsWithTrailingCause(t *testing.T) {
	cause := errors.New("cause")
	err := NewErr(ErrTest, ErrOther, "key", "value", cause)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, ErrTest) {
		t.Error("expected error to contain ErrTest sentinel")
	}
	if !errors.Is(err, ErrOther) {
		t.Error("expected error to contain ErrOther sentinel")
	}
	if !errors.Is(err, cause) {
		t.Error("expected error to contain cause")
	}
}

func TestNewErr_Valid_OnlySentinels_NoTrailingCause(t *testing.T) {
	// When we only have errors, they're all sentinels, no trailing cause
	err := NewErr(ErrTest, ErrOther)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, ErrTest) {
		t.Error("expected error to contain ErrTest sentinel")
	}
	if !errors.Is(err, ErrOther) {
		t.Error("expected error to contain ErrOther sentinel")
	}
}

func TestNewErr_Valid_TrailingCauseWithNoMetadata(t *testing.T) {
	cause := errors.New("cause")
	err := NewErr(ErrTest, cause)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, ErrTest) {
		t.Error("expected error to contain ErrTest sentinel")
	}
	if !errors.Is(err, cause) {
		t.Error("expected error to contain cause")
	}
}

// Test that validation errors use structured metadata

func TestValidationErrors_UseStructuredMetadata(t *testing.T) {
	// Test trailing key error has structured metadata
	err := NewErr(ErrTest, "key", "value", "orphan")
	if err == nil {
		t.Fatal("expected validation error")
	}

	// ErrMeta() now unwraps automatically to find the first doterr node
	kvs := ErrMeta(err)
	if len(kvs) == 0 {
		t.Fatal("expected validation error to have structured metadata")
	}

	// Verify metadata keys
	var foundKey, foundPosition bool
	for _, kv := range kvs {
		if kv.Key() == "key" && kv.Value() == "orphan" {
			foundKey = true
		}
		if kv.Key() == "position" {
			foundPosition = true
		}
	}
	if !foundKey {
		t.Error("expected metadata to contain 'key' with value 'orphan'")
	}
	if !foundPosition {
		t.Error("expected metadata to contain 'position'")
	}
}

// Test that validation errors themselves use sentinels

func TestValidationErrors_UseSentinels(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		sentinel error
	}{
		{
			name:     "empty args",
			args:     []any{},
			sentinel: ErrMissingSentinel,
		},
		{
			name:     "no sentinel",
			args:     []any{"key", "value"},
			sentinel: ErrMissingSentinel,
		},
		{
			name:     "trailing key",
			args:     []any{ErrTest, "key", "value", "orphan"},
			sentinel: ErrTrailingKey,
		},
		{
			name:     "invalid type",
			args:     []any{ErrTest, "key", "value", 123},
			sentinel: ErrInvalidArgumentType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewErr(tt.args...)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !errors.Is(err, tt.sentinel) {
				t.Errorf("expected error to contain sentinel %v, got: %v", tt.sentinel, err)
			}
		})
	}
}

// Test cross-package error detection
// Note: This test simulates cross-package errors by creating a mock entry
// In reality, this would happen when errors from different doterr.go copies
// are mixed together.

func TestWithErr_DetectsCrossPackageErrors_BaseError(t *testing.T) {
	// Create a legitimate entry from this package
	localErr := NewErr(ErrTest, "local", "data")

	// We can't actually create an entry with a different id in this test
	// because we only have access to our own doterr package.
	// This test documents the expected behavior when cross-package errors occur.
	// The feature will be tested in integration scenarios where multiple
	// doterr packages are actually present.

	// When a cross-package error is detected, WithErr should prepend
	// ErrCrossPackageError with metadata:
	//   - package_id: the foreign package's uniqueId
	//   - expected_id: this package's uniqueId

	// For now, verify that local errors work without triggering the check
	enriched := WithErr(localErr, "extra", "metadata")
	if errors.Is(enriched, ErrCrossPackageError) {
		t.Error("local error should not trigger cross-package detection")
	}
}

func TestWithErr_DetectsCrossPackageErrors_CauseError(t *testing.T) {
	// Similar to above - documents expected behavior for cross-package causes
	// When WithErr receives a foreign entry as the trailing cause,
	// it should wrap it with ErrCrossPackageError

	localErr := NewErr(ErrTest, "local", "data")
	enriched := WithErr("extra", "metadata", localErr)
	if errors.Is(enriched, ErrCrossPackageError) {
		t.Error("local error should not trigger cross-package detection")
	}
}
