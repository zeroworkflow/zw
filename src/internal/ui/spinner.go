package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

type RightSpinner struct {
	frames   []string
	text     string
	active   bool
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	lastLen  int
}

func NewRightSpinner(text string) *RightSpinner {
	ctx, cancel := context.WithCancel(context.Background())
	return &RightSpinner{
		frames: []string{"|", "/", "-", "\\"},
		text:   text,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *RightSpinner) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.active {
		return // Already running
	}
	
	s.active = true
	s.wg.Add(1)
	go s.animate()
}

func (s *RightSpinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return // Already stopped
	}
	
	s.active = false
	s.cancel()
	s.mu.Unlock()
	
	// Wait for goroutine to finish
	s.wg.Wait()
	s.clearTopRight()
}

func (s *RightSpinner) animate() {
	defer s.wg.Done()
	
	frameIndex := 0
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			if !s.active {
				s.mu.RUnlock()
				return
			}
			s.drawTopRight(s.text, s.frames[frameIndex%len(s.frames)])
			s.mu.RUnlock()
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
