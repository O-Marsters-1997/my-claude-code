---
name: flow-friction-audit
description: Drive a live web flow through a real browser, turn the run into evidence-backed friction findings, then route those findings into the impeccable skill for critique and redesign. Use whenever the user wants to find where a real user gets stuck, confused, or slowed down in an actual working flow — "audit the signup flow", "where's the friction in checkout", "walk through onboarding and tell me what's clunky", "usability audit of the live app", "why do people drop off on this page". Trigger even when the user says "usability", "UX audit", "friction", "drop-off", or "walk the flow" without naming this skill, and even when they only point you at a URL and ask what's wrong with the experience. NOT for static mockups, Figma files, or design critique of something that isn't running (use impeccable directly for those) and NOT for writing Playwright test suites (use playwright-cli).
allowed-tools: Bash(playwright-cli:*)
---

# Flow Friction Audit

You drive a real, running flow through the browser and turn what actually happens into friction evidence a designer can act on. The unique value here is the **middle**: most UX tooling either drives a browser (no judgement) or critiques a design (no live behaviour). You do both — you *observe the running product* and convert each interaction into named, evidence-backed findings, then hand those findings to `impeccable` for the critique and the fixes.

Do not reimplement critique or design generation. Your job ends at evidence + named findings; `impeccable` owns the opinion about what "good" looks like and what to build instead.

## Prerequisites

- **`playwright-cli` (the CLI)** — required. This skill has no other driving layer; if it isn't installed, stop and say so rather than improvising.
- **the `playwright-cli` skill** — optional but recommended. Defer to it for command syntax and driving best-practices; don't reinvent CLI flags here.

## The pipeline

Five stages, in order. Don't skip the scope step to "just start clicking" — an unscoped walk produces a pile of observations nobody can prioritise.

### 1. Scope

