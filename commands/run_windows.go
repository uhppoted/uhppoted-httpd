package commands

import (
	"log"
	"os"
	"path/filepath"

	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-lib/config"
)

var RUN = Run{
	console:       false,
	workdir:       workdir(),
	configuration: filepath.Join(workdir(), "uhppoted.conf"),
}

func (cmd *Run) Execute(args ...interface{}) error {
	log.Printf("%s service %s - %s (PID %d)\n", SERVICE, uhppote.VERSION, "Microsoft Windows", os.Getpid())

	f := func(c config.Config) error {
		return cmd.start(c)
	}

	return cmd.execute(f)
}

func (cmd *Run) start(c config.Config) error {
	return cmd.run(c)
	//     var logger *log.Logger

	//     eventlogger, err := eventlog.Open(SERVICE)
	//     if err != nil {
	//         events := filelogger.Ticker{Filename: r.logFile, MaxSize: r.logFileSize}
	//         logger = log.New(&events, "", log.Ldate|log.Ltime|log.LUTC)
	//     } else {
	//         defer eventlogger.Close()

	//         events := EventLog{eventlogger}
	//         logger = log.New(&events, SERVICE, log.Ldate|log.Ltime|log.LUTC)
	//     }

	//     logger.Printf("%s service - start\n", SERVICE)

	//     if r.console {
	//         interrupt := make(chan os.Signal, 1)

	//         signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	//         r.run(c, logger, interrupt)
	//         return nil
	//     }

	//     uhppoted := service{
	//         name:   SERVICE,
	//         conf:   c,
	//         logger: logger,
	//         cmd:    r,
	//     }

	//     logger.Printf("%s service - starting\n", SERVICE)
	//     err = svc.Run(SERVICE, &uhppoted)
	//     if err != nil {
	//         fmt.Printf("   Unable to execute ServiceManager.Run request (%v)\n", err)
	//         fmt.Println()
	//         fmt.Printf("   To run %s as a command line application, type:\n", SERVICE)
	//         fmt.Println()
	//         fmt.Printf("     > %s --console\n", SERVICE)
	//         fmt.Println()

	//         logger.Fatalf("Error executing ServiceManager.Run request: %v", err)
	//         return err
	//     }

	//     logger.Printf("%s daemon - started\n", SERVICE)
	//     return nil
}

// func (s *service) Execute(args []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (ssec bool, errno uint32) {
//     s.logger.Printf("%s service - Execute\n", SERVICE)

//     const commands = svc.AcceptStop | svc.AcceptShutdown

//     status <- svc.Status{State: svc.StartPending}

//     interrupt := make(chan os.Signal, 1)
//     var wg sync.WaitGroup

//     wg.Add(1)
//     go func() {
//         defer wg.Done()
//         s.cmd.run(s.conf, s.logger, interrupt)

//         s.logger.Printf("exit\n")
//     }()

//     status <- svc.Status{State: svc.Running, Accepts: commands}

// loop:
//     for {
//         select {
//         case c := <-r:
//             s.logger.Printf("%s service - select: %v  %v\n", SERVICE, c.Cmd, c.CurrentStatus)
//             switch c.Cmd {
//             case svc.Interrogate:
//                 s.logger.Printf("%s service - svc.Interrogate %v\n", SERVICE, c.CurrentStatus)
//                 status <- c.CurrentStatus

//             case svc.Stop:
//                 interrupt <- syscall.SIGINT
//                 s.logger.Printf("%s service- svc.Stop\n", SERVICE)
//                 break loop

//             case svc.Shutdown:
//                 interrupt <- syscall.SIGTERM
//                 s.logger.Printf("%s service - svc.Shutdown\n", SERVICE)
//                 break loop

//             default:
//                 s.logger.Printf("%s service - svc.????? (%v)\n", SERVICE, c.Cmd)
//             }
//         }
//     }

//     s.logger.Printf("%s service - stopping\n", SERVICE)
//     status <- svc.Status{State: svc.StopPending}
//     wg.Wait()
//     status <- svc.Status{State: svc.Stopped}
//     s.logger.Printf("%s service - stopped\n", SERVICE)

//     return false, 0
// }

// func (e *EventLog) Write(p []byte) (int, error) {
//     err := e.log.Info(1, string(p))
//     if err != nil {
//         return 0, err
//     }

//     return len(p), nil
// }
