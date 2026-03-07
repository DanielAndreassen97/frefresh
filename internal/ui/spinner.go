package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var spinnerStyle = lipgloss.NewStyle().Foreground(AccentColor)

var frames = []string{"▱▱▱", "▰▱▱", "▰▰▱", "▰▰▰", "▰▰▱", "▰▱▱"}

// Spinner shows an animated spinner with a message.
type Spinner struct {
	message  string
	stop     chan struct{}
	done     sync.WaitGroup
	stopOnce sync.Once
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		stop:    make(chan struct{}),
	}
}

func (s *Spinner) Start() {
	s.done.Add(1)
	go func() {
		defer s.done.Done()
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Print("\r\033[K") // Clear the spinner line
				return
			default:
				frame := spinnerStyle.Render(frames[i%len(frames)])
				fmt.Printf("\r%s %s", frame, s.message)
				i++
				time.Sleep(150 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.stopOnce.Do(func() {
		close(s.stop)
	})
	s.done.Wait()
}
