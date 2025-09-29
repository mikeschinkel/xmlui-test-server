package pathvars

import (
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Note: UUIDFormatConstraint is not registered directly.
// It's handled by DateFormatConstraint.ParseBytes() based on dataType.

var _ Constraint = (*UUIDFormatConstraint)(nil)

// UUIDFormatConstraint validates UUID formats
type UUIDFormatConstraint struct {
	baseConstraint
	format    string
	validator func(string) error
}

func NewUUIDFormatConstraint(format string, validator func(string) error) *UUIDFormatConstraint {
	c := &UUIDFormatConstraint{
		format:    format,
		validator: validator,
	}
	c.baseConstraint = newBaseConstraint(c)
	return c
}

func (c *UUIDFormatConstraint) ValidDateTypes() []PVDataType {
	return []PVDataType{UUIDType}
}

func (c *UUIDFormatConstraint) Type() ConstraintType {
	return FormatConstraintType
}

func (c *UUIDFormatConstraint) Parse(value string, dataType PVDataType) (Constraint, error) {
	return ParseUUIDFormatConstraint(value)
}

func (c *UUIDFormatConstraint) Validate(value string) error {
	return c.validator(value)
}

func (c *UUIDFormatConstraint) String() string {
	return c.format
}

// ParseUUIDFormatConstraint parses UUID format specifications
func ParseUUIDFormatConstraint(spec string) (constraint *UUIDFormatConstraint, err error) {
	var validator func(string) error

	switch strings.ToLower(spec) {
	case "v1":
		validator = validateUUIDv1
	case "v2":
		validator = validateUUIDv2
	case "v3":
		validator = validateUUIDv3
	case "v4":
		validator = validateUUIDv4
	case "v5":
		validator = validateUUIDv5
	case "v6":
		validator = validateUUIDv6
	case "v7":
		validator = validateUUIDv7
	case "v8":
		validator = validateUUIDv8
	case "v1-5", "v1to5":
		validator = validateUUIDv1to5
	case "v6-8", "v6to8":
		validator = validateUUIDv6to8
	case "any", "generic":
		validator = validateUUIDGeneric
	case "ulid":
		validator = validateULID
	case "ksuid":
		validator = validateKSUID
	case "nanoid":
		validator = validateNanoID
	default:
		err = errors.Join(
			ErrInvalidConstraint,
			fmt.Errorf("spec=%q", spec),
			fmt.Errorf("reason=%s", "unsupported UUID format"),
		)
		goto end
	}

	constraint = NewUUIDFormatConstraint(spec, validator)

end:
	return constraint, err
}

// validateUUIDGeneric validates any standard UUID format (versions 1-8)
func validateUUIDGeneric(value string) error {
	version, err := parseStandardUUID(value)
	if err != nil {
		return err
	}
	if version < 1 || version > 8 {
		return errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("version=%d", version),
			fmt.Errorf("reason=%s", "UUID version must be 1-8"),
		)
	}
	return nil
}

// validateUUIDv1to5 validates UUID versions 1-5
func validateUUIDv1to5(value string) error {
	version, err := parseStandardUUID(value)
	if err != nil {
		return err
	}
	if version < 1 || version > 5 {
		return errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("version=%d", version),
			fmt.Errorf("reason=%s", "UUID version must be 1-5"),
		)
	}
	return nil
}

// validateUUIDv6to8 validates UUID versions 6-8
func validateUUIDv6to8(value string) error {
	version, err := parseStandardUUID(value)
	if err != nil {
		return err
	}
	if version < 6 || version > 8 {
		return errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("version=%d", version),
			fmt.Errorf("reason=%s", "UUID version must be 6-8"),
		)
	}
	return nil
}

// validateUUIDv1 validates UUID version 1 (time + MAC)
func validateUUIDv1(value string) error {
	return validateSpecificUUIDVersion(value, 1)
}

// validateUUIDv2 validates UUID version 2 (time + POSIX UID/GID)
func validateUUIDv2(value string) error {
	return validateSpecificUUIDVersion(value, 2)
}

