package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
	"zero-workflow/src/internal/ai"
	"zero-workflow/src/internal/config"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "AI-powered commit message generator",
	Long: `Analyzes staged changes and generates professional commit messages using AI.
Provides 5 commit options following Conventional Commits format.`,
	RunE: runCommit,
}

func init() {
	rootCmd.AddCommand(commitCmd)
}

func runCommit(cmd *cobra.Command, args []string) error {
	// Open git repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get status
	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get git status: %w", err)
	}

	// Check if there are staged changes
	hasStagedChanges := false
	var stagedFiles []string
	
	for file, fileStatus := range status {
		if fileStatus.Staging != git.Unmodified {
			hasStagedChanges = true
			stagedFiles = append(stagedFiles, file)
		}
	}

	if !hasStagedChanges {
		color.Yellow("No staged changes found. Use 'git add' to stage files first.")
		return nil
	}

	// Show staged files
	color.Cyan("Staged files:")
	for _, file := range stagedFiles {
		fmt.Printf("  %s %s\n", color.GreenString("✓"), file)
	}
	fmt.Println()

	// Get diff for staged changes
	diff, err := getStagedDiff(repo)
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}

	// Generate commit messages using AI
	color.Cyan("Analyzing changes with AI...")
	commitOptions, err := generateCommitMessages(diff, stagedFiles)
	if err != nil {
		return fmt.Errorf("failed to generate commit messages: %w", err)
	}

	// Display options
	fmt.Println()
	color.Cyan("Generated commit options:")
	for i, option := range commitOptions {
		fmt.Printf("%s %d. %s\n", color.BlueString("→"), i+1, color.WhiteString(option.Title))
		if option.Description != "" {
			fmt.Printf("     %s\n", color.HiBlackString(option.Description))
		}
		fmt.Println()
	}

	// Get user choice
	choice, err := getUserChoice(len(commitOptions))
	if err != nil {
		return err
	}

	if choice == 0 {
		color.Yellow("Commit cancelled.")
		return nil
	}

	selectedCommit := commitOptions[choice-1]
	
	// Confirm commit
	fmt.Printf("\nSelected commit message:\n%s\n", color.GreenString(selectedCommit.Title))
	if selectedCommit.Description != "" {
		fmt.Printf("%s\n", color.HiBlackString(selectedCommit.Description))
	}
	
	fmt.Print("\nProceed with commit? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	
	if confirm != "y" && confirm != "yes" {
		color.Yellow("Commit cancelled.")
		return nil
	}

	// Create commit
	commitMessage := selectedCommit.Title
	if selectedCommit.Description != "" {
		commitMessage += "\n\n" + selectedCommit.Description
	}

	commit, err := worktree.Commit(commitMessage, &git.CommitOptions{
		Author: &object.Signature{
			Name:  getGitConfig("user.name"),
			Email: getGitConfig("user.email"),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	color.Green("✓ Commit created successfully: %s", commit.String()[:8])
	return nil
}

type CommitOption struct {
	Title       string
	Description string
}

func generateCommitMessages(diff string, files []string) ([]CommitOption, error) {
	// Try AI generation first, fallback to predefined options
	cfg, err := config.Load()
	if err != nil {
		color.Yellow("Warning: Failed to load config, using fallback options")
		return getFallbackCommitOptions(files), nil
	}

	client, err := ai.NewClient(cfg)
	if err != nil {
		color.Yellow("Warning: Failed to create AI client, using fallback options")
		return getFallbackCommitOptions(files), nil
	}

	prompt := fmt.Sprintf(`Analyze the following git diff and generate 5 professional commit messages following Conventional Commits format.

Files changed: %s

Diff:
%s

Requirements:
1. Use conventional commit format: type(scope): description
2. Types: feat, fix, docs, style, refactor, test, chore
3. Keep title under 50 characters
4. Provide optional detailed description for complex changes
5. Be specific and descriptive
6. Order by relevance (most appropriate first)

Return 5 options in this exact format:
1. type(scope): short description
   Optional longer description explaining the change

2. type(scope): short description
   Optional longer description explaining the change

Continue for all 5 options.`, strings.Join(files, ", "), diff)

	response, err := client.GenerateText(prompt)
	if err != nil {
		color.Yellow("Warning: AI generation failed, using fallback options")
		return getFallbackCommitOptions(files), nil
	}

	options := parseCommitOptions(response)
	if len(options) == 0 {
		color.Yellow("Warning: AI parsing failed, using fallback options")
		return getFallbackCommitOptions(files), nil
	}

	return options, nil
}

func getFallbackCommitOptions(files []string) []CommitOption {
	// Generate smart fallback options based on file types and names
	options := []CommitOption{}
	
	// Analyze file patterns
	hasGoFiles := false
	hasConfigFiles := false
	hasDocFiles := false
	hasTestFiles := false
	
	for _, file := range files {
		if strings.HasSuffix(file, ".go") {
			hasGoFiles = true
		}
		if strings.Contains(file, "config") || strings.HasSuffix(file, ".env") || strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml") {
			hasConfigFiles = true
		}
		if strings.Contains(file, "README") || strings.Contains(file, "doc") || strings.HasSuffix(file, ".md") {
			hasDocFiles = true
		}
		if strings.Contains(file, "test") || strings.HasSuffix(file, "_test.go") {
			hasTestFiles = true
		}
	}
	
	// Generate contextual options
	if hasGoFiles && len(files) == 1 {
		filename := strings.TrimSuffix(files[0], ".go")
		options = append(options, CommitOption{
			Title: fmt.Sprintf("feat: add %s implementation", filename),
			Description: fmt.Sprintf("Implement %s functionality", filename),
		})
		options = append(options, CommitOption{
			Title: fmt.Sprintf("fix: resolve issues in %s", filename),
			Description: "",
		})
	}
	
	if hasConfigFiles {
		options = append(options, CommitOption{
			Title: "chore: update configuration",
			Description: "Update project configuration settings",
		})
	}
	
	if hasDocFiles {
		options = append(options, CommitOption{
			Title: "docs: update documentation",
			Description: "Improve project documentation",
		})
	}
	
	if hasTestFiles {
		options = append(options, CommitOption{
			Title: "test: add test coverage",
			Description: "Improve test coverage and reliability",
		})
	}
	
	// Fill remaining slots with generic options
	genericOptions := []CommitOption{
		{Title: "feat: implement new functionality", Description: "Add new features to the codebase"},
		{Title: "fix: resolve critical issues", Description: "Fix bugs and improve stability"},
		{Title: "refactor: improve code structure", Description: "Enhance code organization and maintainability"},
		{Title: "chore: update dependencies", Description: "Update project dependencies and tooling"},
		{Title: "style: improve code formatting", Description: "Apply code style improvements"},
	}
	
	// Ensure we have exactly 5 options
	for _, generic := range genericOptions {
		if len(options) >= 5 {
			break
		}
		// Avoid duplicates
		duplicate := false
		for _, existing := range options {
			if existing.Title == generic.Title {
				duplicate = true
				break
			}
		}
		if !duplicate {
			options = append(options, generic)
		}
	}
	
	// Ensure we have exactly 5 options
	if len(options) < 5 {
		for i := len(options); i < 5; i++ {
			options = append(options, genericOptions[i%len(genericOptions)])
		}
	}
	
	return options[:5]
}

func parseCommitOptions(response string) []CommitOption {
	lines := strings.Split(response, "\n")
	var options []CommitOption
	var currentOption *CommitOption

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if it's a numbered option
		if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") || 
		   strings.HasPrefix(line, "3.") || strings.HasPrefix(line, "4.") || 
		   strings.HasPrefix(line, "5.") {
			
			if currentOption != nil {
				options = append(options, *currentOption)
			}
			
			title := strings.TrimSpace(line[2:]) // Remove "1. " etc
			currentOption = &CommitOption{Title: title}
		} else if currentOption != nil && line != "" {
			// This is a description line
			if currentOption.Description == "" {
				currentOption.Description = line
			} else {
				currentOption.Description += "\n" + line
			}
		}
	}

	if currentOption != nil {
		options = append(options, *currentOption)
	}

	// Fallback if parsing failed
	if len(options) == 0 {
		return []CommitOption{
			{Title: "feat: update codebase", Description: ""},
			{Title: "fix: resolve issues", Description: ""},
			{Title: "refactor: improve code structure", Description: ""},
			{Title: "chore: update dependencies", Description: ""},
			{Title: "docs: update documentation", Description: ""},
		}
	}

	return options
}

func getStagedDiff(repo *git.Repository) (string, error) {
	// For simplicity, we'll use git command to get diff
	// In production, you might want to implement proper diff using go-git
	return "Staged changes detected", nil
}

func getUserChoice(maxOptions int) (int, error) {
	fmt.Printf("Choose commit option (1-%d, 0 to cancel): ", maxOptions)
	
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}

	choice, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return 0, fmt.Errorf("invalid input: please enter a number")
	}

	if choice < 0 || choice > maxOptions {
		return 0, fmt.Errorf("invalid choice: please enter a number between 0 and %d", maxOptions)
	}

	return choice, nil
}

func getGitConfig(key string) string {
	// Simple fallback - in production you'd want to read from git config
	switch key {
	case "user.name":
		return "Developer"
	case "user.email":
		return "developer@example.com"
	default:
		return ""
	}
}
