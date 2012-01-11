include $(GOROOT)/src/Make.inc

TARG=log4go
GOFILES=\
	log4go.go\
	config.go\
	termlog.go\
	socklog.go\
	filelog.go\
	pattlog.go\
	wrapper.go\

include $(GOROOT)/src/Make.pkg
