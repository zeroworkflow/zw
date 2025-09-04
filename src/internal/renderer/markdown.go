package renderer

import (
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/fatih/color"
)

type MarkdownRenderer struct {
	style *chroma.Style
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
		
		// Try to highlight the code
		highlighted := r.highlightCode(code, lang)
		return r.formatCodeBlock(highlighted, lang)
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

// formatCodeBlock formats a code block with borders and language label using simple borders
func (r *MarkdownRenderer) formatCodeBlock(code, lang string) string {
	// Apply syntax highlighting first if language is specified
	highlightedCode := code
	if lang != "" {
		highlightedCode = r.highlightCode(code, lang)
	}
	
	// Split into lines for border rendering
	lines := strings.Split(highlightedCode, "\n")
	
	// Calculate max display width (accounting for ANSI codes)
	maxWidth := 0
	for _, line := range lines {
		displayWidth := r.getDisplayWidth(line)
		if displayWidth > maxWidth {
			maxWidth = displayWidth
		}
	}
	
	// Minimum width of 40, maximum of 80
	if maxWidth < 40 {
		maxWidth = 40
	}
	if maxWidth > 80 {
		maxWidth = 80
	}
	
	var result strings.Builder
	
	// Top border with language
	if lang != "" {
		langLabel := " " + lang + " "
		topBorder := "╭─" + langLabel + strings.Repeat("─", maxWidth-len(langLabel)-2) + "╮"
		result.WriteString(color.New(color.FgCyan, color.Bold).Sprint(topBorder) + "\n")
	} else {
		topBorder := "╭" + strings.Repeat("─", maxWidth) + "╮"
		result.WriteString(color.New(color.FgCyan, color.Bold).Sprint(topBorder) + "\n")
	}
	
	// Content lines
	for _, line := range lines {
		displayWidth := r.getDisplayWidth(line)
		padding := maxWidth - displayWidth
		if padding < 0 {
			padding = 0
		}
		
		result.WriteString(color.New(color.FgCyan).Sprint("│ ") + line + 
			strings.Repeat(" ", padding) + color.New(color.FgCyan).Sprint(" │") + "\n")
	}
	
	// Bottom border
	bottomBorder := "╰" + strings.Repeat("─", maxWidth) + "╯"
	result.WriteString(color.New(color.FgCyan, color.Bold).Sprint(bottomBorder) + "\n")
	
	return result.String()
}

// getDisplayWidth calculates the display width of a string, ignoring ANSI escape sequences
func (r *MarkdownRenderer) getDisplayWidth(text string) int {
	// Remove ANSI escape sequences for width calculation
	cleaned := r.stripAnsiCodes(text)
	return len(cleaned)
}

// stripAnsiCodes removes ANSI escape sequences from text
func (r *MarkdownRenderer) stripAnsiCodes(text string) string {
	// Regex to match ANSI escape sequences
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
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
