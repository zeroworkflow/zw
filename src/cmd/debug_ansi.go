package cmd

import (
	"fmt"
	"os"
	"strings"

	isatty "github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/fatih/color"
	"zero-workflow/src/internal/renderer"
)

var debugANSICmd = &cobra.Command{
	Use:   "debug-ansi",
	Short: "Проверка поддержки ANSI и подсветки синтаксиса",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ANSI Debug - старт")
		fmt.Println(strings.Repeat("-", 60))

		// 1) Базовые ANSI последовательности
		fmt.Println("1) Raw ANSI (должно быть красным):")
		raw := "\x1b[31mRED text\x1b[0m normal"
		fmt.Println(raw)
		fmt.Printf("bytes(raw[0:8]) = %q\n", []byte(raw)[:8])
		fmt.Println(strings.Repeat("-", 60))

		// 2) fatih/color
		fmt.Println("2) fatih/color:")
		fmt.Println(color.New(color.FgGreen, color.Bold).Sprint("Green Bold"))
		fmt.Println(color.New(color.FgCyan).Sprint("Cyan normal"))
		fmt.Println(strings.Repeat("-", 60))

		// 3) Renderer + Chroma
		r := renderer.NewMarkdownRenderer()
		sample := "```go\npackage main\n\nfunc main() {\n    println(\"hi\")\n}\n```"
		out := r.RenderMarkdown(sample)
		fmt.Println("3) Renderer + Chroma (с рамкой):")
		fmt.Print(out)
		fmt.Println(strings.Repeat("-", 60))

		// 4) Состояние окружения
		fmt.Println("4) Окружение:")
		fmt.Printf("TERM=%q, COLORTERM=%q, NO_COLOR=%q\n", os.Getenv("TERM"), os.Getenv("COLORTERM"), os.Getenv("NO_COLOR"))
		stdoutIsTTY := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
		fmt.Printf("stdout is TTY: %v\n", stdoutIsTTY)
		fmt.Printf("color.NoColor: %v\n", color.NoColor)
		fmt.Println(strings.Repeat("-", 60))

		fmt.Println("ANSI Debug - конец")
	},
}

func init() {
	rootCmd.AddCommand(debugANSICmd)
}
