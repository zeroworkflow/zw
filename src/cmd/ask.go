package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"zero-workflow/src/internal/ai"
	"zero-workflow/src/internal/renderer"
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
	client := ai.NewClient()
	renderer := renderer.NewMarkdownRenderer()

	if interactive {
		runInteractiveMode(client, renderer)
		return
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Please provide a question or use -i for interactive mode\n")
		fmt.Fprintf(os.Stderr, "Usage: zw ask \"your question\" or zw ask -i\n")
		os.Exit(1)
	}

	question := strings.Join(args, " ")
	askQuestion(client, renderer, question)
}

func askQuestion(client *ai.Client, renderer *renderer.MarkdownRenderer, question string) {
	fmt.Printf("🤖 Thinking...\n\n")
	
	response, err := client.Chat(question)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Render the markdown response with syntax highlighting
	rendered := renderer.RenderMarkdown(response)
	fmt.Print(rendered)
	fmt.Println()
}

func runInteractiveMode(client *ai.Client, renderer *renderer.MarkdownRenderer) {
	fmt.Println("🚀 ZeroWorkflow AI - Interactive Mode")
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
			fmt.Println("👋 Goodbye!")
			break
		}
		
		askQuestion(client, renderer, input)
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}
