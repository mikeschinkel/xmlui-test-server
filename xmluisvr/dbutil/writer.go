package dbutil

type CLIWriter interface {
	Printf(format string, args ...any)
	Errorf(format string, args ...any)
}
