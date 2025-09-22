package pathvars_test

import (
	"fmt"
	"testing"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
)

func TestParseRangeConstraint(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		dataType pathvars.PVDataType
		wantErr  bool
		wantType string
	}{
		// Integer range constraints
		{"int-valid-range", "1..10", pathvars.IntegerType, false, "*pathvars.IntegerRangeConstraint"},
		{"int-negative-range", "-10..10", pathvars.IntegerType, false, "*pathvars.IntegerRangeConstraint"},
		{"int-large-range", "1000..9999", pathvars.IntegerType, false, "*pathvars.IntegerRangeConstraint"},
		{"int-invalid-format", "1-10", pathvars.IntegerType, true, ""},
		{"int-invalid-order", "10..1", pathvars.IntegerType, true, ""},
		{"int-non-numeric", "abc..def", pathvars.IntegerType, true, ""},

		// Decimal/Real range constraints
		{"decimal-valid-range", "0.5..9.5", pathvars.DecimalType, false, "*pathvars.DecimalRangeConstraint"},
		{"decimal-negative-range", "-5.5..5.5", pathvars.DecimalType, false, "*pathvars.DecimalRangeConstraint"},
		{"decimal-integer-values", "1..10", pathvars.DecimalType, false, "*pathvars.DecimalRangeConstraint"},
		{"real-valid-range", "0.1..99.9", pathvars.RealType, false, "*pathvars.DecimalRangeConstraint"},
		{"real-scientific", "1e-3..1e3", pathvars.RealType, false, "*pathvars.DecimalRangeConstraint"},
		{"decimal-invalid-format", "1.5-9.5", pathvars.DecimalType, true, ""},
		{"decimal-invalid-order", "9.5..1.5", pathvars.DecimalType, true, ""},

		// Date range constraints
		{"date-valid-range", "2023-01-01..2023-12-31", pathvars.DateType, false, "*pathvars.DateRangeConstraint"},
		{"date-different-years", "2020-01-01..2025-12-31", pathvars.DateType, false, "*pathvars.DateRangeConstraint"},
		{"date-invalid-format", "2023/01/01..2023/12/31", pathvars.DateType, true, ""},
		{"date-invalid-order", "2023-12-31..2023-01-01", pathvars.DateType, true, ""},
		{"date-malformed", "not-a-date..also-not-a-date", pathvars.DateType, true, ""},

		// Unsupported data types
		{"string-unsupported", "a..z", pathvars.StringType, true, ""},
		{"boolean-unsupported", "true..false", pathvars.BooleanType, true, ""},
		{"uuid-unsupported", "uuid1..uuid2", pathvars.UUIDType, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pathvars.ParseRangeConstraint(tt.spec, tt.dataType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseRangeConstraint(%q, %v) expected error, got nil", tt.spec, tt.dataType)
				}
				if constraint != nil {
					t.Errorf("ParseRangeConstraint(%q, %v) expected nil constraint on error, got %T", tt.spec, tt.dataType, constraint)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseRangeConstraint(%q, %v) unexpected error: %v", tt.spec, tt.dataType, err)
				return
			}

			if constraint == nil {
				t.Errorf("ParseRangeConstraint(%q, %v) returned nil constraint", tt.spec, tt.dataType)
				return
			}

			// Check constraint type
			constraintType := fmt.Sprintf("%T", constraint)
			if constraintType != tt.wantType {
				t.Errorf("ParseRangeConstraint(%q, %v) returned %s, want %s", tt.spec, tt.dataType, constraintType, tt.wantType)
			}

			// Check that constraint type is correct
			if constraint.Type() != pathvars.RangeConstraintType {
				t.Errorf("ParseRangeConstraint(%q, %v) constraint.Type() = %v, want %v", tt.spec, tt.dataType, constraint.Type(), pathvars.RangeConstraintType)
			}
		})
	}
}

func TestParseRangeConstraintValidation(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		dataType pathvars.PVDataType
		testVal  string
		wantErr  bool
	}{
		// Test that parsed constraints actually validate correctly
		{"int-range-valid-min", "10..20", pathvars.IntegerType, "10", false},
		{"int-range-valid-max", "10..20", pathvars.IntegerType, "20", false},
		{"int-range-valid-middle", "10..20", pathvars.IntegerType, "15", false},
		{"int-range-invalid-low", "10..20", pathvars.IntegerType, "9", true},
		{"int-range-invalid-high", "10..20", pathvars.IntegerType, "21", true},

		{"decimal-range-valid", "1.5..2.5", pathvars.DecimalType, "2.0", false},
		{"decimal-range-invalid", "1.5..2.5", pathvars.DecimalType, "3.0", true},

		{"real-range-valid", "0.1..0.9", pathvars.RealType, "0.5", false},
		{"real-range-invalid", "0.1..0.9", pathvars.RealType, "1.0", true},

		{"date-range-valid", "2023-01-01..2023-12-31", pathvars.DateType, "2023-06-15", false},
		{"date-range-invalid", "2023-01-01..2023-12-31", pathvars.DateType, "2024-01-01", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pathvars.ParseRangeConstraint(tt.spec, tt.dataType)
			if err != nil {
				t.Fatalf("ParseRangeConstraint(%q, %v) failed: %v", tt.spec, tt.dataType, err)
			}

			err = constraint.Validate(tt.testVal)
			if tt.wantErr && err == nil {
				t.Errorf("constraint.Validate(%q) expected error, got nil", tt.testVal)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("constraint.Validate(%q) unexpected error: %v", tt.testVal, err)
			}
		})
	}
}
