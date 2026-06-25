package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mibienpanjoe/genius/internal/generate"
	"github.com/mibienpanjoe/genius/internal/render"
	"github.com/mibienpanjoe/genius/internal/workspace"
)

var (
	solveSet  string
	solveEx   []string
	solveSave bool
)

func newSolveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "solve <course>",
		Short: "Work exercises from a set, grounded in the course",
		Long: "Work exercises from an ingested set, grounded ONLY in the course.\n\n" +
			"The set lives under exercises/<course>/<set>.md (see `ingest --kind\n" +
			"exercise`). Without --ex, the exercises are enumerated and listed so you\n" +
			"can pick; with --ex they are solved and the worked solution is printed.\n" +
			"Grounding is the whole course; the tutor solves each exercise as stated\n" +
			"and flags any gap in the material rather than fabricating.",
		Example: "  # list the exercises in a set\n" +
			"  genius solve algebra --set td1\n\n" +
			"  # solve specific exercises (and sub-parts)\n" +
			"  genius solve algebra --set td1 --ex 2,3.1\n\n" +
			"  # solve and also save to exercises/algebra/td1.solutions.md\n" +
			"  genius solve algebra --set td1 --ex 2 --save",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSolve(cmd, args[0])
		},
	}
	cmd.Flags().StringVar(&solveSet, "set", "",
		"exercise set name under exercises/<course>/ (required)")
	cmd.Flags().StringSliceVar(&solveEx, "ex", nil,
		"exercises to solve, e.g. 2,3.1,3.a; omit to list the set")
	cmd.Flags().BoolVar(&solveSave, "save", false,
		"also write solutions to <set>.solutions.md (source never modified)")
	_ = cmd.MarkFlagRequired("set")
	return cmd
}

func runSolve(cmd *cobra.Command, course string) error {
	ws, eng, err := setup(cmd)
	if err != nil {
		return err
	}

	setMD, err := ws.ReadExerciseSet(course, solveSet)
	if err != nil {
		if errors.Is(err, workspace.ErrNoMaterial) {
			return unknownSetError(ws, course, solveSet)
		}
		return err
	}

	exs, err := generate.Enumerate(setMD)
	if err != nil {
		if !errors.Is(err, generate.ErrNoExercises) {
			return err
		}
		// ERR-101: nothing enumerable — offer the whole set as one item.
		exs = []generate.Exercise{{Label: "Whole set", Num: "all", Text: setMD}}
	}

	// No selection: enumerate and exit so the learner can pick (FR-103).
	if len(solveEx) == 0 {
		printExerciseList(course, solveSet, exs)
		return nil
	}

	selected, err := generate.Select(exs, solveEx)
	if err != nil {
		return err // ERR-102: unknown exercise/sub-part
	}

	material, err := courseMaterial(ws, course, nil)
	if err != nil {
		return groundingError(course, err) // ERR-041
	}

	fmt.Printf("solving %s/%s using %s…\n", course, solveSet, eng.Name())
	out, err := generate.Solve(cmd.Context(), eng, course, material, selected)
	if err != nil {
		return err
	}

	rendered, rerr := render.Markdown(out, 100)
	if rerr != nil {
		rendered = out // fall back to raw markdown
	}
	fmt.Print(rendered)

	if solveSave {
		path := ws.Path("exercises", course, solveSet+".solutions.md")
		if err := writeWithConfirm(ws, path, []byte(out+"\n")); err != nil {
			return err
		}
		fmt.Printf("✓ solutions saved → %s\n", path)
	}
	return nil
}

// unknownSetError reports a missing set, listing the sets the course does have.
func unknownSetError(ws workspace.Workspace, course, set string) error {
	sets, _ := ws.ExerciseSets(course)
	if len(sets) == 0 {
		return fmt.Errorf("course %q has no exercise sets — ingest one: "+
			"genius ingest <file> --kind exercise --course %s", course, course)
	}
	return fmt.Errorf("no exercise set %q under course %q — available: %s",
		set, course, strings.Join(sets, ", "))
}

// printExerciseList prints the enumerated exercises (and sub-parts) with the
// tokens to pass to --ex.
func printExerciseList(course, set string, exs []generate.Exercise) {
	fmt.Printf("%s / %s — %d exercise(s):\n\n", course, set, len(exs))
	for _, e := range exs {
		fmt.Printf("  %-14s (--ex %s)\n", e.Label, e.Num)
		for _, p := range e.Parts {
			fmt.Printf("      %s.%s\n", e.Num, p.Num)
		}
	}
	fmt.Printf("\nsolve with: genius solve %s --set %s --ex <list>\n", course, set)
}
