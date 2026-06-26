package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mibienpanjoe/genius/internal/engine"
	"github.com/mibienpanjoe/genius/internal/generate"
	"github.com/mibienpanjoe/genius/internal/workspace"
)

var guideFiles []string

func newGuideCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "guide <course>",
		Short: "Generate a study guide for a course",
		Long: "Generate a study guide for a course.\n\n" +
			"By default the guide is grounded on the WHOLE course — every .md under\n" +
			"courses/<course>/ (all ingested chapters) — and written to\n" +
			"guides/<course>.md. Use --files to ground on specific chapter files\n" +
			"instead; the scoped guide is filed separately at\n" +
			"guides/<course>/<scope>.md (joined chapter slugs) so it never overwrites\n" +
			"the whole-course one.",
		Example: "  # whole-course guide (all chapters)\n" +
			"  genius guide algebra\n\n" +
			"  # guide grounded on one chapter only\n" +
			"  genius guide algebra --files chap03.md\n\n" +
			"  # guide over a span of chapters\n" +
			"  genius guide algebra --files chap01.md,chap02.md",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGuide(cmd, args[0])
		},
	}
	cmd.Flags().StringSliceVar(&guideFiles, "files", nil,
		"ground on specific chapter files under courses/<course>/ (default: all)")
	return cmd
}

var (
	qaCount int
	qaScope string
	qaFiles []string
)

func newQACmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "qa <course>",
		Short: "Generate revision Q&A for a course",
		Long: "Generate revision Q&A for a course.\n\n" +
			"Grounding is the WHOLE course by default (written to qa/<course>.md);\n" +
			"--files narrows it to specific chapter files and files the result\n" +
			"separately at qa/<course>/<scope>.md, never overwriting the whole-course\n" +
			"Q&A. --scope is different: it is a free-text focus instruction passed to\n" +
			"the model (e.g. \"Karnaugh maps\"), not a filename — the grounding is\n" +
			"unchanged. Combine them: --files picks the source, --scope the topic.",
		Example: "  # 10 Q&A over the whole course\n" +
			"  genius qa algebra\n\n" +
			"  # 15 Q&A, narrowed to a topic (still grounded on whole course)\n" +
			"  genius qa algebra --count 15 --scope \"Karnaugh maps\"\n\n" +
			"  # Q&A grounded on one chapter only\n" +
			"  genius qa algebra --files chap03.md",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQA(cmd, args[0])
		},
	}
	cmd.Flags().IntVar(&qaCount, "count", 0, "number of Q&A pairs (default 10)")
	cmd.Flags().StringVar(&qaScope, "scope", "", "free-text topic focus (not a filename)")
	cmd.Flags().StringSliceVar(&qaFiles, "files", nil,
		"ground on specific chapter files under courses/<course>/ (default: all)")
	return cmd
}

func runGuide(cmd *cobra.Command, course string) error {
	ws, eng, err := setup(cmd)
	if err != nil {
		return err
	}
	material, err := courseMaterial(ws, course, guideFiles)
	if err != nil {
		return groundingError(course, err)
	}
	fmt.Printf("generating guide for %s using %s…\n", course, eng.Name())
	out, err := generate.Guide(cmd.Context(), eng, course, material)
	if err != nil {
		return err
	}
	path := ws.GuideTarget(course, guideFiles)
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
	material, err := courseMaterial(ws, course, qaFiles)
	if err != nil {
		return groundingError(course, err)
	}
	fmt.Printf("generating Q&A for %s using %s…\n", course, eng.Name())
	out, err := generate.QA(cmd.Context(), eng, course, material,
		generate.QAOpts{Count: qaCount, Scope: qaScope})
	if err != nil {
		return err
	}
	path := ws.QATarget(course, qaFiles)
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

// courseMaterial picks the grounding blob: specific chapter files when --files
// is given, else the whole course (every .md under courses/<course>/).
func courseMaterial(ws workspace.Workspace, course string, files []string) (string, error) {
	if len(files) > 0 {
		return ws.MaterialFromFiles(course, files)
	}
	return ws.CourseMaterial(course)
}

// groundingError maps the no-material case to a clear refusal (ERR-041).
func groundingError(course string, err error) error {
	if errors.Is(err, workspace.ErrNoMaterial) {
		return fmt.Errorf("course %q has no source material — ingest a document first", course)
	}
	return err
}
