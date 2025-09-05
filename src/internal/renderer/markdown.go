package renderer

import (
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
)

type MarkdownRenderer struct {
	style      *chroma.Style
	regexCache *RegexCache
}

// processOpenCodeBlock detects a trailing, not-yet-closed code block and renders it as a box
// so that users can see syntax highlighting during streaming before the closing ``` arrives.
func (r *MarkdownRenderer) processOpenCodeBlock(content string) string {
	// Quick check: odd number of ``` suggests an open block
	if strings.Count(content, "```")%2 == 0 {
		return content
	}
	// Find the last opening ``` position
	idx := strings.LastIndex(content, "```")
	if idx == -1 {
		return content
	}

	// Extract language line after ``` up to newline
	rest := content[idx+3:]
	newline := strings.IndexByte(rest, '\n')
	if newline == -1 {
		// No newline after language, cannot determine code yet
		return content
	}
	lang := strings.TrimSpace(rest[:newline])
	code := rest[newline+1:]

	// Render everything before the open block as usual, and the open block as a formatted box
	before := content[:idx]
	rendered := r.formatCodeBlock(code, lang)
	return before + rendered
}

// HighlightCode exposes syntax highlighting to be used by streaming renderer
func (r *MarkdownRenderer) HighlightCode(code, lang string) string {
	return r.highlightCode(code, lang)
}

func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{
		style:      styles.Get("github"),
		regexCache: NewRegexCache(),
	}
}

// RenderMarkdown renders markdown with syntax highlighting for code blocks
func (r *MarkdownRenderer) RenderMarkdown(content string) string {
	// Process code blocks first
	content = r.processCodeBlocks(content)
	// Handle an open (not yet closed) trailing code block for streaming
	content = r.processOpenCodeBlock(content)
	
	// Process other markdown elements
	content = r.processMarkdownElements(content)
	
	return content
}

// processCodeBlocks finds and syntax highlights code blocks
func (r *MarkdownRenderer) processCodeBlocks(content string) string {
	// Use cached regex for better performance
	codeBlockRegex := r.regexCache.Get("```([a-zA-Z0-9_+-]*)\\n([\\s\\S]*?)\\n?```")
	
	return codeBlockRegex.ReplaceAllStringFunc(content, func(match string) string {
		parts := codeBlockRegex.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		
		lang := strings.TrimSpace(parts[1])
		code := parts[2]
		
		// If no language specified, return as-is with basic formatting
		if lang == "" {
			return r.formatCodeBlock(code, "")
		}
		
		// Pass raw code to formatter; it will apply highlighting when lang is provided
		return r.formatCodeBlock(code, lang)
	})
}

// highlightCode applies syntax highlighting to code
func (r *MarkdownRenderer) highlightCode(code, lang string) string {
	// Get lexer for the language
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)
	
	// Get terminal formatter
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		return code // Fallback to plain text
	}
	
	// Tokenize and format
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}
	
	var buf strings.Builder
	err = formatter.Format(&buf, r.style, iterator)
	if err != nil {
		return code
	}
	
	return buf.String()
}

