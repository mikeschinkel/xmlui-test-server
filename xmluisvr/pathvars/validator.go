// Package pathvars/validator provides data type validation functions for
// parameter values. It validates values against specific data types like
// integers, UUIDs, email addresses, and other supported parameter types.
package pathvars

//import (
//	"fmt"
//	"regexp"
//	"strconv"
//
//	"github.com/xmlui-org/localsvr/xmluisvr/pathvars/dtclassifiers"
//)

//// validateDataType validates a value against a specified data type.
//// Returns an error if the value doesn't conform to the expected type format.
//func (p NameSpecProps) validateDataType(value string, dataType PVDataType) (err error) {
//	switch dataType {
//	case StringType:
//		// Always valid
//	case IntegerType:
//		_, err = strconv.ParseInt(value, 10, 64)
//		if err != nil {
//			err = NewErr(ErrInvalidIntegerFormat, "value", value, err)
//		}
//	case DecimalType:
//		_, err = strconv.ParseFloat(value, 64)
//		if err != nil {
//			err = NewErr(ErrInvalidDecimalFormat, "value", value, err)
//		}
//	case RealType:
//		_, err = strconv.ParseFloat(value, 64)
//		if err != nil {
//			err = NewErr(ErrInvalidRealFormat, "value", value, err)
//		}
//	case IdentifierType:
//		err = p.validateIdentifier(value)
//	case DateType:
//		// Validates date format (YYYY-MM-DD) by default; format constraints can enforce other formats
//		err = p.validateDate(value)
//	case UUIDType:
//		err = p.validateUUID(value)
//	case AlphanumericType:
//		err = p.validateAlphanum(value)
//	case SlugType:
//		err = p.validateSlug(value)
//	case BooleanType:
//		err = p.validateBoolean(value)
//	case EmailType:
//		err = p.validateEmail(value)
//	default:
//		err = ErrUnsupportedDataType
//	}
//	if err != nil {
//		err = WithErr(err,
//			ErrInvalidDataType,
//			"data_type", dataType.Slug(),
//			"value", value,
//		)
//	}
//	return err
//}
//
//// validateIdentifier checks that the value conforms to identifier format:
//// lowercase letters, leading alpha, then alphanumeric or underscore characters.
//func (p NameSpecProps) validateIdentifier(value string) (err error) {
//	var matched bool
//	var regex *regexp.Regexp
//
//	regex = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
//	matched = regex.MatchString(value)
//	if !matched {
//		err = NewErr(
//			ErrParameterValidationFailed,
//			ErrInvalidIdentifierFormat,
//			"regex", "^[a-z][a-z0-9_]*$",
//		)
//	}
//
//	return err
//}
//
//// validateDate validates date values requiring date format (YYYY-MM-DD) by
//// default. If multisegment then YYYY/MM/DD or partial dates (YYYY or YYYY/MM).
//// More specific date format validation can be enforced via format constraints.
//func (p NameSpecProps) validateDate(value string) (err error) {
//	var matched bool
//	var year, month, day int
//	var scanFormat = "%d-%d-%d"
//	var errMsgFormat = "yyyy-mm-dd"
//
//	if p.MultiSegment {
//		// Use regex with capture groups to match yyyy, yyyy/mm, or yyyy/mm/dd
//		matches := dtclassifiers.multisegmentDateRegex.FindStringSubmatch(value)
//		if matches == nil {
//			err = NewErr(
//				ErrFailedToMatchDateFormat,
//				"expected_format", "yyyy, yyyy/mm, or yyyy/mm/dd",
//			)
//			goto end
//		}
//
//		// matches[0] is full match, matches[1] is year, matches[2] is month, matches[3] is day
//		// Parse year (always present)
//		_, err = fmt.Sscanf(matches[1], "%d", &year)
//		if err != nil {
//			err = NewErr(
//				ErrFailedToParseDateFormat,
//				"expected_format", "yyyy",
//				err,
//			)
//			goto end
//		}
//
//		// Validate year
//		if year < 1000 || year > 9999 {
//			err = NewErr(
//				ErrInvalidYearInDate,
//				"invalid_year", year,
//			)
//			goto end
//		}
//
//		// Parse month if present
//		if matches[2] != "" {
//			_, err = fmt.Sscanf(matches[2], "%d", &month)
//			if err != nil {
//				err = NewErr(
//					ErrFailedToParseDateFormat,
//					"expected_format", "mm",
//					err,
//				)
//				goto end
//			}
//
//			// Validate month range
//			if month < 1 || month > 12 {
//				err = NewErr(
//					ErrInvalidMonthInDate,
//					"invalid_month", month,
//				)
//				goto end
//			}
//		}
//
//		// Parse day if present
//		if matches[3] != "" {
//			_, err = fmt.Sscanf(matches[3], "%d", &day)
//			if err != nil {
//				err = NewErr(
//					ErrFailedToParseDateFormat,
//					"expected_format", "dd",
//					err,
//				)
//				goto end
//			}
//
//			// Validate day range
//			if day < 1 || day > 31 {
//				err = NewErr(
//					ErrInvalidDayInDate,
//					"invalid_day", day,
//				)
//				goto end
//			}
//		}
//
//		goto end
//	}
//
//	// Non-multisegment: require full date
//	matched = dtclassifiers.shortISO8601DateRegex.MatchString(value)
//	if !matched {
//		err = NewErr(
//			ErrFailedToMatchDateFormat,
//			"expected_format", errMsgFormat,
//		)
//		goto end
//	}
//
//	// Parse and validate date components
//	_, err = fmt.Sscanf(value, scanFormat, &year, &month, &day)
//	if err != nil {
//		err = NewErr(
//			ErrFailedToParseDateFormat,
//			"expected_format", errMsgFormat,
//			err,
//		)
//		goto end
//	}
//
//	// Validate year
//	if year < 1000 || year > 9999 {
//		err = NewErr(
//			ErrInvalidYearInDate,
//			"invalid_year", year,
//		)
//		goto end
//	}
//
//	// Validate month range
//	if month < 1 || month > 12 {
//		err = NewErr(
//			ErrInvalidMonthInDate,
//			"invalid_month", month,
//		)
//		goto end
//	}
//
//	// Validate day range
//	if day < 1 || day > 31 {
//		err = NewErr(
//			ErrInvalidDayInDate,
//			"invalid_day", day,
//		)
//		goto end
//	}
//
//end:
//	if err != nil {
//		err = WithErr(err,
//			ErrParameterValidationFailed,
//			ErrInvalidDateFormat,
//		)
//	}
//	return err
//}
//
//// validateUUID validates that the value conforms to standard UUID format
//// (8-4-4-4-12 hexadecimal digits with hyphens).
//func (p NameSpecProps) validateUUID(value string) (err error) {
//	var matched bool
//	var regex *regexp.Regexp
//
//	// Basic UUID pattern: 8-4-4-4-12 hex digits
//	regex = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
//	matched = regex.MatchString(value)
//	if !matched {
//		err = NewErr(
//			ErrParameterValidationFailed,
//			ErrInvalidUUIDFormatBasic,
//			"pattern", "8-4-4-4-12 hex digits",
//		)
//	}
//
//	return err
//}
//
//// validateAlphanum validates that the value contains only alphanumeric characters
//// (letters and digits, no spaces or special characters).
//func (p NameSpecProps) validateAlphanum(value string) (err error) {
//	var matched bool
//	var regex *regexp.Regexp
//
//	regex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
//	matched = regex.MatchString(value)
//	if !matched {
//		err = NewErr(
//			ErrParameterValidationFailed,
//			ErrInvalidAlphanumericFormat,
//			"regex", "^[a-zA-Z0-9]+$",
//		)
//	}
//
//	return err
//}
//
//// validateSlug validates that the value conforms to URL-safe slug format:
//// lowercase letters, numbers, and hyphens, with no leading/trailing hyphens.
//func (p NameSpecProps) validateSlug(value string) (err error) {
//	var matched bool
//	var regex *regexp.Regexp
//
//	// Lowercase letters, numbers, hyphens, no leading/trailing hyphens
//	regex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
//	matched = regex.MatchString(value)
//	if !matched {
//		err = NewErr(
//			ErrParameterValidationFailed,
//			ErrInvalidSlugFormat,
//			"regex", "^[a-z0-9]+(?:-[a-z0-9]+)*$",
//		)
//	}
//
//	return err
//}
//
//// validateBoolean validates that the value is exactly "true" or "false".
//func (p NameSpecProps) validateBoolean(value string) (err error) {
//	if value != "true" && value != "false" {
//		err = NewErr(
//			ErrParameterValidationFailed,
//			ErrInvalidBooleanFormat,
//			"allowedValues", "true,false",
//		)
//	}
//	return err
//}
//
//// validateEmail validates that the value conforms to basic email format
//// using a simple regex pattern (local@domain with basic validation).
//func (p NameSpecProps) validateEmail(value string) (err error) {
//	var matched bool
//	var regex *regexp.Regexp
//
//	// Basic email regex: local@domain with basic validation
//	regex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
//	matched = regex.MatchString(value)
//	if !matched {
//		err = NewErr(
//			ErrParameterValidationFailed,
//			ErrInvalidEmailFormat,
//			"pattern", "local@domain",
//		)
//	}
//
//	return err
//}
