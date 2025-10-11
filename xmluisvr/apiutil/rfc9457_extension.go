package apiutil

import (
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/rfc9457"
)

type RFC9457Extension struct {
	Parameter        string            `json:"parameter,omitempty"`
	ExpectedType     string            `json:"expected_type,omitempty"`
	ReceivedValue    string            `json:"received_value,omitempty"`
	Location         LocationType      `json:"location,omitempty"`
	Constraint       any               `json:"constraint,omitempty"`
	Suggestion       string            `json:"suggestion,omitempty"`
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
}

type RFC9457ExtensionArgs struct {
	Parameter        string
	ExpectedType     string
	ReceivedValue    string
	Location         LocationType
	Constraint       any
	Suggestion       string
	ValidationErrors []ValidationError
}

func NewRFC9457Extension(args RFC9457ExtensionArgs) *RFC9457Extension {
	return &RFC9457Extension{
		Parameter:        args.Parameter,
		ExpectedType:     args.ExpectedType,
		ReceivedValue:    args.ReceivedValue,
		Location:         args.Location,
		Constraint:       args.Constraint,
		Suggestion:       args.Suggestion,
		ValidationErrors: args.ValidationErrors,
	}
}

func (r *RFC9457Extension) Error() string {
	return rfc9457.SprintfMany("RFC9547 Extension:", "\n",
		"parameter=%s", r.Parameter,
		"expected_type=%s", r.ExpectedType,
		"received_value=%s", r.ReceivedValue,
		"error_location=%s", r.Location,
		"constraint=%v", r.Constraint,
		"suggestion=%s", r.Suggestion,
		"validation_errors=%v", r.ValidationErrors,
		"received_value=%s", r.ReceivedValue,
	)
}

func (r *RFC9457Extension) SetSuggestion(suggestion string) {
	r.Suggestion = suggestion
}

func (r *RFC9457Extension) SetConstraint(constraint string) {
	r.Constraint = constraint
}

func (r *RFC9457Extension) AddValidationError(ve ValidationError) {
	if r.ValidationErrors == nil {
		r.ValidationErrors = []ValidationError{}
	}
	r.ValidationErrors = append(r.ValidationErrors, ve)
}