// formatCodeBlock formats a code block using go-pretty table for perfect alignment
func (r *MarkdownRenderer) formatCodeBlock(code, lang string) string {
	// Apply syntax highlighting first if language is specified
	highlightedCode := code
	if lang != "" {
		highlightedCode = r.highlightCode(code, lang)
		
		// Special handling for Python - clean up complex ANSI sequences
		if strings.ToLower(lang) == "python" || strings.ToLower(lang) == "py" {
			highlightedCode = r.cleanPythonAnsiCodes(highlightedCode)
		}
	}
	
	// Normalize tabs to 4 spaces
	highlightedCode = strings.ReplaceAll(highlightedCode, "\t", "    ")
	
	// Remove trailing empty lines
	lines := strings.Split(highlightedCode, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	
	if len(lines) == 0 {
		return ""
	}
	
	// Detect shell and configure text width calculation
	r.configureTextWidth()
	
	// Create table with single column for code content
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	
	// Set fixed row length to prevent width calculation issues
	termWidth := r.getTerminalWidth()
	if termWidth > 0 {
		// Use SetAllowedRowLength for precise control
		allowedLength := termWidth - 10
		if r.isZsh() {
			allowedLength -= 6 // Extra conservative for zsh
		}
		if allowedLength > 30 {
			t.SetAllowedRowLength(allowedLength)
		}
		
		// Also set column config as fallback
		maxWidth := allowedLength - 4
		if maxWidth > 20 {
			t.SetColumnConfigs([]table.ColumnConfig{
				{Number: 1, WidthMax: maxWidth, WidthMin: 20},
			})
		}
	}
	
	// Add language header if provided
	if lang != "" {
		t.AppendHeader(table.Row{" " + lang + " "})
	}
	
	// Add each line as a row
	for _, line := range lines {
		t.AppendRow(table.Row{line})
	}
	
	return t.Render() + "\n"
}



// isZsh detects if running under zsh shell
func (r *MarkdownRenderer) isZsh() bool {
	shell := os.Getenv("SHELL")
	zshVersion := os.Getenv("ZSH_VERSION")
	return strings.Contains(shell, "zsh") || zshVersion != ""
}

// configureTextWidth configures text width calculation for different shells
func (r *MarkdownRenderer) configureTextWidth() {
	// Configure text width handling based on shell
	// This is handled by go-pretty internally
}

// cleanPythonAnsiCodes cleans up complex ANSI sequences specific to Python highlighting
func (r *MarkdownRenderer) cleanPythonAnsiCodes(text string) string {
	// Remove complex nested ANSI sequences that can cause width calculation issues
	complexAnsiRegex := r.regexCache.Get(`\x1b\[\d+;\d+;\d+m`)
	text = complexAnsiRegex.ReplaceAllString(text, "")
	
	// Normalize remaining ANSI codes to simpler forms
	simpleAnsiRegex := r.regexCache.Get(`\x1b\[(\d+)m`)
	text = simpleAnsiRegex.ReplaceAllStringFunc(text, func(match string) string {
		// Keep only basic color codes
		if strings.Contains(match, "0m") || // reset
		   strings.Contains(match, "1m") || // bold
		   strings.Contains(match, "3") ||  // foreground colors 30-37
		   strings.Contains(match, "9") {   // bright colors 90-97
			return match
		}
		return "" // Remove other codes
	})
	
	return text
}

// getTerminalWidth returns current terminal width or a sensible default
func (r *MarkdownRenderer) getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80
	}
	return width
}

// processMarkdownElements processes other markdown elements like bold, italic, etc.
func (r *MarkdownRenderer) processMarkdownElements(content string) string {
	// Bold text **text** or __text__
	boldRegex := r.regexCache.Get(`\*\*([^*]+)\*\*|__([^_]+)__`)
	content = boldRegex.ReplaceAllStringFunc(content, func(match string) string {
		text := strings.Trim(match, "*_")
		return color.New(color.Bold).Sprint(text)
	})
	
	// Italic text *text* or _text_
	italicRegex := r.regexCache.Get(`\*([^*]+)\*|_([^_]+)_`)
	content = italicRegex.ReplaceAllStringFunc(content, func(match string) string {
		text := strings.Trim(match, "*_")
		return color.New(color.Italic).Sprint(text)
	})
	
	// Inline code `code`
	inlineCodeRegex := r.regexCache.Get("`([^`]+)`")
	content = inlineCodeRegex.ReplaceAllStringFunc(content, func(match string) string {
		code := strings.Trim(match, "`")
		return color.New(color.BgHiBlack, color.FgWhite).Sprint(" " + code + " ")
	})
	
	// Headers
	headerRegex := r.regexCache.Get(`^(#{1,6})\s+(.+)$`)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if matches := headerRegex.FindStringSubmatch(line); matches != nil {
			level := len(matches[1])
			text := matches[2]
			
			var headerColor *color.Color
			switch level {
			case 1:
				headerColor = color.New(color.FgRed, color.Bold)
			case 2:
				headerColor = color.New(color.FgYellow, color.Bold)
			case 3:
				headerColor = color.New(color.FgGreen, color.Bold)
			default:
				headerColor = color.New(color.FgCyan, color.Bold)
			}
			
			lines[i] = headerColor.Sprint(text)
		}
	}
	content = strings.Join(lines, "\n")
	
	return content
}
