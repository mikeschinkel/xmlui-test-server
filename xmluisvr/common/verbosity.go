package common

import (
	"errors"
)

type Verbosity int

const DefaultVerbosity = LowVerbosity

const (
	NoVerbosity Verbosity = iota
	LowVerbosity
	MediumVerbosity
	HighVerbosity
)

var (
	ErrInvalidateVerbosity = errors.New("invalid verbosity level")
	ErrVerbosityTooLow     = errors.New("verbosity too low; must be between 0..3 inclusive")
	ErrVerbosityTooHigh    = errors.New("verbosity too high; must be between 0..3 inclusive")
)

func ParseVerbosity(verbosity int) (v Verbosity, err error) {
	v = Verbosity(verbosity)
	switch {
	case v < NoVerbosity:
		err = ErrVerbosityTooLow

	case v > HighVerbosity:
		err = ErrVerbosityTooHigh
	}
	if err != nil {
		v = -1
		err = NewErr(
			ErrInvalidIdentifier,
			err,
			"verbosity", v,
		)
	}
	return v, err
}
