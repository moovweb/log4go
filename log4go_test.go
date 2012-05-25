// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
	"time"
)

const testLogFile = "_logtest.log"

var (
	now, _ = time.Parse("02 Jan 2006 15:04:05", "13 Feb 2009 23:31:30")
)

func newLogRecord(lvl LogLevel, src string, msg string) *LogRecord {
	return &LogRecord{
		Level:   lvl,
		Source:  src,
		Created: now,
		Message: msg,
	}
}

func TestELog(t *testing.T) {
	fmt.Printf("Testing %s\n", L4G_VERSION)
	lr := newLogRecord(CRITICAL, "log4go_test", "message")
	if lr.Level != CRITICAL {
		t.Errorf("Incorrect level: %d should be %d", lr.Level, CRITICAL)
	}
	if lr.Source != "log4go_test" {
		t.Errorf("Incorrect source: %s should be %s", lr.Source, "log4go_test")
	}
	if lr.Message != "message" {
		t.Errorf("Incorrect message: %s should be %s", lr.Source, "message")
	}
}

var formatTests = []struct {
	Test    string
	Record  *LogRecord
	Formats map[string]string
}{
	{
		Test: "Standard formats",
		Record: &LogRecord{
			Level:   ERROR,
			Source:  "log4go_test",
			Message: "message",
			Created: now,
		},
		Formats: map[string]string{
			// TODO(kevlar): How can I do this so it'll work outside of PST?
			FORMAT_DEFAULT: "[2009/02/13 23:31:30 UTC] [EROR] (log4go_test) message\n",
			FORMAT_SHORT:   "[23:31 02/13/09] [EROR] message\n",
			FORMAT_ABBREV:  "[EROR] message\n",
		},
	},
}

func TestFormatLogRecord(t *testing.T) {
	for _, test := range formatTests {
		name := test.Test
		for fmt, want := range test.Formats {
			if got := FormatLogRecord(fmt, test.Record); got != want {
				t.Errorf("%s - %s:", name, fmt)
				t.Errorf("   got %q", got)
				t.Errorf("  want %q", want)
			}
		}
	}
}

var logRecordWriteTests = []struct {
	Test    string
	Record  *LogRecord
	Console string
}{
	{
		Test: "Normal message",
		Record: &LogRecord{
			Level:   CRITICAL,
			Source:  "log4go_test",
			Message: "message",
			Created: now,
		},
		Console: "CRIT Fri, 13 Feb 2009 23:31:30 UTC : message\n",
	},
}

func TestConsoleLogWriter(t *testing.T) {
	console := make(ConsoleLogWriter)

	r, w := io.Pipe()
	go console.run(w)
	defer console.Close()

	buf := make([]byte, 1024)

	for _, test := range logRecordWriteTests {
		name := test.Test

		console.LogWrite(test.Record)
		n, _ := r.Read(buf)

		if got, want := string(buf[:n]), test.Console; got != want {
			t.Errorf("%s:  got %q", name, got)
			t.Errorf("%s: want %q", name, want)
		}
	}
}

func TestFileLogWriter(t *testing.T) {
	defer func(buflen int) {
		LogBufferLength = buflen
	}(LogBufferLength)
	LogBufferLength = 0

	w := NewFileLogWriter(testLogFile, false)
	if w == nil {
		t.Fatalf("Invalid return: w should not be nil")
	}
	defer os.Remove(testLogFile)

	w.LogWrite(newLogRecord(CRITICAL, "log4go_test", "message"))
	w.Close()
	runtime.Gosched()

	if contents, err := ioutil.ReadFile(testLogFile); err != nil {
		t.Errorf("read(%q): %s", testLogFile, err)
	} else if len(contents) != 55 {
		t.Errorf("malformed filelog: %q (%d bytes)", string(contents), len(contents))
	}
}

func TestSysLog(t *testing.T) {
	w := NewSysLogWriter(LOCAL4)
	if w == nil {
		t.Fatalf("Invalid return: w should not be nil")
	}

	sl := make(Logger)
	sl.AddFilter("stdout", DEBUG, w)
	sl.Log(INFO, "TestSysLog", "This message is level INFO")
	sl.Debug("This message is level %s", DEBUG)
	w.Close()
	runtime.Gosched()
}

func TestSysLogWriter(t *testing.T) {
	defer func(buflen int) {
		LogBufferLength = buflen
	}(LogBufferLength)
	LogBufferLength = 0

	w := NewSysLogWriter(LOCAL4)
	if w == nil {
		t.Fatalf("Invalid return: w should not be nil")
	}

	w.LogWrite(newLogRecord(CRITICAL, "TestSysLogWriter", "message"))
	w.Close()
	runtime.Gosched()
}

