# Contributing to genius

Thanks for wanting to make genius better. It's a personal study tool grown in
the open — small, opinionated, and built around one loop: **ingest → generate →
revise**, with every generation grounded strictly in the supplied course
material. Contributions that sharpen that loop are very welcome.

## Guiding principles

Keep these in mind before adding anything:

- **Grounded, never invented.** Generation (guides, Q&A, solutions) must stay
  grounded in the course material. The tutor solves problems *as stated* and
  flags gaps rather than fabricating. Don't add features that loosen this.
- **The TUI is the home.** genius is *a place to study*, not a CLI with a menu.
  New capability should be reachable in the TUI, with a discoverable key and a
  matching entry in the `?` help overlay.
- **Match the house style.** The codebase hand-rolls its list/picker rendering
  and follows the patterns in [`docs/`](docs/) (esp.
  [`07_visual_identity.md`](docs/07_visual_identity.md) for palette, status bar,
  and notice glyphs). Read the design docs before reshaping UI.
- **Small, sharp changes.** One concern per PR. If a change is large, open an
  issue first so we can agree on the shape.

## Getting set up

genius shells out to external tools — installing the binary is not enough. See
[Requirements](README.md#requirements) in the README and install:

- **Go 1.25+**
- **`markitdown`** — `pip install 'markitdown[pdf]'` (PDF/PPT → markdown)
- **poppler** (`pdfimages`, `pdftoppm`, `pdfinfo`) — figure extraction
- **`claude`** *or* **`codex`** — generation engine (vision repair needs `codex`)

Then:

```sh
git clone https://github.com/mibienpanjoe/genius && cd genius
go build -o genius .
go test ./...
```

## Before you open a PR

Run all three and make sure they pass clean:

```sh
gofmt -l .        # must print nothing (run `gofmt -w .` to fix)
go vet ./...
go test ./...
```

- **Tests don't hit the network or a real engine.** Use the fake engine
  (`internal/engine/fake.go`) to drive generation/solve logic in tests; don't
  call `claude`/`codex` from a test.
- **Cover behaviour, not internals.** The TUI tests in
  `internal/tui/tui_test.go` drive the model through `Update` and assert on
  rendered output / state — follow that approach for new TUI work.
- **Touched the UI? Update the help.** A new key needs an entry in
  `helpKeys` (`internal/tui/help.go`), the status-bar hints
  (`viewStatusBar`), and the Keys table in the README.

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

## Commits & PRs

- **Conventional commits.** Match the existing history:
  `feat(generate): …`, `fix(tui): …`, `docs(readme): …`. Subject in the
  imperative, lower-case, no trailing period.
- **Keep PRs focused** and describe *what* changed and *why*. Link any related
  issue.
- **Note new external dependencies** (a new CLI tool, a new Go module) in the PR
  — genius keeps its dependency surface small on purpose.

## Reporting bugs / ideas

Open an issue with:

- what you ran (TUI key or CLI command), the engine (`claude`/`codex`), and OS;
- what you expected vs. what happened;
- for ingest/generation bugs, the kind of source document (PDF/PPT/…) — notation
  fidelity (e.g. Boolean complement bars) is a known sharp edge.

## License

By contributing you agree your work is licensed under the project's
[MIT License](LICENSE).
