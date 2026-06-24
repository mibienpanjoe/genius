# genius — Software Requirements Specification
Version: v1.0, 2026-06-24

## Normative Vocabulary
- **MUST / REQUIRED**: Absolute requirement. The system fails if not met.
- **SHOULD**: Recommended. Deviation is permitted with documented justification.
- **MAY**: Optional capability.

## Actors
| Actor | Description |
|-------|-------------|
| Learner | The single human user running genius locally |
| TUI | The Bubble Tea terminal interface genius presents |
| CLI | The non-interactive subcommand surface (`genius ingest`, etc.) |
| Engine | A generation backend subprocess (`claude` or `codex`) behind one interface |
| Converter | The `markitdown` subprocess that turns documents into markdown |
| Workspace | The study home directory and its course/guide/qa/exercise tree |

## Functional Requirements

### FR-010: Workspace Resolution
- **FR-011**: The system MUST resolve the workspace root from `$GENIUS_HOME` if
  set, otherwise from `study_root` in the config file, otherwise default to
  `~/study`.
- **FR-012**: The system MUST create the workspace root and the
  `courses/`, `guides/`, `qa/`, `exercises/` subdirectories if they do not exist.
- **FR-013**: The system MUST treat each directory under `courses/` as one course
  identified by its directory name.

### FR-020: Environment & Navigation
- **FR-021**: Running `genius` with no subcommand MUST launch the TUI at the home
  dashboard.
- **FR-022**: The home dashboard MUST list every course with its guide, Q&A, and
  exercise counts and MUST display the active engine and workspace root.
- **FR-023**: The TUI MUST allow navigation from home to a per-course view and
  from there to a markdown reader, and back, via keyboard.
- **FR-024**: The TUI MUST quit cleanly on the documented quit key, restoring the
  terminal state.

### FR-030: Ingest
- **FR-031**: The system MUST convert a learner-specified document file to
  markdown by invoking the Converter.
- **FR-032**: The system MUST write the converted markdown under
  `courses/<name>/` where `<name>` is the learner-specified or filename-derived
  course name.
- **FR-033**: The system MUST detect when the Converter is not installed and MUST
  report a clear message including the install command, without crashing.
- **FR-034**: The system MUST NOT overwrite an existing course markdown file
  without the learner's confirmation.

### FR-035: Image & Figure Handling (during ingest)
- **FR-035a**: During ingest, the system MUST extract embedded images/figures from
  the source document to `courses/<name>/assets/` (and `exercises/<course>/assets/`
  for exercise sets), skipping images below a configurable size threshold
  (decorative/bullets/logos).
- **FR-035b**: By default, the system MUST describe each extracted figure using a
  vision-capable Engine and insert the description inline in the markdown, marked
  as a model transcription with a link to the saved asset (provenance — INV-12).
- **FR-035c**: When no vision-capable Engine is available, or describing is
  disabled, the system MUST insert a placeholder referencing the saved asset
  (e.g. `[image omitted — assets/<file>]`) and MUST NOT silently drop the figure.
- **FR-035d**: The system MUST detect documents/pages with no extractable text
  layer (scanned/image-only) and, when `--ocr` is enabled, MUST transcribe them
  via OCR or a vision Engine; otherwise it MUST report that the document has no
  extractable text and SHOULD suggest `--ocr`.
- **FR-035e**: Figure descriptions and OCR transcriptions MUST describe only what
  is in the source image (INV-04); they MUST NOT invent content not present.
- **FR-035f**: Ingest MUST preserve semantically critical math/logic notation —
  Boolean complement bars (`X̄`), quantifiers (`∀ ∃`), connectives (`¬ ∧ ∨ → ↔`),
  subscripts (INV-12). For notation-dense pages where the text layer is lossy
  (LaTeX-typeset PDFs), the system SHOULD transcribe via a vision Engine to
  Unicode/LaTeX and/or flag uncertain notation, rather than emit corrupted symbols
  (e.g. losing a complement bar, which inverts meaning).

