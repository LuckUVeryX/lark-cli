package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yjwong/lark-cli/internal/api"
	"github.com/yjwong/lark-cli/internal/output"
)

var (
	lookupUserEmails  []string
	lookupUserMobiles []string
)

var lookupUserCmd = &cobra.Command{
	Use:   "lookup-user",
	Short: "Look up user IDs by email or mobile",
	Long: `Look up user open_ids by email address or mobile number.

Use the returned open_id with other commands like common-freetime.

Examples:
  lark cal lookup-user --email alice@company.com
  lark cal lookup-user --email alice@company.com --email bob@company.com
  lark cal lookup-user --mobile +6512345678
  lark cal lookup-user --email alice@company.com --mobile +6512345678`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(lookupUserEmails) == 0 && len(lookupUserMobiles) == 0 {
			output.Fatalf("VALIDATION_ERROR", "At least one --email or --mobile is required")
		}

		// Validate limits
		if len(lookupUserEmails) > 50 {
			output.Fatalf("VALIDATION_ERROR", "Maximum 50 emails allowed per request")
		}
		if len(lookupUserMobiles) > 50 {
			output.Fatalf("VALIDATION_ERROR", "Maximum 50 mobile numbers allowed per request")
		}

		client := api.NewClient()

		users, err := client.LookupUsers(api.UserLookupOptions{
			Emails:  lookupUserEmails,
			Mobiles: lookupUserMobiles,
		})
		if err != nil {
			output.Fatal("API_ERROR", err)
		}

		// Convert to output format
		outputUsers := make([]api.OutputUserInfo, len(users))
		for i, u := range users {
			outputUsers[i] = api.OutputUserInfo{
				UserID: u.UserID,
				Email:  u.Email,
				Mobile: u.Mobile,
			}
		}

		result := api.OutputUserLookup{
			Users: outputUsers,
		}

		output.JSON(result)
	},
}

func init() {
	lookupUserCmd.Flags().StringArrayVar(&lookupUserEmails, "email", nil, "Email address to look up (can be repeated)")
	lookupUserCmd.Flags().StringArrayVar(&lookupUserMobiles, "mobile", nil, "Mobile number to look up (can be repeated)")
}
