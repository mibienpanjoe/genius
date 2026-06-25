package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mibienpanjoe/genius/internal/convert"
	"github.com/mibienpanjoe/genius/internal/workspace"
)

var (
	ingestName     string
	ingestKind     string
	ingestCourse   string
	ingestDescribe bool
	ingestOCR      bool
	ingestMinPx    int
)

func newIngestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ingest <file>",
		Short: "Convert a PDF/PPT/document to markdown in the workspace",
		Long: "Convert a document to markdown and file it in the workspace.\n\n" +
			"A course is a directory of markdown: genius reads EVERY .md under\n" +
			"courses/<name>/ as grounding. The course name and the document filename\n" +
			"are separate — --name sets the course, the file's own slug names the .md.\n" +
			"So pointing several PDFs at one --name builds a multi-chapter course\n" +
			"(use zero-padded names like chap01, chap02 so they sort in order).\n\n" +
			"You can also drop your own hand-written .md straight into courses/<name>/;\n" +
			"genius reads it the same way. Never copy a raw .pdf there — only ingested\n" +
			"or hand-written markdown counts as grounding.",
		Example: "  # single-file course (course name = filename slug)\n" +
			"  genius ingest lecture.pdf\n\n" +
			"  # multi-chapter course: many PDFs into one course, ordered\n" +
			"  genius ingest chap01.pdf --name algebra\n" +
			"  genius ingest chap02.pdf --name algebra\n\n" +
			"  # exercise set filed under a course\n" +
			"  genius ingest td1.pdf --kind exercise --course algebra",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIngest(cmd, args[0])
		},
	}
	cmd.Flags().StringVar(&ingestName, "name", "",
		"course name (course kind) or set name (exercise kind); default: filename slug")
	cmd.Flags().StringVar(&ingestKind, "kind", "course", "course|exercise")
	cmd.Flags().StringVar(&ingestCourse, "course", "",
		"target course for --kind exercise")
	cmd.Flags().BoolVar(&ingestDescribe, "describe-images", true,
		"vision-caption extracted figures (needs a vision engine)")
	cmd.Flags().BoolVar(&ingestOCR, "ocr", false,
		"attempt OCR/vision on pages with no text layer")
	cmd.Flags().IntVar(&ingestMinPx, "min-image-size", convert.DefaultMinImagePx,
		"skip images smaller than this (px) in both dimensions")
	return cmd
}

func runIngest(cmd *cobra.Command, file string) error {
	ctx := cmd.Context()

	// FR-033/ERR-031: converter presence.
	if !convert.Available() {
		return errors.New(convert.InstallHint)
	}
	// ERR-032: source must exist.
	if info, err := os.Stat(file); err != nil || info.IsDir() {
		return fmt.Errorf("cannot ingest %q: file does not exist or is a directory", file)
	}

	ws, eng, err := setup(cmd)
	if err != nil {
		return err
	}

	// Resolve the target markdown path and the figure assets directory.
	var target, assetsDir string
	switch ingestKind {
	case "course":
		name := ingestName
		if name == "" {
			name = workspace.Slug(file)
		}
		target = ws.CourseDocPath(name, workspace.Slug(file))
		assetsDir = ws.Path("courses", name, "assets")
	case "exercise":
		if ingestCourse == "" {
			return errors.New("--kind exercise requires --course <name>")
		}
		set := ingestName
		if set == "" {
			set = workspace.Slug(file)
		}
		target = ws.ExerciseSetPath(ingestCourse, set)
		assetsDir = ws.Path("exercises", ingestCourse, "assets")
	default:
		return fmt.Errorf("unknown --kind %q (want course|exercise)", ingestKind)
	}

	res, err := convert.Ingest(ctx, file, assetsDir, eng, convert.IngestOpts{
		Describe:   ingestDescribe,
		OCR:        ingestOCR,
		MinImagePx: ingestMinPx,
	})
	if err != nil {
		return err
	}

	if err := writeWithConfirm(ws, target, []byte(res.Markdown+"\n")); err != nil {
		return err
	}
	fmt.Printf("✓ ingested %s → %s\n", file, target)
	if len(res.Assets) > 0 {
		fmt.Printf("  %d figure(s) extracted → %s\n", len(res.Assets), assetsDir)
	}
	return nil
}

// writeWithConfirm applies the overwrite-confirmation rule (FR-034/ERR-034):
// on an existing target it prompts y/N and only overwrites on explicit yes.
func writeWithConfirm(ws workspace.Workspace, path string, data []byte) error {
	err := ws.WriteArtifact(path, data, false)
	if errors.Is(err, workspace.ErrExists) {
		if !confirm(fmt.Sprintf("%s exists — overwrite?", path)) {
			return errors.New("aborted: existing file kept")
		}
		return ws.WriteArtifact(path, data, true)
	}
	return err
}

func confirm(prompt string) bool {
	fmt.Printf("%s [y/N] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes"
}