### FR-040: Study Guide Generation
- **FR-041**: The system MUST read all markdown under `courses/<name>/` and pass
  it to the Engine as grounding when generating a guide.
- **FR-042**: The system MUST produce a guide containing summary, key concepts,
  formulas/definitions, common traps, and worked highlights.
- **FR-043**: The system MUST write the guide to `guides/<name>.md`.
- **FR-044**: The system MUST NOT overwrite an existing guide without the
  learner's confirmation.

### FR-050: Revision Q&A Generation
- **FR-051**: The system MUST read all markdown under `courses/<name>/` and pass
  it to the Engine as grounding when generating Q&A.
- **FR-052**: The system MUST produce Q&A as markdown in the learner's revision
  style: a `## Q<n>. <question>` heading followed by a rich answer (paragraphs,
  comparison tables, `>` key-rule callouts, `$$…$$` formulas) until the next
  heading/rule. The exact format is documented in the CLI spec (`qa/<course>.md`)
  and exemplified by `samples/Fundamentals_QA.md`.
- **FR-053**: The system MUST write the Q&A to `qa/<name>.md`.
- **FR-054**: The system MUST let the learner specify the **number** of Q&A pairs
  to generate. It MUST apply a sensible default when unspecified and MUST treat
  the number as a target the Engine is instructed to honor.
- **FR-055**: The system SHOULD let the learner **scope** generation to part of
  the course — either a free-text focus (e.g. a topic or section) passed to the
  Engine as a constraint, and/or a restriction to specific source markdown files
  under `courses/<name>/`. When a scope is given, the Engine MUST be instructed to
  draw only from that part (grounding still applies — INV-04).
- **FR-056**: In the TUI, choosing Q&A generation MUST prompt for the count and an
  optional scope before running; in the CLI these are flags.

### FR-060: Interactive Revise / Quiz
- **FR-061**: The system MUST load and parse a Q&A file into an ordered list of
  question/answer pairs.
- **FR-062**: The quiz MUST present one question at a time, accept a free-text
  answer, then reveal the model answer for self-grading.
- **FR-063**: The quiz MUST let the learner advance to the next question and MUST
  detect the end of the set.
- **FR-064**: The quiz MUST present an end-of-session summary (questions seen, and
  self-graded result if recorded).
- **FR-065**: The quiz MUST allow the learner to exit early and return to the TUI
  without corrupting any file.

### FR-070: Rendering
- **FR-071**: The system MUST render generated markdown to styled terminal output
  in-process (no external pager process required).
- **FR-072**: The reader MUST support scrolling through content longer than the
  viewport.

### FR-080: Engine Selection
- **FR-081**: The system MUST route all generation through the `Engine` interface.
- **FR-082**: The system MUST default the engine to `claude` and MUST allow
  overriding it per invocation via `--engine`.
- **FR-083**: The system MUST allow a default engine and model to be set in config.
- **FR-084**: The Engine implementation MUST return only the final assistant text,
  stripping any backend log/framing output.

### FR-090: CLI Subcommands
- **FR-091**: The system MUST expose `ingest`, `guide`, `qa`, and (post-MVP)
  `solve` as non-interactive subcommands mirroring the TUI actions.
- **FR-092**: Subcommands MUST report success/failure with non-zero exit codes on
  failure for scriptability.

### FR-100: Exercise Solving *(next iteration — post-MVP)*
- **FR-101**: The system MUST let the learner add an exercise document (TD,
  exercise set, past test) to a course by converting it to markdown and filing it
  under `exercises/<course>/<set>.md`.
- **FR-102**: The system MUST enumerate the individual exercises within a set into
  an addressable, ordered list (by exercise number/label/heading) for selection.
