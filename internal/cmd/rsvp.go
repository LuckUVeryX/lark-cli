package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yjwong/lark-cli/internal/api"
	"github.com/yjwong/lark-cli/internal/output"
)

var (
	rsvpAccept    bool
	rsvpDecline   bool
	rsvpTentative bool
)

var rsvpCmd = &cobra.Command{
	Use:   "rsvp <event-id>",
	Short: "Reply to an event invitation",
	Long: `Reply to an event invitation with accept, decline, or tentative.

You must specify exactly one of --accept, --decline, or --tentative.

Examples:
  lark cal rsvp abc123 --accept
  lark cal rsvp abc123 --decline
  lark cal rsvp abc123 --tentative`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		eventID := args[0]

		// Validate that exactly one status flag is set
		count := 0
		var status string
		if rsvpAccept {
			count++
			status = "accept"
		}
		if rsvpDecline {
			count++
			status = "decline"
		}
		if rsvpTentative {
			count++
			status = "tentative"
		}

		if count == 0 {
			output.Fatal("VALIDATION_ERROR", fmt.Errorf("must specify one of --accept, --decline, or --tentative"))
		}
		if count > 1 {
			output.Fatal("VALIDATION_ERROR", fmt.Errorf("cannot specify multiple status flags"))
		}

		client := api.NewClient()

		// Get primary calendar
		cal, err := client.GetPrimaryCalendar()
		if err != nil {
			output.Fatal("CALENDAR_ERROR", err)
		}

		// Send RSVP reply
		if err := client.ReplyToEvent(cal.CalendarID, eventID, status); err != nil {
			output.Fatal("API_ERROR", err)
		}

		output.JSON(map[string]interface{}{
			"success":     true,
			"message":     fmt.Sprintf("RSVP sent: %s", status),
			"event_id":    eventID,
			"rsvp_status": status,
		})
	},
}

func init() {
	rsvpCmd.Flags().BoolVar(&rsvpAccept, "accept", false, "Accept the invitation")
	rsvpCmd.Flags().BoolVar(&rsvpDecline, "decline", false, "Decline the invitation")
	rsvpCmd.Flags().BoolVar(&rsvpTentative, "tentative", false, "Mark as tentative")
}
