# genius — Visual Identity Guide (Terminal UI)
Version: v1.0, 2026-06-24

> genius is a terminal app, so this design system targets **Lip Gloss / ANSI**,
> not web CSS. Values are truecolor hex (degrading to 256/16-color), and spacing
> is in terminal **cells**, not pixels. The goal: a contributor can build a new
> TUI state that fits seamlessly, with no designer present.

## Brand Essence
**Name:** *genius* — the study companion that makes you feel sharp. Lowercase,
unadorned: a tool, not a brand shouting. It evokes focus and quiet capability.

**Personality:** Focused, calm, capable, warm, modern. The opposite of a noisy
dashboard.

**Design principles:**
- **A place, not a prompt** — every screen reinforces "you are here to study."
- **Calm by default** — restrained color; emphasis is earned, not sprayed.
- **Charm-native** — consistent with the Charm ecosystem (glow/Glamour) the user
  already knows.
- **Legible first** — content (the study material) is the hero; chrome recedes.
- **Keyboard-driven** — every action has a visible, discoverable key.

## Color System
Terminal truecolor hex with graceful degradation. Honor `NO_COLOR` and adapt to
light/dark terminals via Lip Gloss `AdaptiveColor` where noted.

### Primary Palette
| Token | Name | Hex | Use |
|-------|------|-----|-----|
| `--c-primary` | Genius Violet | `#7C6FF0` | Active selection, focus, brand accents |
| `--c-primary-dim` | Violet Dim | `#5B50C2` | Hover/secondary accent, borders on focus |
| `--c-bg` | Ink | `#0E0E14` | App background (dark terminals) |
| `--c-surface` | Panel | `#1A1A24` | Cards, panels, status bar background |
| `--c-text` | Paper | `#E6E6F0` | Primary text |
| `--c-text-muted` | Slate | `#8A8AA0` | Secondary text, hints, inactive items |

### Brand Gradient (wordmark / splash)
A left-to-right violet→pink ramp applied per-cell across the ASCII wordmark and
other hero accents. Interpolate in RGB across the terminal width of the glyphs.
| Stop | Hex |
|------|-----|
| 0% | `#6BA8F5` (sky) |
| 35% | `#7C6FF0` (Genius Violet) |
| 70% | `#B06FE6` (orchid) |
| 100% | `#E66FB0` (pink) |
Degrade to a single `--c-primary` fill on 16-color terminals or under `NO_COLOR`.

### Semantic Colors
| Token | Hex | Use |
|-------|-----|-----|
| `--c-success` | `#3FB950` | Correct self-grade, completed counts, saved state |
| `--c-warning` | `#D6A33E` | Pending/empty courses, caution, overwrite prompts |
| `--c-error` | `#E5534B` | Errors, missing dependencies, failed generation |
| `--c-info` | `#58A6FF` | Neutral notices, the active-engine indicator |

### Neutral Scale (borders/dividers)
| Token | Hex |
|-------|-----|
| `--c-line` | `#2A2A38` |
| `--c-line-strong` | `#3A3A4C` |

### Adaptive (light terminals)
`AdaptiveColor`: text `#1A1A24` on bg `#FAFAFB`, surface `#EFEFF4`, primary stays
Genius Violet `#7C6FF0`. Semantic hues unchanged (they read on both).

## Typography
A terminal cannot choose fonts — it inherits the user's monospace face. "Type" is
therefore expressed through **weight, color, case, and Glamour styles**, with a
consistent hierarchy.

| Role | Treatment | Use |
|------|-----------|-----|
| Title | Bold, `--c-primary`, optional uppercase | Screen titles ("genius", "ALGEBRA") |
| Section header | Bold, `--c-text` | Group labels, panel headings |
| Body | Regular, `--c-text` | Questions, content, list labels |
| Muted | Regular, `--c-text-muted` | Hints, counts, footers |
| Emphasis | Bold or `--c-primary` | The selected item, key answer terms |
| Code/ID | `--c-info` on `--c-surface` | Paths, course names, keybindings |

