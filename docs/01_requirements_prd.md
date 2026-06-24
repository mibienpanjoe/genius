# genius — Product Requirements Document
Version: v1.0, 2026-06-24

## 1. Problem Statement
Studying from raw lecture material is slow and fragmented. Course content arrives
as PDF and PowerPoint files that are hard to skim, search, or revise from.
Turning that material into study guides and revision Q&A is manual labor, and
asking a generic AI chatbot for help produces answers ungrounded in the actual
course — often wrong or off-syllabus. There is no single "study mode": material,
notes, and tooling live in different places, so starting a study session means
assembling an environment every time.

genius removes that friction. It is a dedicated terminal environment a learner
*enters* to study. It owns a study workspace, ingests course files into
markdown, generates study guides and Q&A grounded in that material, runs
interactive revision quizzes, and tutors exercises against the course — all in
one cohesive terminal UI.

## 2. Personas
- **Primary — Self-directed learner (the user).** A student or professional who
  studies from course PDFs/PPTs, is fluent in the terminal, and wants a focused,
  repeatable study loop. Comfortable with the Charm/Bubble Tea interaction model.
  Values aesthetics and a sense of "place" while studying.
- **Secondary — Future learners.** People who adopt genius for their own
  material. Same workflow, different content. Should be able to install and run
  without bespoke setup beyond the documented prerequisites.
- **Tertiary — Contributors.** Engineers extending genius (new commands, engine
  backends). Need a legible architecture and clear component boundaries.

## 3. Solution Overview
genius is a single Go binary presenting a Charm/Bubble Tea terminal UI rooted at
a study workspace. It converts course files to markdown (via `markitdown`),
generates study guides and revision Q&A grounded in that markdown (via a
swappable `claude`/`codex` engine), renders the results as styled scrollable
markdown in-process (via Glamour), and runs an interactive revision quiz over a
Q&A file. Running `genius` with no arguments opens a home dashboard showing
courses and progress; subcommands expose the same actions for scripting.

## 4. MVP Scope

### Environment & Navigation
- Running `genius` (no args) opens a TUI **home dashboard** rooted at the
  workspace, listing courses with per-course progress (guide / Q&A / exercise
  counts) and the active engine.
- Keyboard navigation between home, a per-course view, and a markdown reader.

### Ingest
- Convert a PDF/PPT (and other `markitdown`-supported formats) to markdown.
- File the output under `courses/<name>/` in the workspace.
- Detect when `markitdown` is missing and tell the user how to install it.

### Study Guide Generation
- Generate a structured study guide from a course's markdown.
- Sections: summary, key concepts, formulas/definitions, common traps, worked
  highlights.
- Write the result to `guides/<name>.md`.

### Revision Q&A Generation
- Generate revision Q&A pairs (markdown) from a course's markdown.
- Write the result to `qa/<name>.md`.

### Interactive Revise / Quiz
- Load a Q&A file and run a stateful loop: present a question, capture the
  learner's answer, reveal the model answer for self-grading, advance to the
  next question, and summarize at the end.

### Rendering
- Render any generated guide/Q&A as styled, scrollable markdown in-process
  (Glamour), inside the environment.

### Engine Selection
- Generation runs through a swappable engine: `claude` (default) or `codex`,
  selectable by flag and configurable as a default.

## 5. Out of Scope (for MVP)
- Multi-user support, accounts, authentication, or any networked/server component.
- Chunking / vector retrieval / RAG (the full course is stuffed into the prompt).
- Spaced-repetition scheduling algorithms (the quiz is a simple linear loop).
- Editing or authoring course content inside genius (it reads and generates; it
  is not a document editor).
- Cloud sync or sharing of the workspace.
- Non-terminal interfaces (web or desktop GUI).
- Generating images, diagrams, or non-text artifacts.
- The `solve` exercise-tutor command and persistent weak-spot tracking — planned
  for the next iteration, not MVP.

## 6. Success Criteria
- A learner can go from a raw PDF to a rendered study guide in a single session
  without leaving the environment.
- Generated guides, Q&A, and (later) exercise help are grounded in the supplied
  course material rather than generic knowledge — no fabricated facts beyond the
  source.
- The interactive quiz loop runs end-to-end (ask → answer → reveal → advance →
  summary) without losing state.
- Switching engine (`claude` ↔ `codex`) changes nothing about the workflow or
  the file outputs' locations.
- First-run to first useful artifact (ingest → guide) takes only a few commands
  and no manual file shuffling.
- Launching `genius` is experienced as entering a dedicated study space, not as
  running a utility.
