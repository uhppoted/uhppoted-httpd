package commands

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-lib/config"
	"github.com/uhppoted/uhppoted-lib/eventlog"
)

var RUN = Run{
	console:       false,
	workdir:       "/usr/local/var/com.github.uhppoted",
	configuration: "/usr/local/etc/com.github.uhppoted/uhppoted.conf",
	logFile:       fmt.Sprintf("/usr/local/var/com.github.uhppoted/logs/%s.log", SERVICE),
	logFileSize:   10,
}

func (cmd *Run) Execute(args ...interface{}) error {
	log.Printf("%s service %s - %s (PID %d)\n", SERVICE, uhppote.VERSION, "MacOS", os.Getpid())

	f := func(c config.Config) error {
		return cmd.exec(c)
	}

	return cmd.execute(f)
}

func (cmd *Run) exec(conf config.Config) error {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags)

	interrupt := make(chan os.Signal, 1)

	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	if !cmd.console {
		events := eventlog.Ticker{Filename: cmd.logFile, MaxSize: cmd.logFileSize}

		log.SetOutput(&events)
		log.SetFlags(log.Ldate | log.Ltime | log.LUTC)

		rotate := make(chan os.Signal, 1)

		signal.Notify(rotate, syscall.SIGHUP)

		go func() {
			for {
				<-rotate
				log.Printf("Rotating %s log file '%s'\n", SERVICE, cmd.logFile)
				events.Rotate()
			}
		}()
	}

	return cmd.run(conf, interrupt)
}
