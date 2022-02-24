package commands

import (
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/eventlog"

	"path/filepath"
)

type EventLog struct {
	log *eventlog.Log
}

func (e *EventLog) Write(p []byte) (int, error) {
	err := e.log.Info(1, string(p))
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func workdir() string {
	programData, err := windows.KnownFolderPath(windows.FOLDERID_ProgramData, windows.KF_FLAG_DEFAULT)
	if err != nil {
		return `C:\uhppoted`
	}

	return filepath.Join(programData, "uhppoted")
}
