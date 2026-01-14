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
	commonFreetimeFrom            string
	commonFreetimeTo              string
	commonFreetimeUsers           string
	commonFreetimeOnlyBusy        bool
	commonFreetimeIncludeExternal bool
	commonFreetimeWorkHours       bool
	commonFreetimeMinLength       int
	commonFreetimeLimit           int
)

var commonFreetimeCmd = &cobra.Command{
	Use:   "common-freetime",
	Short: "Find common free time for multiple users",
	Long: `Query common free time slots for one or more users.

Returns time slots when all specified users are available.

Examples:
  lark cal common-freetime --from 2026-01-05 --to 2026-01-09 --users ou_abc123
  lark cal common-freetime --from 2026-01-05 --to 2026-01-09 --users "ou_abc123,ou_def456"
  lark cal common-freetime --from 2026-01-05T09:00:00+08:00 --to 2026-01-05T18:00:00+08:00 --users ou_abc123 --work-hours --min-length 30`,
	Run: func(cmd *cobra.Command, args []string) {
		if commonFreetimeFrom == "" || commonFreetimeTo == "" {
			output.Fatalf("VALIDATION_ERROR", "--from and --to are required")
		}
		if commonFreetimeUsers == "" {
			output.Fatalf("VALIDATION_ERROR", "--users is required")
		}

		// Parse user IDs
		userIDs := strings.Split(commonFreetimeUsers, ",")
		for i := range userIDs {
			userIDs[i] = strings.TrimSpace(userIDs[i])
		}

		// Validate user count
		if len(userIDs) == 0 || len(userIDs) > 10 {
			output.Fatalf("VALIDATION_ERROR", "Must specify 1-10 users")
		}

		// Parse timezone
		tz := config.GetTimezone()
		loc, err := time.LoadLocation(tz)
		if err != nil {
			loc = time.Local
			tz = loc.String()
		}

		// Parse time range
		startTime, err := timex.Parse(commonFreetimeFrom, loc)
		if err != nil {
			output.Fatalf("PARSE_ERROR", "Failed to parse --from: %v", err)
		}

		endTime, err := timex.Parse(commonFreetimeTo, loc)
		if err != nil {
			output.Fatalf("PARSE_ERROR", "Failed to parse --to: %v", err)
		}

		// If input doesn't contain a specific time, use start/end of day
		if !containsTimeSpec(commonFreetimeFrom) {
			startTime = timex.StartOfDay(startTime)
		}
		if !containsTimeSpec(commonFreetimeTo) {
			endTime = timex.EndOfDay(endTime)
		}

		// Validate time range (max 14 days for this API)
		if endTime.Sub(startTime) > 14*24*time.Hour {
			output.Fatalf("VALIDATION_ERROR", "Time range cannot exceed 14 days")
		}

		if endTime.Before(startTime) {
			output.Fatalf("VALIDATION_ERROR", "--to must be after --from")
		}

		// Set defaults
		limit := commonFreetimeLimit
		if limit == 0 {
			limit = 10
		}

		// Convert min-length from minutes to seconds
		minLengthSeconds := commonFreetimeMinLength * 60

		client := api.NewClient()

		slots, err := client.GetCommonFreeTime(api.CommonFreeTimeOptions{
			UserIDs:                 userIDs,
			StartTime:               startTime,
			EndTime:                 endTime,
			Timezone:                tz,
			OnlyBusy:                commonFreetimeOnlyBusy,
			IncludeExternalCalendar: commonFreetimeIncludeExternal,
			EnableWorkHour:          commonFreetimeWorkHours,
			MinTimeLength:           minLengthSeconds,
			Limit:                   limit,
		})
		if err != nil {
			output.Fatal("API_ERROR", err)
		}

		// Convert to output format
		freeSlots := make([]api.OutputFreeTimeSlot, len(slots))
		for i, s := range slots {
			freeSlots[i] = api.OutputFreeTimeSlot{
				Start:         s.StartTime,
				End:           s.EndTime,
				LengthMinutes: s.Length / 60,
			}
		}

		result := api.OutputCommonFreeTime{
			Query: api.OutputCommonFreeTimeQuery{
				Users:    userIDs,
				From:     startTime.Format(time.RFC3339),
				To:       endTime.Format(time.RFC3339),
				Timezone: tz,
			},
			FreeSlots: freeSlots,
		}

		output.JSON(result)
	},
}

func init() {
	commonFreetimeCmd.Flags().StringVar(&commonFreetimeFrom, "from", "", "Start time (required, ISO 8601 or date)")
	commonFreetimeCmd.Flags().StringVar(&commonFreetimeTo, "to", "", "End time (required, ISO 8601 or date)")
	commonFreetimeCmd.Flags().StringVar(&commonFreetimeUsers, "users", "", "Comma-separated user open_ids (required, max 10)")
	commonFreetimeCmd.Flags().BoolVar(&commonFreetimeOnlyBusy, "only-busy", true, "Only consider busy events")
	commonFreetimeCmd.Flags().BoolVar(&commonFreetimeIncludeExternal, "include-external", false, "Include external calendars")
	commonFreetimeCmd.Flags().BoolVar(&commonFreetimeWorkHours, "work-hours", false, "Respect work hour settings")
	commonFreetimeCmd.Flags().IntVar(&commonFreetimeMinLength, "min-length", 0, "Minimum slot length in minutes")
	commonFreetimeCmd.Flags().IntVar(&commonFreetimeLimit, "limit", 10, "Maximum results to return")
}
