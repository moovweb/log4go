package main

import (
	"time"
)

import l4g "log4go.googlecode.com/svn/stable"

func main() {
	log := l4g.NewLogger()
	log.AddFilter("stdout", l4g.DEBUG, l4g.NewConsoleLogWriter())
	log.Info("The time is now: %s", time.LocalTime().Format("15:04:05 MST 2006/01/02"))
}
