# ZeroWorkflow

[🇷🇺 RU](doc/lang/README.ru.md)

<img src="assets/image/logo/light_logo.png" alt="ZeroWorkflow Logo" width="310"/>

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-brightgreen?style=flat-square)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-success?style=flat-square)](https://github.com/derxanax/ZeroWorkflow)
[![Version](https://img.shields.io/badge/Version-1.0.0-blue?style=flat-square)](https://github.com/derxanax/ZeroWorkflow/releases)

> AI-powered developer tools suite for streamlined workflow automation

## 🧶 About

ZeroWorkflow is a collection of AI-powered command-line utilities designed to automate common development tasks. Built with Go for maximum performance and cross-platform compatibility.

**Key Features:**
- **AI Chat Assistant** - Interactive AI conversations with syntax highlighting
- **Modular Architecture** - Easy to extend with new commands
- **Beautiful Terminal UI** - Rich markdown rendering with code highlighting
- **Cross-Platform** - Single binary deployment

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/derxanax/ZeroWorkflow.git
cd ZeroWorkflow

# Build the binary
go build -o zw src/main.go

# Or install globally
go install
```

### Setup

1. Set your AI token in environment variables:
```bash
export AI_TOKEN="your_ai_token_here"
```

2. Or create a `.env` file:
```bash
echo "AI_TOKEN=your_ai_token_here" > .env
```

### Usage

```bash
# Ask AI a question
./zw ask "How to create a REST API in Go?"

# Interactive mode for continuous conversation
./zw ask -i

# Get help
./zw --help
./zw ask --help
```

## 🛠 Commands

### `zw ask` - AI Assistant

Interactive AI assistant with markdown rendering and syntax highlighting.

**Examples:**
```bash
# Single question
./zw ask "Explain Go interfaces"

# Interactive mode
./zw ask -i
> What are Go channels?
> How to handle errors in Go?
> exit
```

**Features:**
- ✨ Syntax highlighting for code blocks
- 📝 Rich markdown rendering
- 🔄 Interactive conversation mode
- 🎨 Beautiful terminal formatting

## 📁 Project Structure

```text
ZeroWorkflow/
├── src/                   # Source code
│   ├── cmd/               # CLI commands
│   │   ├── root.go        # Root command setup
│   │   └── ask.go         # AI assistant command
│   ├── internal/          # Internal packages
│   │   ├── ai/            # AI client implementation
│   │   │   └── client.go  # Z.ai API client
│   │   └── renderer/      # Output rendering
│   │       └── markdown.go # Markdown renderer with syntax highlighting
│   └── main.go            # Application entry point
├── assets/                # Static assets
│   └── image/logo/        # Logo files
├── doc/                   # Documentation
│   ├── lang/              # Localized documentation
│   └── ask.md             # Command documentation
├── go.mod                 # Go module definition
└── .env                   # Environment variables (create manually)
```

## 🎨 Features

### Syntax Highlighting
- **50+ Languages** supported via Chroma
- **Terminal-optimized** color schemes
- **Code block borders** with language labels
- **Inline code** highlighting

### AI Integration
- **Z.ai GLM-4.5** model integration
- **Streaming responses** for real-time output
- **Context preservation** in interactive mode
- **Error handling** with graceful fallbacks

### Terminal UI
- **Rich formatting** with colors and styles
- **Responsive design** adapts to terminal width
- **Progress indicators** during AI processing
- **Clean, modern** interface design

## 🔧 Development

### Prerequisites
- Go 1.21 or higher
- Terminal with 256-color support

### Building from Source
```bash
# Clone repository
git clone https://github.com/derxanax/ZeroWorkflow.git
cd ZeroWorkflow

# Install dependencies
go mod tidy

# Build
go build -o zw src/main.go

# Run tests
go test ./...
```

### Adding New Commands
1. Create new command file in `src/cmd/`
2. Implement command logic
3. Register with root command
4. Add documentation

## 🌐 Supported Platforms

[![Linux](https://img.shields.io/badge/Linux-FCC624?style=flat-square&logo=linux&logoColor=black)](https://www.linux.org/)
[![macOS](https://img.shields.io/badge/macOS-000000?style=flat-square&logo=apple&logoColor=white)](https://www.apple.com/macos/)
[![Windows](https://img.shields.io/badge/Windows-0078D6?style=flat-square&logo=windows&logoColor=white)](https://www.microsoft.com/windows/)

## 📚 Documentation

- [Command Reference](doc/commands.md) - Complete command documentation
- [AI Assistant Guide](doc/ask.md) - Detailed guide for the ask command
- [Configuration](doc/config.md) - Environment setup and configuration
- [Contributing](doc/contributing.md) - Development guidelines

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](doc/contributing.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Z.ai** for providing the AI API
- **Chroma** for syntax highlighting
- **Cobra** for CLI framework
- **Go Community** for excellent tooling

---

<div align="center">
  <strong>Built with ❤️ by <a href="https://github.com/derxanax">@derxanax</a></strong>
</div>