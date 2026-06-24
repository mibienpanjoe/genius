package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mibienpanjoe/genius/internal/engine"
	"github.com/mibienpanjoe/genius/internal/generate"
	"github.com/mibienpanjoe/genius/internal/workspace"
)

func newGuideCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "guide <course>",
		Short: "Generate a study guide for a course",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGuide(cmd, args[0])
		},
	}
}

var (
	qaCount int
	qaScope string
)

func newQACmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "qa <course>",
		Short: "Generate revision Q&A for a course",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQA(cmd, args[0])
		},
	}
	cmd.Flags().IntVar(&qaCount, "count", 0, "number of Q&A pairs (default 10)")
	cmd.Flags().StringVar(&qaScope, "scope", "", "limit Q&A to a topic/section")
	return cmd
}

func runGuide(cmd *cobra.Command, course string) error {
	ws, eng, err := setup(cmd)
	if err != nil {
		return err
	}
	material, err := ws.CourseMaterial(course)
	if err != nil {
		return groundingError(course, err)
	}
	fmt.Printf("generating guide for %s using %s…\n", course, eng.Name())
	out, err := generate.Guide(cmd.Context(), eng, course, material)
	if err != nil {
		return err
	}
	path := ws.GuidePath(course)
	if err := writeWithConfirm(ws, path, []byte(out+"\n")); err != nil {
		return err
	}
	fmt.Printf("✓ guide written → %s\n", path)
	return nil
}

func runQA(cmd *cobra.Command, course string) error {
	ws, eng, err := setup(cmd)
	if err != nil {
		return err
	}
	material, err := ws.CourseMaterial(course)
	if err != nil {
		return groundingError(course, err)
	}
	fmt.Printf("generating Q&A for %s using %s…\n", course, eng.Name())
	out, err := generate.QA(cmd.Context(), eng, course, material,
		generate.QAOpts{Count: qaCount, Scope: qaScope})
	if err != nil {
		return err
	}
	path := ws.QAPath(course)
	if err := writeWithConfirm(ws, path, []byte(out+"\n")); err != nil {
		return err
	}
	fmt.Printf("✓ Q&A written → %s\n", path)
	return nil
}

// setup opens the workspace and constructs the active engine (config default,
// overridable by --engine).
func setup(cmd *cobra.Command) (workspace.Workspace, engine.Engine, error) {
	cfg, err := workspace.LoadConfig()
	if err != nil {
		return workspace.Workspace{}, nil, fmt.Errorf("reading config: %w", err)
	}
	ws, err := workspace.Open(cfg)
	if err != nil {
		return workspace.Workspace{}, nil, err
	}
	name := engineFlag
	if !cmd.Flags().Changed("engine") && cfg.DefaultEngine != "" {
		name = cfg.DefaultEngine
	}
	eng, err := engine.New(name, cfg.Model)
	if err != nil {
		return ws, nil, err
	}
	return ws, eng, nil
}

// groundingError maps the no-material case to a clear refusal (ERR-041).
func groundingError(course string, err error) error {
	if errors.Is(err, workspace.ErrNoMaterial) {
		return fmt.Errorf("course %q has no source material — ingest a document first", course)
	}
	return err
}
