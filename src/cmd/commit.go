package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
	zeroconfig "zero-workflow/src/internal/config"
	"zero-workflow/src/pkg/ai/zai"
	"zero-workflow/src/pkg/errors"
)

var (
	commitLang string
	autoPush   bool
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "AI-powered commit message generator",
	Long: `Analyzes staged changes and generates professional commit message using AI.
Follows Conventional Commits format.

Supported languages: ru (Russian), en (English), uk (Ukrainian), kz (Kazakh)`,
	RunE: runCommit,
}

func init() {
	rootCmd.AddCommand(commitCmd)
	commitCmd.Flags().StringVarP(&commitLang, "lang", "l", "en", "Language for commit messages (ru, en, uk, kz)")
	commitCmd.Flags().BoolVarP(&autoPush, "push", "p", false, "Automatically push after commit (if remote exists)")
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

	var selectedCommit CommitOption
	for {
		// Generate commit message using AI
		color.Cyan("Analyzing changes with AI...")
		commitOptions, err := generateCommitMessages(diff, stagedFiles)
		if err != nil {
			return fmt.Errorf("failed to generate commit messages: %w", err)
		}
		if len(commitOptions) == 0 {
			color.Yellow("AI failed to generate a commit message.")
			fmt.Print("Retry? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			retry, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(retry)) == "y" {
				continue
			}
			return fmt.Errorf("no commit message generated")
		}
		selectedCommit = commitOptions[0]

		// Display generated commit
		fmt.Println()
		color.Cyan("Generated commit message:")
		fmt.Printf("%s %s\n", color.GreenString("→"), color.WhiteString(selectedCommit.Title))
		if selectedCommit.Description != "" {
			fmt.Printf("  %s\n", color.HiBlackString(selectedCommit.Description))
		}

		// Ask for user confirmation
		fmt.Print("\nProceed with commit? (y/N/r to regenerate): ")
		reader := bufio.NewReader(os.Stdin)
		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))

		if confirm == "y" || confirm == "yes" {
			break // Proceed to commit
		}
		if confirm == "r" {
			color.Cyan("Regenerating commit message...")
			continue // Loop to regenerate
		}

		// Any other input cancels
		color.Yellow("Commit cancelled.")
		return nil
	}


	// Create commit
	commitMessage := selectedCommit.Title
	if selectedCommit.Description != "" {
		commitMessage += "\n\n" + selectedCommit.Description
	}

	userName, err := getGitConfig("user.name")
	if err != nil {
		return err
	}
	userEmail, err := getGitConfig("user.email")
	if err != nil {
		return err
	}

	commit, err := worktree.Commit(commitMessage, &git.CommitOptions{
		Author: &object.Signature{
			Name:  userName,
			Email: userEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	color.Green("✓ Commit created successfully: %s", commit.String()[:8])
	
	// Auto push if requested and remote exists
	if autoPush {
		if err := pushToRemote(repo); err != nil {
			color.Yellow("Warning: Failed to push: %v", err)
		} else {
			color.Green("✓ Pushed to remote successfully")
		}
	}
	
	return nil
}

type CommitOption struct {
	Title       string
	Description string
}

func generateCommitMessages(diff string, files []string) ([]CommitOption, error) {
	token, err := zeroconfig.GetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI token: %w", err)
	}

	client, err := zai.NewClient(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	// Validate language
	if !isValidLanguage(commitLang) {
		return nil, fmt.Errorf("unsupported language: %s. Supported: ru, en, uk, kz", commitLang)
	}

	langInstructions := getLanguageInstructions(commitLang)
	
	prompt := fmt.Sprintf(`%s

Analyze the following git diff and generate 1 professional commit message following Conventional Commits format.

Files changed: %s

Diff:
%s

Requirements:
1. Use conventional commit format: type(scope): description
2. Types: feat, fix, docs, style, refactor, test, chore
3. Keep title under 50 characters
4. Provide optional detailed description for complex changes
5. Be specific and descriptive
6. Choose the most appropriate commit type and description

Return in this exact format:
type(scope): short description
Optional longer description explaining the change`, langInstructions, strings.Join(files, ", "), diff)

	ctx := context.Background()
	response, err := client.Chat(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	options := parseCommitOptions(response)
	if len(options) == 0 {
		return nil, fmt.Errorf("failed to parse AI response")
	}

	return options, nil
}

func isValidLanguage(lang string) bool {
	validLangs := []string{"ru", "en", "uk", "kz"}
	for _, valid := range validLangs {
		if lang == valid {
			return true
		}
	}
	return false
}

func getLanguageInstructions(lang string) string {
	switch lang {
	case "ru":
		return "Генерируй коммиты на русском языке. Используй русские слова для описания изменений."
	case "en":
		return "Generate commits in English. Use clear and concise English descriptions."
	case "uk":
		return "Генеруй коміти українською мовою. Використовуй українські слова для опису змін."
	case "kz":
		return "Коммиттерді қазақ тілінде жасаңыз. Өзгерістерді сипаттау үшін қазақ сөздерін пайдаланыңыз."
	default:
		return "Generate commits in Russian. Use Russian words for describing changes."
	}
}


func parseCommitOptions(response string) []CommitOption {
	lines := strings.Split(response, "\n")
	var options []CommitOption
	var title string
	var description strings.Builder
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// First non-empty line is the title
		if title == "" {
			title = line
		} else {
			// Rest is description
			if description.Len() > 0 {
				description.WriteString("\n")
			}
			description.WriteString(line)
		}
	}
	
	if title != "" {
		options = append(options, CommitOption{
			Title:       title,
			Description: description.String(),
		})
	}
	
	return options
}

func getStagedDiff(repo *git.Repository) (string, error) {
	// Validate git command for security
	args := []string{"--staged"}
	if err := errors.ValidateGitCommand("diff", args); err != nil {
		return "", fmt.Errorf("git command validation failed: %w", err)
	}

	// Use the system's `git diff` command for a reliable and standard diff.
	cmd := exec.Command("git", "diff", "--staged")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) == 0 {
			return "", errors.NewGitError("diff", args, "failed to get staged diff", err)
		}
	}

	result := string(output)

	// Limit diff size to avoid overwhelming the AI model and hitting token limits.
	if len(result) > 8000 {
		result = result[:8000] + "\n... (diff truncated)"
	}

	return result, nil
}



