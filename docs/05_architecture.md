# genius — System Architecture
Version: v1.0, 2026-06-24

## Architectural Style
**Modular monolith — a single Go binary with clearly bounded internal packages.**
genius is one process that either launches a Bubble Tea TUI or runs a Cobra
subcommand and exits. Internal packages (`workspace`, `engine`, `convert`,
`generate`, `render`, `tui`) have crisp responsibilities and one-directional
dependencies.

Chosen over alternatives because:
- It is a **local single-user tool** — there is no scale, tenancy, or network
  surface that would justify multiple services.
- External capabilities (`markitdown`, `claude`, `codex`) are already separate
  processes; genius orchestrates them, so its own code can be one cohesive unit.
- A single static binary is the simplest possible distribution (`go build`, copy,
  run) and matches the "drop it anywhere" goal.

## Component Architecture

### Workspace (`internal/workspace`)
**Responsibility:** Own the study home — root resolution, directory creation,
course discovery, progress counts, and all artifact reads/writes.
**Owned invariants:** INV-01, INV-02, INV-03.
**Inputs:** `$GENIUS_HOME`/config/default; course names; bytes to persist.
**Outputs:** Resolved paths, course list with counts, course markdown, write
results.
**Key behaviors:**
1. Resolve root once: `$GENIUS_HOME` → config `study_root` → `~/study`.
2. Ensure `courses/ guides/ qa/ exercises/` exist.
3. List courses (dirs under `courses/`) and compute guide/Q&A/exercise counts.
4. Read all `*.md` under `courses/<name>/` as a single grounding blob.
5. Write a declared output file; if it exists, require confirmation before
   replacing it.
**Must NOT:** Touch any path outside the root, or modify any file other than the
current operation's declared output.

### Engine (`internal/engine`)
**Responsibility:** Generate text from a prompt via a swappable backend.
**Owned invariants:** INV-06, INV-07, INV-11.
**Interface:** `Engine.Generate(ctx, system, user string) (string, error)` and
`Engine.Describe(ctx, imagePath, prompt string) (string, error)` (vision; may
return `ErrNoVision` if the backend has no image support).
**Implementations:**
- `claudeEngine` → `claude -p "<user>" --model <m> --append-system-prompt "<system>"`;
  stdout (default text format) is the answer. Image path for `Describe` validated
  at build (see ADR-06); returns `ErrNoVision` if unsupported.
- `codexEngine` → `codex exec "<combined system+user>" -m <m>` (prompt via arg or
  stdin); `Describe` uses `codex exec -i <image>`; extract the final assistant
  message, discarding any log framing.
- `fakeEngine` (tests) → deterministic canned output for both methods.
**Must NOT:** Decide file locations, or send content to anything but the selected
backend.

### Converter (`internal/convert`)
**Responsibility:** Convert a document to markdown via `markitdown`, and handle
its images/figures and scanned pages.
**Owned/Shared invariants:** INV-09 (conversion dependency), INV-12 (figure
fidelity & provenance).
**Behavior:**
1. Check `markitdown` on PATH (clear message + install hint if missing); run
   `markitdown <path>` for the text layer.
2. Extract embedded images (PyMuPDF for PDF, python-pptx for PPTX) to an
   `assets/` directory, skipping images under the size threshold.
3. For each kept image, call `Engine.Describe` (vision) and splice a
   provenance-marked caption inline; if the engine has no image support or
   describing is disabled, splice an asset-referencing placeholder instead
   (INV-12 / EXC-12).
4. Detect no-text-layer (scanned) documents; with `--ocr`, transcribe via vision
   Engine or tesseract; without it, stop and suggest `--ocr` (ERR-035).
5. Preserve semantically critical notation (complement bars, quantifiers,
   connectives, subscripts; INV-12/FR-035f). For notation-dense pages where the
   text layer is lossy, prefer a vision transcription to Unicode/LaTeX and/or flag
   uncertain spots — never emit a silently inverted/dropped symbol.
Returns markdown bytes plus the list of extracted asset paths.
**Must NOT:** Decide where the markdown/assets are filed (Workspace does that), or
silently drop a figure.

