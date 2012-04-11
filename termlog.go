// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"io"
	"os"
	"fmt"
	"sync"
)

var stdout io.Writer = os.Stdout
var lock sync.Mutex

// This is the standard writer that prints to standard output.
type ConsoleLogWriter struct{
	writer io.Writer
}

// This creates a new ConsoleLogWriter
func NewConsoleLogWriter() *ConsoleLogWriter {
	return new(ConsoleLogWriter)
}

// This is the ConsoleLogWriter's output method.  This will block if the output
// buffer is full.
func (w ConsoleLogWriter) LogWrite(rec *LogRecord) {
	lock.Lock()
	defer lock.Unlock()
	if w.writer == nil {
		w.writer = stdout
		println("setting the writer to stdout")
	}
	var timestr string
	var timestrAt int64

	if rec.Created != timestrAt {
		tm := TimeConversionFunction(rec.Created / 1e9)
		timestr, timestrAt = fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d %s", tm.Year, tm.Month, tm.Day, tm.Hour, tm.Minute, tm.Second, tm.Zone), rec.Created/1e9
	}
	fmt.Fprint(w.writer, levelStrings[rec.Level], " ", timestr, " ", rec.Prefix, ": ", rec.Message, "\n")
}

// Close stops the logger from sending messages to standard output.  Attempts to
// send log messages to this logger after a Close have undefined behavior.
func (w ConsoleLogWriter) Close() {
}
