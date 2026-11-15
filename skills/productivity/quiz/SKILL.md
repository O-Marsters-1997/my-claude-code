---
name: quiz
description: >-
  Generate quiz questions for the user's Obsidian vault on demand by running the vault-questions
  CLI. Use when the user types /quiz, or says "generate vault questions", "quiz me on my notes",
  "make questions for my vault", or "what questions are pending". Not for answering questions from
  the vault (vault-recall) or filing notes (to-vault).
---

Run the `vault-questions` CLI to generate vault questions, then report what it printed. This skill
never writes questions itself: the CLI makes each model call outside this conversation.

## Inputs

- **Vault**: `~/Documents/personal/obsidian/my-obsidian-notes`. If it is missing, stop and say so.
- **CLI**: `vault-questions`. Check `command -v vault-questions` first. If it is not on `PATH`, stop
  and reply with the install command, `cd ~/Documents/personal/obsidian/olly-obsidian-plugin-utils && bun link`,
  and nothing else. Never write questions yourself as a fallback.

## Steps

1. Run `vault-questions status --vault <vault>` and show the pending pages and count. If the count
   is 0, say so and stop.
2. Ask what to generate: a page limit, a topic or a single page. Run `vault-questions generate --help` to find the flags for a
   topic or page; use `--max-pages N` for the limit.
3. Run `vault-questions generate --vault <vault> --pending` plus the chosen flags.
4. Report the CLI's output as printed, including any failed pages.

## Never

- Write or edit question files, or anything else in the vault.
- Generate before showing the pending count.
