# genius — CLI & Internal Interface Specification
Version: v1.0, 2026-06-24

> genius exposes no HTTP API. Its "interfaces" are three contracts: the **CLI
> surface** (commands/flags a learner or script calls), the **internal Go
> interfaces** between components, and the **subprocess invocations** of external
> tools. This document specifies all three so a contributor can implement against
> them without guessing.

## Conventions
- **Invocation:** `genius [global flags] [subcommand] [args]`. No subcommand →
  launch the TUI.
- **Exit codes:** `0` success; non-zero on any failure (scriptable).
- **Output:** subcommands write artifacts to the workspace and print a one-line
  result to stdout; errors go to stderr.
- **Config:** `~/.config/genius/config.toml` — keys `study_root` (string),
  `default_engine` (`claude`|`codex`), `model` (string).
- **Workspace root resolution:** `$GENIUS_HOME` → config `study_root` → `~/study`.

## Global Flags
| Flag | Type | Default | Meaning |
|------|------|---------|---------|
| `--engine` | `claude`\|`codex` | config or `claude` | Generation backend for this run |
| `--model` | string | config or backend default | Model passed to the engine |
| `--root` | path | resolved root | Override workspace root for this run |
| `--yes` | bool | false | Assume "yes" to overwrite confirmations (scripting) |

## CLI Commands

### `genius` (no subcommand)
Launch the TUI at the home dashboard.
Args: none. Exit: `0` on clean quit.

### `genius ingest <file> [--name <course>] [--kind course|exercise] [--course <course>]`
Convert a document to markdown and file it.
- **Args:** `<file>` — path to a `markitdown`-supported document (required).
- **`--kind`** — `course` (default) files course material; `exercise` files an
  exercise set (TD / problem set / past test).
- **`--name`** — course name for `--kind course`; defaults to the file's base name
  (slugified).
- **`--course`** — required with `--kind exercise`: the course the set belongs to.
  The set name defaults to the file's base name.
- **`--describe-images` / `--no-describe-images`** — caption extracted figures
  with a vision Engine (default: on when the engine supports images; FR-035b/c).
- **`--ocr`** — transcribe scanned/no-text-layer documents via vision/tesseract
  (default: off; FR-035d).
- **`--min-image-size <px>`** — skip images smaller than this (default e.g.
  `64`), to drop logos/bullets.
- **Behavior:** verify `markitdown` present (ERR-031) → convert text → extract
  images to `assets/` → caption or placeholder per figure (INV-12) → handle
  scanned docs per `--ocr` (ERR-035) → write to `courses/<name>/<file>.md` +
  `courses/<name>/assets/*` (course) or the `exercises/<course>/` equivalents;
  confirm before overwrite unless `--yes` (ERR-034).
- **Success (stdout):** `ingested <file> → courses/<name>/<file>.md (4 figures
  described)`  (or `→ exercises/<course>/<set>.md`)
- **Errors:**
  | Exit | Condition |
  |------|-----------|
  | 1 | `markitdown` not installed (ERR-031) — message includes `pip install 'markitdown[all]'` |
  | 1 | File missing/unsupported (ERR-032) |
  | 1 | Output exists and overwrite declined (ERR-034) |

### `genius guide <course>`
Generate a study guide for the course.
- **Args:** `<course>` — course name (required).
- **Behavior:** read grounding markdown (ERR-041 if none) → assemble guide prompt
  → `Engine.Generate` → strip framing (INV-07) → write `guides/<course>.md`
  (confirm/`--yes`).
- **Success (stdout):** `wrote guides/<course>.md (engine=<engine>)`
- **Errors:**
  | Exit | Condition |
  |------|-----------|
  | 1 | No source markdown for course (ERR-041) |
  | 1 | Engine binary missing (ERR-081) |
  | 1 | Engine failed/timed out (ERR-082) |
  | 1 | Output exists and overwrite declined (ERR-034) |

### `genius qa <course> [--count N] [--scope <text>] [--file <md>...]`
Generate revision Q&A for the course. Same grounding/output/error contract as
`guide` (output `qa/<course>.md`), plus:
- **`--count N`** — number of Q&A pairs to target (FR-054). Default: `15`.
  Instructed to the Engine as a target count.
- **`--scope <text>`** — free-text focus, e.g. `--scope "chapter 3: limits"`
  (FR-055). Passed to the Engine as a constraint to draw only from that part.
- **`--file <md>...`** — restrict grounding to specific markdown files under
  `courses/<course>/` instead of the whole course (FR-055).
