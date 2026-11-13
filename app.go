package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
)

const frameInterval = time.Second / 20

func run(ctx context.Context, screen tcell.Screen, population population) error {
	defer screen.Fini()
	if err := screen.Init(); err != nil {
		return fmt.Errorf("initialize terminal screen: %w", err)
	}
	screen.SetStyle(tcell.StyleDefault)
	screen.HideCursor()

	events := make(chan tcell.Event, 16)
	quit := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		screen.ChannelEvents(events, quit)
	}()
	defer func() {
		close(quit)
		<-done
	}()

	width, height := screen.Size()
	tank := newTank(width, height, rand.New(rand.NewSource(time.Now().UnixNano())), population)
	tank.draw(screen)
	screen.Show()

	ticker := time.NewTicker(frameInterval)
	defer ticker.Stop()
	last := time.Now()
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-events:
			if !ok {
				return errors.New("terminal event stream closed unexpectedly")
			}
			exit, err := handleEvent(screen, tank, event)
			if err != nil || exit {
				return err
			}
			if _, resized := event.(*tcell.EventResize); resized {
				last = time.Now()
			}
		case <-ticker.C:
			now := time.Now()
			tank.update(now.Sub(last).Seconds())
			last = now
			tank.draw(screen)
			screen.Show()
		}
	}
}

func handleEvent(screen tcell.Screen, tank *tank, event tcell.Event) (bool, error) {
	switch event := event.(type) {
	case *tcell.EventKey:
		return event.Key() == tcell.KeyEscape ||
			event.Key() == tcell.KeyCtrlC ||
			(event.Key() == tcell.KeyRune && event.Rune() == 'q'), nil
	case *tcell.EventResize:
		screen.Sync()
		width, height := screen.Size()
		tank.resize(width, height)
		tank.draw(screen)
		screen.Show()
	case *tcell.EventError:
		return false, fmt.Errorf("terminal event: %w", event)
	}
	return false, nil
}
