package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/sirrobot01/protodex/internal/config"

	"github.com/sirrobot01/protodex/internal/client"
)

func init() {
	loginCmd.Flags().StringP("username", "u", "", "Username (required)")
	loginCmd.Flags().StringP("password", "p", "", "Password (will prompt if not provided)")
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the protodex registry",
	Long: `Authenticate with the protodex registry by creating an API token.

This command will create a session token and save it to your configuration file.

Examples:
  protodex login --username johndoe
  protodex login --username johndoe --password mypass`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		// If username not provided, show a prompt
		if username == "" {
			fmt.Print("Username: ")
			_, err := fmt.Scanln(&username)
			if err != nil {
				return fmt.Errorf("failed to read username: %w", err)
			}
		}

		// Prompt for password if not provided
		if password == "" {
			fmt.Print("Password: ")
			bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			password = string(bytePassword)
			fmt.Println()
		}

		if username == "" || password == "" {
			return fmt.Errorf("username and password are required")
		}

		c, err := client.New()
		if err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}

		response, err := c.Login(username, password)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		// Saving to config
		cfg := config.Get()
		cfg.HashedToken = response.Token
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("Successfully authenticated as %s\n", response.User.Username)
		fmt.Printf("Session token saved to configuration\n")

		return nil
	},
}
