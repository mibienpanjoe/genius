# genius — Project Overview

## Product name and core mission
**genius** is a dedicated terminal study environment. You run `genius`, and you
are immediately *in a place built for studying* — not issuing commands inside a
general-purpose shell or chat agent. Its mission: take a learner's raw course
material (PDF/PPT lectures) and turn it into study guides, revision Q&A, and
grounded exercise help, all inside one cohesive, beautiful terminal UI.

## The problem being solved
Studying from raw lecture files is manual and fragmented:
- Course files are PDF/PPT — not skimmable, not searchable, not revisable.
- Making study guides and Q&A by hand is slow.
- Getting help on exercises from a generic chatbot produces answers ungrounded
  in the actual course, so they can be wrong or off-syllabus.
- There is no single "study mode" — material, notes, and tooling are scattered.

genius fixes this by owning a study workspace and providing one immersive tool
that ingests material, generates guides/Q&A grounded in that material, runs
interactive revision quizzes, and tutors exercises against the course content.

## Users (personas)
- **Primary — the self-directed learner (the user).** A student or
  professional who studies from course PDFs/PPTs, is comfortable in the
  terminal, and wants a focused, repeatable study loop. Comfortable with the
  Charm/Bubble Tea (Elm) interaction model.
- **Secondary — future learners** who clone genius to study their own
  material. Same workflow, different content.

## Core capabilities (MVP)
1. **Ingest** — convert PDF/PPT/other course files to markdown via `markitdown`,
   filed into a per-course workspace folder.
2. **Generate study guides** — turn course markdown into a structured study
   guide (summary, key concepts, formulas/definitions, common traps).
3. **Generate revision Q&A** — turn course markdown into Q&A pairs for revision.
4. **Interactive revise/quiz** — a stateful quiz loop over a Q&A file: ask a
   question, capture the answer, reveal + self-grade, advance. The signature
   feature.
5. **Render** — display generated guides/Q&A as styled, scrollable markdown
   in-process (Glamour), never leaving the environment.
6. **Dedicated environment** — launching `genius` with no arguments opens a TUI
   home dashboard rooted at the study workspace, showing courses and progress.

## Tech stack
- **Language:** Go 1.25.
- **TUI:** Charm stack — Bubble Tea (Elm architecture), Lip Gloss (styling),
  Bubbles (list, viewport, spinner, textinput), Huh (in-TUI prompts).
- **Markdown rendering:** Glamour (in-process; the engine inside `glow`).
- **CLI plumbing:** Cobra for subcommands (optionally Charm Fang for styled help).
- **Generation engines (subprocess, swappable):** `claude -p` (default) and
  `codex exec`, behind a single `Engine` interface.
- **Document conversion (subprocess):** `markitdown` (Python CLI).
- **Config:** TOML at `~/.config/genius/config.toml`.

## Deployment context
A single static Go binary `genius`, run locally on the user's machine (Linux).
No server, no network service of its own. It shells out to local CLIs
(`claude`, `codex`, `markitdown`) which may themselves call remote APIs. Single
user, single machine — no multi-tenancy, no authentication.

## Workspace model
genius owns a study home directory, resolved as `$GENIUS_HOME`, else `~/study`:
```
~/study/
  courses/<name>/*.md      ingested/source markdown
  courses/<name>/assets/    figures extracted from PDFs/PPTs during ingest
  guides/<name>.md         generated study guides
  qa/<name>.md             generated revision Q&A
  exercises/<name>/*        exercise sets (+ assets/ for their figures)
  .genius/progress.json    quiz weak-spot tracking (post-MVP)
```
Courses are referenced by name; genius reads all markdown under
`courses/<name>/` as grounding for generation.

## Grounding model
MVP stuffs the **full course markdown** into the generation prompt (typical
lectures fit in model context). Chunking/retrieval is deferred until a course
overflows context.

## Monetization model
None. Personal/educational tool. Generation cost is borne by the user's own
`claude`/`codex` subscriptions.

## Success criteria
- A learner can go from a raw PDF to a rendered study guide in one session
  without leaving the environment.
- Generated guides/Q&A and exercise help are grounded in the supplied course
  material, not generic knowledge.
- The interactive quiz loop runs reliably: ask → answer → reveal → advance.
- Engine is swappable (`claude` ↔ `codex`) with identical workflow.
- Launching `genius` *feels* like entering a dedicated study space.

## Out of scope (for MVP)
- Multi-user, accounts, or any networked/server component.
- Chunking / vector retrieval / RAG (full course fits in prompt).
- Spaced-repetition scheduling algorithms (simple quiz only).
- Editing course content inside genius (it reads/generates, not a doc editor).
- Cloud sync of the workspace.
- Non-terminal interfaces (web/desktop GUI).
- Generating diagrams or images.
