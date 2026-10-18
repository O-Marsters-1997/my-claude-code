# Pattern catalogue

Categories to hunt through in Step 2. Each lists where to look and what a skill-worthy
instance looks like. Skip categories the repo has no footprint in.

## Adding a unit of the main abstraction

The highest-yield category: whatever the repo is mostly made of, and how a new one is added.

- **Look at:** route tables, handler or controller dirs, component dirs, CLI command dirs,
  plugin or provider dirs, job or worker dirs, GraphQL resolvers, gRPC services.
- **Count:** files matching the dominant naming pattern in that dir.
- **Skill-worthy when:** a new unit needs files in several places (handler + route + schema +
  test + docs), or a registration step lives far from the unit itself.
- **Find the steps:** `git log --diff-filter=A --stat -- <dir>` and read two or three commits
  that added a unit. The other files those commits touched are the hidden steps.

## Schema and data changes

- **Look at:** `migrations/`, `db/`, `prisma/`, `alembic/`, `schema.sql`, ORM model dirs,
  seed and fixture data.
- **Skill-worthy when:** a schema change also needs model updates, regenerated types, a
  backfill, fixture changes or a down migration, and instances vary in which they include.
- **Inconsistency tells:** migrations without a matching down step, model fields with no
  migration, generated files out of step with the schema.

## Code generation and registration

Steps that are easy to forget because nothing fails loudly until later.

- **Look at:** `//go:generate`, `*.proto`, OpenAPI specs, `codegen` config, `generated/` or
  `*_gen.*` files, registries (`init()` registration, `index.ts` barrel exports, plugin
  manifests, DI containers, enum-to-handler maps).
- **Skill-worthy when:** editing a source of truth requires regenerating and also hand-editing
  something the generator does not touch. Pure "run the generator" belongs to a script or hook.

## Test construction

- **Look at:** test helpers, fixtures, factories, `testdata/`, `conftest.py`, test containers,
  mocks and fakes dirs, golden files.
- **Count:** tests using each helper or fixture style.
- **Skill-worthy when:** a correct test for a given layer needs a specific setup sequence
  (spin up a db, seed, use a particular fake, clean up) that differs by layer.
- **Inconsistency tells:** two or more competing fixture styles, tests that hand-roll what a
  helper already does.

## Error handling and logging

- **Look at:** error types and wrappers, logger setup, middleware, `catch`/`except`/`if err`
  sites, metrics and tracing calls.
- **Usually belongs elsewhere:** "wrap errors with X" or "use the structured logger" is a
  rules line; banning `console.log` or `fmt.Println` is a linter rule.
- **Skill-worthy when:** handling an error means deciding between several paths (retry,
  surface to user, map to status code, alert) according to repo-specific rules.

## Integrations with internal or external services

- **Look at:** client packages, SDK wrappers, `adapters/`, `integrations/`, queue producers
  and consumers, webhook handlers, feature flag calls, config for third-party keys.
- **Skill-worthy when:** adding a new integration or call follows a set recipe (client
  wrapper, config entry, secret, retry policy, fake for tests, contract test).

## Configuration and feature flags

- **Look at:** config structs, env var loaders, `.env.example`, flag definitions, Helm values.
- **Skill-worthy when:** a new setting must be added in several places (struct, loader,
  example env, deploy manifests, docs) and instances miss some.

## Release, deploy and versioning

- **Look at:** CI deploy jobs, `release` scripts, changelogs, version files, tags, Dockerfiles,
  IaC, `RELEASING.md`, runbooks in `docs/`.
- **Usually belongs elsewhere:** fully scripted pipelines need no skill.
- **Skill-worthy when:** releasing involves judgement (choosing a version bump, writing the
  changelog entry, deciding on a migration window) or manual steps documented in prose.

## Docs that must track code

- **Look at:** API reference docs, generated docs, READMEs per package, ADRs, changelogs.
- **Skill-worthy when:** a code change of a given kind is expected to update a specific doc,
  and the docs show drift from the code.

## Cross-cutting checklists in docs

- **Look at:** `CONTRIBUTING.md`, PR templates, `docs/` how-tos, checklists in issue templates.
- These are often a procedure already written in prose. A how-to that describes a multi-step
  change is strong evidence for a skill, unless an existing skill already encodes it.

## Style and formatting

Nearly always "belongs elsewhere".

- **Look at:** naming, import order, file layout, quote style, comment style.
- Route to a linter or formatter rule when a tool can check it, otherwise to a CLAUDE.md line.
  Note when an existing config already enforces it, and do not list it at all in that case.
