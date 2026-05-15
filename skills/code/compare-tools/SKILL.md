---
name: compare-tools
description: >
  Help the user select the right technology, library, or tool for a specific use case in the
  current project. Explores the codebase first to understand its shape, then aligns with the
  user on high-level priorities (community support, maintenance, vendor lock-in, extensibility,
  etc.), researches the market, and outputs a structured recommendation report to ./ideas/ that
  includes every viable option with doc links, a side-by-side comparison table, and a ranked
  recommendation list with a compatibility score out of 100 derived from a rubric built from the
  user's own priorities. Use whenever the user asks "which library should I use", "help me pick
  a tool", "what's the best X for this project", "compare these technologies", "should I use X
  or Y", "what are my options for Z", or any similar technology selection question. Trigger even
  if the user doesn't phrase it as a comparison — if they need to *choose* a technology for their
  current project, this skill applies.
---

# Compare Tools

Help the user pick the right technology for a specific use case. The output is a grounded,
auditable recommendation — not a generic listicle, but a score derived from *this codebase's*
shape and *this user's* stated priorities.

## Operating principles

- **Codebase first.** Technology decisions are context-dependent. Understand what's already here
  before evaluating what might be added.
- **Priorities are the rubric.** Don't rank tools on your own taste. Derive the ranking from the
  user's stated priorities weighted against observed options. Show the arithmetic.
- **Push back when it matters.** If a stated priority contradicts the observed codebase (e.g.
  "avoid vendor lock-in" in a repo already deep in AWS), name that tension and ask them to
  reconcile. Don't push back reflexively — only when the conflict is real and material.
- **Research, don't invent.** Don't claim a tool has a feature unless you can verify it. If
  unsure, say so or omit the claim.
- **One question at a time.** Use `AskUserQuestion` for every clarification. Never batch
  questions — each answer may change what you ask next.

---

## Phase 1 — Alignment

### 1a. Explore the codebase

Read the following in order, stopping when you have enough to write a one-paragraph
characterisation:

1. `README.md` and `CLAUDE.md` (or `CLAUDE.local.md`) — project intent and conventions
2. Top-level `docs/` if it exists
3. Dependency manifests — whichever of these exist:
   - `package.json` / `package-lock.json` (Node/TS)
   - `pyproject.toml` / `requirements.txt` (Python)
   - `go.mod` (Go)
   - `Cargo.toml` (Rust)
   - `Gemfile` (Ruby)
4. Recent git log: `git log --oneline -20`

Characterise the codebase in one paragraph: *type of project, primary language/runtime,
what's already installed, who it's for, what it does.* Surface this so the user can correct
a wrong framing before the research phase.

### 1b. Confirm the use case

State your understanding of what the user wants to choose a tool for and ask them to confirm
or correct it. If the use case is under-specified — "I need a database", "pick a queue",
"something for caching" — invoke the `grill-me` skill to resolve the decision tree:
scale, read/write mix, latency, hosting constraints, team familiarity, existing
infrastructure, etc. Skip `grill-me` if the use case is already crisp.

### 1c. Elicit priorities

Present a short list of common high-level priorities from `references/priority-catalogue.md`
and ask the user to pick 3–5 that matter most to them, then rank them. Use `AskUserQuestion`
— one question at a time. Rankings become the rubric weights in Phase 3.

Good default priorities to offer:
- Community support
- Maintenance activity
- Avoiding vendor lock-in
- Extensibility / composability
- Ecosystem fit (language + framework)
- License
- Performance
- Learning curve
- Cost
- Operational burden

### 1d. Sensible push-back

After the priorities are in, check them against what you observed in Phase 1a. If a stated
priority contradicts the codebase:

- **Real conflict**: "You said 'avoid vendor lock-in', but the repo already uses AWS Secrets
  Manager, CloudWatch, and S3. Adding a vendor-neutral queue would be swimming upstream — is
  that still the goal, or should we score SQS more favorably?" Ask via `AskUserQuestion`.
- **No conflict**: say nothing. Don't manufacture friction.

---

## Phase 2 — Research

Use `WebSearch` to survey the market. Aim for **4–7 candidates** covering:

- The mature/default choice (the boring, safe option)
- 1–2 newer or popular challengers
- At least one left-field option that could win on the user's specific priorities
- Any option already in the dependency manifests (it may already be installed)

For each candidate, record:

| Field | What to capture |
|-------|----------------|
| One-liner | What it is in one sentence |
| Strengths | Primary advantages, typical use cases |
| Weaknesses | Known pain points, limitations |
| Maintenance | Last release date, release cadence, # of active maintainers |
| License | MIT / Apache / GPL / commercial / etc. |
| Doc URL | Canonical documentation URL |

Don't invent features or claims. If you can't verify something, note "unverified" or omit it.

---

## Phase 3 — Output

### Resolve the output path

```bash
mkdir -p ./ideas
DATE=$(date -u +%Y-%m-%d)
SLUG="<short-kebab-case-topic>"   # e.g. "job-queue", "vector-db", "state-manager"
BASE="./ideas/${DATE}-tech-${SLUG}"
OUT="${BASE}.md"; N=2
while [ -e "$OUT" ]; do OUT="${BASE}-${N}.md"; N=$((N+1)); done
echo "$OUT"
```

Write the report to the echoed path using `references/report-template.md`.

### Report structure

The report must contain exactly these five sections in order:

1. **Context** — one paragraph: codebase characterisation + the confirmed use case. If `grill-me`
   ran, summarise the key answers that shaped the shortlist.

2. **Priorities & rubric** — the ranked priorities, their weights (must sum to 100), and scoring
   anchors for each priority (what 0 / 50 / 100 looks like). See `references/rubric.md` for how
   to derive weights and write anchors.

3. **Options** — one sub-section per candidate:
   ```
   ### [Tool name]
   [One-liner]
   - **Strengths**: ...
   - **Weaknesses**: ...
   - **License**: ... | **Last release**: ... | **Docs**: [url]
   ```

4. **Comparison table** — rows = tools, columns = each priority + a final "Score /100" column.
   Show the raw score per priority (0–100), then the weighted total. This lets the user audit
   your arithmetic.

   | Tool | Priority 1 (W%) | Priority 2 (W%) | ... | Score /100 |
   |------|----------------|----------------|-----|-----------|
   | A | 80 | 60 | ... | 72 |

5. **Ranked recommendations** — tools sorted descending by compatibility score. For each:
   - Score /100
   - 2–3 sentences tying the score back to observed codebase facts and the aligned priorities
   - Flag the **top recommendation** prominently

   End with: "Next step: feed this decision into `write-a-prd` or open a design doc."

---

## Quality bar (self-check before writing)

- [ ] Characterisation is grounded in files actually read, not assumed
- [ ] Use case was confirmed or clarified (via `grill-me` if fuzzy)
- [ ] Priorities were user-stated, not inferred from your own preferences
- [ ] Rubric weights sum to 100
- [ ] All scores are derivable from the stated anchors — no magic numbers
- [ ] Every option has a verified doc URL
- [ ] No invented features or unverified claims
- [ ] Push-back was offered only when a real conflict existed
- [ ] The report can be scanned in under 5 minutes
