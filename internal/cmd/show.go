package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yjwong/lark-cli/internal/api"
	"github.com/yjwong/lark-cli/internal/output"
)

var showCmd = &cobra.Command{
	Use:   "show <event-id>",
	Short: "Show event details",
	Long: `Show details of a specific calendar event.

Example:
  lark cal show efa67a98-06a8-4df5-8559-746c8f4477ef_0`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		eventID := args[0]
		client := api.NewClient()

		// Get primary calendar
		cal, err := client.GetPrimaryCalendar()
		if err != nil {
			output.Fatal("CALENDAR_ERROR", err)
		}

		// Get event
		event, err := client.GetEvent(cal.CalendarID, eventID)
		if err != nil {
			output.Fatal("EVENT_NOT_FOUND", err)
		}

		// If there are more attendees, fetch the full list
		if event.HasMoreAttendee {
			attendees, err := client.ListEventAttendees(cal.CalendarID, eventID)
			if err == nil {
				event.Attendees = attendees
			}
			// Silently ignore errors - use partial list from GetEvent
		}

		// Convert to output format
		outputEvent := api.ConvertToOutputEvent(*event)
		output.JSON(outputEvent)
	},
}
