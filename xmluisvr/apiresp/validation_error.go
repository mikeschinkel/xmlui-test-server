package apiresp

type ValidationError struct {
	Parameter string       `json:"parameter"`
	Location  LocationType `json:"location"`
	Expected  string       `json:"expected"`
	Received  string       `json:"received"`
	Message   string       `json:"message"`
}

func NewValidationError(args ValidationErrorArgs) ValidationError {
	return ValidationError{
		Parameter: args.Parameter,
		Location:  args.Location,
		Expected:  args.Expected,
		Received:  args.Received,
		Message:   args.Message,
	}
}

type ValidationErrorArgs struct {
	Parameter string       `json:"parameter"`
	Location  LocationType `json:"location"`
	Expected  string       `json:"expected"`
	Received  string       `json:"received"`
	Message   string       `json:"message"`
}
