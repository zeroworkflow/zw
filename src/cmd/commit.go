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
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
	zeroconfig "zero-workflow/src/internal/config"
	"zero-workflow/src/pkg/ai/zai"
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

	// Generate commit message using AI
	color.Cyan("Analyzing changes with AI...")
	commitOptions, err := generateCommitMessages(diff, stagedFiles)
	if err != nil {
		return fmt.Errorf("failed to generate commit messages: %w", err)
	}

	if len(commitOptions) == 0 {
		return fmt.Errorf("no commit message generated")
	}

	selectedCommit := commitOptions[0]
	
	// Display generated commit
	fmt.Println()
	color.Cyan("Generated commit message:")
	fmt.Printf("%s %s\n", color.GreenString("→"), color.WhiteString(selectedCommit.Title))
	if selectedCommit.Description != "" {
		fmt.Printf("  %s\n", color.HiBlackString(selectedCommit.Description))
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
	// Get HEAD commit
	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	headCommit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	// Get HEAD tree
	headTree, err := headCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD tree: %w", err)
	}

	// Get worktree and index
	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get status to find staged files
	status, err := worktree.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	var diffOutput strings.Builder
	
	// Process each staged file
	for file, fileStatus := range status {
		if fileStatus.Staging == git.Unmodified {
			continue
		}
		
		diffOutput.WriteString(fmt.Sprintf("=== %s ===\n", file))
		diffOutput.WriteString(fmt.Sprintf("Status: %s\n", getStagingStatus(fileStatus.Staging)))
		
		// Get staged content from index
		if fileStatus.Staging == git.Modified || fileStatus.Staging == git.Added {
			// Get file from HEAD (if exists)
			var headContent string
			if fileStatus.Staging == git.Modified {
				headFile, err := headTree.File(file)
				if err == nil {
					headContent, _ = headFile.Contents()
				}
			}
			
			// Get staged content from index
			idx, err := repo.Storer.Index()
			if err == nil {
				for _, entry := range idx.Entries {
					if entry.Name == file {
						// Get blob content from index
						obj, err := repo.Storer.EncodedObject(plumbing.BlobObject, entry.Hash)
						if err == nil {
							reader, err := obj.Reader()
							if err == nil {
								defer reader.Close()
								
								var stagedContent strings.Builder
								buffer := make([]byte, 1024)
								for {
									n, err := reader.Read(buffer)
									if n > 0 {
										stagedContent.Write(buffer[:n])
									}
									if err != nil {
										break
									}
								}
								
								stagedStr := stagedContent.String()
								
								// Show diff
								if fileStatus.Staging == git.Modified {
									diffOutput.WriteString("Changes:\n")
									diffOutput.WriteString(fmt.Sprintf("- Old: %d chars\n", len(headContent)))
									diffOutput.WriteString(fmt.Sprintf("+ New: %d chars\n", len(stagedStr)))
									
									// Show actual line changes
									oldLines := strings.Split(headContent, "\n")
									newLines := strings.Split(stagedStr, "\n")
									
									diffOutput.WriteString("Diff:\n")
									
									// Simple diff algorithm
									maxLines := 20
									changeCount := 0
									
									for i := 0; i < len(newLines) && changeCount < maxLines; i++ {
										if i < len(oldLines) {
											if oldLines[i] != newLines[i] {
												diffOutput.WriteString(fmt.Sprintf("-%d: %s\n", i+1, oldLines[i]))
												diffOutput.WriteString(fmt.Sprintf("+%d: %s\n", i+1, newLines[i]))
												changeCount++
											}
										} else {
											// New line added
											diffOutput.WriteString(fmt.Sprintf("+%d: %s\n", i+1, newLines[i]))
											changeCount++
										}
									}
									
									if changeCount >= maxLines {
										diffOutput.WriteString("... (more changes)\n")
									}
								} else {
									// New file
									diffOutput.WriteString("New file content:\n")
									lines := strings.Split(stagedStr, "\n")
									maxLines := 15
									if len(lines) > maxLines {
										lines = lines[:maxLines]
										diffOutput.WriteString("(showing first 15 lines)\n")
									}
									for i, line := range lines {
										diffOutput.WriteString(fmt.Sprintf("+%d: %s\n", i+1, line))
									}
								}
							}
						}
						break
					}
				}
			}
		} else if fileStatus.Staging == git.Deleted {
			diffOutput.WriteString("File deleted\n")
		}
		
		diffOutput.WriteString("\n")
	}
	
	result := diffOutput.String()
	if strings.TrimSpace(result) == "" {
		return "No staged changes found", nil
	}
	
	// Limit diff size to avoid overwhelming AI
	if len(result) > 8000 {
		lines := strings.Split(result, "\n")
		if len(lines) > 200 {
			result = strings.Join(lines[:200], "\n") + "\n... (diff truncated)"
		}
	}
	
	return result, nil
}

func getStagingStatus(status git.StatusCode) string {
	switch status {
	case git.Added:
		return "Added"
	case git.Modified:
		return "Modified"
	case git.Deleted:
		return "Deleted"
	case git.Renamed:
		return "Renamed"
	case git.Copied:
		return "Copied"
	default:
		return "Unknown"
	}
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
	
	// Use system git command for push to leverage existing auth
	// This avoids authentication issues with go-git
	cmd := exec.Command("git", "push", "origin", head.Name().Short())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push: %s", string(output))
	}
	
	return nil
}

func getGitConfig(key string) string {
	// Read actual git config
	cmd := exec.Command("git", "config", "--global", key)
	output, err := cmd.Output()
	if err != nil {
		// Fallback to local config if global fails
		cmd = exec.Command("git", "config", key)
		output, err = cmd.Output()
		if err != nil {
			// Final fallbacks
			switch key {
			case "user.name":
				return "Developer"
			case "user.email":
				return "developer@example.com"
			default:
				return ""
			}
		}
	}
	
	return strings.TrimSpace(string(output))
}
