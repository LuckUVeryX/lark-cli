package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yjwong/lark-cli/internal/api"
	"github.com/yjwong/lark-cli/internal/output"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <event-id>",
	Short: "Delete an event",
	Long: `Delete a calendar event.

Example:
  lark cal delete efa67a98-06a8-4df5-8559-746c8f4477ef_0`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		eventID := args[0]
		client := api.NewClient()

		// Get primary calendar
		cal, err := client.GetPrimaryCalendar()
		if err != nil {
			output.Fatal("CALENDAR_ERROR", err)
		}

		// Delete event
		if err := client.DeleteEvent(cal.CalendarID, eventID); err != nil {
			output.Fatal("API_ERROR", err)
		}

		output.JSON(map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Event deleted: %s", eventID),
		})
	},
}