### Rendered markdown (Glamour)
Guides/Q&A render through a custom Glamour style aligned to this palette:
headings in Genius Violet, code blocks on `--c-surface`, links `--c-info`, block
quotes with a `--c-primary-dim` left bar. One shared style for the whole app.
**Figure captions** (the provenance-marked blockquotes from ingest) inherit the
blockquote style; the `[Figure N]` label is bold `--c-primary` and the
`(transcribed from …)` provenance is muted, so model transcriptions read as
clearly distinct from original course text (INV-12).

## Spacing & Layout
Unit = **1 terminal cell**. Layout is responsive to terminal width/height
(queried from Bubble Tea `WindowSizeMsg`).

| Token | Cells | Use |
|-------|-------|-----|
| `space-1` | 1 | Tight padding inside chips/status |
| `space-2` | 2 | Standard horizontal panel padding |
| `space-3` | 4 | Section separation |

- **Min supported size:** 80×24; below that, show a "resize terminal" notice.
- **Layout:** full-screen alt-buffer; a 1-row **status bar** pinned to the
  bottom; content area fills the rest. Reader/quiz use a centered column capped at
  ~100 cells for readability.
- **Borders:** Lip Gloss `RoundedBorder` for panels/cards; focused panel border
  uses `--c-primary`, unfocused `--c-line`.

## Component Styling

### Wordmark / splash banner (home header)
Inspired by the Gemini-CLI launch screen. A large block-ASCII **`genius`**
wordmark rendered with the Brand Gradient (per-cell violet→pink across its
width), drawn at the top of the home screen.
- **Source:** a `//go:embed` block-letter template, or generated with a figlet-
  style font; colored per-cell via Lip Gloss, not a static escape blob.
- **Responsive:** full banner at ≥ 100 cells wide; a compact single-line gradient
  `genius` below that; hidden under ~24 rows to protect the list.
