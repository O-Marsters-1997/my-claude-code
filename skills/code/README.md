# Code skills

Seventeen skills for writing, designing and reviewing code. Unlike the workflow suite they are
mostly independent: each one loads on its own trigger, and several standards skills are meant to
load side by side. This file is the map for humans working on them.

## Installing

Install the ones that match your stack. Nothing breaks if you skip one, but a standards skill
that names a partner (`htmx-go-templates` → `go-idiomatic`) works best with the partner installed.

```
npx skills add O-Marsters-1997/my-claude-code --skill <name> -g -y
```

**Skills must stay at `skills/code/<name>/`.** The `skills` CLI discovers
`skills/<category>/<name>/SKILL.md` and nothing deeper; a skill nested one level further is
silently skipped by `npx skills add`.

**Keep every `description` a folded block scalar (`description: >`).** A plain one-line description
containing `": "` is invalid YAML, and the skill is dropped from `npx skills add --list` with no
error. After editing a description:

```
grep -l '^description: [^>|].*: ' skills/*/*/SKILL.md   # must print nothing
```

**Skills must not reference this file.** Installation is flat, so `../README.md` does not resolve
once installed. When one skill hands off to another, name the skill, not its path.

## Standards skills

These load *before* code is written, triggered by the file being touched rather than by the user
asking. Each description uses the same `TRIGGER — load BEFORE … / SKIP …` shape so they fire on
plain feature requests ("add an endpoint", "style this") and stay out of each other's way.

| Skill | Fires on | Loads alongside |
|---|---|---|
| `go-idiomatic` | any `.go`, `go.mod`, Go named in the prompt | `go-testing`, `tdd`, `htmx-go-templates`, `goquery-parsing` |
| `go-testing` | `_test.go`, `go test`, fakes, golden files, fuzzing, testcontainers | `go-idiomatic`, `tdd` |
| `htmx-go-templates` | `.tmpl`/`.gohtml`, `hx-*` attributes, `HX-*` headers, a handler calling `ExecuteTemplate` | `go-idiomatic` |
| `goquery-parsing` | writing a Go HTML parser with goquery | `go-idiomatic` |
| `frontend-best-practices` | `.tsx`/`.jsx`/`.vue`/`.svelte`, components, hooks | `tailwind-shadcn` |
| `tailwind-shadcn` | `class=`/`className=` edits, `components/ui/` | `frontend-best-practices` |
| `tailwind-design-system` | `@theme`, design tokens, Tailwind v4 config | `tailwind-shadcn` |
| `postgres-best-practices` | `.sql` files, migrations, schema, index, constraint or query changes on Postgres | `go-idiomatic` |
| `tdd` | any implementation or bug fix | everything |

Install `postgres-best-practices` with `npx skills add O-Marsters-1997/my-claude-code --skill postgres-best-practices -g -y`.

Boundaries that keep them from colliding:

- **Server templates vs components.** `hx-*` in a Go template routes to `htmx-go-templates`;
  `frontend-best-practices` covers React and Solid only and says so in its framework detection.
- **Go language vs Go web rendering.** `go-idiomatic` owns style, errors and concurrency;
  it routes `html/template` + htmx work to `htmx-go-templates` from its reference table.
- **Go code vs Go tests.** `go-testing` owns how tests are written, including test databases
  through its `pgtest` helper; a repo's `testing-policy` decides which tests each layer needs.
- **Tokens vs usage.** `tailwind-design-system` owns the theme file; `tailwind-shadcn` owns
  day-to-day classes inside an existing token set.

## Design and architecture

| Skill | Takes | Produces |
|---|---|---|
| `codebase-design` | a module or seam under discussion | shared deep-module vocabulary; a routine other skills call |
| `architecture-review` | the whole codebase | a diagnosis and prioritised roadmap; never refactors |
| `improve-codebase-architecture` | the whole codebase | an HTML report of deepening opportunities, then a grilling on one (slash-only) |
| `record-decision` | a decision just made | an ADR in `docs/adr/`, plus `CONTEXT.md` terms |
| `compare-tools` | a choice between libraries or tools | a ranked comparison for this project |

## Review and docs

| Skill | Takes | Produces |
|---|---|---|
| `code-review` | a branch, PR or range since a fixed point | Standards and Contract findings side by side |
| `clean-comments` | staged files | comments removed per `~/.claude/rules/comments.md` |
| `update-project-docs` | the codebase | synced docs and token-lean codemaps |
| `readme-generator` | the project | a plain README (slash-only) |

## Adding a code skill

- Standards skill: copy the `TRIGGER — load BEFORE … / SKIP …` description shape, keep
  `SKILL.md` to rules plus a "where to look" table, and push detail into `references/`.
- Ship `evals/evals.json` with made-up fixtures under `evals/fixtures/`, not code copied from
  another repo.
- Add a row to the right table above, and a routing line in any standards skill it borders.