// validateUUIDv3 validates UUID version 3 (name-based, MD5)
func validateUUIDv3(value string) error {
	return validateSpecificUUIDVersion(value, 3)
}

// validateUUIDv4 validates UUID version 4 (random)
func validateUUIDv4(value string) error {
	return validateSpecificUUIDVersion(value, 4)
}

// validateUUIDv5 validates UUID version 5 (name-based, SHA-1)
func validateUUIDv5(value string) error {
	return validateSpecificUUIDVersion(value, 5)
}

// validateUUIDv6 validates UUID version 6 (reordered v1)
func validateUUIDv6(value string) error {
	return validateSpecificUUIDVersion(value, 6)
}

// validateUUIDv7 validates UUID version 7 (Unix timestamp + random)
func validateUUIDv7(value string) error {
	return validateSpecificUUIDVersion(value, 7)
}

// validateUUIDv8 validates UUID version 8 (custom/experimental)
func validateUUIDv8(value string) error {
	return validateSpecificUUIDVersion(value, 8)
}

// validateSpecificUUIDVersion validates a specific UUID version
func validateSpecificUUIDVersion(value string, expectedVersion int) error {
	version, err := parseStandardUUID(value)
	if err != nil {
		return err
	}
	if version != expectedVersion {
		return errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("expectedVersion=%d", expectedVersion),
			fmt.Errorf("actualVersion=%d", version),
			fmt.Errorf("reason=%s", "UUID version mismatch"),
		)
	}
	return nil
}

// parseStandardUUID parses and validates a standard UUID, returning the version
func parseStandardUUID(value string) (version int, err error) {
	var b [16]byte
	var hexStr string
	var variant int

	// Check basic shape: 36 chars with hyphens at correct positions
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		err = errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("reason=%s", "invalid UUID shape (expected 8-4-4-4-12 format)"),
		)
		goto end
	}

	// Remove hyphens and decode hex
	hexStr = strings.ReplaceAll(value, "-", "")
	_, err = hex.Decode(b[:], []byte(hexStr))
	if err != nil {
		err = errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("reason=%s", "invalid hex encoding in UUID"),
		)
		goto end
	}

	// Check variant bits (must be RFC 4122/9562: bits 10xx)
	variant = int((b[8] & 0xC0) >> 6)
	if variant != 0b10 {
		err = errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("variant=%08b", variant),
			fmt.Errorf("reason=%s", "invalid UUID variant (must be RFC 4122/9562)"),
		)
		goto end
	}

	// Extract version from upper 4 bits of byte 6
	version = int(b[6] >> 4)
	if version < 1 || version > 8 {
		err = errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("version=%d", version),
			fmt.Errorf("reason=%s", "invalid UUID version (must be 1-8)"),
		)
		goto end
	}

end:
	return version, err
}

// validateULID validates ULID format (26 chars, Crockford Base32)
func validateULID(value string) error {
	// ULID: 26 characters, Crockford Base32 alphabet
	ulidRegex := regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{26}$`)
	if !ulidRegex.MatchString(value) {
		return errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("pattern=%s", "26 chars of Crockford Base32"),
			fmt.Errorf("reason=%s", "invalid ULID format"),
		)
	}
	return nil
}

// validateKSUID validates KSUID format (27 chars, Base62)
func validateKSUID(value string) error {
	// KSUID: 27 characters, Base62 alphabet
	ksuidRegex := regexp.MustCompile(`^[0-9A-Za-z]{27}$`)
	if !ksuidRegex.MatchString(value) {
		return errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("pattern=%s", "27 chars of Base62"),
			fmt.Errorf("reason=%s", "invalid KSUID format"),
		)
	}
	return nil
}

// validateNanoID validates NanoID format (21 chars by default, URL-safe)
func validateNanoID(value string) error {
	// NanoID: 21 characters (default), URL-safe alphabet
	nanoidRegex := regexp.MustCompile(`^[A-Za-z0-9_-]{21}$`)
	if !nanoidRegex.MatchString(value) {
		return errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("pattern=%s", "21 chars of URL-safe alphabet"),
			fmt.Errorf("reason=%s", "invalid NanoID format"),
		)
	}
	return nil
}
