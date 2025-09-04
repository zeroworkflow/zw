package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

type RightSpinner struct {
	frames   []string
	text     string
	active   bool
	stopChan chan bool
	doneChan chan bool
}

func NewRightSpinner(text string) *RightSpinner {
	return &RightSpinner{
		frames:   []string{"|", "/", "-", "\\"},
		text:     text,
		stopChan: make(chan bool),
		doneChan: make(chan bool),
	}
}

func (s *RightSpinner) Start() {
	s.active = true
	go s.animate()
}

func (s *RightSpinner) Stop() {
	if s.active {
		s.active = false
		s.stopChan <- true
		<-s.doneChan
		fmt.Print("\r\033[K") // Clear the line
	}
}

func (s *RightSpinner) animate() {
	frameIndex := 0
	dotIndex := 0
	dots := []string{"", ".", "..", "..."}
	
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			s.doneChan <- true
			return
		case <-ticker.C:
			// Get terminal width
			width := s.getTerminalWidth()
			
			// Create left part with text and dots
			leftText := s.text + dots[dotIndex%len(dots)]
			
			// Create right part with spinner
			rightText := fmt.Sprintf("[%s]", s.frames[frameIndex%len(s.frames)])
			
			// Calculate spaces needed
			spacesNeeded := width - len(leftText) - len(rightText) - 1
			if spacesNeeded < 1 {
				spacesNeeded = 1
			}
			
			// Create the full line
			line := fmt.Sprintf("\r%s%s%s", leftText, strings.Repeat(" ", spacesNeeded), rightText)
			fmt.Print(line)
			
			frameIndex++
			if frameIndex%3 == 0 { // Change dots every 3 frame changes
				dotIndex++
			}
		}
	}
}

func (s *RightSpinner) getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80 // Default width if can't detect
	}
	return width
}
