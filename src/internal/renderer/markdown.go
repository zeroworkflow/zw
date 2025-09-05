package renderer

import (
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
)

type MarkdownRenderer struct {
	style *chroma.Style
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
		style: styles.Get("github"),
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
	// Regex to match code blocks with language specification
	codeBlockRegex := regexp.MustCompile("```([a-zA-Z0-9_+-]*)\n([\\s\\S]*?)\n```")
	
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

// formatCodeBlock formats a code block with borders and language label using box drawing characters
func (r *MarkdownRenderer) formatCodeBlock(code, lang string) string {
	// Apply syntax highlighting first if language is specified
	highlightedCode := code
	if lang != "" {
		highlightedCode = r.highlightCode(code, lang)
	}
	
	// Split into lines for border rendering
	lines := strings.Split(highlightedCode, "\n")
	
	// Remove empty trailing lines
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	
	if len(lines) == 0 {
		return ""
	}
	
	// Calculate max display width (accounting for ANSI codes)
	maxWidth := 0
	for _, line := range lines {
		displayWidth := r.getDisplayWidth(line)
		if displayWidth > maxWidth {
			maxWidth = displayWidth
		}
	}
	
	// Determine inner content width with better sizing
	innerWidth := maxWidth + 2 // Add some padding
	if innerWidth < 50 {
		innerWidth = 50
	}
	if innerWidth > 100 {
		innerWidth = 100
	}
	
	var result strings.Builder
	borderColor := color.New(color.FgCyan, color.Bold)
	
	// Top border with language
	if lang != "" {
		langLabel := " " + lang + " "
		dashCount := innerWidth - len(langLabel)
		if dashCount < 0 {
			dashCount = 0
		}
		topBorder := "╭─" + langLabel + strings.Repeat("─", dashCount) + "╮"
		result.WriteString(borderColor.Sprint(topBorder) + "\n")
	} else {
		topBorder := "╭" + strings.Repeat("─", innerWidth) + "╮"
		result.WriteString(borderColor.Sprint(topBorder) + "\n")
	}
	
	// Content lines with proper padding
	for _, line := range lines {
		displayWidth := r.getDisplayWidth(line)
		padding := innerWidth - displayWidth
		if padding < 0 {
			// Truncate line if too long
			line = r.truncateLine(line, innerWidth)
			padding = 0
		}
		
		leftBorder := borderColor.Sprint("│ ")
		rightBorder := borderColor.Sprint(" │")
		result.WriteString(leftBorder + line + strings.Repeat(" ", padding) + rightBorder + "\n")
	}
	
	// Bottom border
	bottomBorder := "╰" + strings.Repeat("─", innerWidth) + "╯"
	result.WriteString(borderColor.Sprint(bottomBorder) + "\n")
	
	return result.String()
}

// truncateLine truncates a line to fit within the specified width, preserving ANSI codes
func (r *MarkdownRenderer) truncateLine(text string, maxWidth int) string {
	if r.getDisplayWidth(text) <= maxWidth {
		return text
	}
	
	// Simple truncation for now - could be improved to preserve ANSI codes better
	cleaned := r.stripAnsiCodes(text)
	if len(cleaned) > maxWidth-3 {
		return cleaned[:maxWidth-3] + "..."
	}
	return cleaned
}

// getDisplayWidth calculates the display width of a string, ignoring ANSI escape sequences
func (r *MarkdownRenderer) getDisplayWidth(text string) int {
	// Remove ANSI escape sequences for width calculation
	cleaned := r.stripAnsiCodes(text)
	// Expand tabs to 4 spaces to stabilize width
	cleaned = strings.ReplaceAll(cleaned, "\t", "    ")
	// Use runewidth to account for Unicode display width
	return runewidth.StringWidth(cleaned)
}

// stripAnsiCodes removes ANSI escape sequences from text
func (r *MarkdownRenderer) stripAnsiCodes(text string) string {
	// Regex to match ANSI escape sequences including more complex ones
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)
	return ansiRegex.ReplaceAllString(text, "")
}

// processMarkdownElements processes other markdown elements like bold, italic, etc.
func (r *MarkdownRenderer) processMarkdownElements(content string) string {
	// Bold text **text** or __text__
	boldRegex := regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__`)
	content = boldRegex.ReplaceAllStringFunc(content, func(match string) string {
		text := strings.Trim(match, "*_")
		return color.New(color.Bold).Sprint(text)
	})
	
	// Italic text *text* or _text_
	italicRegex := regexp.MustCompile(`\*([^*]+)\*|_([^_]+)_`)
	content = italicRegex.ReplaceAllStringFunc(content, func(match string) string {
		text := strings.Trim(match, "*_")
		return color.New(color.Italic).Sprint(text)
	})
	
	// Inline code `code`
	inlineCodeRegex := regexp.MustCompile("`([^`]+)`")
	content = inlineCodeRegex.ReplaceAllStringFunc(content, func(match string) string {
		code := strings.Trim(match, "`")
		return color.New(color.BgHiBlack, color.FgWhite).Sprint(" " + code + " ")
	})
	
	// Headers
	headerRegex := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
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
