# genius — Transition: Requirements to Architecture
Version: v1.0, 2026-06-24

## Method
Every guarantee from `03_design_contract_invariant.md` is assigned to exactly one
component owner — a conceptual responsibility, not a library. After this mapping,
any contract violation points to a single component whose code is opened.

## Component Definitions
- **Workspace** — Owns the study home: root resolution, directory creation,
  course discovery, progress counting, and all filesystem reads/writes of
  artifacts. It is the only component that touches study files on disk. It knows
  nothing about the TUI, engines, or markdown rendering.
- **Engine** — Owns generation. An interface with `claude` and `codex`
  implementations. It builds the prompt (system + course grounding + task),
  invokes the backend subprocess, and returns clean final text. It knows nothing
  about where files live.
- **Converter** — Owns document-to-markdown conversion via the `markitdown`
  subprocess, including dependency detection. It produces markdown bytes; it does
  not decide where they are filed (Workspace does).
- **Generate** — Owns prompt assembly and the generation use-cases (guide, qa,
  solve). It reads course material through Workspace, asks Engine to generate,
  and hands the result back to Workspace to persist.
- **Render** — Owns turning markdown into styled terminal output via Glamour.
- **TUI** — Owns the interactive environment: the Bubble Tea program, all
  states (home, course, reader, generate, quiz), keyboard handling, spinners,
  terminal setup/teardown, and the quiz loop logic.
- **CLI/Root** — Owns process entry: Cobra command tree, flag parsing, config
  loading, engine selection, and dispatch to either the TUI or a subcommand.

## Invariant Assignments

### Workspace (owns: INV-01, INV-02, INV-03)
It is the sole component touching study files, so single-rooted access (INV-01),
course-identity consistency (INV-02), and non-destructive writes incl. the
overwrite-confirmation gate (INV-03) are enforced here by construction — no other
component has a path to the filesystem for artifacts.

### Generate (owns: INV-04, INV-05)
Course grounding (INV-04) and grounding provenance (INV-05) are properties of how
the prompt is assembled and when generation is allowed. Generate refuses to call
Engine when Workspace reports no material, and always supplies the material as
context with grounding instructions.

### Engine (owns: INV-06, INV-07, INV-11)
Substitutability (INV-06) is the interface's reason to exist. Clean output
(INV-07) is enforced where the subprocess is read. Outbound confinement (INV-11)
lives here because Engine is the only component that sends content off-box — it
sends only to the selected backend.

### Converter (owns: INV-09 for `markitdown`, INV-12)
Detects the absence/failure of `markitdown` and surfaces it (shares INV-09 with
CLI/Root, scoped to the conversion dependency). Owns figure fidelity & provenance
(INV-12): it extracts embedded images, calls the Engine's `Describe` for captions,
writes provenance-marked descriptions or asset placeholders into the markdown, and
handles scanned-document detection. The *content* of a caption comes from the
Engine; the *guarantee that figures are never silently dropped and descriptions
are marked as transcriptions* is enforced here.

### TUI (owns: INV-08, INV-10)
Terminal restoration (INV-08) is owned by the only component that puts the
terminal into raw mode. No-content-execution (INV-10) is enforced wherever
external text is handled for display/quizzing — it is rendered/printed, never
shelled.

### CLI/Root (owns: INV-09)
Top-level graceful-degradation (INV-09): any dependency error or component error
bubbling to the entry point becomes a clear message and a non-zero exit.

## Invariant Coverage Table
| Invariant | Owner | Enforcement Point |
|-----------|-------|-------------------|
| INV-01 Single rooted workspace | Workspace | Root resolved once; all paths derived from it |
| INV-02 Course identity stability | Workspace | `<name>` derived from `courses/` dir; reused for all artifacts |
| INV-03 Non-destructive writes | Workspace | Write only declared output; confirm-before-overwrite |
| INV-04 Course grounding | Generate | Course markdown injected as context + grounding instruction |
| INV-05 Grounding provenance | Generate | Abort if Workspace returns no material |
| INV-06 Engine substitutability | Engine | Single `Engine` interface; impls are interchangeable |
| INV-07 Clean engine output | Engine | Final-answer extraction when reading subprocess stdout |
| INV-08 Terminal restoration | TUI | Bubble Tea teardown + recover on panic/interrupt |
| INV-09 Graceful dependency absence | CLI/Root (+ Converter) | Errors mapped to messages + exit codes |
| INV-10 No content execution | TUI | External text only rendered/printed |
| INV-11 Outbound confinement | Engine | Content sent only to selected backend subprocess |
| INV-12 Figure fidelity & provenance | Converter | Extract images; caption via Engine.Describe; mark transcriptions; placeholder fallback |

## Coupling & Cohesion Decisions
- **Workspace is the only filesystem owner.** Centralizing artifact I/O is what
  makes INV-01/02/03 enforceable by signature rather than by discipline. Generate
  and Converter produce bytes; Workspace decides placement.
- **Converter is separate from Generate.** Conversion (`markitdown`) and
  generation (`claude`/`codex`) are different external dependencies with different
  failure modes; splitting them keeps INV-09 ownership crisp and lets ingest work
  without an engine present.
- **Render is separate from TUI.** Markdown-to-styled-string is a pure transform
  reusable by both the TUI reader and any future CLI preview, so it is not buried
  inside Bubble Tea view code.
- **Engine owns confinement (INV-11), not Generate.** Generate decides *what* to
  send; Engine is the only place content actually leaves the process, so the
  "only the selected backend" guarantee belongs with it.
