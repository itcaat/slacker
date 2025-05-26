package cmd

import (
	"fmt"
	"os"

	"github.com/itcaat/slacker/internal/ui"
	"github.com/spf13/cobra"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive text-based user interface",
	Long: `Launch the interactive TUI (Text User Interface) for Slacker. This provides
a full-screen interface for browsing channels and viewing messages with keyboard navigation.

The TUI interface allows you to:
- Browse and select channels
- View message history with threading
- Navigate with keyboard shortcuts
- Refresh data in real-time

Keyboard shortcuts:
- ↑/↓ or k/j: Navigate up/down
- Enter: Select channel or expand thread
- Esc: Go back to previous view
- r: Refresh current view
- q or Ctrl+C: Quit

Examples:
  slacker tui                    # Launch the TUI interface`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runTUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI() error {
	return ui.RunTUI()
}
