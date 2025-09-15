package common

import (
	"errors"
	"strconv"
	"time"
)

var (
	ErrURLPathMustNotBeEmpty = errors.New("URL path must not be empty")
	ErrHostMustNotBeEmpty    = errors.New("host must not be empty")
	ErrPortMustNotBeZero     = errors.New("port must not be zero")
	ErrPortMustBeZero        = errors.New("port must be zero")
)

func ParseURLPath(p string) (up URLPath, err error) {
	if p == "" {
		err = ErrURLPathMustNotBeEmpty
	}
	// TODO Add some validation here
	up = URLPath(p)
	return up, err
}
func ParseHost(h string) (_ Host, err error) {
	if h == "" {
		err = ErrHostMustNotBeEmpty
	}
	// TODO Add some validation here
	return Host(h), err
}

type ZeroHandling int

const (
	UnspecifiedZeroHandling ZeroHandling = iota
	ZeroOk
	ZeroInvalid
	ZeroRequired
)

func ParseServerPort(p int, zh ZeroHandling) (sp ServerPort, err error) {
	switch zh {
	case ZeroOk:
		// S'all good, man
	case ZeroInvalid:
		if p == 0 {
			err = ErrPortMustNotBeZero
		}
	case ZeroRequired:
		// I cannot imagine this will ever be used, but here for symmetry
		if p != 0 {
			err = ErrPortMustBeZero
		}
	case UnspecifiedZeroHandling:
		fallthrough
	default:
		logger.Error("ParseServerPort: invalid zero handling value", "value", zh)
		goto end
	}
	// TODO Add some validation here
	sp = ServerPort(p)
end:
	return sp, err
}

func ParseFilepath(f string) (fp Filepath, err error) {
	// TODO Add some validation here
	fp = Filepath(f)
	return fp, err
}

func ParseFilepaths(files []string) (fps []Filepath, _ error) {
	var errs []error
	fps = make([]Filepath, 0, len(files))
	for _, file := range files {
		fp, err := ParseFilepath(file)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		fps = append(fps, fp)
	}
	return fps, errors.Join(errs...)
}

func ParseConnectString(s string) (cs ConnectString, err error) {
	// TODO MAYBE add some validation here
	cs = ConnectString(s)
	return cs, err
}

func ParseDirPath(f string) (dp DirPath, err error) {
	// TODO Add some validation here
	dp = DirPath(f)
	return dp, err
}

func ParseIdentifier(s string) (id Identifier, err error) {
	// TODO Add some validation here
	id = Identifier(s)
	return id, err
}

// ParseTimeDurationEx parses a string as EITHER a Go duration format
// (like "3s", "10m", "1h30m") OR as an integer representing seconds.
func ParseTimeDurationEx(s string) (td time.Duration, err error) {
	var seconds int
	var errs []error

	// First try parsing as integer seconds
	seconds, err = strconv.Atoi(s)
	if err == nil {
		td = time.Duration(seconds) * time.Second
		goto end
	}
	errs = append(errs, err)

	// If that fails, try parsing as a standard Go duration
	td, err = time.ParseDuration(s)
	if err == nil {
		goto end
	}
	errs = append(errs, err)

end:
	return td, errors.Join(errs...)
}