- **Below the wordmark (empty/first-run state only):**
  - **Getting-started tips** — a short numbered list (muted text, keywords in
    `--c-primary`), e.g. _1. ingest a course: `genius ingest lecture.pdf` · 2.
    generate a guide: press `g` · 3. revise: press `r` · 4. `?` for help_.
  - **Notice box** — a `RoundedBorder` panel in `--c-warning` when the workspace
    is empty (mirrors Gemini's home-dir warning): _"No courses yet — ingest a PDF
    or PPT to begin. Workspace: ~/study"_.
  - **Muted "using" line** — `using: engine:<name> · ~/study` in
    `--c-text-muted`, echoing Gemini's `Using:` line.
Once the workspace has courses, the banner stays but tips/notice collapse and the
**course list** becomes the focus.

### Status bar (bottom, always visible)
Background `--c-surface`, height 1 row. Left: workspace root (muted). Center:
context hints/keys. Right: `engine:<name>` in `--c-info`. This is the persistent
"you are in genius" anchor.

### Course list (home)
Bubbles `list`. Each row: course name (`--c-text`, bold when selected) + count
chips `g·N q·N e·N`. Counts colored `--c-success` when > 0, `--c-text-muted` when
0. Selected row: left bar `▌` in `--c-primary`, background `--c-surface`.

### Cards / panels (course view, reader frame)
`RoundedBorder`, padding `space-2`, title in bold `--c-primary`. Focused card
border `--c-primary`; others `--c-line`.

### Quiz card
Centered card. Question in body weight; a divider (`--c-line`) before the revealed
answer; the answer fades in styled with `--c-text`. Self-grade keys shown as
chips: `[y] knew it` (`--c-success`), `[n] missed` (`--c-error`), `[space] next`.
Progress shown as `Q <i>/<n>` (muted) top-right of the card.

### Inputs (Huh / textinput)
Prompt label bold; input underline `--c-line`, caret `--c-primary`. Validation
errors below in `--c-error`.

### Spinner / progress
Bubbles `spinner` in `--c-primary` with a muted label (e.g. `generating guide…`)
shown for any Engine call. Long generations MAY show an indeterminate dot pulse.

### Notices
Inline rows prefixed by a glyph + semantic color: `✓` success, `!` warning, `✗`
error, `i` info — never color alone (glyph carries the meaning too).

## Motion & Animation
Terminal-appropriate, restrained.
| Token | Value | Use |
|-------|-------|-----|
| spinner tick | ~100 ms | Engine activity |
| state transition | instant (≤ 1 frame) | Navigating between TUI states |
| answer reveal | ~150 ms ease | Quiz answer appearing |
Respect a "reduced motion" preference: if `GENIUS_NO_ANIM=1` (or `NO_COLOR`),
replace spinners with a static `working…` label and disable reveals.

## Screen-Level Patterns

### Home dashboard
Gradient wordmark banner on top (per-cell violet→pink), then — depending on
state — the getting-started tips/notice (empty workspace) or the course list
(populated). Status bar pinned at the bottom.

**Populated state:**
```
   ▟█ ▟██▖ ▟██▖ ▜█▛ ▜█▛ ▟██▖ ▜█▛  ▟██▖      ← block wordmark, violet→pink gradient
   ▜█▛ █▄▄  █  █  █   █   █  █  █   █   ▄▄█
   ▝▀  ▀▀▀  ▀  ▀ ▝▀▘  ▀   ▀▀▀  ▝▀▘  ▀▀▀

   using: engine:claude · ~/study

   COURSES               g    q    e
   ▌algebra              3    2    1
    organic-chem         1    0    4
    history-ww2          0    0    0

   enter open · g guide · q qa · r revise
─────────────────────────────────────────────────────────────
 ~/study            i ingest · s solve              engine:claude
```

**Empty / first-run state** (mirrors the Gemini splash):
```
   ▟█ ▟██▖ ▟██▖ ▜█▛ ▜█▛ ▟██▖ ▜█▛  ▟██▖
   ▜█▛ █▄▄  █  █  █   █   █  █  █   █   ▄▄█
   ▝▀  ▀▀▀  ▀  ▀ ▝▀▘  ▀   ▀▀▀  ▝▀▘  ▀▀▀

   getting started:
   1. ingest a course:  genius ingest lecture.pdf
   2. generate a guide: press g   3. revise: press r
   4. ? for help

   ┌──────────────────────────────────────────────────────┐
   │ No courses yet — ingest a PDF or PPT to begin.         │
   │ Workspace: ~/study                                     │
   └──────────────────────────────────────────────────────┘

   using: engine:claude · ~/study
─────────────────────────────────────────────────────────────
 ~/study            i ingest · s solve              engine:claude
```

### Reader
Centered column, Glamour-rendered content in a scrollable `viewport`; footer
shows scroll % (muted) and `q back`.

### Quiz
Centered quiz card as specified above; status bar shows `Q i/n` and grade keys.

### Exercises / solve *(post-MVP)*
Reached from the course view. A two-pane flow:
1. **Set list** — exercise sets under the course (`td3`, `partiel-2024`, …) as a
   Bubbles `list`.
2. **Exercise picker** — the chosen set's enumerated exercises (`▢ Ex 1`, `▢ Ex
   2`, …), multi-select with `space`, each row showing the exercise's label and a
   truncated prompt. Selected rows marked `▣` in `--c-primary`.
Pressing solve runs generation (spinner) and shows the worked solution in the
reader. A solution that hits a course gap renders the flagged note as a
`--c-warning` callout. `--save` writes a separate solutions file (source set is
never touched).

## Accessibility Checklist
- [ ] Foreground/background pairs meet ~4.5:1 contrast on the default dark theme.
- [ ] `NO_COLOR` is honored — UI stays fully legible with color stripped.
- [ ] Color is **never** the sole signal: counts, grades, and notices also carry a
      glyph or label.
- [ ] `AdaptiveColor` keeps the UI legible on light terminals.
- [ ] Reduced-motion mode (`GENIUS_NO_ANIM`/`NO_COLOR`) disables spinners/reveals.
- [ ] Every action shows its key; nothing is mouse-only.
- [ ] Minimum size handled gracefully with a resize notice (no broken layout).

## Language & Tone Guidelines
- Lowercase, plain, encouraging. Verbs over nouns: `generating guide…`, not
  `Guide Generation In Progress`.
- Errors are actionable and blameless: `markitdown not found — run: pip install
  markitdown`.
- Quiz copy is supportive, never punitive: `missed — review and move on`.
- Never expose internal jargon (component names, invariant IDs) in the UI.
