package cmd

import (
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yjwong/lark-cli/internal/api"
	"github.com/yjwong/lark-cli/internal/config"
	"github.com/yjwong/lark-cli/internal/output"
	timex "github.com/yjwong/lark-cli/internal/time"
)

var (
	freebusyFrom string
	freebusyTo   string
	freebusyUser string
	freebusyRoom string
)

var freebusyCmd = &cobra.Command{
	Use:   "freebusy",
	Short: "Query availability",
	Long: `Query busy/free information for yourself, a user, or a meeting room.

By default, queries your own availability. Use --user or --room to check others.

Examples:
  lark cal freebusy --from 2026-01-03T09:00:00+08:00 --to 2026-01-03T18:00:00+08:00
  lark cal freebusy --from 2026-01-03 --to 2026-01-03 --user ou_xxxxxxxxxx
  lark cal freebusy --from 2026-01-06 --to 2026-01-10 --room omm_xxxxxxxxxx`,
	Run: func(cmd *cobra.Command, args []string) {
		if freebusyFrom == "" || freebusyTo == "" {
			output.Fatalf("VALIDATION_ERROR", "--from and --to are required")
		}

		// Parse timezone
		tz := config.GetTimezone()
		loc, err := time.LoadLocation(tz)
		if err != nil {
			loc = time.Local
		}

		// Parse time range
		startTime, err := timex.Parse(freebusyFrom, loc)
		if err != nil {
			output.Fatalf("PARSE_ERROR", "Failed to parse --from: %v", err)
		}

		endTime, err := timex.Parse(freebusyTo, loc)
		if err != nil {
			output.Fatalf("PARSE_ERROR", "Failed to parse --to: %v", err)
		}

		// If input doesn't contain a specific time, use start/end of day
		if !containsTimeSpec(freebusyFrom) {
			startTime = timex.StartOfDay(startTime)
		}
		if !containsTimeSpec(freebusyTo) {
			endTime = timex.EndOfDay(endTime)
		}

		// Validate time range (max 90 days)
		if endTime.Sub(startTime) > 90*24*time.Hour {
			output.Fatalf("VALIDATION_ERROR", "Time range cannot exceed 90 days")
		}

		if endTime.Before(startTime) {
			output.Fatalf("VALIDATION_ERROR", "--to must be after --from")
		}

		client := api.NewClient()

		// If no user or room specified, get current user's open_id
		userID := freebusyUser
		if userID == "" && freebusyRoom == "" {
			user, err := client.GetCurrentUser()
			if err != nil {
				output.Fatal("USER_ERROR", err)
			}
			userID = user.OpenID
			if userID == "" {
				output.Fatalf("USER_ERROR", "Could not determine current user ID")
			}
		}

		// Query freebusy
		periods, err := client.GetFreebusy(api.FreebusyOptions{
			StartTime: startTime,
			EndTime:   endTime,
			UserID:    userID,
			RoomID:    freebusyRoom,
		})
		if err != nil {
			output.Fatal("API_ERROR", err)
		}

		// Convert to output format
		busyPeriods := make([]api.OutputFreebusyPeriod, len(periods))
		for i, p := range periods {
			busyPeriods[i] = api.OutputFreebusyPeriod{
				Start: p.StartTime,
				End:   p.EndTime,
			}
		}

		result := api.OutputFreebusy{
			Query: api.OutputFreebusyQuery{
				From:   startTime.Format(time.RFC3339),
				To:     endTime.Format(time.RFC3339),
				UserID: userID,
				RoomID: freebusyRoom,
			},
			BusyPeriods: busyPeriods,
		}

		output.JSON(result)
	},
}

func init() {
	freebusyCmd.Flags().StringVar(&freebusyFrom, "from", "", "Start time (required, ISO 8601)")
	freebusyCmd.Flags().StringVar(&freebusyTo, "to", "", "End time (required, ISO 8601)")
	freebusyCmd.Flags().StringVar(&freebusyUser, "user", "", "User open_id to check (default: self)")
	freebusyCmd.Flags().StringVar(&freebusyRoom, "room", "", "Meeting room room_id to check")
}

// containsTimeSpec checks if a time string contains a time specification
func containsTimeSpec(s string) bool {
	s = strings.ToLower(s)
	// Check for time patterns like "10:00", "T15:00"
	if strings.Contains(s, ":") {
		return true
	}
	// Check for ISO 8601 time separator (T followed by digit)
	for i := 0; i < len(s)-1; i++ {
		if s[i] == 't' && s[i+1] >= '0' && s[i+1] <= '9' {
			return true
		}
	}
	return false
}
