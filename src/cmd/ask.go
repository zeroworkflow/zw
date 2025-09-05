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
	"zero-workflow/src/internal/handlers"
	"zero-workflow/src/internal/renderer"
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
	errorHandler := handlers.NewErrorHandler()
	
	token, err := config.GetToken()
	if err != nil {
		errorHandler.HandleFatalError(err, "token retrieval")
	}

	client, err := zai.NewClient(token)
	if err != nil {
		errorHandler.HandleFatalError(err, "AI client creation")
	}

	renderer := renderer.NewMarkdownRenderer()

	if interactive {
		runInteractiveMode(client, renderer, errorHandler)
		return
	}

	if len(args) == 0 {
		errorHandler.HandleFatalError(fmt.Errorf("please provide a question or use -i for interactive mode"), "argument validation")
	}

	question := strings.Join(args, " ")
	askQuestion(client, renderer, question, fileList, errorHandler)
}

func askQuestion(client *zai.Client, renderer *renderer.MarkdownRenderer, question string, filePaths []string, errorHandler *handlers.ErrorHandler) {
	spinnerHandler := handlers.NewSpinnerHandler("Thinking")
	var response string
	var err error

	err = spinnerHandler.WithSpinner(func() error {
		// Process files if provided
		fileContext, processErr := processFiles(filePaths, errorHandler)
		if processErr != nil {
			return processErr
		}

		// Combine question with file context
		fullQuestion := question + fileContext

		ctx := context.Background()
		// The callback is now empty, as we are not printing live deltas.
		// The final complete response is returned by ChatStream.
		response, err = client.ChatStream(ctx, fullQuestion, func(delta string) {
			// Do nothing here to prevent raw output
		})
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		errorHandler.HandleFatalError(err, "question processing")
	}

	// After the spinner has stopped, render and print the complete response
	if response != "" {
		finalRendered := renderer.RenderMarkdown(response)
		fmt.Println(finalRendered)
	}
}

// processFiles handles file processing logic
func processFiles(filePaths []string, errorHandler *handlers.ErrorHandler) (string, error) {
	if len(filePaths) == 0 {
		return "", nil
	}

	fileReader := files.NewReader()

	if err := fileReader.ValidateFiles(filePaths); err != nil {
		return "", fmt.Errorf("file validation error: %w", err)
	}

	fileContents, err := fileReader.ReadFiles(filePaths)
	if err != nil {
		return "", fmt.Errorf("error reading files: %w", err)
	}

	fmt.Printf("\n[*] Loaded %d file(s) for context\n", len(fileContents))
	return fileReader.FormatFilesForAI(fileContents), nil
}

func runInteractiveMode(client *zai.Client, renderer *renderer.MarkdownRenderer, errorHandler *handlers.ErrorHandler) {
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
		askQuestion(client, renderer, input, []string{}, errorHandler)
	}
	
	if err := scanner.Err(); err != nil {
		errorHandler.HandleFatalError(err, "input reading")
	}
}
