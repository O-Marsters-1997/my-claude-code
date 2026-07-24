---
name: flow-friction-audit
description: Drive a live web flow through a real browser, turn the run into evidence-backed friction findings, then route those findings into the impeccable skill for critique and redesign. Use whenever the user wants to find where a real user gets stuck, confused, or slowed down in an actual working flow — "audit the signup flow", "where's the friction in checkout", "walk through onboarding and tell me what's clunky", "usability audit of the live app", "why do people drop off on this page". Trigger even when the user says "usability", "UX audit", "friction", "drop-off", or "walk the flow" without naming this skill, and even when they only point you at a URL and ask what's wrong with the experience. NOT for static mockups, Figma files, or design critique of something that isn't running (use impeccable directly for those) and NOT for writing Playwright test suites (use playwright-cli).
allowed-tools: Bash(playwright-cli:*), Bash(od:*), Bash(odo:*)
---

# Flow Friction Audit

Drive a real, running flow through the browser and turn what happens into friction evidence a designer can act on. The value is the **middle**: most UX tooling either drives a browser (no judgement) or critiques a design (no live behaviour). Do both — observe the running product, convert each interaction into named evidence, then hand it to `impeccable` for the critique and the fixes.

Your job ends at evidence and named problems. `impeccable` owns the opinion about what "good" looks like and what to build instead.

## Prerequisites

- **A browser driver** — required. Default to `playwright-cli`; behind enterprise SSO, drive the user's authenticated Chrome session with `claude-in-chrome` (stage 1b). If neither is available, stop and say so rather than improvising.
- **the `playwright-cli` skill** — optional. Defer to it for command syntax rather than reinventing flags here.
- **Open Design (`od`/`odo`)** — optional; stage 5's wireframe renderer. If it's absent or the renderer is unavailable, fall back to ASCII and say so — never hard-fail an audit for a missing design tool. [references/open-design.md](references/open-design.md).

## The pipeline

Five stages with one human gate at 3b. Don't skip scope to "just start clicking" — an unscoped walk produces observations nobody can prioritise. Don't skip the gate to get to the designs; that's the more expensive mistake.

### 1. Scope

