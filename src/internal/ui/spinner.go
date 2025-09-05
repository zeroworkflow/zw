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
	lastLen  int
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
		s.clearTopRight()
	}
}

func (s *RightSpinner) animate() {
	frameIndex := 0
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			s.doneChan <- true
			return
		case <-ticker.C:
			s.drawTopRight(s.text, s.frames[frameIndex%len(s.frames)])
			frameIndex++
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

func (s *RightSpinner) getTerminalHeight() int {
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 24
	}
	return height
}

func (s *RightSpinner) drawTopRight(label, frame string) {
	width := s.getTerminalWidth()
	content := fmt.Sprintf("%s [%s]", label, frame)
	col := width - len(content) + 1
	if col < 1 {
		col = 1
	}

	// Save cursor, move to row 1, col, draw, restore
	fmt.Print("\x1b7")
	fmt.Printf("\x1b[%d;%dH", 1, col)
	fmt.Print(content)
	// Track last length to clear on Stop
	s.lastLen = len(content)
	fmt.Print("\x1b8")
}

func (s *RightSpinner) clearTopRight() {
	width := s.getTerminalWidth()
	col := width - s.lastLen + 1
	if col < 1 {
		col = 1
	}
	fmt.Print("\x1b7")
	fmt.Printf("\x1b[%d;%dH", 1, col)
	fmt.Print(strings.Repeat(" ", s.lastLen))
	fmt.Print("\x1b8")
}

func (s *RightSpinner) setScrollRegion() {
	rows := s.getTerminalHeight()
	if rows < 2 {
		return
	}
	// Reserve top line (row 1) outside of the scroll region so it never scrolls
	fmt.Printf("\x1b[%d;%dr", 2, rows)
}

func (s *RightSpinner) resetScrollRegion() {
	// Reset scroll region to full screen
	fmt.Print("\x1b[r")
}