// Skipping this test.
func testXMLLogWriter(t *testing.T) {
	defer func(buflen int) {
		LogBufferLength = buflen
	}(LogBufferLength)
	LogBufferLength = 0

	w := NewXMLLogWriter(testLogFile, false)
	if w == nil {
		t.Fatalf("Invalid return: w should not be nil")
	}
	defer os.Remove(testLogFile)

	w.LogWrite(newLogRecord(CRITICAL, "log4go_test", "message"))
	w.Close()
	runtime.Gosched()

	if contents, err := ioutil.ReadFile(testLogFile); err != nil {
		t.Errorf("read(%q): %s", testLogFile, err)
	} else if len(contents) != 190 {
		t.Errorf("malformed xmllog: %q (%d bytes)", string(contents), len(contents))
	}
}

func TestLogger(t *testing.T) {
	sl := NewDefaultLogger(WARNING)
	if sl == nil {
		t.Fatalf("NewDefaultLogger should never return nil")
	}
	if lw, exist := sl["stdout"]; lw == nil || exist != true {
		t.Fatalf("NewDefaultLogger produced invalid logger (DNE or nil)")
	}
	if sl["stdout"].Level != WARNING {
		t.Fatalf("NewDefaultLogger produced invalid logger (incorrect level)")
	}
	if len(sl) != 1 {
		t.Fatalf("NewDefaultLogger produced invalid logger (incorrect map count)")
	}

	//func (l *Logger) AddFilter(name string, level int, writer LogWriter) {}
	l := make(Logger)
	l.AddFilter("stdout", DEBUG, NewConsoleLogWriter())
	if lw, exist := l["stdout"]; lw == nil || exist != true {
		t.Fatalf("AddFilter produced invalid logger (DNE or nil)")
	}
	if l["stdout"].Level != DEBUG {
		t.Fatalf("AddFilter produced invalid logger (incorrect level)")
	}
	if len(l) != 1 {
		t.Fatalf("AddFilter produced invalid logger (incorrect map count)")
	}

	//func (l *Logger) Warn(format string, args ...interface{}) os.Error {}
	if err := l.Warn("%s %d %#v", "Warning:", 1, []int{}); err.Error() != "Warning: 1 []int{}" {
		t.Errorf("Warn returned invalid error: %s", err)
	}

	//func (l *Logger) Error(format string, args ...interface{}) os.Error {}
	if err := l.Error("%s %d %#v", "Error:", 10, []string{}); err.Error() != "Error: 10 []string{}" {
		t.Errorf("Error returned invalid error: %s", err)
	}

	//func (l *Logger) Critical(format string, args ...interface{}) os.Error {}
	if err := l.Critical("%s %d %#v", "Critical:", 100, []int64{}); err.Error() != "Critical: 100 []int64{}" {
		t.Errorf("Critical returned invalid error: %s", err)
	}

	// Already tested or basically untestable
	//func (l *Logger) Log(level int, source, message string) {}
	//func (l *Logger) Logf(level int, format string, args ...interface{}) {}
	//func (l *Logger) intLogf(level int, format string, args ...interface{}) string {}
	//func (l *Logger) Finest(format string, args ...interface{}) {}
	//func (l *Logger) Fine(format string, args ...interface{}) {}
	//func (l *Logger) Debug(format string, args ...interface{}) {}
	//func (l *Logger) Trace(format string, args ...interface{}) {}
	//func (l *Logger) Info(format string, args ...interface{}) {}
}

func TestLogOutput(t *testing.T) {
	const (
		expected = "85895942723382e03f559e8ddc12da20"
	)

	// Unbuffered output
	defer func(buflen int) {
		LogBufferLength = buflen
	}(LogBufferLength)
	LogBufferLength = 0

	l := make(Logger)

	// Delete and open the output log without a timestamp (for a constant md5sum)
	l.AddFilter("file", DEBUG, NewFileLogWriter(testLogFile, false).SetFormat("[%L] %M"))
	defer os.Remove(testLogFile)

	// Send some log messages
	l.Log(CRITICAL, "testsrc1", fmt.Sprintf("This message is level %d", int(CRITICAL)))
	l.Logf(ERROR, "This message is level %v", ERROR)
	l.Logf(WARNING, "This message is level %s", WARNING)
	l.Logc(INFO, func() string { return "This message is level INFO" })
	l.Notice("This message is level %d", int(NOTICE))
	l.Debug("This message is level %s", DEBUG)
	l.Alert(func() string { return fmt.Sprintf("This message is level %v", ALERT) })
	l.Emergency("This message is level %v", EMERGENCY)
	l.Emergency(EMERGENCY, "is also this message's level")

	l.Close()

	contents, err := ioutil.ReadFile(testLogFile)
	if err != nil {
		t.Fatalf("Could not read output log: %s", err)
	}

	sum := md5.New()
	sum.Write(contents)
	if sumstr := hex.EncodeToString(sum.Sum(nil)); sumstr != expected {
		t.Errorf("--- Log Contents:\n%s---", string(contents))
		t.Fatalf("Checksum does not match: %s (expecting %s)", sumstr, expected)
	}
}