Settle three things before opening a browser (ask only for what you can't infer):

- **The goal.** *Friction discovery* (where does anyone get stuck?), *conversion* (where do people drop before the key action?), or *concept validation* (is the core idea legible?). The goal decides what counts — a slow-but-clear step is a finding for conversion, noise for concept validation.
- **One flow.** A flow can span several screens; auditing 3–5 flows at once cannot. This skill is precise and re-invoked per scenario, not run as a sprawling sweep.
- **An explicit URL**, plus credentials and test data. The URL must come from the user — **never infer it, never default to `localhost`, never start a dev server.** Environments differ in seeded data, feature flags and auth, so a guessed URL silently audits the wrong product. If none was given, ask before anything else.

State the scope back in a line or two and start.

### 1b. Auth precondition

Confirm the app is authenticated before driving — a walk that starts logged-out audits the login screen instead of the flow. Open the URL, `snapshot`, and check the target UI rendered rather than a sign-in page or a stuck loading overlay.

**Never automate the login.** Google and friends block automated sign-in (CAPTCHA, 2FA, bot detection), so scripting it is flaky and burns the run. If the flow can't continue, stop and ask the user to log in manually in the open session, then resume. One session runs the whole audit (stage 2), so a single manual login carries through every step.

**Behind enterprise SSO**, a fresh `playwright-cli` profile can't get past at all — no cookies, no device trust, and manual login inside a throwaway profile is often blocked outright. Drive the user's already-authenticated browser with `claude-in-chrome` instead; everything downstream is unchanged. Record which driver was used in `findings.md` — evidence from a live account has different provenance than a clean profile, and a reader needs to know which they're looking at.

Plain's `support-app` default and the logged-out state: [references/auth.md](references/auth.md).

### 2. Drive

Attach to the user's running browser at the stage-1 URL rather than launching a fresh profile — the running session carries the auth and app state a real user arrives with. Keep that one session for the whole audit; teardown happens once, at stage 5.

Work from snapshot element refs, never hand-rolled selectors: `snapshot` to read the page, act on refs like `e15`, re-`snapshot` after anything that changes it. Save each snapshot into the run's artifacts directory (stage 4) so the evidence outlives the browser.

**Walk one step at a time, logging each as a triple:**

```
element (ref + label) → expected outcome → actual outcome
```

The gap between expected and actual is where much of the friction lives:

- **Mismatch** (actual ≠ expected) → confusion. The label promised one thing, the app did another.
- **Retry** (you acted more than once, or hunted for the right element) → "wait, how do I…".
- **Long wait / no feedback** → invisible system status.
- **Dead end** (expected a next step, got nothing) → the flow breaks here.

But the triple is one lens, not the whole game. A step can do exactly what its label promised and still overload: a dense screen with no obvious starting point, a primary action buried among competing controls, information you must carry from an earlier screen. So at every step also ask "would a first-time user know where to start?" and "is everything needed to decide visible right now?" — the *recognition-over-recall* and *aesthetic/minimalist* questions the rubric scores, and the friction automated walks miss most.

Record the wait qualitatively (instant / a beat / had-to-wait / stalled). You don't need milliseconds; you need to know which steps made you wait without telling you why.

**Audit intended behaviour, not defects.** The triple lens surfaces anomalies, and most anomalies are bugs — not what this skill exists to find. A design finding is one where the product worked as built and still failed the user. So re-test every anomaly (dead click, no response, error banner, blank region) on a fresh load before promoting it:

- Reproduces identically → may be a design issue; score it in stage 3.
- Intermittent, or gone on the retry → a bug or a browser-automation artifact. Raise it with the user so it isn't lost, but keep it out of the report — never score it as friction.

A flaky click scored as major friction crowds out the real findings and sends `impeccable` designing around a defect someone should just fix.

### 3. Score

Turn the step log into **named findings** with the timed rubric in [references/rubric.md](references/rubric.md) — read it now. It converts Nielsen's heuristics into concrete questions ("was the result visible within ~1s?") so a finding is a rubric failure with evidence attached, not a vibe.

Two heuristics earn extra weight, being the ones automated walks surface best and teams miss most:

- **User control & freedom** — can the user undo, cancel, go back, escape a state they landed in by mistake? Watch every point where you felt trapped.
- **Recognition rather than recall** — is what's needed visible at the moment of decision? Watch every point where you had to hold something in your head.

Work at the **micro level**: one interaction ("the email field only errors after submit"), not a whole-flow grade ("signup is confusing"). Whole-flow verdicts can't be fixed or re-tested; micro findings can.

Give every finding a **stable, descriptive ID** — `<flow>-<what>`, e.g. `signup-email-late-validation`. The ID describes the friction, not its position, so a re-run diffs cleanly by ID and shows exactly what closed. Never bare sequence numbers.

Then **synthesise the findings into 2–4 named problems.** A finding is an observation; a problem is what a cluster of them *means* — "arrival is empty of purpose", not ten separate complaints about one screen. Each problem cites the finding IDs behind it. Problems are what gets designed against; findings are what prove them.

### 3b. Align the problems with the user

**Stop and get agreement before spending anything downstream.**

Stages 2 and 3 run **blind** — don't ask the user what they think is wrong while walking or scoring, because a steered walk finds the friction they already suspected and misses the rest. This gate is the first point of human input, placed after the evidence exists and before any design work does.

Present each problem with its findings, then ask:

- Anything missing — friction you know about that the walk didn't reach?
- Anything you disagree with, or that looks mis-rated?
- Anything mis-clustered — one problem that's really two, or two that are really one?

Revise until the user agrees, then carry the aligned problems into stages 4 and 5. Everything after this — critique, research, fix directions, wireframes — is expensive and shaped entirely by this list.

### 4. Report

One artifacts directory per run: `friction-audit/<flow>-<YYYYMMDD>/` under the working directory, holding the snapshots and `findings.md`. A stable home lets a re-run diff cleanly against the last.

Write `findings.md` with [assets/findings-template.md](assets/findings-template.md), which fixes the section order. Two rules that order enforces:

- **Lead with a summary** — goal, flow, driver, a one-paragraph verdict, the aligned problems as bullets, headline counts. A reader should get the story in thirty seconds without reconstructing it from a ranked list.
- **Problems before findings**, each citing its IDs; the findings follow as the layer that proves them, then the step log. State the conclusion, then show the work — and state it **once**, so no restatement further down can drift out of sync.

This is what `impeccable` reads and what a re-run diffs against; keep the structure stable so diffs stay legible.

### 5. Route to impeccable, then close

Close the browser — the audit is done driving. Then hand off to `impeccable`, which owns both the critique and the fixes:

- **Critique:** invoke `impeccable` in `critique` mode against the audited surface with `findings.md` as evidence, so its review is grounded in observed behaviour rather than a static read.
- **Research the precedent, before generating anything.** Take the aligned problems as-is; don't re-derive pain points here. For each, research how well-regarded products solve *that specific* problem, and read enough to say what each does right. A fix direction invented in a vacuum is a guess; the precedent is what makes it an argument.
- **Fan out fix directions:** invoke `impeccable` in `shape` mode with each problem plus its precedent, asking for **three divergent directions per problem**. Render each as a low-fi wireframe grounded in the surface's current code. Fidelity and framing: [references/wireframe-style.md](references/wireframe-style.md), read before drawing. Render mechanics and the fallback: [references/open-design.md](references/open-design.md). (Not `design-an-interface`; deprecated.)
- **Close with the trail:** end `findings.md` with `## Inspiration — where to look next` — per problem, the products worth studying and one line on what each does right. The research is worth more as a trail the user can follow than as a citation buried in a fix direction.

Report what you handed off and what came back. Don't editorialise a second critique on top.

## Re-runs (did the fix land?)

Re-drive the same flow, regenerate `findings.md`, and diff by finding ID: an ID that's gone is fixed, one that survived isn't, a new one is a regression. Report the delta, not a fresh wall of findings — that one-glance answer is the whole point of stable IDs.

## Scope boundaries

- **You don't critique or redesign** — `impeccable` does. Rendering its fix direction is drawing someone else's idea, not forming your own.
- **You don't hunt bugs** — defects and automation artifacts are listed out of scope (stage 2), never scored as friction. A broken thing needs fixing, not redesigning.
- **You don't write tests** — the walk is throwaway observation. Regression tests are `playwright-cli`'s job.
- **You don't do participant recruiting, fidelity ladders, video, or network mocking** — one agent, one live flow.
