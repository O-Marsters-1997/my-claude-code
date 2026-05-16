---
name: readme-generator
model: haiku
description: Generate a simple, honest README.md for the current project — focused on what it does and why, how to install, and how to use it with a minimal example. Use whenever the user wants to write, generate, regenerate, or refresh a README, or runs /readme. Trigger on phrases like "write a readme", "generate the readme", "make me a readme", "the project needs docs", "no readme yet", "refresh the readme", or whenever the user is at the root of a project with no README and is talking about documentation.
---

# README Generator

Write a README that helps a stranger answer three questions in under 30 seconds: *what is this, how do I install it, how do I use it.* Everything else is optional.

## Quick start

```
/readme
```

Optionally point at a project root:

```
/readme path/to/project
```

## Workflow

1. **Discover** — figure out what kind of project this is (see [Project discovery](#project-discovery)).
2. **Read existing README** if present. Don't overwrite without showing a diff first.
3. **Get the "why"** — find a one-line description of what the project does and why it exists. Sources, in order of preference:
   - `description` field in `package.json`, `pyproject.toml`, `Cargo.toml`, `setup.py`
   - The opening paragraph of an existing README
   - Top-level docstring or comment in the main entry point
   - **If none of the above exist or they're a generic stub** (`"A new project"`, `"TODO"`, etc.), ask the user: *"One sentence — what does this project do, and who is it for?"* Wait for the answer. Don't invent one.
4. **Ask for project-specific context** — see [Project-specific input](#project-specific-input). A good README often hinges on one or two things only the author knows.
5. **Draft** the README following the [Structure](#structure) below. Required sections always; optional sections only when there's something real to say.
6. **Show the draft to the user** before writing. Offer to adjust tone, trim sections, or add things.
7. **Write** to `README.md` once they're happy. If a README already exists, back it up to `README.md.bak` first.

## Structure

GitHub-flavoured markdown. Sentence case in headings. No emojis unless the user asks. No badges unless the project has real CI/coverage/version info to show.

### Required sections (always include)

1. **Title** — `# Project Name` (from `package.json` name, repo folder, etc.)
2. **One-line tagline** — the "why" from step 3 above, italicised under the title.
3. **What it does** — 2–4 sentences. What problem it solves, who it's for, what makes it different. No marketing language; plain English.
4. **Installation** — the actual command(s). Fenced code block with the right language hint:
   ```bash
   npm install my-package
   ```
   If install requires multiple steps (clone, env vars, build), list them as a short numbered list.
5. **Usage** — minimal, runnable examples. Use as few as possible while still showing the *scope* of what the tool can do — often one is enough, sometimes two are needed to hint at a second use case. Each example should be the smallest copy-pasteable thing that demonstrates that capability. If it's a library, show import + a call. If it's a CLI, show command + expected output. Resist the urge to demo every flag or feature here.

### Optional sections (only if they apply)

Include these only when the project actually has something to say. Empty placeholder sections are worse than no sections.

- **Prerequisites** — only if there's something non-obvious (specific Node/Python version, system libraries, accounts to create).
- **Configuration** — environment variables, config files. Use a table if there are more than three.
- **Development** — how to set up locally for contributing. Skip for libraries with a one-step install.
- **Testing** — the test command, only if there is one.
- **API / Reference** — for libraries with a non-trivial public API. Keep it brief and link to fuller docs if they exist.
- **Contributing** — only if there's a real contribution flow. Link to `CONTRIBUTING.md` if present rather than duplicating.
- **Future development** — short bullet list of upcoming work or known gaps, only if the user actually has plans worth sharing. Skip if there's nothing concrete; aspirational filler is worse than no section.
- **License** — name and link. Only include if a `LICENSE` file exists.

### Things to leave out

- Long feature bullet lists ("✨ Lightning fast! ✨ Type safe! ✨ Easy to use!") — these say nothing
- Acknowledgments, "made with love" footers — unless the user asks
- Tables of contents in short READMEs (under ~150 lines)
- Generated badge soup
- Screenshots/GIFs (note where one would go with `<!-- screenshot here -->` if a UI exists, but don't fabricate)

## Project-specific input

The code tells you *what* the project does mechanically. It rarely tells you the things that actually make a README useful: the constraint that drove the design choices, the gotcha a new user always hits, the comparison to an obvious alternative, the audience assumption. Ask the user before drafting.

Ask one short, open question:

> *"Before I draft this — is there anything specific about this project a reader needs to know that the code won't tell me? E.g. a non-obvious gotcha, the reason it exists rather than using \<obvious alternative\>, who it's NOT for, a known limitation, or anything you'd want a stranger to see in the first 30 seconds. Skip if there's nothing."*

Then incorporate whatever they say using the strategy below — don't dump it into a "notes" section.

### Where each kind of input goes

| User says something about… | Goes into… |
|---|---|
| Why this exists vs an obvious alternative | The 2–4 sentence "what it does" paragraph, as a contrast sentence ("Unlike X, this…") |
| A non-obvious gotcha or footgun | A short callout under Usage, or a `> **Note:**` blockquote next to the relevant example |
| Audience / who it's for / who it's not for | The "what it does" paragraph — explicit audience statement |
| A known limitation | A new optional **Limitations** section (one-line bullets) |
| Performance characteristics or benchmarks | Inline in the "what it does" paragraph if a headline number; otherwise a brief **Performance** section |
| Required external setup (API keys, accounts) | The **Prerequisites** optional section, made required for this project |
| A specific install gotcha | Inline in Installation as a one-line note under the command |
| Future direction the user cares about | The **Future development** optional section |
| Anything that doesn't fit above | Ask the user where they'd like it; don't invent a new top-level section silently |

If the user says "nothing specific", trust them and move on — don't fish.

## Project discovery

Use `Glob` and `Read` to figure out what you're looking at. The goal isn't a full audit — just enough to write the right install/usage commands.

| Detect | Look for | Pulls into |
|---|---|---|
| Node.js / TypeScript | `package.json`, `tsconfig.json` | name, description, scripts, install command |
| Python | `pyproject.toml`, `setup.py`, `requirements.txt` | name, description, install command |
| Rust | `Cargo.toml` | name, description, install command |
| Go | `go.mod` | module name, install command |
| CLI tool | `bin/` entry in `package.json`, `[project.scripts]` in `pyproject.toml`, `cmd/` in Go | usage example shape |
| Library | exports / public API surface | usage example shape |
| Web app | `next.config`, `vite.config`, `index.html` | dev command, build command |
| Tests | `__tests__/`, `*_test.go`, `tests/`, `vitest.config`, `pytest.ini` | test command in optional Testing section |
| CI | `.github/workflows/`, `.circleci/` | only relevant if user wants badges |

For unknown stacks, ask the user what the install and run commands are rather than guessing.

## Writing style

- **Clear over clever.** "Install with npm" beats "Get up and running in seconds!"
- **Active voice, present tense.** "The CLI reads from stdin" not "Stdin will be read by the CLI".
- **Show, don't tell.** A 4-line code example is worth a paragraph of prose.
- **Code blocks always get a language hint** (` ```bash `, ` ```ts `, ` ```python `) — that's what makes them GitHub-flavoured rather than plain.
- **One idea per paragraph.** If a paragraph is more than four lines, it's probably two paragraphs.

## When the README already exists

1. Read it first.
2. Ask the user what they want: *refresh* (keep structure, update commands/versions/details from the code), *rewrite* (start fresh, the current one is stale), or *augment* (add a missing section).
3. For a refresh, preserve any hand-written prose under the standard headings — only regenerate factual content (install commands, scripts, config tables) from source.
4. Back up the original to `README.md.bak` before writing.

## Examples

### Example: a small CLI tool

````markdown
# tinygrep

*A faster `grep` for huge log files, written in Rust.*

tinygrep streams through gzipped log files without decompressing them to disk,
making it 3–5x faster than `zgrep` on multi-GB files. Built for ops engineers
who spend their day searching production logs.

## Installation

```bash
cargo install tinygrep
```

## Usage

```bash
tinygrep "error" logs/*.gz
```

Prints matching lines with filename and line number, same as `grep -Hn`.
````

### Example: a library

````markdown
# date-fns-lite

*A 4kb subset of date-fns for projects that only need formatting and parsing.*

If you're using date-fns just for `format()` and `parseISO()`, you're shipping
80kb you don't need. date-fns-lite gives you those two functions and nothing
else, with the same API.

## Installation

```bash
npm install date-fns-lite
```

## Usage

```ts
import { format } from "date-fns-lite";

format(new Date(), "yyyy-MM-dd"); // "2026-05-16"
```
````

Both examples follow the same shape: title, italic tagline, 2–3 sentence "what & why", install, one minimal usage example. That's the whole template for 90% of projects.