func TestCountMallocs(t *testing.T) {
	const N = 1	
	// Console logger
	sl := NewDefaultLogger(INFO)
	for i := 0; i < N; i++ {
		sl.Log(WARNING, "here", "This is a WARNING message")
	}
	for i := 0; i < N; i++ {
		sl.Logf(WARNING, "%s is a log message with level %s", "This", WARNING)
	}
	for i := 0; i < N; i++ {
		sl.Log(DEBUG, "here", "This is a DEBUG log message")
	}
	for i := 0; i < N; i++ {
		sl.Logf(DEBUG, "%s is a log message with level %s", "This", DEBUG)
	}
}

func BenchmarkFormatLogRecord(b *testing.B) {
	const updateEvery = 1
	interval, _ := time.ParseDuration(fmt.Sprintf("%ds", 1./updateEvery))
	rec := &LogRecord{
		Level:   CRITICAL,
		Created: now,
		Source:  "log4go_test",
		Message: "message",
	}
	for i := 0; i < b.N; i++ {
		rec.Created = rec.Created.Add(interval)
		if i%2 == 0 {
			FormatLogRecord(FORMAT_DEFAULT, rec)
		} else {
			FormatLogRecord(FORMAT_SHORT, rec)
		}
	}
}

func BenchmarkConsoleLog(b *testing.B) {
	/*
	sink, err := os.Open(os.DevNull)
	if err != nil {
		panic(err)
	}

	// Linux
	if err, _ := syscall.Dup2(sink.Fd(), syscall.Stdout); err != 0 {
		panic(os.Errno(err))
	// Mac
	if err = syscall.Dup2(int(sink.Fd()), syscall.Stdout); err != nil {
		panic(err)
	}
	*/
	stdout = ioutil.Discard
	sl := NewDefaultLogger(INFO)
	for i := 0; i < b.N; i++ {
		sl.Log(WARNING, "here", "This is a log message")
	}
}

func BenchmarkConsoleNotLogged(b *testing.B) {
	sl := NewDefaultLogger(INFO)
	for i := 0; i < b.N; i++ {
		sl.Log(DEBUG, "here", "This is a log message")
	}
}

func BenchmarkConsoleUtilLog(b *testing.B) {
	sl := NewDefaultLogger(INFO)
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
}

func BenchmarkConsoleUtilNotLog(b *testing.B) {
	sl := NewDefaultLogger(INFO)
	for i := 0; i < b.N; i++ {
		sl.Debug("%s is a log message", "This")
	}
}

func BenchmarkFileLog(b *testing.B) {
	sl := make(Logger)
	b.StopTimer()
	sl.AddFilter("file", INFO, NewFileLogWriter("benchlog.log", false))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Log(WARNING, "here", "This is a log message")
	}
	b.StopTimer()
	os.Remove("benchlog.log")
}

func BenchmarkFileNotLogged(b *testing.B) {
	sl := make(Logger)
	b.StopTimer()
	sl.AddFilter("file", INFO, NewFileLogWriter("benchlog.log", false))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Log(DEBUG, "here", "This is a log message")
	}
	b.StopTimer()
	os.Remove("benchlog.log")
}

func BenchmarkFileUtilLog(b *testing.B) {
	sl := make(Logger)
	b.StopTimer()
	sl.AddFilter("file", INFO, NewFileLogWriter("benchlog.log", false))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
	b.StopTimer()
	os.Remove("benchlog.log")
}

func BenchmarkFileUtilNotLog(b *testing.B) {
	sl := make(Logger)
	b.StopTimer()
	sl.AddFilter("file", INFO, NewFileLogWriter("benchlog.log", false))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Debug("%s is a log message", "This")
	}
	b.StopTimer()
	os.Remove("benchlog.log")
}

// Benchmark results (darwin amd64 6g)
//elog.BenchmarkConsoleLog           100000       22819 ns/op
//elog.BenchmarkConsoleNotLogged    2000000         879 ns/op
//elog.BenchmarkConsoleUtilLog        50000       34380 ns/op
//elog.BenchmarkConsoleUtilNotLog   1000000        1339 ns/op
//elog.BenchmarkFileLog              100000       26497 ns/op
//elog.BenchmarkFileNotLogged       2000000         821 ns/op
//elog.BenchmarkFileUtilLog           50000       33945 ns/op
//elog.BenchmarkFileUtilNotLog      1000000        1258 ns/op
