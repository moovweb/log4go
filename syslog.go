// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"os"
	"fmt"
	"net"
)

// This log writer sends output to a socket
type SysLogWriter chan *LogRecord

// This is the SocketLogWriter's output method
func (w SysLogWriter) LogWrite(rec *LogRecord) {
	w <- rec
}

func (w SysLogWriter) Close() {
	close(w)
}

func connectSyslogDaemon() (sock net.Conn, err os.Error) {
	logTypes := []string{"unixgram", "unix"}
	logPaths := []string{"/dev/log", "/var/run/syslog"}
	var raddr string
	for _, network := range logTypes {
		for _, path := range logPaths {
			raddr = path
			sock, err = net.Dial(network, raddr)
			if err != nil {
				continue
			} else {
				return
			}
		}
	}
	if err != nil {
		err = os.NewError("cannot connect to Syslog Daemon")
	}
	return
}

func NewSysLogWriter() (w SysLogWriter) {
	sock, err := connectSyslogDaemon()
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewSysLogWriter: %s\n", err)
		return
	}
	w = SysLogWriter(make(chan *LogRecord, LogBufferLength))
	go func() {
		defer func() {
			if sock != nil {
				sock.Close()
			}
		}()
		var timestr string
		var timestrAt int64
		for rec := range w {
			if rec.Created != timestrAt {
				tm := TimeConversionFunction(rec.Created / 1e9)
				timestr, timestrAt = fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d %s", tm.Year, tm.Month, tm.Day, tm.Hour, tm.Minute, tm.Second, tm.Zone), rec.Created/1e9
			}
			fmt.Fprint(sock, rec.Prefix, ":", levelStrings[rec.Level], " ", timestr, " ", rec.Prefix, ": ", rec.Message, "\n")
		}
	}()
	return
}
