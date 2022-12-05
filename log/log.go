package log

import (
	"fmt"
	"os/exec"

	syslog "log"

	"github.com/uhppoted/uhppoted-lib/log"
)

var queue = make(chan string, 8)

func init() {
	go say()
}

func Default() *syslog.Logger {
	return syslog.Default()
}

func SetFatalHook(f func()) {
	log.AddFatalHook(f)
}

func Sayf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)

	queue <- s
}

func Debugf(format string, args ...any) {
	log.Debugf(format, args...)
}

func Infof(format string, args ...any) {
	log.Infof(format, args...)
}

func Warnf(format string, args ...any) {
	log.Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	log.Errorf(format, args...)
}

func Fatalf(format string, args ...any) {
	log.Fatalf(format, args...)
}

func say() {
	for {
		s := <-queue
		exec.Command("say", s).Run()
	}
}
