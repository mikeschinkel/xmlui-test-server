package errutil

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type ParsedError struct {
	Messages       []string
	MessageDetails []map[string]string
	Details        map[string][]string
	CustomErrors   map[reflect.Type][]error
}

func NewParsedError() ParsedError {
	return ParsedError{
		Messages:       make([]string, 0),
		CustomErrors:   make(map[reflect.Type][]error),
		MessageDetails: make([]map[string]string, 0),
		Details:        make(map[string][]string),
	}
}

func (pe ParsedError) GetMessageDetail(name, message string) (value string, err error) {
	var ok bool
	idx := -1
	for i, m := range pe.Messages {
		if m != message {
			continue
		}
		idx = i
		break
	}
	if idx == -1 {
		err = ErrMessageNotFoundInParsedError
		goto end
	}
	value, ok = pe.MessageDetails[idx][name]
	if !ok {
		err = errors.Join(
			ErrDetailNotFoundInParsedError,
			fmt.Errorf("detail_name=%s", name),
		)
		goto end
	}
end:
	if err != nil {
		err = errors.Join(err,
			fmt.Errorf("message=%s", message),
		)
	}
	return value, err
}

func (pe ParsedError) GetMessageDetailByIndex(name string, index int) (value string, err error) {
	var ok bool
	if index < 0 {
		err = errors.Join(err,
			ErrInvalidMessageIndex,
			ErrMessageExceedsMinimum,
		)
	}
	if index >= len(pe.MessageDetails) {
		err = errors.Join(err,
			ErrInvalidMessageIndex,
			ErrMessageExceedsMaximum,
		)
	}
	value, ok = pe.MessageDetails[index][name]
	if !ok {
		err = errors.Join(err,
			ErrMessageDetailNotFoundInParsedError,
			fmt.Errorf("detail_name=%s", name),
		)
		goto end
	}
end:
	if err != nil {
		err = errors.Join(err,
			fmt.Errorf("message_index=%d", index),
			fmt.Errorf("valid_range=0..%d", len(pe.MessageDetails)-1),
		)
	}
	return value, err
}

func (pe ParsedError) MaybeGetCustomError(typ reflect.Type) (err error) {
	err, _ = pe.GetCustomError(typ)
	return err
}

func (pe ParsedError) GetCustomError(typ reflect.Type) (_ error, err error) {
	var errs []error
	var ok bool
	errs, ok = pe.CustomErrors[typ]
	if len(errs) == 0 {
		errs = append(errs, nil)
	}
	if !ok {
		err = errors.Join(
			ErrCustomErrorNotFoundInParsedError,
			fmt.Errorf("error_type=%s", typ.String()),
		)
		goto end
	}
end:
	return errs[0], err
}

func (pe ParsedError) GetCustomErrors(typ reflect.Type) (errs []error, err error) {
	var ok bool
	errs, ok = pe.CustomErrors[typ]
	if !ok {
		err = errors.Join(err,
			ErrCustomErrorTypeNotFoundInParsedError,
			fmt.Errorf("error_type=%s", typ.String()),
		)
	}
	return errs, err
}

func (pe ParsedError) GetIntDetail(name string) (value int, err error) {
	s, err := pe.GetDetail(name)
	if err != nil {
		goto end
	}
	value, err = strconv.Atoi(s)
end:
	return value, err
}

func (pe ParsedError) MaybeGetDetail(name string) (value string) {
	value, _ = pe.GetDetail(name)
	return value
}

func (pe ParsedError) MaybeGetIntDetail(name string) (value int) {
	value, _ = pe.GetIntDetail(name)
	return value
}

func (pe ParsedError) GetDetail(name string) (value string, err error) {
	values, ok := pe.Details[name]
	if !ok {
		err = ErrDetailNotFoundInParsedError
		goto end
	}
	if len(values) == 0 {
		err = ErrDetailEmptyInParsedError
		goto end
	}
	value = values[0]
end:
	if err != nil {
		err = errors.Join(err,
			fmt.Errorf("detail_name=%s", name),
		)
	}
	return value, err
}

func (pe ParsedError) GetDetails(name string) (values []string, err error) {
	values, ok := pe.Details[name]
	if !ok {
		err = ErrDetailNotFoundInParsedError
		goto end
	}
	if len(values) == 0 {
		err = ErrDetailEmptyInParsedError
		goto end
	}
end:
	return values, err
}

func (pe ParsedError) HasErrors() (has bool) {
	has = true
	if len(pe.Messages) != 0 {
		goto end
	}
	if len(pe.Details) != 0 {
		goto end
	}
	if len(pe.MessageDetails) != 0 {
		goto end
	}
	has = false
end:
	return has
}

func ParseError(err error) (pe ParsedError, _ error) {
	var lines []string
	pe = NewParsedError()
	if err == nil {
		goto end
	}

	for _, ue := range UnwrapAllJoinErrors(err) {
		if !IsCustomError(ue) {
			continue
		}
		t := reflect.TypeOf(ue)
		_, ok := pe.CustomErrors[t]
		if !ok {
			pe.CustomErrors[t] = make([]error, 0)
		}
		pe.CustomErrors[t] = append(pe.CustomErrors[t], ue)
	}
	lines = strings.Split(err.Error(), "\n")
	for _, line := range lines {
		matches := captureDetailsRegex.FindStringSubmatch(line)
		if len(matches) != 0 {
			name := matches[1]
			value := matches[2]

			pe.Details[name] = append(pe.Details[name], value)

			if len(pe.MessageDetails) == 0 {
				pe.MessageDetails = append(pe.MessageDetails, make(map[string]string))
			}
			pe.MessageDetails[len(pe.MessageDetails)-1][name] = value

			continue
		}
		pe.Messages = append(pe.Messages, line)
	}
end:
	return pe, nil
}