func pushToRemote(repo *git.Repository) error {
	// Get current branch
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}
	
	// Check if remote exists
	_, err = repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("no remote 'origin' found: %w", err)
	}
	
	// Validate git command for security
	args := []string{"origin", head.Name().Short()}
	if err := errors.ValidateGitCommand("push", args); err != nil {
		return fmt.Errorf("git command validation failed: %w", err)
	}
	
	// Use system git command for push to leverage existing auth
	// This avoids authentication issues with go-git
	cmd := exec.Command("git", "push", "origin", head.Name().Short())
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Sanitize output to prevent token leakage
		sanitizedOutput := errors.SanitizeForLog(fmt.Errorf(string(output)))
		return errors.NewGitError("push", args, "failed to push", fmt.Errorf(sanitizedOutput))
	}
	
	return nil
}

func getGitConfig(key string) (string, error) {
	// Validate git command for security
	args := []string{key}
	if err := errors.ValidateGitCommand("config", args); err != nil {
		return "", fmt.Errorf("git command validation failed: %w", err)
	}

	cmd := exec.Command("git", "config", key)
	output, err := cmd.Output()
	if err != nil {
		return "", errors.NewGitError("config", args, fmt.Sprintf("git config %s is not set. Please configure it using git config --global %s \"Your Name/Email\"", key, key), err)
	}
	
	value := strings.TrimSpace(string(output))
	if value == "" {
		return "", errors.NewGitError("config", args, fmt.Sprintf("git config %s is empty. Please configure it", key), nil)
	}
	
	return value, nil
}