- **Success (stdout):** `wrote qa/<course>.md (15 pairs, engine=<engine>)`
- **Errors:** same set as `guide`, plus exit 1 if `--file` names a path not under
  the course, or `--count` is not a positive integer.

### `genius solve <course> --set <set> [--ex <list>] [--save]`  *(post-MVP)*
Solve specific exercises from a previously-ingested set, grounded in the course.
- **Args:** `<course>` — the course (grounding); **`--set`** — exercise set name
  under `exercises/<course>/`.
- **`--ex <list>`** — comma-separated exercise numbers/labels to solve, including
  sub-parts: `--ex 2,4` or `--ex 3.1,3.a` (FR-103). Omitted → print the enumerated
  list of exercises (and sub-parts) in the set and exit (so the learner can pick).
- **`--save`** — also write the solution(s) to `exercises/<course>/<set>.solutions.md`
  (never modifies the source set; INV-03 / FR-106).
- **Behavior:** read the set → enumerate exercises (ERR-101 if none) → resolve the
  selection (ERR-102 if a label is missing) → read course grounding (ERR-041 /
  INV-05) → `Generate.solve` → `Engine.Generate` → render to stdout/reader. Engine
  is instructed to solve *as stated* and flag course gaps (FR-105, ERR-105).
- **Success (stdout):** the worked solution(s), or the exercise list when `--ex`
  is omitted.
- **Errors:** ERR-041, ERR-081, ERR-082, plus ERR-101 (no exercises), ERR-102
  (unknown exercise), ERR-105 (gap flagged, not an error exit).

## Internal Go Interfaces

### `engine.Engine`
```go
// Generate returns only the final assistant text (INV-07).
// Describe captions a single image (vision); returns ErrNoVision if unsupported.
type Engine interface {
    Generate(ctx context.Context, system, user string) (string, error)
    Describe(ctx context.Context, imagePath, prompt string) (string, error)
}
var ErrNoVision = errors.New("engine has no image support")
```
- `claudeEngine` and `codexEngine` implement it; `fakeEngine` backs tests.
- Implementations MUST send content only to their backend (INV-11) and MUST strip
  backend log/framing from the returned string (INV-07).
- `Describe` for `codexEngine` uses `codex exec -i <image>`; `claudeEngine`'s image
  path is validated at build (ADR-06) and returns `ErrNoVision` if unsupported, so
  the Converter falls back to placeholders (EXC-12).

### `workspace.Workspace`
```go
type Workspace interface {
    Root() string
    EnsureLayout() error
    Courses() ([]Course, error)              // includes guide/qa/exercise counts
    ReadCourse(name string) (string, error)  // concatenated *.md grounding; "" if none
    WriteCourse(name, file string, md []byte, opts WriteOpts) error
    WriteGuide(name, content string, opts WriteOpts) error
    WriteQA(name, content string, opts WriteOpts) error
    WriteExercise(course, set string, md []byte, opts WriteOpts) error
    ExerciseSets(course string) ([]string, error)        // set names under exercises/<course>/
    ReadExerciseSet(course, set string) (string, error)  // raw set markdown
    Read(path string) (string, error)
}
type WriteOpts struct{ Overwrite bool } // Overwrite=false → confirm if exists (INV-03)
type Course struct {
    Name string
    Guides, QA, Exercises int
}
```

### `convert.Converter`
```go
type Converter interface {
    Available() error // nil if markitdown on PATH, else actionable error (INV-09)
    // ToMarkdown returns the markdown plus the paths of extracted image assets.
    // It captions figures via eng.Describe (or placeholders on ErrNoVision; INV-12)
    // and applies opts (describe on/off, ocr, min image size).
    ToMarkdown(ctx context.Context, path string, eng Engine, opts IngestOpts) (md []byte, assets []string, err error)
}
type IngestOpts struct {
    Describe     bool // caption figures (default true)
    OCR          bool // transcribe scanned/no-text docs (default false)
    MinImagePx   int  // skip images smaller than this (default 64)
}
```

