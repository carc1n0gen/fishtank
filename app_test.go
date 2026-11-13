package main

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

type observedScreen struct {
	tcell.SimulationScreen
	ready     chan struct{}
	once      sync.Once
	finalized bool
	initError error
}

func (s *observedScreen) Init() error {
	if s.initError != nil {
		return s.initError
	}
	if err := s.SimulationScreen.Init(); err != nil {
		return err
	}
	s.SetSize(80, 24)
	return nil
}

func (s *observedScreen) Show() {
	s.SimulationScreen.Show()
	s.once.Do(func() { close(s.ready) })
}

func (s *observedScreen) Fini() {
	s.SimulationScreen.Fini()
	s.finalized = true
}

func TestRunExitAndRestoration(t *testing.T) {
	tests := []struct {
		name    string
		event   tcell.Event
		cancel  bool
		wantErr string
	}{
		{name: "q", event: tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)},
		{name: "escape", event: tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)},
		{name: "ctrl-c", event: tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModNone)},
		{name: "signal context", cancel: true},
		{name: "terminal failure", event: tcell.NewEventError(errors.New("read failed")), wantErr: "read failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := &observedScreen{
				SimulationScreen: tcell.NewSimulationScreen("UTF-8"),
				ready:            make(chan struct{}),
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			result := make(chan error, 1)
			go func() { result <- run(ctx, screen, population{}) }()
			select {
			case <-screen.ready:
			case err := <-result:
				t.Fatalf("exited before first frame: %v", err)
			case <-time.After(3 * time.Second):
				t.Fatal("first frame timed out")
			}
			if tt.cancel {
				cancel()
			} else if err := screen.PostEvent(tt.event); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-result:
				if tt.wantErr == "" && err != nil {
					t.Fatal(err)
				}
				if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
					t.Fatalf("got error %v; want %q", err, tt.wantErr)
				}
				if !screen.finalized {
					t.Fatal("terminal was not finalized")
				}
			case <-time.After(3 * time.Second):
				t.Fatal("shutdown timed out")
			}
		})
	}
}

func TestRunInitializationFailure(t *testing.T) {
	screen := &observedScreen{
		SimulationScreen: tcell.NewSimulationScreen("UTF-8"),
		initError:        errors.New("no terminal"),
	}
	err := run(context.Background(), screen, population{})
	if err == nil || !strings.Contains(err.Error(), "initialize terminal screen: no terminal") {
		t.Fatalf("initialization error not reported: %v", err)
	}
	if !screen.finalized {
		t.Fatal("initialization failure did not finalize terminal")
	}
}

func TestUnrelatedKeyDoesNotQuit(t *testing.T) {
	screen := simulationScreen(t, 20, 8)
	exit, err := handleEvent(screen, testTank(20, 8), tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	if err != nil || exit {
		t.Fatalf("unrelated key returned exit=%v error=%v", exit, err)
	}
}
