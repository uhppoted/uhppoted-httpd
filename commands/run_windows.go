package commands

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"golang.org/x/sys/windows/svc"
	syslog "golang.org/x/sys/windows/svc/eventlog"

	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-lib/config"
	"github.com/uhppoted/uhppoted-lib/eventlog"
)

var RUN = Run{
	console:       false,
	workdir:       workdir(),
	configuration: filepath.Join(workdir(), "uhppoted.conf"),
	logFile:       filepath.Join(workdir(), "logs", fmt.Sprintf("%s.log", SERVICE)),
	logFileSize:   10,
}

type service struct {
	name string
	conf config.Config
	cmd  *Run
}

func (cmd *Run) Execute(args ...interface{}) error {
	log.Printf("%s service %s - %s (PID %d)\n", SERVICE, uhppote.VERSION, "Microsoft Windows", os.Getpid())

	f := func(c config.Config) {
		cmd.start(c)
	}

	return cmd.execute(f)
}

func (cmd *Run) start(conf config.Config) {
	if cmd.console {
		log.SetOutput(os.Stdout)
		log.SetFlags(log.LstdFlags)

		interrupt := make(chan os.Signal, 1)

		signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
		cmd.run(conf, interrupt)
		return
	}

	if eventlogger, err := syslog.Open(SERVICE); err != nil {
		events := eventlog.Ticker{Filename: cmd.logFile, MaxSize: cmd.logFileSize}

		log.SetOutput(&events)
	} else {
		defer eventlogger.Close()

		log.SetOutput(&EventLog{eventlogger})
	}

	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	log.Printf("%s service - start\n", SERVICE)

	uhppoted := service{
		name: SERVICE,
		conf: conf,
		cmd:  cmd,
	}

	log.Printf("%s service - starting\n", SERVICE)

	if err := svc.Run(SERVICE, &uhppoted); err != nil {
		fmt.Printf("   Unable to execute ServiceManager.Run request (%v)\n", err)
		fmt.Println()
		fmt.Printf("   To run %s as a command line application, type:\n", SERVICE)
		fmt.Println()
		fmt.Printf("     > %s --console\n", SERVICE)
		fmt.Println()

		log.Printf("   Unable to execute ServiceManager.Run request (%v)\n", err)
		log.Println()
		log.Printf("   To run %s as a command line application, type:\n", SERVICE)
		log.Println()
		log.Printf("     > %s --console\n", SERVICE)
		log.Println()

		log.Panicf("Error executing ServiceManager.Run request: %v", err)

		return
	}

	log.Printf("%s daemon - started\n", SERVICE)
}

func (s *service) Execute(args []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (ssec bool, errno uint32) {
	log.Printf("%s service - Execute\n", SERVICE)

	const commands = svc.AcceptStop | svc.AcceptShutdown

	status <- svc.Status{State: svc.StartPending}

	interrupt := make(chan os.Signal, 1)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.cmd.run(s.conf, interrupt)

		log.Printf("exit\n")
	}()

	status <- svc.Status{State: svc.Running, Accepts: commands}

loop:
	for {
		select {
		case c := <-r:
			log.Printf("%s service - select: %v  %v\n", SERVICE, c.Cmd, c.CurrentStatus)
			switch c.Cmd {
			case svc.Interrogate:
				log.Printf("%s service - svc.Interrogate %v\n", SERVICE, c.CurrentStatus)
				status <- c.CurrentStatus

			case svc.Stop:
				interrupt <- syscall.SIGINT
				log.Printf("%s service- svc.Stop\n", SERVICE)
				break loop

			case svc.Shutdown:
				interrupt <- syscall.SIGTERM
				log.Printf("%s service - svc.Shutdown\n", SERVICE)
				break loop

			default:
				log.Printf("%s service - svc.????? (%v)\n", SERVICE, c.Cmd)
			}
		}
	}

	log.Printf("%s service - stopping\n", SERVICE)
	status <- svc.Status{State: svc.StopPending}
	wg.Wait()
	status <- svc.Status{State: svc.Stopped}
	log.Printf("%s service - stopped\n", SERVICE)

	return false, 0
}
