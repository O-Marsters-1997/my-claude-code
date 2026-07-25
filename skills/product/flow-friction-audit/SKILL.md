---
name: flow-friction-audit
description: Drive a live web flow through a real browser, turn the run into evidence-backed friction findings, then route those findings into the impeccable skill for critique and redesign. Use whenever the user wants to find where a real user gets stuck, confused, or slowed down in an actual working flow — "audit the signup flow", "where's the friction in checkout", "walk through onboarding and tell me what's clunky", "usability audit of the live app", "why do people drop off on this page". Trigger even when the user says "usability", "UX audit", "friction", "drop-off", or "walk the flow" without naming this skill, and even when they only point you at a URL and ask what's wrong with the experience. Also use when the problems are already agreed and the user wants the design half — "shape fix directions from this problems doc", "we've aligned on the problems, now design against them" — which re-enters this skill at its fan-out stage. NOT for static mockups, Figma files, or design critique of something that isn't running (use impeccable directly for those) and NOT for writing Playwright test suites (use playwright-cli).
allowed-tools: Bash(playwright-cli:*), Bash(od:*), Bash(odo:*)
---

# Flow Friction Audit

Drive a real, running flow through the browser and turn what happens into friction evidence a designer can act on. The value is the **middle**: most UX tooling either drives a browser (no judgement) or critiques a design (no live behaviour). Do both — observe the running product, convert each interaction into named evidence, agree what those problems are, and only then shape fixes against them.

Evidence and agreed problems are the load-bearing output; they stand on their own even if nothing is designed. `impeccable` owns the critique and sharpens the directions — it doesn't decide what's wrong.

## Two modes, one skill

The pipeline splits at the 3b alignment gate, and each half is invoked separately. Running all five stages in one pass makes a single audit expensive, and it leaves the agreed problem space living only in that conversation — so the next run re-derives what's wrong instead of designing against it.

- **`audit` — stages 1 to 4.** Walk the flow, score it, agree the problems, publish the alignment doc and `findings.md`. Ends at alignment; no design work.
- **`shape` — stage 5.** Take an already-aligned problem set and produce researched fix directions. Re-runnable: the problem space is settled, so each invocation spends its budget on ideas rather than rediscovering the friction.

**Route on entry:** the user points at or names an aligned-problems doc → `shape`. Otherwise → `audit`, from stage 1. **Never go looking for a doc the user didn't name** — auto-discovery is the same mistake as reading a prior audit ([Fresh runs](#fresh-runs)): it quietly makes someone else's stale problem set the frame for a flow you haven't looked at yet.

## Prerequisites

- **A browser driver** — required in `audit` mode, irrelevant in `shape`. Default to `playwright-cli`; behind enterprise SSO, drive the user's authenticated Chrome session with `claude-in-chrome` (stage 1b). If neither is available, stop and say so rather than improvising.
- **the `playwright-cli` skill** — optional. Defer to it for command syntax rather than reinventing flags here.
- **Notion access** — for the stage-3b alignment doc. Without it, write the doc as markdown alongside `findings.md` and tell the user it needs publishing by hand; the shared artifact matters more than where it lives.
- **Open Design (`od`/`odo`)** — optional; `shape` mode's wireframe renderer. If it's absent or the renderer is unavailable, fall back to ASCII and say so — never hard-fail an audit for a missing design tool. [references/open-design.md](references/open-design.md).

## The pipeline

Five stages, hinged on the human gate at 3b. Don't skip scope to "just start clicking" — an unscoped walk produces observations nobody can prioritise. Don't skip the gate to get to the designs; that's the more expensive mistake.

### 1. Scope

