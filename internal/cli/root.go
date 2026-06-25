package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/mibienpanjoe/genius/internal/tui"
	"github.com/mibienpanjoe/genius/internal/workspace"
)

var engineFlag string

// NewRootCmd builds the genius root command. With no subcommand it launches the
// TUI at the home dashboard (FR-021).
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "genius",
		Short: "genius — a dedicated terminal study environment",
		Long: "genius — a dedicated terminal study environment.\n\n" +
			"With no subcommand it opens the TUI at the home dashboard. The subcommands\n" +
			"script the same loop: ingest → generate → revise, all grounded strictly in\n" +
			"your own course material under ~/study (or $GENIUS_HOME).\n\n" +
			"Workflow:\n" +
			"  1. ingest a lecture          genius ingest lecture.pdf\n" +
			"  2. build a study guide       genius guide lecture\n" +
			"  3. build revision Q&A        genius qa lecture\n" +
			"  4. revise (quiz)             genius   → pick course → r\n\n" +
			"Multi-chapter course (one course, many PDFs):\n" +
			"  genius ingest chap01.pdf --name algebra\n" +
			"  genius ingest chap02.pdf --name algebra\n" +
			"  genius guide algebra                 # grounded on all chapters\n" +
			"  genius guide algebra --files chap02.md   # just one chapter",
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
