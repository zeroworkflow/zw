package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"zero-workflow/src/internal/ai"
	"zero-workflow/src/internal/renderer"
	"zero-workflow/src/internal/ui"
)

var (
	interactive bool
)

var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask AI a question",
	Long: `Ask the AI assistant a question. You can either provide the question as an argument
or use the interactive mode with the -i flag.

Examples:
  zw ask "How to create a Go struct?"
  zw ask -i  # Interactive mode`,
	Args: cobra.ArbitraryArgs,
	Run:  runAsk,
}

func init() {
	rootCmd.AddCommand(askCmd)
	askCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode for continuous conversation")
}

func runAsk(cmd *cobra.Command, args []string) {
	service, err := ai.NewService()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Hint: run 'zw ai login' to configure your AI token.\n")
		os.Exit(1)
	}
	renderer := renderer.NewMarkdownRenderer()

	if interactive {
		runInteractiveMode(service, renderer)
		return
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Please provide a question or use -i for interactive mode\n")
		fmt.Fprintf(os.Stderr, "Usage: zw ask \"your question\" or zw ask -i\n")
		os.Exit(1)
	}

	question := strings.Join(args, " ")
	askQuestion(service, renderer, question)
}

func askQuestion(service *ai.Service, renderer *renderer.MarkdownRenderer, question string) {
	// Fixed top-right spinner during real streaming
	spinner := ui.NewRightSpinner("Thinking")
	spinner.Start()

	var rawBuilder strings.Builder

	// Stream deltas and print them as raw text during streaming
	response, err := service.AskStream(question, func(delta string) {
		rawBuilder.WriteString(delta)
		fmt.Print(delta)
	})

	// Stop spinner
	spinner.Stop()

	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}

	// After streaming is complete, replace raw output with rendered markdown
	if response != "" {
		// Count lines in the raw output to know how much to clear
		rawOutput := rawBuilder.String()
		lines := strings.Count(rawOutput, "\n")
		
		// Move cursor up and clear the raw output
		if lines > 0 {
			fmt.Printf("\x1b[%dA", lines)
		}
		fmt.Print("\r\x1b[J")
		
		// Print the beautifully rendered version
		finalRendered := renderer.RenderMarkdown(response)
		fmt.Print(finalRendered)
	}
	
	fmt.Println("")
}

func runInteractiveMode(service *ai.Service, renderer *renderer.MarkdownRenderer) {
	fmt.Println("ZeroWorkflow AI - Interactive Mode")
	fmt.Println("Type your questions and press Enter. Type 'exit' or 'quit' to leave.")
	fmt.Println(strings.Repeat("─", 60))
	
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Print("\n> ")
		
		if !scanner.Scan() {
			break
		}
		
		input := strings.TrimSpace(scanner.Text())
		
		if input == "" {
			continue
		}
		
		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			break
		}
		
		askQuestion(service, renderer, input)
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}
