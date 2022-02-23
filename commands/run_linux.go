package commands

import (
	"log"
	"os"

	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-lib/config"
)

var RUN = Run{
	console:       false,
	workdir:       "/var/uhppoted",
	configuration: "/etc/uhppoted/uhppoted.conf",
}

func (cmd *Run) Execute(args ...interface{}) error {
	log.Printf("%s service %s - %s (PID %d)\n", SERVICE, uhppote.VERSION, "Linux", os.Getpid())

	f := func(c config.Config) error {
		return cmd.exec(c)
	}

	return cmd.execute(f)
}

func (cmd *Run) exec(c config.Config) error {
	return cmd.run(c)

	// logger := log.New(os.Stdout, "", log.LstdFlags)
	// interrupt := make(chan os.Signal, 1)
	//
	// signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	//
	// if !r.console {
	// 	events := eventlog.Ticker{Filename: r.logFile, MaxSize: r.logFileSize}
	// 	logger = log.New(&events, "", log.Ldate|log.Ltime|log.LUTC)
	// 	rotate := make(chan os.Signal, 1)
	//
	// 	signal.Notify(rotate, syscall.SIGHUP)
	//
	// 	go func() {
	// 		for {
	// 			<-rotate
	// 			log.Printf("Rotating %s log file '%s'\n", SERVICE, r.logFile)
	// 			events.Rotate()
	// 		}
	// 	}()
	// }

	// r.run(c, logger, interrupt)
	//
	// return nil
}
