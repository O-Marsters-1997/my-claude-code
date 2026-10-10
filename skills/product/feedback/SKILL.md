---
name: feedback
description: >
  Turns the user's unstructured feedback about how a product or system behaves into verified,
  agent-ready GitHub issues. Splits a complaint into points, checks each against the code and
  recorded intent (ADRs, CONTEXT.md, git history), plays back what it found until the user
  agrees, then files one ticket per agreed finding under a shared label for /implement or
  /fleet dispatch. Use when the user says "/feedback", "I don't like how X works", "this isn't
  usable because…", "X feels wrong", "this behaviour is annoying", "push back on how Y
  behaves", or otherwise complains about the repo's behaviour or UX without a concrete bug
  report or spec. Not for feedback on a skill (use skill-feedback-collector), a single
  well-understood bug (triage-issue) or a ready-to-file follow-up (file-issue).
disable-model-invocation: true
---

# Feedback

The user's feedback is a claim about the system, not an instruction. Your job is to find out
what the system actually does and why, agree with the user on what is really wrong, and only
then file work an agent can pick up cold. Three stages: **verify**, **clarify**, **file**.

## Shape the input

Feedback arrives loose: "onboarding isn't usable because the form resets, errors are vague and
it's too slow". Scoping it is the user's job; splitting it is yours.

- Name the **area** in a short slug (`onboarding`). It becomes the label in stage 3.
- Number the **points** (P1, P2, P3), keeping the user's own words for each.

Don't ask the user to restructure anything. If something is too vague to verify, that is a
stage 1 result, not a reason to stop.

## 1. Verify

Explore the area once, then judge each point against what you found. For anything broader
than a couple of lookups, use an Explore subagent and take its summary. Brief it with the area
and every point, and ask for:

- the code paths behind each point, as files and symbols
- recorded intent: ADRs in `docs/adr/`, `CONTEXT.md`, README, and `git log` on the files
  involved, for why the behaviour is the way it is
- existing or in-flight work: open issues (`gh issue list --search`) and code that already
  does what the user wants, searched by concept, not by the user's wording

Give each point one verdict:

| Verdict | Meaning | Carries |
|---|---|---|
| Confirmed | The system does what the user says | code path; the recorded reason, if any |
| Intended | It does, deliberately | the ADR, doc or commit that decided it |
| Already handled | Built, or an open issue covers it | where, or the issue number |
| Unverified | Can't establish it from here | what's missing |

Points can merge ("P2 and P3 share one cause") or split ("P1 is two problems"). Do the regrouping
now and show it in the play-back.

Label every fact by source: **User** (what they said) or **Code** (what you found, with a
path). An assumption is labelled as one. This stops an inference being filed as a finding.

You may disagree. If the evidence says the behaviour is right, say so and show why. Agreeing to
be agreeable files tickets that the next agent will rightly push back on.

## 2. Clarify

Play back **every point at once**, in one message:

```markdown
## Feedback: <area>

**P1 — <user's words>** · Confirmed
- Code: `path/file.ts` `fn` does …
- Why: no recorded reason / ADR-0004 chose this because …
- Your objection, as I read it: …
- Proposed finding: <what should change, one line>

**P2 + P3 — merged** · …

Correct anything I've got wrong. Anything you don't mention, I'll take as agreed.
```

The user replies once, about everything they disagree with. **Silence on a point means they
accept your reading and reasoning**, including any pushback. Never ask about points one at a
time, and never re-ask about a point they let stand.

If the reply has corrections, re-verify **all** of them together in one pass, then play back
only the changed points plus a one-line list of what's settled. Repeat until a reply carries
no corrections. A reply like "fine" or "go" settles everything.

## 3. File

Each agreed finding becomes one GitHub issue. Points the user dropped, or Intended points they
accepted, are not filed; list them in the final report with a one-line reason.

**Route each finding:**

- **A clear change** → `status:ready`.
- **Reverses a recorded decision**, with the new direction settled in stage 2 → run
  `record-decision` for the superseding ADR first, then file the ticket linking it, `status:ready`.
- **Too big, or the direction is still open** → `needs-design` and no `status:*` label, so fleet
  leaves it alone until someone designs it.
- **Intended, accepted, but the reason was written nowhere** → one small docs ticket to record
  it, so the next person doesn't raise the same feedback.

**Ticket body**, `file-issue`'s format plus the evidence from stage 1, so the implementing agent
starts from verified facts instead of re-investigating:

```markdown
## Context

<1–3 sentences: what's wrong and why it matters, in the user's terms.>

## Evidence

- Code: `path/file.ts` `fn` — <what it does today>
- Intent: <ADR/commit/doc, or "none recorded">
- Agreed in feedback session: <the settled reading, one line>

## Where to look

- `path/to/file.ts` — `Symbol`: <what's relevant>

## Acceptance criteria

- [ ] <observable outcome, checkable without reading the diff>

## Blocked by

- #<n> — needs <named type, file or function> from it
```

Omit "Blocked by" unless a finding needs a nameable artifact from another. Being in the same
area, or the order the user listed them, is not a dependency.

**Labels.** Every ticket gets `feedback:<area>` plus its route label. Create missing labels first:

```bash
gh label create "feedback:<area>" -c "#5319e7" -d "Feedback session on <area>"
gh issue create --title "<imperative outcome>" --body-file <tmpfile> \
  --label "feedback:<area>" --label "status:ready"
```

File blockers before the tickets they block, so the `#<n>` exists. A blocked ticket gets
`status:backlog`, not `status:ready`.

## Report

End with the ticket URLs, one per line, each with its route; the dropped or accepted points with
their reason; and the one command that starts the work:

- one ready ticket → `/implement <n>`
- several → `/fleet dispatch feedback:<area>`
