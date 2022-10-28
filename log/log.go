package log

import (
	"fmt"
	"os/exec"

	syslog "log"
)

var queue = make(chan string, 8)

func init() {
	go say()
}

func Default() *syslog.Logger {
	return syslog.Default()
}

func Sayf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)

	queue <- s
}

func Debugf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)

	syslog.Printf("%-5v  %v", "DEBUG", s)
}

func Infof(format string, args ...any) {
	s := fmt.Sprintf(format, args...)

	syslog.Printf("%-5v  %v", "INFO", s)
}

func Warnf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)

	syslog.Printf("%-5v  %v", "WARN", s)
}

func Fatalf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)

	syslog.Printf("%-5v  %v", "FATAL", s)
	panic(s)
}

func say() {
	for {
		s := <-queue
		exec.Command("say", s).Run()
	}
}