Settle three things before opening a browser (ask only for what you can't infer):

- **The goal.** *Friction discovery* (where does anyone get stuck?), *conversion* (where do people drop before the key action?), or *concept validation* (is the core idea legible?). The goal decides what counts — a slow-but-clear step is a finding for conversion, noise for concept validation.
- **One flow.** A flow can span several screens; auditing 3–5 flows at once cannot. This skill is precise and re-invoked per scenario, not run as a sprawling sweep.
- **An explicit URL**, plus credentials and test data. The URL must come from the user — **never infer it, never default to `localhost`, never start a dev server.** Environments differ in seeded data, feature flags and auth, so a guessed URL silently audits the wrong product. If none was given, ask before anything else.

**Start from nothing** — don't go looking for previous audits of this flow, and don't read one you happen to know about ([Fresh runs](#fresh-runs)). Then state the scope back in a line or two and start.

### 1b. Auth precondition

Confirm the app is authenticated before driving — a walk that starts logged-out audits the login screen instead of the flow. Open the URL, `snapshot`, and check the target UI rendered rather than a sign-in page or a stuck loading overlay.

**Never automate the login.** Google and friends block automated sign-in (CAPTCHA, 2FA, bot detection), so scripting it is flaky and burns the run. If the flow can't continue, stop and ask the user to log in manually in the open session, then resume. One session runs the whole audit (stage 2), so a single manual login carries through every step.

**Behind enterprise SSO**, a fresh `playwright-cli` profile can't get past at all — no cookies, no device trust, and manual login in a throwaway profile is often blocked outright. Drive the user's already-authenticated browser with `claude-in-chrome` instead; everything downstream is unchanged. Record the driver in `findings.md` — evidence from a live account has different provenance than a clean profile.

Plain's `support-app` default and the logged-out state: [references/auth.md](references/auth.md).

### 2. Drive

Attach to the user's running browser at the stage-1 URL rather than launching a fresh profile — the running session carries the auth and app state a real user arrives with. Keep that one session for the whole walk; close it once, at the end of stage 4, since `shape` never drives.

Work from snapshot element refs, never hand-rolled selectors: `snapshot` to read the page, act on refs like `e15`, re-`snapshot` after anything that changes it. **Don't persist screenshots** — the step log below is the evidence record, and images cost more context than they return. Prefer `read_page`/`get_page_text` for structure; screenshot only to see something a text read can't show, and discard it after.

**Mutating live state.** **Never** act in a way that notifies a third party or spends money — invites, outbound mail, OAuth grants, anything billable. **Freely** take genuinely reversible UI-state actions (expand, dismiss, open a composer you don't submit); no need to ask. Where only an irreversible action would answer a question, answer it from source instead (3a) — and name the steps you never exercised in the report, so a reader can tell an observed claim from an inferred one.

**Walk one step at a time, logging each as a triple:**

```
element (ref + label) → expected outcome → actual outcome
```

The gap between expected and actual is where much of the friction lives:

- **Mismatch** (actual ≠ expected) → confusion. The label promised one thing, the app did another.
- **Retry** (you acted more than once, or hunted for the right element) → "wait, how do I…".
- **Long wait / no feedback** → invisible system status.
- **Dead end** (expected a next step, got nothing) → the flow breaks here.

The triple is one lens, not the whole game — a step can do exactly what its label promised and still overload. So at every step also ask "would a first-time user know where to start?" and "is everything needed to decide visible right now?" — the *recognition-over-recall* and *aesthetic/minimalist* questions the rubric scores, and the friction automated walks miss most.

Record the wait qualitatively (instant / a beat / had-to-wait / stalled) — which steps made you wait without telling you why.

**Audit intended behaviour, not defects.** A design finding is one where the product worked as built and still failed the user. Browser automation is janky by nature, so an anomaly (dead click, error banner, blank region) is a signal to re-test on a fresh load, never a result. Reproduces identically → may be design; score it. Intermittent or gone on retry → a bug; mention it to the user in conversation and stop there.

**Defects never enter the report** — not as a finding, not as evidence inside one, and not as an out-of-scope section. A defect list beside the findings competes for attention and sends `impeccable` designing around something someone should just fix.

### 3. Score

Turn the step log into **named findings** with the timed rubric in [references/rubric.md](references/rubric.md) — read it now. It converts Nielsen's heuristics into concrete questions ("was the result visible within ~1s?") so a finding is a rubric failure with evidence attached, not a vibe.

Two heuristics earn extra weight, being the ones automated walks surface best and teams miss most: **user control & freedom** — watch every point where you felt trapped with no undo, cancel or way back — and **recognition rather than recall** — watch every point where you had to hold something in your head instead of reading it off the screen.

Work at the **micro level**: one interaction ("the email field only errors after submit"), not a whole-flow grade ("signup is confusing"). Whole-flow verdicts can't be fixed or re-tested; micro findings can.

Give every finding a **stable, descriptive ID** — `<flow>-<what>`, e.g. `signup-email-late-validation` — describing the friction, not its position. Never bare sequence numbers. **Every finding cites its evidence:** the step-log row number(s) it rests on, plus `file:line` for anything read from source. A claim a reader can't trace isn't a finding.

Then **synthesise the findings into 2–4 named problems.** A finding is an observation; a problem is what a cluster of them *means* — "arrival is empty of purpose", not ten separate complaints about one screen. **Each finding belongs to exactly one problem.** If a finding seems to belong to three, the problems aren't yet distinct — re-cluster until each owns its own evidence. Problems are what gets designed against; findings are what prove them.

### 3a. Corroborate against source

When the audited product's source is available, read the components behind each finding before scoring is final. It turns "the progress bar never moved" into "the criterion measures threads, not connected channels", and answers questions that would otherwise need a live mutation. **State no claim the code doesn't support** — if a criterion ticks under some condition you didn't hit, name that condition rather than rounding it up to "never".

### 3b. Align the problems with the user

**Stop and get agreement before spending anything downstream.** Stages 2 and 3 run **blind** — a steered walk finds the friction the user already suspected and misses the rest — so this gate is the first point of human input, after the evidence exists and before any design work does.

Align **on a document, not in chat.** Publish the problems as a new child page under the growth team's Notion home, **📈 Home**, and iterate there: a settled problem space is worth more than an agreement one person remembers — the team reads the same statement of what's wrong, and `shape` re-enters from it as many times as the ideas need reworking.

The page carries a descriptive title, the flow's goal with an honest assessment of how well the flow serves it, and a `## problems` list — things to solve outright or areas to mitigate. **No solution ideas on this page.** An idea smuggled in as a problem gets agreed by proxy and quietly becomes the only direction anyone designs.

Template, the questions to ask while aligning, and the bounded re-entry context the page must carry: [references/alignment-doc.md](references/alignment-doc.md).

Revise until the user agrees, then carry the aligned problems into stage 4. Everything after this — critique, research, fix directions, wireframes — is expensive and shaped entirely by this list.

### 4. Report

One artifacts directory per run: **`<audited-repo>/.claude/friction-audit/<flow>-<YYYYMMDD>/`** — the repo of the product you audited, not wherever you were invoked from. An audit belongs next to its product. It holds `findings.md` and the wireframe HTML; no screenshots, and no throwaway scripts you used to generate them.

**Two homes, one link each.** Notion owns the summary and the problems — the product layer the team aligns on. This directory owns the evidence, and later the critique and fix directions. Link each to the other exactly once rather than restating either inside the other.

Write `findings.md` with [assets/findings-template.md](assets/findings-template.md), which fixes the section order. Two rules that order enforces: **lead with a summary** (goal, flow, driver, one-paragraph verdict, the problems as bullets, headline counts) so a reader gets the story in thirty seconds; and **problems before findings**, each citing its IDs, with the findings as the layer that proves them and the step log after. State the conclusion, then show the work — and state it **once**, so no restatement drifts out of sync.

### 5. Shape the fix directions (`shape` mode)

Entered with an aligned problem set — straight from 3b, or from a doc the user named. Read the doc and its re-entry context and take the problems as given; re-deriving them here throws away the agreement they represent. Close the browser first if this run drove one.

- **Critique:** invoke `impeccable` in `critique` mode against the audited surface with `findings.md` as evidence, so its review is grounded in observed behaviour rather than a static read. **Fold what comes back into `findings.md`** — score, specificity verdict and P0s as a `## Critique` section. A critique in a separate file gets read separately or not at all; the point is that its emphasis visibly drives the directions.
- **One research-and-ideation agent per problem.** This is the fan-out unit: each agent owns a single problem end to end. It researches how well-regarded products solve that problem — **at least one primary source** (the product's own docs, changelog, release notes, or a first-hand teardown), naming the specific surface examined; aggregator listicles find candidates but are never a citation, and **no statistic without a primary source**, since a vendor's conversion claim is marketing until you can point at where it's measured. Two or three worked examples beat a survey. It then ideates against the findings and returns **three divergent ideas argued in prose**, each naming the problem and the critique point it answers. Route them through `impeccable` in `shape` mode to sharpen. (Not `design-an-interface`; deprecated.)
- **Wireframes illustrate ideas; they never produce them.** Draw only once the ideas exist as prose, grounded in the surface's current code. The test: **if an idea can't be written as a paragraph before it's drawn, it isn't an idea yet.** Sketching first inverts the work — you end up reading rationale off your own drawings and shipping nine sketches with justifications retrofitted onto them, which is exactly the failure this ordering prevents. Fidelity and framing: [references/wireframe-style.md](references/wireframe-style.md), read before drawing. Render mechanics and the fallback: [references/open-design.md](references/open-design.md).
- **Close with the trail:** end `findings.md` with `## Inspiration — where to look next` — per problem, the products worth studying and one line on what each does right. The research is worth more as a trail the user can follow than as a citation buried in a fix direction.

Report what you handed off and what came back. Don't editorialise a second critique on top.

## Execution shape

Stages 2 and 3 stay one context, one session, serial — each step informs the next, and the 3b gate needs one context holding the whole evidence set. The rest is fan-out shaped:

- **3a corroboration and the stage-5 research agents** — independent per file and per problem, so run them as parallel subagents. Serial shallow searches are what produce listicle-grade precedent.
- **`impeccable critique`'s two assessments** — isolation is a *correctness* requirement here, not a speed-up: run inline and the critique has already absorbed your findings, so its agreement proves nothing. If subagents are unavailable, say so in the report and state that agreement is not corroboration.
- **Wireframe authoring** — single-context. One shared style pass beats the same fix made four times.

You stay the synthesis point; subagents return evidence, not conclusions.

**Match the model to the work.** An all-Opus run costs several times a tiered one and buys nothing on the mechanical stages, so spend the expensive model only where judgement actually happens:

- **Sonnet** — 3a source corroboration, precedent gathering, and wireframe authoring and rendering. These are read-heavy and style-constrained; the reference files already carry the judgement.
- **Opus** — scoring and clustering (stage 3), each problem's idea synthesis, and `impeccable critique`. Deciding what a cluster of findings *means*, and what to build instead, is the part that degrades when you economise.

## Fresh runs

**Every invocation starts clean.** Don't search for prior `friction-audit` directories, don't read one you know exists, don't present cached findings as current. Old evidence goes stale, and a prior run's fix directions quietly become the ceiling on the new ones — the second audit reproduces the first instead of seeing the flow again. Compare across runs **only when the user asks**; then diff by finding ID (gone is fixed, surviving isn't, new is a regression) and say which prior IDs you couldn't re-test.

An alignment doc the user hands you in `shape` mode isn't an exception to this — they named it, so it's their current problem set, not a cached one you went looking for.

## Scope boundaries

- **You don't own the critique** — `impeccable` does, and its verdict gets folded in rather than argued with. Ideation is shared: the per-problem research agents form the ideas, `impeccable` sharpens them.
- **You don't decide the problems alone** — the 3b gate does, with the user. A problem set you never got agreed is a guess with a document around it.
- **You don't hunt bugs** — defects and automation artifacts never reach the report (stage 2). A broken thing needs fixing, not redesigning.
- **You don't write tests** — the walk is throwaway observation. Regression tests are `playwright-cli`'s job.
- **You don't do participant recruiting, fidelity ladders, video, or network mocking** — one agent, one live flow.
