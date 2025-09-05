package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sirrobot01/protodex/internal/client"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear authentication token",
	Long:  `Logout and clear the stored authentication token from configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}

		err = c.Logout()
		if err != nil {
			return fmt.Errorf("logout failed: %w", err)
		}

		fmt.Printf("Successfully logged out\n")
		fmt.Printf("Authentication token cleared\n")

		return nil
	},
}
