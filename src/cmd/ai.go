package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"zero-workflow/src/internal/ai"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "Manage AI settings (token, status)",
}

var aiLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store AI token securely",
	Run: func(cmd *cobra.Command, args []string) {
		token, _ := cmd.Flags().GetString("token")
		token = strings.TrimSpace(token)
		if token == "" {
			fmt.Print("Enter AI token (input hidden): ")
			b, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading token: %v\n", err)
				os.Exit(1)
			}
			token = strings.TrimSpace(string(b))
		}
		loc, err := ai.SaveToken(token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error saving token: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("AI token saved in %s\n", loc)
	},
}

var aiLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored AI token",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ai.DeleteToken(); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting token: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("AI token removed")
	},
}

var aiStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show AI token status",
	Run: func(cmd *cobra.Command, args []string) {
		tok, src, err := ai.ResolveToken()
		if err != nil {
			fmt.Println("Status: no token configured")
			fmt.Println("Hint: run 'zw ai login'")
			return
		}
		masked := maskToken(tok)
		fmt.Printf("Status: token from %s\n", src)
		fmt.Printf("Token: %s\n", masked)
	},
}

func maskToken(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 8 {
		return "********"
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

func init() {
	rootCmd.AddCommand(aiCmd)
	aiCmd.AddCommand(aiLoginCmd)
	aiCmd.AddCommand(aiLogoutCmd)
	aiCmd.AddCommand(aiStatusCmd)

	aiLoginCmd.Flags().String("token", "", "Token value (not recommended, prefer interactive input)")
}
