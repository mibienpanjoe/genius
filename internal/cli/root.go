package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"genius/internal/tui"
	"genius/internal/workspace"
)

var engineFlag string

// NewRootCmd builds the genius root command. With no subcommand it launches the
// TUI at the home dashboard (FR-021).
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "genius",
		Short:         "genius — a dedicated terminal study environment",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTUI(cmd.Flags().Changed("engine"))
		},
	}
	root.PersistentFlags().StringVar(&engineFlag, "engine", "claude",
		"generation engine (claude|codex)")
	root.AddCommand(newIngestCmd())
	root.AddCommand(newGuideCmd())
	root.AddCommand(newQACmd())
	return root
}

// runTUI opens the workspace, scans courses, and launches the Bubble Tea
// program in the alt-screen buffer.
func runTUI(engineChanged bool) error {
	cfg, err := workspace.LoadConfig()
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}
	ws, err := workspace.Open(cfg)
	if err != nil {
		return fmt.Errorf("opening workspace: %w", err)
	}
	courses, err := ws.Courses()
	if err != nil {
		return fmt.Errorf("scanning courses: %w", err)
	}

	engine := engineFlag
	if !engineChanged && cfg.DefaultEngine != "" {
		engine = cfg.DefaultEngine
	}

	m := tui.New(engine, ws, courses)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// Execute runs the root command and maps errors to exit codes (FR-092).
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "genius:", err)
		os.Exit(1)
	}
}
