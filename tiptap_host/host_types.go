package tiptaphost

import (
	"log"
	"net"
	"os"
)

type Logger struct {
	_stdout *log.Logger
	_stderr *log.Logger
}

type TipTapHost struct {
	listener net.Listener
	logger   Logger
}

func (logger *Logger) Loginfo(msg string) {
	if logger._stdout == nil {
		logger._stdout = log.Default()
	}
	logger._stdout.Println(msg)
}

func (logger *Logger) Logerr(msg string) {
	if logger._stderr == nil {
		logger._stderr = log.Default()
		logger._stderr.SetOutput(os.Stderr)
	}
	logger._stderr.Println(msg)
}