### Generate (`internal/generate`)
**Responsibility:** Assemble prompts and run the guide / qa / solve use-cases,
including enumerating exercises within an exercise set.
**Owned invariants:** INV-04, INV-05.
**Behavior:** Ask Workspace for `<name>`'s grounding markdown; if empty, abort
(INV-05); otherwise build `(system, user)` from an embedded template + grounding +
task; call `Engine.Generate`; hand the result to Workspace to persist (or to the
reader, for solve). For **solve**: parse an exercise set into an ordered,
addressable list (by number/label/heading); for the selected exercise(s), build a
prompt with the exercise text + course grounding, instructing the Engine to solve
it *as stated* and flag any course gap (INV-04, FR-105).
**Must NOT:** Write files directly, or call a backend directly (only via Engine).

### Render (`internal/render`)
**Responsibility:** Convert markdown to a styled terminal string via Glamour.
**Behavior:** Configure a Glamour renderer matched to the theme; return a styled
string for a Bubbles `viewport`.
**Must NOT:** Perform any I/O beyond rendering.

### TUI (`internal/tui`)
**Responsibility:** The interactive environment — the Bubble Tea program and all
states.
**Owned invariants:** INV-08, INV-10.
**States (Elm model):** `home`, `course`, `reader`, `generate`, `quiz`,
`exercises` (pick a set → enumerated exercises → select → solve → reader).
**Behavior:** Render the active state; handle keys; show a spinner during Engine
calls; run the quiz loop (present → answer → reveal → advance → summary); set up
and tear down raw mode; recover on panic/interrupt and restore the terminal. The
`home` state renders a gradient ASCII wordmark banner (see doc 07); when the
workspace is empty it shows getting-started tips + a notice box, otherwise the
course list.
**Must NOT:** Bypass Workspace for files, or execute any external text.

### CLI/Root (`main.go`)
**Responsibility:** Process entry — Cobra tree, flag/config parsing, engine
selection, dispatch.
**Owned invariant:** INV-09.
**Behavior:** Load config; resolve engine (`--engine` over config default `claude`);
no subcommand → launch TUI; subcommand → run `ingest`/`guide`/`qa`/(`solve`) and
exit with a status code; map errors to clear messages.

## Data Architecture

### Entities (filesystem-backed, no database)
- **Workspace root** — one directory; resolution order `$GENIUS_HOME` → config →
  `~/study`.
- **Course** — a directory `courses/<name>/` containing one or more `*.md` plus an
  `assets/` subfolder of figures extracted during ingest. The name is the identity
  (INV-02). 1 course → N source markdown files + N image assets.
- **Asset** — an image extracted from a source document, stored under the owner's
  `assets/`, referenced by a provenance-marked caption in the markdown (INV-12).
- **Guide** — `guides/<name>.md`, 0..1 per course.
- **Q&A** — `qa/<name>.md`, 0..1 per course; an ordered list of question/answer
  pairs in the documented markdown format.
- **Exercise set** — `exercises/<name>/*`, 0..N files per course.
- **Progress** (post-MVP) — `.genius/progress.json`, quiz weak-spot records keyed
  by course.

### Key constraints
- Course name uniqueness = directory uniqueness under `courses/`.
- Guide/Q&A filenames are derived from the course name; one canonical path each.
- All paths are children of the resolved root (INV-01).

## Flow Architecture

### Ingest
```
genius ingest lecture.pdf  (or TUI: i)
  └─► Converter.check()                 → ERR-031 if markitdown missing
        ↓ present
  └─► Converter.toMarkdown(path)
        ├─ markitdown → text layer
        ├─ extract embedded images → assets/  (skip < threshold)
        ├─ for each figure: Engine.Describe → provenance caption
        │     └─ no vision / disabled → asset placeholder (EXC-12)
        └─ no text layer? --ocr → transcribe, else ERR-035
        ↓ markdown bytes + asset paths
  └─► Workspace.writeCourse(name, md, assets)  → confirm if exists (INV-03)
        → courses/<name>/lecture.md  +  courses/<name>/assets/*
```

