package cliutil

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

type WriterLogger struct {
	Writer
	*slog.Logger
	v2 *WriterLogger
	v3 *WriterLogger
}

func NewWriterLogger(writer Writer, logger *slog.Logger) WriterLogger {
	return WriterLogger{Writer: writer, Logger: logger}
}

func (wl WriterLogger) Info(format string, args ...any) {
	wl.Logger.Info(fmt.Sprintf(format, args...))
}

func (wl WriterLogger) Printf(format string, args ...any) {
	wl.Writer.Printf(format, args...)
}

func (wl WriterLogger) V2() WriterLogger {
	if wl.v2 == nil {
		wl.v2 = &WriterLogger{
			Writer: wl.Writer.V2(),
			Logger: wl.Logger,
		}
	}
	return *wl.v2
}

func (wl WriterLogger) V3() WriterLogger {
	if wl.v3 == nil {
		wl.v3 = &WriterLogger{
			Writer: wl.Writer.V3(),
			Logger: wl.Logger,
		}
	}
	return *wl.v3
}

func (wl WriterLogger) ErrorError(msg string, args ...any) (err error) {
	var ok bool
	wl.Logger.Error(msg, args...)
	msg = wl.concatMsgAndArgs("ErrorError", msg, args...)
	wl.Writer.Errorf(msg + "\n")
	if len(args) == 0 {
		err = errors.New(msg)
		goto end
	}
	err, ok = args[len(args)-1].(error)
	if !ok {
		err = errors.New(msg)
		goto end
	}
	if strings.HasSuffix(msg, err.Error()) {
		err = errors.New(msg)
		goto end
	}
	err = NewErr(errors.New(msg), err)
end:
	return err
}

func (wl WriterLogger) WarnError(msg string, args ...any) {
	wl.Logger.Warn(msg, args...)
	wl.Writer.Errorf(wl.concatMsgAndArgs("WarnError", msg, args...) + "\n")
}

func (wl WriterLogger) InfoPrint(msg string, args ...any) {
	wl.Logger.Info(msg, args...)
	wl.Writer.Printf(wl.concatMsgAndArgs("InfoPrint", msg, args...) + "\n")
}

func (wl WriterLogger) InfoLoud(msg string, args ...any) {
	wl.Logger.Info(msg, args...)
	wl.Writer.Loud().Printf(wl.concatMsgAndArgs("InfoPrint", msg, args...) + "\n")
}

func (wl WriterLogger) concatMsgAndArgs(caller string, msg string, args ...any) string {
	var sb strings.Builder
	last := len(args) - 1
	sb.WriteString(msg)
	for i := 0; i < len(args); i += 2 {
		if i == last && i == len(args)-1 {
			sb.WriteString(fmt.Sprintf(" %v", args[i]))
			goto end
		}
		sb.WriteString(fmt.Sprintf(" %s=%v", args[i], args[i+1]))
	}
end:
	return sb.String()
}
