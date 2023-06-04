package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/ooni/probe-engine/pkg/model"
)

// loggerSingleton is the logger singleton.
var loggerSingleton = &prettyLogger{}

// prettyLogger is a logger emitting colorized output.
type prettyLogger struct{}

var _ model.Logger = &prettyLogger{}

// Debug implements model.Logger.
func (pl *prettyLogger) Debug(msg string) {
	// nothing
}

// Debugf implements model.Logger.
func (pl *prettyLogger) Debugf(format string, v ...interface{}) {
	// nothing
}

// Info implements model.Logger.
func (pl *prettyLogger) Info(msg string) {
	pl.emit(color.BlueString("INFO"), msg)
}

// Infof implements model.Logger.
func (pl *prettyLogger) Infof(format string, v ...interface{}) {
	pl.emit(color.BlueString("INFO"), fmt.Sprintf(format, v...))
}

// Warn implements model.Logger.
func (pl *prettyLogger) Warn(msg string) {
	pl.emit(color.RedString("WARN"), msg)
}

// Warnf implements model.Logger.
func (pl *prettyLogger) Warnf(format string, v ...interface{}) {
	pl.emit(color.RedString("WARN"), fmt.Sprintf(format, v...))
}

func (pl *prettyLogger) emit(level, message string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", level, message)
}