### Generate guide / Q&A
```
genius guide algebra  (or TUI: g)
  └─► Workspace.readCourse("algebra")   → grounding md (ERR-041 / INV-05 if empty)
        ↓
  └─► Generate.guide(grounding)         → (system, user) prompt
        ↓
  └─► Engine.Generate(ctx, system, user)   ← spinner in TUI
        │   claude -p ... | codex exec ...
        └─► clean final text (INV-07)
        ↓
  └─► Workspace.writeGuide("algebra", text)  → guides/algebra.md (confirm if exists)
        ↓
  └─► (TUI) progress recount → home refresh
```
**Latency budget:** sync filesystem/path work < 50 ms; generation bounded by the
Engine subprocess (spinner shown for any op > 1 s).

For Q&A, the TUI first shows a Huh form collecting **count** (default 15) and an
optional **scope** (focus text and/or source-file subset) before running the
generate flow above; the CLI takes these as `--count`/`--scope`/`--file`.

### Read / render
```
TUI reader ← Workspace.read(path) → Render.markdown(md) → viewport (scroll)
```

### Quiz (revise)
```
genius (TUI) → quiz state
  └─► Workspace.read(qa/<name>.md) → parse pairs (ERR-061 if invalid)
        ↓ loop
  present question → capture answer → reveal model answer → advance
        ↓ end
  summary → back to home
```

### Solve an exercise *(post-MVP)*
```
1. add the set:  genius ingest td3.pdf --kind exercise --course algebra
      └─► Converter.toMarkdown → Workspace.writeExercise("algebra","td3", md)
            → exercises/algebra/td3.md

2. solve:  genius solve algebra --set td3 --ex 2,4   (or TUI: course → exercises → pick)
      └─► Workspace.readExerciseSet("algebra","td3")     → set markdown
            ↓
      └─► Generate.enumerate(set)   → ordered list [Ex1, Ex2, …]  (ERR-101 if none)
            ↓ select Ex2, Ex4 (ERR-102 if missing)
      └─► Workspace.readCourse("algebra")   → grounding (INV-05)
            ↓
      └─► Generate.solve(exercise, grounding)   → prompt: solve as stated, flag gaps
            ↓
      └─► Engine.Generate(ctx, system, user)    ← spinner
            └─► clean final text (INV-07)
            ↓
      └─► Render → reader   (optionally Workspace.write solution on request; INV-03)
```

## Technology Mapping
| Capability | Technology | Component |
|-----------|------------|-----------|
| Language/runtime | Go 1.25, single static binary | all |
| CLI tree / flags | Cobra (optionally Charm Fang for help) | CLI/Root |
| TUI framework | Bubble Tea (Elm architecture) | TUI |
| Styling | Lip Gloss | TUI, Render |
| Components | Bubbles (list, viewport, spinner, textinput) | TUI |
| In-TUI prompts | Huh | TUI |
| Markdown rendering | Glamour | Render |
| Document conversion | `markitdown` (Python CLI, subprocess) | Converter |
| Image extraction | PyMuPDF (PDF), python-pptx (PPTX), via subprocess/helper | Converter |
| Figure captioning | vision Engine (`Engine.Describe`) | Converter + Engine |
| Scanned-doc OCR (opt-in) | vision Engine or `tesseract` | Converter |
| Generation backends | `claude -p`, `codex exec` (subprocess) | Engine |
| Config | TOML at `~/.config/genius/config.toml` | CLI/Root |
| Prompt templates | `//go:embed` of `prompts/` | Generate |

## Deployment Architecture
A single binary `genius` on the learner's Linux machine. Runtime prerequisites on
PATH: `markitdown` (for ingest), and at least one of `claude`/`codex` (for
generation). No daemon, no ports, no server. External network calls happen only
inside the backend subprocesses the learner invoked.

## Project Structure
```
genius/
  go.mod                      module genius
  main.go                     Cobra root; no args → TUI; subcommands
  internal/
    workspace/                root resolution, scan, counts, artifact I/O (+tests)
    engine/                   Engine iface + claude.go + codex.go + fake.go
    convert/                  markitdown wrapper, image extraction, captioning, OCR
    generate/                 prompt assembly: guide / qa / solve
    render/                   glamour helpers
    tui/                      model.go home.go course.go reader.go generate.go quiz.go styles.go
  prompts/                    embedded templates (guide.md, qa.md, solve.md)
  docs/                       this documentation suite
  project_overview.md         source overview
```

