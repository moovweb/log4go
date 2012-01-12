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
		for rec := range w {
			var timestr string
			//var timestrAt int64

			//if rec.Created != timestrAt {
			tm := TimeConversionFunction(rec.Created / 1e9)
			//timestr, timestrAt = tm.Format("01/02/06 15:04:01"), rec.Created/1e9
			timestr = tm.Format("01/02/06 15:04:01")
			//}
			fmt.Printf("[%v] [%v] %v --------  !!!!!\n", timestr, levelStrings[rec.Level], rec.Message)
			fmt.Fprintf(sock, "%v [%v] %v !!\n", timestr, levelStrings[rec.Level], rec.Message)
			//}
			
		}
	}()
	return
}
