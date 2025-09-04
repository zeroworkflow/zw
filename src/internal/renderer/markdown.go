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

// formatCodeBlock formats a code block with borders and language label
func (r *MarkdownRenderer) formatCodeBlock(code, lang string) string {
	var result strings.Builder
	
	// Clean up code - remove ANSI escape sequences if any
	cleanCode := r.stripAnsiCodes(code)
	lines := strings.Split(cleanCode, "\n")
	
	// Calculate max line length for proper border sizing
	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}
	
	// Minimum width of 40, maximum of 80
	borderWidth := maxLen + 4
	if borderWidth < 44 {
		borderWidth = 44
	}
	if borderWidth > 84 {
		borderWidth = 84
	}
	
	borderColor := color.New(color.FgCyan, color.Bold)
	
	// Top border with language label
	topBorder := "┌"
	if lang != "" {
		langLabel := " " + lang + " "
		topBorder += "─" + langLabel + strings.Repeat("─", borderWidth-len(langLabel)-3) + "┐"
	} else {
		topBorder += strings.Repeat("─", borderWidth-2) + "┐"
	}
	
	result.WriteString(borderColor.Sprint(topBorder) + "\n")
	
	// Code content with side borders
	contentWidth := borderWidth - 4 // Account for "│ " and " │"
	for _, line := range lines {
		// Truncate if too long
		if len(line) > contentWidth {
			line = line[:contentWidth-3] + "..."
		}
		
		// Pad to content width
		paddedLine := line + strings.Repeat(" ", contentWidth-len(line))
		result.WriteString(borderColor.Sprint("│ ") + paddedLine + borderColor.Sprint(" │") + "\n")
	}
	
	// Bottom border
	bottomBorder := "└" + strings.Repeat("─", borderWidth-2) + "┘"
	result.WriteString(borderColor.Sprint(bottomBorder) + "\n")
	
	return result.String()
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
