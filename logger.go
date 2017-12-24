package logger

import (
	"fmt"
	stdlog "log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// New creates a *Logger. Panics if logger is nil.
func New(logger *stdlog.Logger, flags ...Flag) *Logger {
	if logger == nil {
		panic(ErrNilLogger)
	}
	var buffer []interface{}
	for _, v := range flags {
		buffer = append(buffer, v)
	}
	return &Logger{
		Logger: logger,
		flags:  buffer,
	}
}

// Logger adds word pair logs to the standard logger
type Logger struct {
	*stdlog.Logger
	flags []interface{}
}

// Fatalw .
func (l *Logger) Fatalw(v ...interface{}) {
	v = append(v, l.flags...)
	l.Logger.Fatal(logfmt(v...))
}

// Panicw .
func (l *Logger) Panicw(v ...interface{}) {
	v = append(v, l.flags...)
	l.Logger.Panic(logfmt(v...))
}

// Printw .
func (l *Logger) Printw(v ...interface{}) {
	v = append(v, l.flags...)
	l.Logger.Print(logfmt(v...))
}

func logfmt(pairs ...interface{}) string {
	var res []string

	var buf []interface{}
	for _, v := range pairs {
		switch v {
		case LTime:
			res = append(res, fmt.Sprintf("time=%v", time.Now().Format("2006-01-02T15:04:05")))
		case LApp:
			res = append(res, fmt.Sprintf("app=%v", filepath.Base(os.Args[0])))
		case LCaller:
			var name string
			funcName, fileName, fileLine, err := here(3)
			if err != nil {
				name = "N/A"
			} else {
				name = fmt.Sprintf(markerFormat, fileName, fileLine, funcName)
			}
			res = append(res, fmt.Sprintf("location=%v", name))
		default:
			buf = append(buf, v)
		}
	}
	pairs = buf

	if len(pairs)&1 == 1 {
		res = append(res, "msg="+quote(pairs[0]))
		pairs = pairs[1:]
	}
	for i := 0; i < len(pairs)-1; i = i + 2 {
		k := fmt.Sprint(pairs[i])
		v := quote(fmt.Sprint(pairs[i+1]))
		res = append(res, fmt.Sprintf("%v=%v", k, v))
	}

	return strings.Join(res, " ")
}

func here(skip ...int) (funcName, fileName string, fileLine int, callerErr error) {
	sk := 1
	if len(skip) > 0 && skip[0] > 1 {
		sk = skip[0]
	}
	var pc uintptr
	var ok bool
	pc, fileName, fileLine, ok = runtime.Caller(sk)
	if !ok {
		callerErr = ErrNotAvailable
		return
	}
	fn := runtime.FuncForPC(pc)
	name := fn.Name()
	ix := strings.LastIndex(name, ".")
	if ix > 0 && (ix+1) < len(name) {
		name = name[ix+1:]
	}
	funcName = name
	nd, nf := filepath.Split(fileName)
	fileName = filepath.Join(filepath.Base(nd), nf)
	return
}

// Flag .
type Flag int

// valid values for FlagFlag
const (
	LTime Flag = iota
	LCaller
	LApp
)

func quote(v interface{}) string {
	s := fmt.Sprint(v)
	if strings.Contains(s, " ") {
		s = strings.Replace(s, `"`, "__", -1)
		s = `"` + s + `"`
	}
	return s
}

const markerFormat = "%s/%02d:%s()"

func errorf(format string, a ...interface{}) error {
	return sentinelErr(fmt.Sprintf(format, a...))
}

// errors
var (
	ErrNotAvailable = errorf("N/A")
	ErrNilLogger    = errorf("LOGGER MUST BE NON-NIL")
)

type sentinelErr string

func (v sentinelErr) Error() string { return string(v) }
