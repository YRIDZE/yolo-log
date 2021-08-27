package yolo_log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	timeFormat = "2006/01/02 15:04:05"

	boldRed string = "\x1b[31;1m"
	magenta string = "\x1b[35m"
	yellow  string = "\x1b[33m"
	green   string = "\x1b[32m"
	blue    string = "\x1b[34m"
	dim     string = "\x1b[2m"
	normal  string = "\x1b[0m"
)

// region Supplementary definitions

type Severity int

const (
	DEBUG Severity = iota
	INFO
	WARNING
	ERROR
	FATAL
)

func (l Severity) String() string {
	return [...]string{
		"[DEBUG]  ",
		"[INFO]   ",
		"[WARNING]",
		"[ERROR]  ",
		"[FATAL]  ",
	}[l]
}

type LoggerOutput struct {
	Writer   io.Writer
	Mutex    sync.Mutex
	LogLevel Severity
}

func (lo *LoggerOutput) SyncedPrint(line string) {
	lo.Mutex.Lock()
	lo.Writer.Write([]byte(line))
	lo.Mutex.Unlock()
}

type LoggerParams struct {
	ConsoleOutputStream io.Writer
	ConsoleLogLevel     Severity
	LogFileName         string
	FileLogLevel        Severity
}

// endregion

// region Logger definition and methods

type Logger struct {
	ConsoleOutput *LoggerOutput
	FileOutput    *LoggerOutput
}

func NewLogger(params LoggerParams) (*Logger, error) {
	var consoleOutput *LoggerOutput = nil
	if params.ConsoleOutputStream != nil {
		consoleOutput = new(LoggerOutput)
		consoleOutput.LogLevel = params.ConsoleLogLevel
		consoleOutput.Mutex = sync.Mutex{}
		consoleOutput.Writer = params.ConsoleOutputStream
	}

	var fileOutput *LoggerOutput = nil
	if params.LogFileName != "" {
		fileOutput = new(LoggerOutput)
		fileOutput.LogLevel = params.FileLogLevel
		fileOutput.Mutex = sync.Mutex{}

		var err error
		fileOutput.Writer, err = os.OpenFile(params.LogFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
		if err != nil {
			return nil, err
		}
	}

	return &Logger{
		consoleOutput, fileOutput,
	}, nil
}

func (l *Logger) getExecutionLocation() (fileName string, funcName string, line int) {
	pc, fileName, line, ok := runtime.Caller(3)
	if !ok {
		fileName = "?"
		line = 0
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		funcName = "?()"
	} else {
		dotName := filepath.Ext(fn.Name())
		funcName = strings.TrimLeft(dotName, ".") + "()"
	}

	return
}

func (l *Logger) output(severity Severity, msg string, color string) {
	fileName, funcName, line := l.getExecutionLocation()
	now := time.Now().Format(timeFormat)

	if l.ConsoleOutput != nil && severity >= l.ConsoleOutput.LogLevel {
		var consoleLine string = fmt.Sprintf("%s%s%s %s%s%s %s:%d %s: %s\n",
			color, severity.String(), normal, dim, now, normal, fileName, line, funcName, msg)

		l.ConsoleOutput.SyncedPrint(consoleLine)
	}

	if l.FileOutput != nil && severity >= l.FileOutput.LogLevel {
		var fileLine string = fmt.Sprintf("%s %s %s:%d %s: %s\n",
			severity.String(), now, fileName, line, funcName, msg)

		l.FileOutput.SyncedPrint(fileLine)
	}
}

// region Convenience log function shortcuts

func (l *Logger) Debug(msg ...interface{}) {
	l.output(DEBUG, fmt.Sprint(msg...), blue)
}
func (l *Logger) Debugf(format string, msg ...interface{}) {
	l.output(DEBUG, fmt.Sprintf(format, msg...), blue)
}

func (l *Logger) Info(msg ...interface{}) {
	l.output(INFO, fmt.Sprint(msg...), green)
}
func (l *Logger) Infof(format string, msg ...interface{}) {
	l.output(INFO, fmt.Sprintf(format, msg...), green)
}

func (l *Logger) Error(msg ...interface{}) {
	l.output(ERROR, fmt.Sprint(msg...), magenta)
}
func (l *Logger) Errorf(format string, msg ...interface{}) {
	l.output(ERROR, fmt.Sprintf(format, msg...), magenta)
}

func (l *Logger) Warning(msg ...interface{}) {
	l.output(WARNING, fmt.Sprint(msg...), yellow)
}
func (l *Logger) Warningf(format string, msg ...interface{}) {
	l.output(WARNING, fmt.Sprintf(format, msg...), yellow)
}

func (l *Logger) Fatal(msg ...interface{}) {
	l.output(FATAL, fmt.Sprint(msg...), boldRed)
}

func (l *Logger) Fatalf(format string, msg ...interface{}) {
	l.output(FATAL, fmt.Sprintf(format, msg...), boldRed)
}

// endregion

//endregion