- **FR-103**: The system MUST let the learner select one or more specific
  exercises from a set to solve (by number/label), rather than only solving the
  whole sheet.
- **FR-104**: For each selected exercise, the system MUST produce a worked
  solution grounded in the course material (INV-04): the answer with reasoning
  steps, referencing the relevant course concepts.
- **FR-105**: The system MUST solve the exercise **as stated** in the set; it MUST
  NOT substitute a different problem, and MUST flag when the course material does
  not cover what the exercise requires rather than fabricating.
- **FR-106**: The system MUST NOT modify the uploaded exercise file when solving
  (INV-03); solutions are presented in the reader and MAY be saved to a separate
  output file on request.
- **FR-107**: In the TUI, the learner MUST be able to open a course, pick an
  exercise set, see its enumerated exercises, select one or more, and view the
  solution in the reader; the CLI MUST expose the same via flags.

## Business Rules (BR)
- **BR-01**: All generated artifacts derive from course material in the
  workspace; the Engine receives that material as grounding context.
- **BR-02**: A course's identity is its directory name under `courses/`; all of
  that course's guides and Q&A use the same `<name>`.
- **BR-03**: The workspace is the single source of truth for all study artifacts;
  genius does not store study state elsewhere.

## Non-Functional Constraints

### Performance
- TUI input-to-render latency SHOULD be < 50 ms for navigation actions.
- Generation latency is bounded by the Engine subprocess; the TUI MUST show a
  spinner/progress indicator for any operation expected to exceed 1 second.

### Availability
- genius is a local single-process tool; it MUST degrade gracefully (clear error,
  clean exit) when an external dependency (`markitdown`, `claude`, `codex`) is
  absent or fails.

### Security & Safety
- The system MUST NOT delete or truncate workspace files except the specific
  output it is asked to write, and only after the confirmation rules above.
- The system MUST NOT execute arbitrary content from course files or Engine
  output as shell commands.

### Data Privacy
- Course content is sent to whichever Engine subprocess the learner selects; the
  system MUST NOT send workspace content to any destination other than the
  selected Engine.

### Portability
- The system MUST build and run as a single static binary on Linux with Go 1.25.
  It SHOULD avoid OS-specific assumptions beyond a POSIX filesystem and a TTY.

### Usability
- First run MUST auto-create the workspace so the learner is never blocked by a
  missing directory.

## Error Cases (ERR)
| ID | Trigger | Required Behavior |
|----|---------|-------------------|
| ERR-031 | `markitdown` not on PATH during ingest | Report missing-dependency message with install hint; exit non-zero (CLI) or show styled error (TUI) |
| ERR-032 | Ingest target file does not exist or is unsupported | Report the path/format problem; do not write partial output |
| ERR-035 | Document has no extractable text and `--ocr` not set | Report it; suggest re-running with `--ocr`; write nothing |
| ERR-036 | Vision describe requested but engine lacks image support | Fall back to placeholders; warn once; continue ingest |
| ERR-034 | Output file already exists (ingest/guide/qa) | Prompt for confirmation; abort write if declined |
| ERR-041 | No markdown found under `courses/<name>/` for generation | Report that the course has no material to ground on; do not call Engine |
| ERR-081 | Selected Engine binary not on PATH | Report which engine is missing; do not hang |
| ERR-082 | Engine subprocess exits non-zero or times out | Surface the error, write nothing, leave prior artifacts intact |
| ERR-084 | Engine output contains log/framing around the answer | Extract and keep only the final assistant text |
| ERR-061 | Q&A file is unparseable for the quiz | Report the parse problem with the offending location; do not enter a broken quiz |
| ERR-101 | Exercise set has no enumerable exercises | Report it; let the learner solve the whole set as one item instead |
| ERR-102 | Selected exercise number/label not found in the set | Report the available items; do not call Engine |
| ERR-105 | Course material does not cover the selected exercise | Engine instructed to state the gap explicitly; no fabricated solution |
