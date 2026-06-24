```
░█▀▀░█▀▀░█▀█░▀█▀░█░█░█▀▀
░█░█░█▀▀░█░█░░█░░█░█░▀▀█
░▀▀▀░▀▀▀░▀░▀░▀▀▀░▀▀▀░▀▀▀
```

# genius

**A dedicated terminal study environment.** Run `genius` and you are *in a place
built for studying* — not issuing commands inside a general-purpose shell. It
takes your raw course material (PDF/PPT lectures) and turns it into study
guides, revision Q&A, and interactive quizzes, all inside one cohesive,
gradient-styled terminal UI.

---

## Why

Studying from raw lecture files is manual and fragmented: PDFs aren't skimmable
or revisable, hand-making guides and Q&A is slow, and a generic chatbot gives
answers ungrounded in *your* course — off-syllabus or wrong. genius owns a study
workspace and gives you one immersive loop: **ingest → generate → revise**, with
every generation grounded strictly in the supplied material.
I designed it the way I study most of the time for my CS tests to get the best marks with low effort. 

## Features

- **Ingest** — convert PDF/PPT/documents to markdown (via `markitdown`), filed
  per course. Figures are extracted (poppler) and vision-captioned.
- **Notation-faithful** — text extraction silently drops Boolean complement
  bars (`X̄`); genius detects this and re-transcribes the affected pages with a
  vision model, since the bars are graphics no text extractor can recover.
- **Study guides** — structured guides (summary, key concepts, formulas, common
  traps) generated from the full course markdown.
- **Revision Q&A** — Q&A pairs for revision, in a format the quiz loop consumes.
- **Interactive quiz** — the signature feature: ask → answer → reveal →
  self-grade → advance, with a running score.
- **In-place rendering** — guides/Q&A shown as styled, scrollable markdown
  (Glamour), never leaving the environment.
- **Swappable engines** — `claude` (default) or `codex`, behind one interface,
  identical workflow.

## Requirements

| Tool | Purpose | Install |
|------|---------|---------|
| Go 1.25+ | build | — |
| `markitdown` | PDF/PPT → markdown | `pip install 'markitdown[pdf]'` |
| poppler (`pdfimages`, `pdftoppm`, `pdfinfo`) | figure extraction, page rasterize | system package |
| `claude` *or* `codex` | generation engine | their respective CLIs |

> `markitdown` installs to `~/.local/bin` — make sure it's on `PATH`.
> `claude` has no vision support; figure captioning / notation repair needs
> `codex`.

## Install

```sh
# install the binary directly
go install github.com/mibienpanjoe/genius@latest

# …or build from a clone
git clone https://github.com/mibienpanjoe/genius && cd genius
go build -o genius .
install -m 0755 genius ~/.local/bin/genius   # onto your PATH
```

> genius is a personal study setup: it shells out to `markitdown`, `claude`, and
> `codex`, and assumes the workspace convention below. Installing the binary is
> not enough on its own — see [Requirements](#requirements).

## Usage

Launch the environment (home dashboard rooted at your study workspace):

```sh
genius
```

Or script the actions with subcommands:

```sh
genius ingest lecture.pdf              # → courses/lecture/lecture.md (+ assets)
genius ingest td.pdf --kind exercise --course logic
genius guide logic                     # → guides/logic.md
genius qa logic --count 15 --scope "Boolean algebra"
genius guide logic --engine codex      # swap the generation engine
```

### Keys (TUI)

| Key | Action |
|-----|--------|
| `↑`/`↓` `j`/`k` | move between courses |
| `g` / `enter` | open the study guide |
| `q` | open the Q&A |
| `r` | start a revision quiz |
| `esc` | back |
| `ctrl+c` | quit |

In a quiz: type your answer → `enter` reveals → `y`/`n` (or `space`) self-grades
→ advance.

## Workspace

genius owns a study home, resolved as `$GENIUS_HOME`, else `~/study`:

```
~/study/
  courses/<name>/*.md       ingested / source markdown
  courses/<name>/assets/    figures extracted during ingest
  guides/<name>.md          generated study guides
  qa/<name>.md              generated revision Q&A
  exercises/<name>/*         exercise sets (+ assets/)
```

Courses are referenced by name; genius reads all markdown under
`courses/<name>/` as grounding. Config lives at `~/.config/genius/config.toml`
(`study_root`, `default_engine`, `model`).

## How it works

- **Grounding** — the full course markdown is stuffed into the generation prompt
  (typical lectures fit in context). Chunking/retrieval is deferred until a
  course overflows.
- **Engines** — `claude -p` / `codex exec` run as swappable subprocesses behind
  an `Engine` interface; only the final assistant text is captured.
- **Architecture** — Go + the Charm stack (Bubble Tea / Lip Gloss / Bubbles),
  Glamour for rendering, Cobra for subcommands. Design docs live in [`docs/`](docs/).

## Project layout

```
main.go                 entrypoint → cli.Execute
internal/
  cli/                  cobra subcommands + TUI launch
  workspace/            study-root resolution, course scan, writes
  engine/               Engine interface + claude / codex / fake
  convert/              markitdown ingest, figure extraction, notation repair
  generate/             grounded guide / qa prompt assembly
  render/               Glamour helper (markdown → styled string)
  tui/                  Bubble Tea: home / reader / quiz
  quiz/                 qa-markdown parser
docs/                   01–07 design docs (PRD, SRS, architecture, …)
```

## Status

MVP complete: ingest, guide, qa, reader, and quiz all work end-to-end against
real material. Post-MVP: exercise tutoring (`solve`), TUI-driven generation,
and weak-spot tracking.

## License

[MIT](LICENSE). A personal / educational tool — generation cost is borne by your
own `claude` / `codex` subscriptions.