## Invariant Traceability Matrix
| Invariant | Owner (doc 04) | Architecture enforcement |
|-----------|----------------|--------------------------|
| INV-01 | Workspace | Root resolved once; all paths derived from it |
| INV-02 | Workspace | `<name>` from `courses/` dir reused for guide/qa/exercise |
| INV-03 | Workspace | Writes limited to declared output; confirm-before-overwrite |
| INV-04 | Generate | Grounding md injected with grounding instruction in prompt |
| INV-05 | Generate | Abort generation when `readCourse` returns empty |
| INV-06 | Engine | One `Engine` interface; `claude`/`codex` interchangeable |
| INV-07 | Engine | Final-answer extraction from subprocess stdout |
| INV-08 | TUI | Bubble Tea teardown + panic/interrupt recovery |
| INV-09 | CLI/Root, Converter | Dependency/component errors → message + exit code |
| INV-10 | TUI | External text rendered/printed, never shelled |
| INV-11 | Engine | Content sent only to the selected backend subprocess |
| INV-12 | Converter | Extract figures; caption via Engine.Describe; mark transcriptions; placeholder fallback; never drop |

## Functional Requirement Traceability (selected)
| FR group | Component |
|----------|-----------|
| FR-010 Workspace resolution | Workspace |
| FR-020 Environment & navigation | TUI, CLI/Root |
| FR-030 Ingest | Converter + Workspace |
| FR-040 Guide generation | Generate + Engine + Workspace |
| FR-050 Q&A generation | Generate + Engine + Workspace |
| FR-060 Quiz | TUI + Workspace |
| FR-070 Rendering | Render + TUI |
| FR-080 Engine selection | Engine + CLI/Root |
| FR-090 Subcommands | CLI/Root |

## Architectural Constraints & ADRs
- **ADR-01 — Go + Charm over a Claude Code plugin.** A plugin is claude-only and
  gives no dedicated "place"; a Go/Charm TUI delivers the immersive environment
  and keeps both `claude` and `codex` as swappable engines. Cost: more code than
  a plugin. Accepted.
- **ADR-02 — Glamour in-process over shelling to `glow`.** Same render engine,
  one fewer external dependency, content stays inside the TUI viewport.
- **ADR-03 — Stuff full course markdown (no RAG) for MVP.** Typical lectures fit
  in model context; chunking/retrieval is deferred until a course overflows.
- **ADR-04 — Workspace as sole filesystem owner.** Makes INV-01/02/03 enforceable
  by construction rather than convention.
- **ADR-05 — Subprocess engines behind one interface.** Backends evolve
  independently; genius depends only on the `Engine` contract, and must strip
  backend framing to satisfy INV-07.
- **ADR-06 — Extract + vision-caption images on ingest.** Study figures (diagrams,
  graphs, equation images) are core content; text-only extraction loses them. We
  extract embedded images and caption them via a vision Engine, inserting
  provenance-marked transcriptions (INV-12). Cost: extra vision calls per figure
  and an image-extraction dependency. **Build risk:** the `claude` CLI's image
  input path is unconfirmed — `Engine.Describe` is validated per backend at build;
  unsupported backends fall back to placeholders (EXC-12). OCR for scanned docs is
  opt-in (`--ocr`) to avoid taxing every ingest.
- **ADR-07 — Notation fidelity is a first-class ingest concern (INV-12/FR-035f).**
  Validated against `samples/`: the course and TD are LaTeX-typeset logic where
  `markitdown` text extraction tends to drop Boolean complement bars (`X̄`→`X`,
  an inverted meaning) and mangle quantifiers/connectives. Mitigation: for
  notation-dense pages, prefer vision transcription to Unicode/LaTeX and/or flag
  uncertain notation; never emit a silently corrupted symbol. This is the highest
  grounding risk for this domain and must be exercised against the samples before
  the ingest pipeline is considered done.
- **ADR-08 — Q&A house style matches the learner's own.** The `qa` prompt
  template emulates `samples/Fundamentals_QA.md`: `## Q<n>.` headings, rich answers
  with comparison tables, `>` key-rule callouts, and `$$` formulas. The quiz
  loader parses exactly that shape (see CLI spec).
