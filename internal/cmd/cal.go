package cmd

import (
	"github.com/spf13/cobra"
)

var calCmd = &cobra.Command{
	Use:   "cal",
	Short: "Calendar commands",
	Long:  "Manage Lark calendar events",
}

func init() {
	calCmd.AddCommand(listCmd)
	calCmd.AddCommand(showCmd)
	calCmd.AddCommand(createCmd)
	calCmd.AddCommand(updateCmd)
	calCmd.AddCommand(deleteCmd)
	calCmd.AddCommand(searchCmd)
	calCmd.AddCommand(freebusyCmd)
	calCmd.AddCommand(rsvpCmd)
	calCmd.AddCommand(lookupUserCmd)
	calCmd.AddCommand(commonFreetimeCmd)
	calCmd.AddCommand(attendeeCmd)
}
