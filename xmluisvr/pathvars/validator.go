package pathvars

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// validateDataType validates a value against a data type
func validateDataType(value string, dataType PVDataType) (err error) {
	switch dataType {
	case StringType:
		// Always valid
	case IntegerType:
		_, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			err = errors.Join(
				err,
				fmt.Errorf("value=%q", value),
				fmt.Errorf("data_type=%v", dataType),
				fmt.Errorf("reason=%s", "invalid integer format"),
			)
		}
	case DecimalType:
		_, err = strconv.ParseFloat(value, 64)
		if err != nil {
			err = errors.Join(
				err,
				fmt.Errorf("value=%q", value),
				fmt.Errorf("data_type=%v", dataType),
				fmt.Errorf("reason=%s", "invalid decimal format"),
			)
		}
	case RealType:
		_, err = strconv.ParseFloat(value, 64)
		if err != nil {
			err = errors.Join(
				err,
				fmt.Errorf("value=%q", value),
				fmt.Errorf("data_type=%v", dataType),
				fmt.Errorf("reason=%s", "invalid real number format"),
			)
		}
	case IdentifierType:
		err = validateIdentifier(value)
	case DateType:
		// Basic date validation, constraints handle format
		err = validateDate(value)
	case UUIDType:
		err = validateUUID(value)
	case AlphanumericType:
		err = validateAlphanum(value)
	case SlugType:
		err = validateSlug(value)
	case BooleanType:
		err = validateBoolean(value)
	case EmailType:
		err = validateEmail(value)
	default:
		err = errors.Join(
			ErrInvalidType,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("data_type=%v", dataType),
			fmt.Errorf("reason=%s", "unsupported data type"),
		)
	}

	return err
}

// validateIdentifier checks lowercase, leading alpha, then alphanumeric or underscore
func validateIdentifier(value string) (err error) {
	var matched bool
	var regex *regexp.Regexp

	regex = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	matched = regex.MatchString(value)
	if !matched {
		err = errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("data_type=%v", IdentifierType),
			fmt.Errorf("regex=%s", "^[a-z][a-z0-9_]*$"),
			fmt.Errorf("reason=%s", "must start with lowercase letter, followed by lowercase letters, digits, or underscores"),
		)
	}

	return err
}

// validateDate validates date values (basic validation, constraints handle format)
func validateDate(value string) (err error) {
	// Very basic date validation - just check if it's not empty for now
	// More specific validation handled by constraints
	if value == "" {
		err = errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("data_type=%v", DateType),
			fmt.Errorf("reason=%s", "date value cannot be empty"),
		)
	}
	return err
}

// validateUUID validates UUID format
func validateUUID(value string) (err error) {
	var matched bool
	var regex *regexp.Regexp

	// Basic UUID pattern: 8-4-4-4-12 hex digits
	regex = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	matched = regex.MatchString(value)
	if !matched {
		err = errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("data_type=%v", UUIDType),
			fmt.Errorf("pattern=%s", "8-4-4-4-12 hex digits"),
			fmt.Errorf("reason=%s", "invalid UUID format (expected 8-4-4-4-12 hex digits)"),
		)
	}

	return err
}

// validateAlphanum validates alphanumeric values
func validateAlphanum(value string) (err error) {
	var matched bool
	var regex *regexp.Regexp

	regex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	matched = regex.MatchString(value)
	if !matched {
		err = errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("data_type=%v", AlphanumericType),
			fmt.Errorf("regex=%s", "^[a-zA-Z0-9]+$"),
			fmt.Errorf("reason=%s", "value must contain only letters and digits"),
		)
	}

	return err
}

// validateSlug validates URL-safe slug format
func validateSlug(value string) (err error) {
	var matched bool
	var regex *regexp.Regexp

	// Lowercase letters, numbers, hyphens, no leading/trailing hyphens
	regex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	matched = regex.MatchString(value)
	if !matched {
		err = errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("data_type=%v", SlugType),
			fmt.Errorf("regex=%s", "^[a-z0-9]+(?:-[a-z0-9]+)*$"),
			fmt.Errorf("reason=%s", "must be lowercase letters/digits with optional hyphens between segments"),
		)
	}

	return err
}

// validateBoolean validates boolean values
func validateBoolean(value string) (err error) {
	if value != "true" && value != "false" {
		err = errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("data_type=%v", BooleanType),
			fmt.Errorf("allowedValues=%s", "true,false"),
			fmt.Errorf("reason=%s", "boolean value must be exactly 'true' or 'false'"),
		)
	}
	return err
}

// validateEmail validates email format
func validateEmail(value string) (err error) {
	var matched bool
	var regex *regexp.Regexp

	// Basic email regex: local@domain with basic validation
	regex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	matched = regex.MatchString(value)
	if !matched {
		err = errors.Join(
			ErrValidationFailed,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("data_type=%v", EmailType),
			fmt.Errorf("pattern=%s", "local@domain"),
			fmt.Errorf("reason=%s", "invalid email format"),
		)
	}

	return err
}
