# `zw ask` Command Documentation

## Overview

The `zw ask` command provides an interactive AI assistant with rich terminal formatting, syntax highlighting, and file context support.

## Usage

```bash
zw ask [question] [flags]
```

## Examples

### Basic Usage
```bash
# Ask a simple question
zw ask "How to create a REST API in Go?"

# Ask about best practices
zw ask "What are Go coding best practices?"
```

### File Context
```bash
# Include a single file for analysis
zw ask "Review this code" --file main.go
zw ask "Explain this function" -f utils.go

# Include multiple files
zw ask "How can I improve this?" -f main.go -f config.go
```

### Interactive Mode
```bash
# Start interactive conversation
zw ask -i

# Interactive mode with initial question
zw ask -i "Help me debug this issue"
```

## Flags

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--file` | `-f` | Include file content in the request | `-f main.go` |
| `--interactive` | `-i` | Start interactive conversation mode | `-i` |
| `--help` | `-h` | Show help information | `-h` |

## Features

### Syntax Highlighting
- Automatic language detection for code blocks
- Support for 100+ programming languages
- Terminal-optimized color schemes

### Markdown Rendering
- Rich text formatting in terminal
- Code block highlighting
- List and table rendering
- Link formatting

### File Context Support
- Safe file reading with size limits (max 1MB per file)
- Automatic encoding detection
- Binary file protection
- Multiple file support

### Interactive Mode
- Persistent conversation context
- Command history
- Easy exit with `quit`, `exit`, or `Ctrl+C`

## Configuration

The AI assistant requires a Z.ai API token. Configure it using:
   - Free keys [zw-free keys](https://github.com/zeroworkflow/zw-keys)

1. **Environment variable:**
   ```bash
   export ZAI_TOKEN="your-token-here"
   ```

2. **`.env` file:**
   ```
   ZAI_TOKEN=your-token-here
   ```

## Error Handling

Common errors and solutions:

### Missing API Token
```
Error: ZAI_TOKEN not found
```
**Solution:** Set your Z.ai API token in environment or `.env` file
   - Free keys [zw-free keys](https://github.com/zeroworkflow/zw-keys)

### File Not Found
```
Error: file not found: example.go
```
**Solution:** Check file path and permissions

### File Too Large
```
Error: file size exceeds limit (1MB)
```
**Solution:** Use smaller files or split large files

## Tips

1. **Use descriptive questions** - The more context you provide, the better the AI response
2. **Include relevant files** - Add source files when asking about code
3. **Interactive mode for debugging** - Use `-i` for back-and-forth troubleshooting
4. **Combine with other tools** - Pipe output or use in scripts

## Examples by Use Case

### Code Review
```bash
zw ask "Review this code for security issues" -f auth.go -f middleware.go
```

### Learning
```bash
zw ask "Explain how this algorithm works" -f sorting.go
```

### Debugging
```bash
zw ask -i "Help me debug this panic"
# Then paste error logs in interactive mode
```

### Architecture
```bash
zw ask "How should I structure this microservice?" -f main.go -f config.go
```
