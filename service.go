package main

import (
	"context"
	"fmt"
	"log"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

// RunUpdaterService implements svc.Handler for operating as a Windows service
type RunUpdaterService struct {
	updater    *RunUpdater
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewRunUpdaterService creates a new instance of the Windows service handler
func NewRunUpdaterService(config Config) *RunUpdaterService {
	ctx, cancel := context.WithCancel(context.Background())
	return &RunUpdaterService{
		updater:    NewRunUpdater(config),
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

// Execute implements svc.Handler
func (s *RunUpdaterService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	changes <- svc.Status{State: svc.StartPending}

	// Start the updater
	err := s.updater.Start(s.ctx)
	if err != nil {
		log.Printf("Error starting updater: %v", err)
		return true, 1
	}

	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	// Wait for service control events
	for {
		c := <-r
		switch c.Cmd {
		case svc.Interrogate:
			changes <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			changes <- svc.Status{State: svc.StopPending}
			s.cancelFunc()
			s.updater.Stop()
			return
		default:
			log.Printf("Unexpected control request #%d", c)
		}
	}
}

// RunAsService starts the program as a Windows service
func RunAsService(name string, isDebug bool, config Config) {
	var err error
	var elog debug.Log

	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			log.Fatalf("Failed to open eventlog: %v", err)
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("Starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, NewRunUpdaterService(config))
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
}
