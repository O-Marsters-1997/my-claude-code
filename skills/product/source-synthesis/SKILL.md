---
name: source-synthesis
description: >
  Deep synthesis of a curated reading list — blog posts, tutorials, docs, articles — tailored to a specific use case or build goal. Use whenever the user provides URLs (public links to educational content) and wants a thorough, opinionated synthesis rather than a raw summary. Trigger on phrases like "synthesise these sources", "give me a lit review of", "I want to read up on X, here are some links", "help me digest these articles", "summarise these tutorials for my use case", "I'm about to build X, here's my reading list", or whenever the user pastes a list of URLs alongside a goal or project description. Also trigger proactively when the user provides multiple URLs and a description of something they're trying to learn or build — even if they don't say "synthesise".
compatibility:
  tools:
    - WebFetch
---

# Source Synthesis

You've been given a list of URLs and a use-case prompt. Your job is to read all of them, extract the most valuable insights, and weave them into a single synthesis — one coherent document that saves the user from reading everything themselves while giving them everything they need to hit the ground running.

This is not summarise-each-source. It is cross-source synthesis: find the patterns, tensions, and actionable takeaways that only emerge when you hold all the sources together.

## Your inputs

The user will give you:

1. **A list of URLs** — public links to tutorials, blog posts, docs pages, or articles. Treat them all as reachable; fetch them all.
2. **A use-case prompt** — what they're trying to build, learn, or decide. This is your north star: everything in the synthesis should serve this goal.

If the user hasn't provided a use-case prompt, ask for one before fetching. A synthesis without a goal is just a summary.

## Reading phase

Fetch every URL. For each source, note:

- What type of source it is (opinionated tutorial, official docs, war story, introductory overview, deep dive, etc.)
- Its level of authority and likely recency
- The 2–5 most important ideas it contributes
- Anything that contradicts or complicates what other sources say

Fetch them in parallel where possible. If a URL fails or returns no useful content, note it and continue — don't stop.

## Synthesis structure

Write the synthesis as a single flowing document with these sections. Adapt section depth to how much material you have — don't pad, don't truncate.

### 1. Use-case frame (2–3 sentences)

Restate the user's goal in your own words. This anchors the whole synthesis and signals to the user that you've understood the real question, not just the surface request.

### 2. The mental model (key concepts)

The 4–8 concepts a newcomer must internalise before anything else makes sense. Define each one briefly but precisely. Prioritise concepts that multiple sources assume without explaining, and concepts that are commonly confused or misnamed.

Write these as a numbered list with a short definition and a one-sentence note on why it matters for *this* use case.

### 3. Consensus view

What do the sources broadly agree on? What is the settled, established wisdom in this space? This gives the user a stable foundation to reason from.

Focus on substance, not source attribution. The user doesn't need "Source 3 says X and Source 5 also says X" — they need to know X is well-established.

### 4. Tensions and trade-offs

Where do the sources disagree, hedge, or come from different schools of thought? What are the decisions the user will genuinely have to make, with no universally right answer? What do experienced practitioners argue about?

Be honest about unresolved questions. A synthesis that papers over real tensions is worse than useless — it creates false confidence.

### 5. Practical patterns for your use case

The concrete techniques, patterns, or approaches that are most directly applicable to what the user is trying to build. Ordered by relevance to their specific goal, not by frequency of mention across sources.

Include code snippets, commands, or configuration examples where sources provide them and they're relevant. Paraphrase and synthesise rather than quoting at length — the user wants a usable distillation, not a paste job.

### 6. What to watch out for

Footguns, common mistakes, things that look obvious but aren't, and gotchas that the sources mention as things beginners consistently get wrong. Keep this tight — only include things that are genuinely non-obvious or that multiple sources flag.

### 7. Gaps in the reading list

What important questions does this set of sources *not* answer? Where would the user need to go deeper? If a source was clearly out of date on something important, say so here.

Be specific: "None of these sources cover X, which matters for your use case because…" is useful. "You may want to read more" is not.

### 8. Recommended next steps (3–5 bullets)

Concrete actions the user can take right now. Not "learn more about X" — specific: what to try first, what to set up, what decision to make before writing code, what to validate early.

## Tone and style

- Write for someone who is intelligent but genuinely new to this area. No condescension, but no skipping steps either.
- Be direct and opinionated where the evidence supports it. "The consensus is clear: do X" is more useful than "some sources suggest X might be worth considering."
- Cite sources sparingly and only when it adds value — "this pattern comes from the Rails doctrine" or "this warning appears in the official migration guide" is useful; footnote-style attribution for every sentence is not.
- Keep the document skimmable. Use section headers, bullet points for lists, and short paragraphs.
- Write in the user's language. If they said "I'm building a webhook handler", use that framing throughout, not "event-driven HTTP callback receiver".

## Length calibration

Match depth to the source material and the complexity of the use case:
- 3–4 short sources on a focused topic → 600–900 words
- 6–10 varied sources on a complex topic → 1000–1800 words
- Longer is not better. A tight 800-word synthesis beats a padded 2000-word one every time.

## Output destination

Write the completed synthesis to `./docs/approach.md` under a `## Background Research` section. This file is the canonical alignment document for the project — the synthesis lives as the first substantive section, before any alignment content (Problem, Goals, Features, etc.) that the `chat-to-approach` skill adds later.

**If `./docs/approach.md` does not exist:**
- Create `./docs/` if needed
- Write a minimal file: the `# Approach` header, the `_Last updated_` line, and the `## Background Research` section containing the synthesis
- Leave a placeholder comment after the section: `<!-- Alignment content (Problem, Goals, Features, etc.) to be added via chat-to-approach -->`

**If `./docs/approach.md` already exists:**
- Replace the content of the `## Background Research` section with the new synthesis
- If the section doesn't exist yet, insert it immediately after the `_Last updated_` line, before any other sections
- Do not touch any other sections

After writing, tell the user the file has been saved to `./docs/approach.md` and that they can run `chat-to-approach` to add or update the alignment content (features, goals, constraints, etc.) below the research section.

## What not to do

- Don't summarise each source in turn (source 1 says… source 2 says…). That's a list of summaries, not a synthesis.
- Don't hedge everything. Pick a view where the evidence supports one.
- Don't include concepts that aren't relevant to the user's use case just because they appeared in the sources.
- Don't fabricate content that wasn't in the sources. If something isn't covered, say so in the Gaps section.
