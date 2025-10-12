package errparsr

import (
	"errors"
)

var (
	ErrMessageNotFoundInParsedError         = errors.New("message not found in parsed error")
	ErrDetailNotFoundInParsedError          = errors.New("detail not found in parsed error")
	ErrMessageDetailNotFoundInParsedError   = errors.New("message detail not found in parsed error")
	ErrDetailEmptyInParsedError             = errors.New("detail empty in parsed error")
	ErrInvalidMessageIndex                  = errors.New("invalid message index")
	ErrMessageExceedsMaximum                = errors.New("message index exceeds maximum")
	ErrMessageExceedsMinimum                = errors.New("message index exceeds minimum")
	ErrCustomErrorTypeNotFoundInParsedError = errors.New("custom error type not found in parsed error")
	ErrCustomErrorNotFoundInParsedError     = errors.New("custom error not found in parsed error")
)
