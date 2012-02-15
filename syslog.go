// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"os"
	"fmt"
	"net"
	"time"
)

const (
	LOCAL0 = 16
	LOCAL1 = 17
	LOCAL2 = 18
	LOCAL3 = 19
	LOCAL4 = 20
	LOCAL5 = 21
	LOCAL6 = 22
	LOCAL7 = 23
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

func NewSysLogWriter(facility int) (w SysLogWriter) {
	offset := facility * 8
	host, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot obtain hostname: %s\n", err.String())
		host = "unknown"
	}
	sock, err := connectSyslogDaemon()
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewSysLogWriter: %s\n", err.String())
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
				timestrAt = rec.Created/1e9
				timestr = time.SecondsToUTC(timestrAt).Format(time.RFC3339)
			}
			fmt.Fprintf(sock, "<%d>%s %s %s: %s\n", offset + int(rec.Level), timestr, host, rec.Prefix, rec.Message)
		}
	}()
	return
}
