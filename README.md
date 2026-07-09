```
░█▀▀░█▀▀░█▀█░▀█▀░█░█░█▀▀
░█░█░█▀▀░█░█░░█░░█░█░▀▀█
░▀▀▀░▀▀▀░▀░▀░▀▀▀░▀▀▀░▀▀▀
```

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
- **Exercise tutoring** — enumerate a problem set, pick exercises (or sub-parts),
  and get worked solutions grounded only in the course; solved *as stated*, with
  gaps in the material flagged rather than fabricated.
- **In-place rendering** — guides/Q&A shown as styled, scrollable markdown
  (Glamour), never leaving the environment.
- **Self-sufficient TUI** — ingest documents, generate guides/Q&A, and scope a
  generation to chosen chapters entirely in-app, without dropping to the CLI.
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

From the home dashboard you can do everything in-app — ingest (`i`), generate a
guide/Q&A (`g`/`q`), scope to chapters (`f`), revise (`r`), solve (`s`); see
[Keys](#keys-tui). Or script the same actions with subcommands:

```sh
genius ingest lecture.pdf              # → courses/lecture/lecture.md (+ assets)
genius ingest td.pdf --kind exercise --course logic
genius guide logic                     # → guides/logic.md
genius qa logic --count 15 --scope "Boolean algebra"
genius solve logic --set td            # list the set's exercises
genius solve logic --set td --ex 2,3.1 # work specific exercises, grounded
genius guide logic --engine codex      # swap the generation engine
```

### Keys (TUI)

| Key | Action |
|-----|--------|
| `↑`/`↓` `j`/`k` | move between courses |
| `g` / `enter` | open the study guide (build it if missing) |
| `q` | open the Q&A (build it if missing) |
| `G` / `Q` | force-regenerate the guide / Q&A |
| `r` | start a revision quiz |
| `s` | solve exercises (pick set → pick exercises) |
| `i` | ingest a document (in-app file browser) |
| `f` | chapter hub — per-chapter guides & Q&A |
| `?` | keybinding help |
| `esc` | back — or cancel a running generation / solve / ingest |
| `ctrl+c` | quit |

The TUI is self-sufficient: ingest material (`i`), generate guides/Q&A on the
spot (`g`/`q`, or `G`/`Q` to redo), and open the **chapter hub** (`f`) — no need
to drop to the CLI. In the hub, `space`-select one or more chapters and `g`/`q`
builds (or opens) a guide/Q&A scoped to just that selection, filed separately
from the whole-course one (see [Per-chapter artifacts](#per-chapter-artifacts)).
Generating Q&A (whole-course or scoped) first asks how many pairs to make —
`enter` accepts the default. While a generation, solve, or ingest is running the
spinner names the exact scope being built, `esc` cancels it and drops you back
where you were, and `ctrl+c` quits genius; closing the reader afterwards returns
you to the screen you opened it from.
When a course has more than one Q&A, `r` first asks which to revise — whole,
a single chapter, or all chapters merged. In a quiz: type your answer → `enter`
reveals → `y`/`n` (or `space`) self-grades → advance. In solve: pick a set →
`space` toggles exercises → `enter` solves the selection (or the highlighted
one) and shows the worked solution in the reader.

## Organising courses

A course is a **directory of markdown** under `courses/<name>/`; genius reads
*every* `.md` there as grounding. The course name and the document filename are
separate, which makes a few layouts fall out naturally.

**Single-file course** — one lecture, one course. The course name defaults to
the filename slug.

```sh
genius ingest lecture.pdf            # → courses/lecture/lecture.md
genius guide lecture
```

**Multi-chapter course** — one course split across several PDFs (a chapter per
file). Point them all at one course with `--name`; each keeps its own `.md`.
Zero-pad the names so they sort in reading order (the grounding concatenates
files lexically).

```sh
genius ingest chap01.pdf --name algebra      # → courses/algebra/chap01.md
genius ingest chap02.pdf --name algebra      # → courses/algebra/chap02.md
genius ingest chap03.pdf --name algebra

genius guide algebra                         # → guides/algebra.md (ALL chapters)
genius guide algebra --files chap03.md       # → guides/algebra/chap03.md
genius guide algebra --files chap01.md,chap02.md   # → guides/algebra/chap01+chap02.md
```

`--files` works the same on `qa`. A scoped run is filed separately under the
course subdir (see [Per-chapter artifacts](#per-chapter-artifacts)) and never
overwrites the whole-course artifact. Without `--files`, generation grounds on —
and writes — the whole course.

### Per-chapter artifacts

The whole-course guide and Q&A always live at `guides/<course>.md` and
`qa/<course>.md`. A **scoped** generation — one chapter or a span — is filed
separately under the course's own subdir, so it never overwrites the
whole-course one:

```
guides/algebra.md                 whole course
guides/algebra/chap01.md          just chapter 1
guides/algebra/chap01+chap02.md   a span (joined chapter slugs)
qa/algebra/chap03.md              chapter 3 Q&A
```

In the TUI this is the **chapter hub** (`f`): `space`-select chapters, then
`g`/`q` to build or open the scoped artifact (`G`/`Q` force-rebuild). Each
chapter row shows whether its guide/Q&A already exists. The home `g·N q·N`
chips count the whole-course artifact **plus** every scoped one. When several
Q&A exist, `r` opens a picker (whole / chapter / all merged) before revising.

**Topic focus vs. file scope** — two different knobs:

- `--files <a.md,b.md>` chooses *which source material* is fed to the model
  (file-level grounding).
- `qa --scope "<text>"` is a *free-text instruction* ("focus on Karnaugh maps"),
  not a filename — the grounding is unchanged, the model just narrows its
  attention.

```sh
genius qa algebra --count 15 --scope "Karnaugh maps"
genius qa algebra --files chap03.md --scope "don't-care conditions"
```

**Your own notes** — drop hand-written `.md` straight into `courses/<name>/`;
genius reads it as grounding alongside ingested files. Only markdown counts —
never copy a raw `.pdf` into the workspace; run it through `ingest` first.

**Exercises** — exercise sets are filed under a course but kept apart from the
lecture material (so problems never pollute the guide/Q&A grounding):

```sh
genius ingest td1.pdf --kind exercise --course algebra   # → exercises/algebra/td1.md
```

Then `solve` enumerates the set and works the exercises you pick, grounded only
in the course. The tutor solves each one **as stated** and flags any gap in the
material rather than inventing a method.

```sh
genius solve algebra --set td1                 # list the exercises (and sub-parts)
genius solve algebra --set td1 --ex 2,3.1      # work exercise 2 and sub-part 3.1
genius solve algebra --set td1 --ex 2 --save   # also write td1.solutions.md (source untouched)
```

Sub-parts are addressed `<exercise>.<part>` — `3.1`, `3.a`. In the TUI, press
`s` to do the same: pick a set, toggle exercises with `space`, `enter` to solve.

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
  generate/             grounded guide / qa / solve prompt assembly + enumerate
  render/               Glamour helper (markdown → styled string)
  tui/                  Bubble Tea: home / reader / quiz / solve / ingest / chapters / generate / help
  quiz/                 qa-markdown parser
docs/                   01–07 design docs (PRD, SRS, architecture, …)
```

## Status

Ingest, guide, qa, reader, quiz, and exercise solving (`solve`) all work
end-to-end against real material, in both the CLI and the TUI — including
in-app ingest, generation, and per-chapter scoping (v0.2.0). v0.3.0 adds a Q&A
count prompt, a scope-aware generating spinner, origin-aware reader back, and
`esc`-cancelable generation. Post-MVP: quiz weak-spot tracking.

## Contributing

Want to help sharpen the study loop? See [CONTRIBUTING.md](CONTRIBUTING.md) for
setup, the build/test commands, and the house style.

## License

[MIT](LICENSE). A personal / educational tool — generation cost is borne by your
own `claude` / `codex` subscriptions.
