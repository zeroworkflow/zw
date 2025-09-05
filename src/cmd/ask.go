package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"zero-workflow/src/internal/config"
	"zero-workflow/src/internal/files"
	"zero-workflow/src/internal/renderer"
	"zero-workflow/src/internal/ui"
	"zero-workflow/src/pkg/ai/zai"
)

var (
	interactive bool
	fileList   []string
)

var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask AI a question",
	Long: `Ask the AI assistant a question. You can either provide the question as an argument
or use the interactive mode with the -i flag. You can also include files for context.

Examples:
  zw ask "How to create a Go struct?"
  zw ask "Explain this code" --file src/main.go
  zw ask "Review my code" -f main.go -f config.go
  zw ask -i  # Interactive mode`,
	Args: cobra.ArbitraryArgs,
	Run:  runAsk,
}

func init() {
	rootCmd.AddCommand(askCmd)
	askCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode for continuous conversation")
	askCmd.Flags().StringSliceVarP(&fileList, "file", "f", []string{}, "Include files for context (can be used multiple times)")
}

func runAsk(cmd *cobra.Command, args []string) {
	token, err := config.GetToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client, err := zai.NewClient(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating AI client: %v\n", err)
		os.Exit(1)
	}

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
	askQuestion(client, renderer, question, fileList)
}

func askQuestion(client *zai.Client, renderer *renderer.MarkdownRenderer, question string, filePaths []string) {
    // Fixed top-right spinner during real streaming
    spinner := ui.NewRightSpinner("Thinking")
    spinner.Start()

    var rawBuilder strings.Builder

    // Process files if provided
    var fileContext string
    if len(filePaths) > 0 {
        fileReader := files.NewReader()
        
        // Validate files first
        if err := fileReader.ValidateFiles(filePaths); err != nil {
            spinner.Stop()
            fmt.Fprintf(os.Stderr, "\nFile validation error: %v\n", err)
            os.Exit(1)
        }
        
        // Read files
        fileContents, err := fileReader.ReadFiles(filePaths)
        if err != nil {
            spinner.Stop()
            fmt.Fprintf(os.Stderr, "\nError reading files: %v\n", err)
            os.Exit(1)
        }
        
        fileContext = fileReader.FormatFilesForAI(fileContents)
        fmt.Printf("\n[*] Loaded %d file(s) for context\n", len(fileContents))
    }
    
    // Combine question with file context
    fullQuestion := question + fileContext
    
    // Stream deltas and print them as raw text during streaming
    ctx := context.Background()
    response, err := client.ChatStream(ctx, fullQuestion, func(delta string) {
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

func runInteractiveMode(client *zai.Client, renderer *renderer.MarkdownRenderer) {
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
		
		// In interactive mode, files are not supported yet
		askQuestion(client, renderer, input, []string{})
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}
