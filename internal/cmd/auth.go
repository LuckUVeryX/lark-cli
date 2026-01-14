package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yjwong/lark-cli/internal/api"
	"github.com/yjwong/lark-cli/internal/auth"
	"github.com/yjwong/lark-cli/internal/output"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  "Manage Lark OAuth authentication",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Lark",
	Long:  "Authenticate with Lark using OAuth browser flow",
	Run: func(cmd *cobra.Command, args []string) {
		if err := auth.Login(); err != nil {
			output.Fatal("AUTH_ERROR", err)
		}
		output.Success("Successfully authenticated with Lark")
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from Lark",
	Long:  "Clear stored authentication credentials",
	Run: func(cmd *cobra.Command, args []string) {
		if err := auth.Logout(); err != nil {
			output.Fatal("AUTH_ERROR", err)
		}
		output.Success("Successfully logged out")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long:  "Display current authentication status and token expiry",
	Run: func(cmd *cobra.Command, args []string) {
		store := auth.GetTokenStore()

		status := api.OutputAuthStatus{
			Authenticated: store.IsValid(),
			ExpiresAt:     store.GetExpiresAt(),
		}

		if !status.Authenticated && store.CanRefresh() {
			// Token expired but we can refresh
			if err := auth.RefreshAccessToken(); err == nil {
				status.Authenticated = true
				status.ExpiresAt = store.GetExpiresAt()
			}
		}

		output.JSON(status)
	},
}

func init() {
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(statusCmd)
}
