package pathvars

// ParseRangeConstraint uses the registry to find and parse the appropriate range constraint based on data type
func ParseRangeConstraint(rangeSpec string, dataType PVDataType) (constraint Constraint, err error) {
	constraint, err = GetConstraint(RangeConstraintType, dataType)
	if err != nil {
		goto end
	}

	constraint, err = constraint.Parse(rangeSpec, dataType)

end:
	if err != nil {
		// Make sure constraint is nil and not a non-nil interface containing a nil.
		constraint = nil
	}
	return constraint, err
}