### `generate` use-cases
```go
func Guide(ctx, ws Workspace, eng Engine, course string) (string, error) // INV-04/05

type QAOpts struct {
    Count int      // target number of pairs; 0 → default (15)
    Scope string   // optional free-text focus, "" → whole course
    Files []string // optional restriction to specific course md files; nil → all
}
func QA(ctx, ws Workspace, eng Engine, course string, opts QAOpts) (string, error)

// solve (post-MVP)
type Exercise struct {
    Label string     // bilingual: "Exercice 3" / "Exercise 3" / "Ex 3" / "Problème 2"
    Text  string     // the exercise statement
    Parts []Exercise // optional sub-questions (1., 2., (a), (b)…) addressable as "3.1", "3.a"
}
// Enumerate handles FR/EN headings (Exercice|Exercise|Ex|Problème|Problem + N) and
// nested numbered/lettered sub-parts. Ordered; ERR-101 if none found.
func Enumerate(setMarkdown string) ([]Exercise, error)
func Solve(ctx, ws Workspace, eng Engine, course string, exs []Exercise) (string, error)
```
`Guide`/`QA` MUST abort before calling `eng` if the resolved grounding (whole
course, or the `Files` subset) is empty (INV-05). `QA` injects `Count` and `Scope`
into the prompt template. `Solve` reads course grounding for `course` (INV-05),
builds a prompt with each selected exercise's text + grounding, and instructs the
Engine to solve as stated and flag gaps (FR-105).

### `render`
```go
func Markdown(md string, width int) (string, error) // Glamour → styled string
```

## Subprocess (Outbound) Invocations
| Tool | Command genius runs | Notes / retry |
|------|---------------------|---------------|
| markitdown | `markitdown <path>` | stdout = markdown text layer; non-zero exit → ERR-032. No retry. |
| image extract | PyMuPDF / python-pptx (helper) | Embedded images → `assets/`; skip < `MinImagePx`. |
| claude | `claude -p "<user>" --model <m> --append-system-prompt "<system>"` | stdout (text) = answer. No retry; surface error. |
| claude (vision) | image input path TBD (ADR-06) | `Describe`; `ErrNoVision` → placeholder fallback. |
| codex | `codex exec "<system+user>" -m <m>` (prompt via arg or stdin) | Extract final assistant message; discard log framing (INV-07). No retry. |
| codex (vision) | `codex exec -i <image> "<prompt>"` | `Describe` for figure captions / OCR. |
| tesseract (opt-in) | `tesseract <image> stdout` | OCR fallback for scanned docs when `--ocr` and no vision. |

All three are the only outbound paths; workspace content reaches none but the
learner-selected engine (INV-11).

## File-Format Reference

### Q&A file (`qa/<course>.md`)
Matches the learner's real revision-guide style (see `samples/Fundamentals_QA.md`).
```markdown
# <Title> – Study Guide          ← optional header block, ignored by the quiz loader
**Course:** ... **Instructor:** ...

## Q1. <question text on the heading line>
<rich answer markdown: paragraphs, tables, > blockquote "key rule" callouts,
$$ LaTeX $$, bold — everything until the next `## ` or `---`>

---

## Q2. <question text>
<answer …>

*<optional motivational footer>*        ← ignored by the loader
```
Loader contract (FR-061):
- A **question** is any `## Q<n>.` heading; the question text is the remainder of
  that heading line.
- The **answer** is all content from after that heading up to the next `## `
  heading, `---` rule, or EOF.
- Pairs are ordered by appearance. A non-`Q` `##` header block and a trailing
  italic footer are ignored, not errors.
- Answers may contain tables, blockquotes, and `$$…$$` LaTeX; the quiz renders
  them via Glamour (LaTeX shown as raw text in-terminal).
- Only a structural break (e.g. a `## Q` heading with no following answer content)
  fails fast with the offending heading (ERR-061).

### Figure caption (inline in course/exercise markdown)
Provenance-marked transcription of an extracted image (INV-12):
```markdown
> **[Figure 3]** _(transcribed from assets/lecture-3.png)_
> Bode plot: gain flat to ω≈1k rad/s, then rolls off at −20 dB/decade.
```
Placeholder form when no vision engine / describing disabled:
```markdown
> **[Figure 3]** _(image — assets/lecture-3.png; not described)_
```
Extracted images live in `courses/<course>/assets/` (or
`exercises/<course>/assets/`).

### Guide file (`guides/<course>.md`)
Free-form markdown with the sections: Summary, Key Concepts,
Formulas/Definitions, Common Traps, Worked Highlights.

### config.toml
```toml
study_root      = "/home/mj/study"
default_engine  = "claude"
model           = ""   # empty → backend default
```

## Command Summary Table
| Command | Reads | Writes | Engine? | Stage |
|---------|-------|--------|---------|-------|
| `genius` | workspace | — | no | MVP |
| `genius ingest` | a document | `courses/<name>/*.md` | no | MVP |
| `genius guide` | `courses/<name>/` | `guides/<name>.md` | yes | MVP |
| `genius qa` | `courses/<name>/` | `qa/<name>.md` | yes | MVP |
| `genius ingest --kind exercise` | a document | `exercises/<course>/<set>.md` | no | post-MVP |
| `genius solve` | exercise set + course | stdout/reader (+`--save`) | yes | post-MVP |
