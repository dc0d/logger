package logger

import (
	"fmt"
	stdlog "log"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
)

// here info about code location with a string representation formatted as
// <dir>/<file>.go@<line>:<package>.<function>()
func here(skip ...int) loc {
	sk := 1
	if len(skip) > 0 && skip[0] > 1 {
		sk = skip[0]
	}
	pc, fileName, fileLine, ok := runtime.Caller(sk)
	fn := runtime.FuncForPC(pc)
	var res loc
	defer func() {
		if res.long != "" {
			return
		}
		res.long = res.FuncName
	}()
	if !ok {
		res.FuncName = "N/A"
		return res
	}
	res.FileName = fileName
	res.FileLine = fileLine
	res.FuncName = fn.Name()
	fileName = filepath.Join(filepath.Base(filepath.Dir(fileName)), filepath.Base(fileName))
	res.long = fmt.Sprintf("%s@%d:%s()", fileName, res.FileLine, res.FuncName)
	res.short = fmt.Sprintf("%s@%d:%s()", fileName, res.FileLine, strings.TrimLeft(filepath.Ext(res.FuncName), "."))
	return res
}

// loc info about code location with a string representation formatted as
// <dir>/<file>.go@<line>:<package>.<function>()
type loc struct {
	FuncName string
	FileName string
	FileLine int

	long  string
	short string
}

func (l loc) String() string {
	return l.long
}

// flags, overwrites default flags.
const (
	Ldate         = 1 << iota // the date in the local time zone: 2009/01/23
	Ltime                     // the time in the local time zone: 01:23:23
	Lmicroseconds             // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                 // long caller description.
	Lshortfile                // short caller description.
	LUTC                      // if Ldate or Ltime is set, use UTC rather than the local time zone
	LDebug                    // level debug
	LInfo                     // level info
	LWarn                     // level warn
	LError                    // level error

	hastime = Ldate | Ltime | Lmicroseconds | LUTC
	hasloc  = Llongfile | Lshortfile
)

// Logger adds word pair logs to the standard logger
type Logger struct {
	*stdlog.Logger
	flags int
}

// New creates a *Logger. Panics if logger is nil.
func New(logger *stdlog.Logger) (res *Logger) {
	if logger == nil {
		panic("logger can not be nil")
	}
	res = new(Logger)
	res.Logger = logger
	res.flags = logger.Flags()
	logger.SetFlags(0)

	switch {
	case res.flags&LError == LError:
		logger.SetPrefix(color.New(color.FgHiRed).Sprintf("level=error") + " ")
	case res.flags&LWarn == LWarn:
		logger.SetPrefix(color.New(color.FgHiYellow).Sprintf("level=warn") + " ")
	case res.flags&LDebug == LDebug:
		logger.SetPrefix(color.New(color.FgWhite).Sprintf("level=debug") + " ")
	case res.flags&LInfo == LInfo:
		fallthrough
	default:
		logger.SetPrefix(color.New(color.FgBlue).Sprintf("level=info") + " ")
	}

	return
}

// Fatalln .
func (l *Logger) Fatalln(v ...interface{}) {
	l.Logger.Println(l.sprint(v...))
}

// Panicln .
func (l *Logger) Panicln(v ...interface{}) {
	l.Logger.Println(l.sprint(v...))
}

// Println needs pairs.
func (l *Logger) Println(v ...interface{}) {
	l.Logger.Println(l.sprint(v...))
}

func (l *Logger) sprint(v ...interface{}) string {
	var parts []string
	{
		h := l.headers()
		if h != "" {
			parts = append(parts, h)
		}
	}
	for i := 0; i < len(v)-1; i += 2 {
		k, v := fmt.Sprint(v[i]), fmt.Sprint(v[i+1])
		if strings.ContainsAny(k, " \n\r\t\"") {
			k = fmt.Sprintf("%q", k)
		}
		if strings.ContainsAny(v, " \n\r\t\"") {
			v = fmt.Sprintf("%q", v)
		}
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	if len(v)&1 == 1 {
		parts = append(parts, fmt.Sprintf("msg=%q", v[len(v)-1]))
	}
	return strings.Join(parts, " ")
}

func (l *Logger) headers() string {
	var parts []string

	if l.flags&hastime > 0 {
		now := time.Now()
		if l.flags&LUTC == LUTC {
			now = now.UTC()
		}
		var format string
		if l.flags&Ldate == Ldate {
			format = "2006-01-02"
		}
		if l.flags&Ltime == Ltime {
			if format != "" {
				format += "T"
			}
			format += "15:04:05"
			if l.flags&Lmicroseconds == Lmicroseconds {
				format += ".000000"
			}
		}
		if format == "" {
			format = time.RFC3339
		}

		parts = append(parts, "time="+now.Format(format))
	}

	if l.flags&hasloc > 0 {
		loc := here(2)
		switch {
		case l.flags&Llongfile == Llongfile:
			parts = append(parts, "loc="+loc.long)
		case l.flags&Lshortfile == Lshortfile:
			parts = append(parts, "loc="+loc.short)
		}
	}

	return strings.Join(parts, " ")
}
