# genius — System Contract & Invariants
Version: v1.0, 2026-06-24

## Actors & Allowed Actions
| Actor | Permitted to |
|-------|--------------|
| Learner | Run commands, choose a course/engine, type quiz answers, confirm overwrites |
| TUI | Read workspace, render markdown, invoke ingest/generate/quiz, display state |
| CLI | Run a single action non-interactively and exit with a status code |
| Engine | Receive a prompt (system + course grounding + task), return generated text |
| Converter | Receive a document path, return markdown |
| Workspace | Hold all study artifacts under a single resolved root |

## System Guarantees (Invariants)

### Workspace Integrity
**INV-01 — Single Rooted Workspace**
All reads and writes of study artifacts MUST occur under one resolved workspace
root. genius MUST NOT read or write study data outside that root.

**INV-02 — Course Identity Stability**
A course is identified by its directory name under `courses/`. Every artifact for
that course (`guides/<name>.md`, `qa/<name>.md`, `exercises/<name>/`) shares that
exact `<name>`. The mapping is consistent across the whole system.

**INV-03 — Non-Destructive Writes**
genius MUST only create or replace the specific output file of the current
operation, and only after the overwrite-confirmation rule is satisfied. No other
file may be modified, deleted, or truncated as a side effect.

### Generation Grounding
**INV-04 — Course Grounding**
Any generated guide, Q&A, or exercise solution MUST be produced from the course
material supplied to the Engine as context. The Engine MUST be instructed to base
its output on that material, not on unrelated outside knowledge, and MUST NOT
fabricate course facts (formulas, definitions, claims) absent from the source.
When solving an exercise, the output MUST address the exercise **as stated** (no
substituted problem), and MUST explicitly flag any gap where the course material
does not cover what the exercise requires rather than inventing it.

**INV-05 — Grounding Provenance**
Generation MUST NOT proceed for a course that has no source markdown. The system
never produces an artifact with no grounding material behind it.

**INV-12 — Source Fidelity & Provenance**
Ingest MUST preserve the meaning of the source, including **semantically critical
notation** — Boolean complement bars (`X̄` vs `X`), quantifiers (`∀`, `∃`),
logical connectives (`¬ ∧ ∨ → ↔`), and subscripts/superscripts. Where the
extracted text layer is lossy or ambiguous for such notation (common with
LaTeX-typeset PDFs), the system MUST recover it via vision transcription and/or
flag the uncertainty for review rather than emit corrupted notation. Likewise,
any figure description or OCR transcription MUST be derived from the actual source
image and MUST be marked as a model-generated transcription referencing the saved
asset. The system MUST NOT present invented content as original course text, MUST
NOT silently invert or drop notation, and MUST NOT discard an extracted figure
without leaving at least a placeholder reference to its asset.

### Engine Abstraction
**INV-06 — Engine Substitutability**
All generation MUST flow through the single `Engine` interface. Swapping the
concrete engine (`claude` ↔ `codex`) MUST NOT change which files are written,
where, or the workflow around them.

**INV-07 — Clean Engine Output**
The text genius persists from an Engine MUST be only the final assistant answer,
with any backend logs, reasoning traces, or framing removed.

### Environment & Safety
**INV-08 — Terminal Restoration**
However genius exits (normal, error, or interrupt), it MUST restore the terminal
to a usable cooked state; it must never leave the user's shell corrupted.

**INV-09 — Graceful Dependency Absence**
A missing or failing external dependency (`markitdown`, `claude`, `codex`) MUST
produce a clear, actionable message and a clean exit — never a hang or a stack
dump as the only output.

**INV-10 — No Content Execution**
Course file contents and Engine output MUST be treated strictly as data. genius
MUST NOT execute them as shell commands or code.

### Confinement
**INV-11 — Outbound Confinement**
Workspace content MUST leave the machine only via the Engine subprocess the
learner explicitly selected. genius MUST NOT transmit study content to any other
destination.

## Absolute Prohibitions (FRB)
| ID | The system MUST NEVER... |
|----|--------------------------|
| FRB-01 | Read or write study artifacts outside the resolved workspace root |
| FRB-02 | Delete or truncate a workspace file that is not the current operation's declared output |
| FRB-03 | Overwrite an existing guide/Q&A/course file without explicit confirmation |
| FRB-04 | Generate a guide/Q&A/answer for a course that has no source material |
| FRB-05 | Persist Engine output that still contains backend log or framing text |
| FRB-06 | Execute course-file contents or Engine output as commands |
| FRB-07 | Send workspace content anywhere other than the learner-selected Engine |
| FRB-08 | Exit leaving the terminal in a raw/unusable state |
| FRB-09 | Silently drop an extracted figure, or present an invented figure description as original course text |
| FRB-10 | Silently corrupt, invert, or drop semantically critical notation (complement bars, quantifiers, connectives) during ingest |

## Exception Handlers (EXC)
| ID | Trigger | Contracted Recovery |
|----|---------|---------------------|
| EXC-01 (INV-03) | Output path already exists | Prompt; on decline, abort the write and keep the existing file |
| EXC-04 (INV-04/05) | Course has no source markdown | Refuse generation, tell the learner to ingest material first |
| EXC-07 (INV-07) | Engine emits framing around the answer | Parse out and keep only the final assistant text |
| EXC-08 (INV-08) | Panic or interrupt mid-TUI | Recover, restore terminal, print a concise error, exit non-zero |
| EXC-09 (INV-09) | Dependency missing/failed | Report the specific dependency and remedy; exit non-zero (CLI) or styled error (TUI) |
| EXC-12 (INV-12) | Vision describe unavailable for the selected engine | Insert asset-referencing placeholder, warn once, continue ingest |
| EXC-35 (INV-12) | Document has no text layer and `--ocr` unset | Stop ingest, report, suggest `--ocr` |
