# Agent Skills

A collection of agent skills that extend capabilities across planning, development, and tooling.

## CLI Reference

| Command | Purpose |
|---|---|
| `npx skills add <source>` | Install a skill |
| `npx skills list` | List all installed skills |
| `npx skills find [query]` | Search/browse available skills |
| `npx skills remove [skill]` | Remove an installed skill |
| `npx skills check` | Check installed skills for updates |
| `npx skills update` | Update all installed skills |
| `npx skills init [name]` | Create a new SKILL.md template |

Browse available skills at [skills.sh](https://skills.sh). Common flags: `-g` (global), `-y` (skip prompts).

To list all skills available in this repo:

```
npx skills add O-Marsters-1997/skills --list
```

## Product workflow suite

Eight skills that take work from a raw idea through to tickets on a board. See
[`skills/workflow/README.md`](skills/workflow/README.md) for the map. Install the whole suite at
once — the skills call each other, so a partial install leaves broken handoffs:

```
for s in artifact-scan ideate chat-to-approach to-roadmap to-prd to-plan to-tickets ticket-tracker; do
  npx skills add O-Marsters-1997/my-claude-code --skill "$s" -g -y
done
```

| Skill | Level | Takes | Produces |
|---|---|---|---|
| `artifact-scan` | — | nothing | a report + one routing recommendation |
| `ideate` | portfolio | the codebase | `./ideas/reports/YYYY-MM-DD-ideate.md` |
| `chat-to-approach` | portfolio | a pasted conversation | `./docs/approach.md` |
| `to-roadmap` | portfolio | `./docs/approach.md` | `./roadmap.html` (cards are features) |
| `to-prd` | feature | one feature | `./docs/prd-<feature>.md` + a `[PRD]` issue |
| `to-plan` | feature | a PRD | `./plans/<feature>.md` |
| `to-tickets` | feature | a plan | GitHub issues |
| `ticket-tracker` | feature | GitHub issues | `status:*` label moves |

The suite calls three routines that are not stages. These sit at depth 2, so they install normally:

```
for s in grilling source-synthesis triage-issue; do
  npx skills add O-Marsters-1997/my-claude-code --skill "$s" -g -y
done
```

## Planning & Design

These skills help you think through problems before writing code.

- **grilling** — The interview loop itself: rounds of questions along the frontier of the design tree, each with a recommended answer. Model-invoked; the suite delegates here.

  ```
  npx skills add O-Marsters-1997/skills --skill grilling
  ```

- **grill-me** — Slash-command front door to `grilling`. `/grill-me`.

  ```
  npx skills add O-Marsters-1997/skills --skill grill-me
  ```

- **grill-me-with-docs** — Grilling that also challenges your plan against `CONTEXT.md`, sharpens terminology, and writes glossary entries and ADRs inline. `/grill-me-with-docs`.

  ```
  npx skills add O-Marsters-1997/skills --skill grill-me-with-docs
  ```

## Development

These skills help you write, refactor, and fix code.

- **code-review** — Review a diff since a fixed point on two axes in parallel — Standards (repo conventions plus a Fowler smell baseline) and Spec (does it match the originating issue) — plus a third Greptile axis on Greptile-enabled repos that triages the PR bot's findings before you push.

  ```
  npx skills add O-Marsters-1997/skills --skill code-review
  ```

- **tdd** — Test-driven development with a red-green-refactor loop. Builds features or fixes bugs one vertical slice at a time.

  ```
  npx skills add O-Marsters-1997/skills --skill tdd
  ```

- **triage-issue** — Investigate a bug by exploring the codebase, identify the root cause, and file a GitHub issue with a TDD-based fix plan.

  ```
  npx skills add O-Marsters-1997/skills --skill triage-issue
  ```

- **improve-codebase-architecture** — Explore a codebase for architectural improvement opportunities, focusing on deepening shallow modules and improving testability.

  ```
  npx skills add O-Marsters-1997/skills --skill improve-codebase-architecture
  ```

- **scaffold-exercises** — Create exercise directory structures with sections, problems, solutions, and explainers.

  ```
  npx skills add O-Marsters-1997/skills --skill scaffold-exercises
  ```

## Tooling & Setup

- **clean-comments** — Set up a `/clean-comments` command and a PreToolUse hook that blocks git commits when staged code files contain self-documenting comments.

  ```
  npx skills add O-Marsters-1997/skills --skill clean-comments
  ```

## Writing & Knowledge

- **edit-article** — Edit and improve articles by restructuring sections, improving clarity, and tightening prose.

  ```
  npx skills add O-Marsters-1997/skills --skill edit-article
  ```

- **ubiquitous-language** — Extract a DDD-style ubiquitous language glossary from the current conversation.

  ```
  npx skills add O-Marsters-1997/skills --skill ubiquitous-language
  ```