Before opening a browser, settle three things (ask the user only for what you can't infer):

- **The goal of this run.** Is it *friction discovery* (where does anyone get stuck?), *conversion* (where do people drop before the key action?), or *concept validation* (is the core idea even legible?)? The goal decides what counts as a finding — a slow-but-clear step is a real finding for conversion, noise for concept validation.
- **The flow.** Pick **one** target flow and audit it well — a flow can span several screens and interactions, but resist widening to 3–5 flows in one run. This skill is meant to be precise and re-invoked per scenario, not run as a sprawling general audit; a multi-flow sweep bloats context and produces observations nobody can prioritise.
- **The starting URL and any credentials/test data** needed to actually complete the flow.

State the scope back in one or two lines and start.

### 1b. Auth precondition

Before driving, confirm the app is actually authenticated — many flows sit behind a login wall, and a walk that starts logged-out audits the login screen instead of the flow. Open the start URL, `snapshot`, and check the target UI rendered — not a sign-in page or a stuck loading overlay.

**Never automate the login.** Providers like Google block automated sign-in (CAPTCHA, 2FA, bot detection), so scripting it is flaky and burns the run. If the app is unauthenticated and the flow can't continue, **stop and ask the user to log in manually in the open browser session**, then resume from where you paused. Because one session is kept for the whole run (stage 2), a single manual login at the start carries through every step.

Plain's `support-app` default and how to recognise the logged-out state: see [references/auth.md](references/auth.md).

### 2. Drive

Open **one** browser session and keep it for the whole audit — teardown happens once, at the very end (stage 5's close), not between flows. A single session preserves login and app state so later flows start where a real returning user would; note that if state from an earlier flow visibly leaks into a later one, that leak is itself a finding, not a reason to restart.

Use `playwright-cli` as the driving layer, working from snapshot element refs (never hand-rolled selectors): `snapshot` to read the page, act on refs like `e15`, and re-`snapshot` after anything that changes the page since refs only survive the snapshot you took them from. For command syntax and driving best-practices, defer to the `playwright-cli` skill rather than memorising flags here.

Save each snapshot into the run's artifacts directory (see stage 4) as you go, so the evidence is reviewable after the browser closes.

**Walk each flow one step at a time, logging every step as a triple:**

```
element (ref + label) → expected outcome → actual outcome
```

The triple is one lens, not the whole game — the clearest one, because the gap between expected and actual is where a lot of friction lives:

- **Mismatch** (actual ≠ expected) → confusion. The label promised one thing, the app did another.
- **Retry** (you had to act more than once, or hunt for the right element) → friction. A real user feels this as "wait, how do I…".
- **Long wait / no feedback** (the page sat there after an action with no spinner, toast, or change) → slow / invisible system status.
- **Dead end** (expected a next step, got nothing) → the flow is broken here.

But some friction never shows as a gap: a step can do exactly what its label promised and still overload — a dense screen with no obvious starting point, a primary action buried among competing controls, information you must carry from an earlier screen. So at every step also ask "would a first-time user know where to start here?" and "is everything needed to decide visible right now?" — these are the *recognition-over-recall* and *aesthetic/minimalist* questions the rubric scores in stage 3, and the class of friction automated walks miss most.

Record the wait qualitatively per step (instant / a beat / had-to-wait / stalled) — you don't need millisecond timing, you need to know which steps made you wait without telling you why. Keep the running log; it becomes the evidence section of the report.

### 3. Score

Turn the raw step log into **named findings** using the timed heuristic rubric in [references/rubric.md](references/rubric.md). Read that file now — it converts Nielsen's heuristics into concrete, timed review questions ("was the result of this action visible within ~1s?") so a finding is a rubric failure with evidence attached, not a vibe.

Two heuristics earn extra weight because they're the ones automated walks surface best and teams miss most:

- **User control & freedom** — can the user undo, cancel, go back, escape a state they landed in by mistake? Watch every point where you felt trapped.
- **Recognition rather than recall** — does the UI show what's needed at the moment of decision, or must the user remember it from a previous screen? Watch every point where you had to hold something in your head.

Work at the **micro level**: a finding is about one interaction ("the email field accepts an invalid address and only errors after submit"), not a whole-flow grade ("signup is confusing"). Whole-flow verdicts can't be fixed or re-tested; micro findings can.

Give every finding a **stable, descriptive ID** — `<flow>-<what>`, e.g. `signup-email-late-validation`, `checkout-no-back-from-payment`. The ID must describe the friction, not its position in a list, so that when the flow is fixed and you re-run, the same underlying issue keeps the same ID and a diff shows exactly what closed. Never use bare sequence numbers.

### 4. Report

Create one artifacts directory per run — `friction-audit/<flow>-<YYYYMMDD>/` under the current working directory — holding the snapshots you saved while driving and the `findings.md` below. A stable home makes the run visualisable and lets a re-run diff cleanly against the last.

Write findings to `findings.md` (in that directory) using [assets/findings-template.md](assets/findings-template.md). The report carries: the run goal, the per-flow step-log evidence, and the named findings (each with ID, flow, the triple that exposed it, the heuristic it fails, and a severity). This is the artifact `impeccable` reads and the baseline a re-run diffs against — keep it human-readable and stable in structure so diffs stay legible.

### 5. Route to impeccable, then close

Close the browser (`playwright-cli close`) — the audit is done driving.

Hand the findings to `impeccable`, which owns both the critique and the fixes:

- **Critique the current experience:** invoke `impeccable` in `critique` mode against the audited surface, feeding it `findings.md` as the evidence so its heuristic review is grounded in observed behaviour, not a static read.
- **Fan out fix directions:** from the findings, name the 2–3 **main** pain points (highest-severity or most-repeated friction). For each, research how well-regarded apps solve the same problem — don't invent in a vacuum — and carry that prior art in as reference. Then invoke `impeccable` in `shape`/redesign mode with each pain point as a problem statement plus its precedent, asking for **three divergent fix directions per pain point** and, if `impeccable` can produce them, low-fidelity wireframes so the directions are legible at a glance. `impeccable` generates the options — you supply the problem, the evidence, and the precedent. (Do not reach for `design-an-interface`; it is deprecated.)

Tell the user what you handed off and what `impeccable` came back with. Don't editorialise a second critique of your own on top.

## Re-runs (did the fix land?)

When asked to re-audit after changes, re-drive the same flows, regenerate `findings.md`, and diff against the previous one **by finding ID**: an ID that's gone is fixed, an ID that survived isn't, a new ID is a regression. Report the delta, not a fresh wall of findings — the point of stable IDs is that the second run answers "did we fix it?" in one glance.

## Scope boundaries

- **You don't critique or redesign** — `impeccable` does. You produce evidence and named findings and route them.
- **You don't write tests** — the walk is throwaway observation, not a suite. If the user wants regression tests, that's `playwright-cli`'s job.
- **You don't do participant recruiting, fidelity ladders, video, or network mocking** — this is a single automated agent walking a live flow, nothing more.
