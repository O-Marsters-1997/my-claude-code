# Idea Template

Use this structure for each idea in the report. Omit fields that genuinely don't apply — this is
a guide, not a form to fill mechanically.

---

## [Idea Title] · *[Tier: adjacent / expansion / moonshot]*

**Opportunity addressed**
Which opportunity from the report this idea maps to. One idea can only map to one primary
opportunity — if it seems to address several, the opportunities probably need sharpening.

**Problem**
Who feels this pain, and what specifically makes it a problem worth solving. Be concrete about the
situation — "developers who need to X find that Y" not "users want better X".

**Why now**
What's changed (or is changing) that makes this a good time to solve this problem. If nothing has
changed, that's a signal this isn't the right moment.

**Existing solutions and their gaps**
What does the market already offer? Where do existing tools fall short for the target user? This
is where the market scan pays off — one specific gap per tool, not a feature matrix.

**Why this codebase is uniquely positioned** *(mandatory — if weak, drop the idea)*
Point to specific observed architectural properties, modules, abstractions, data, or user trust
that give this product an edge other teams couldn't easily replicate. Avoid aspirational claims.
"We already have X which means Y" — concrete and verifiable.

**Architectural leverage**
What existing parts of the codebase this idea builds on. What new pieces it would require. Be
honest about the new-pieces side — acknowledging real costs increases the idea's credibility.

**Risks and unknowns**
The 1–2 things most likely to make this fail or take much longer than expected.

**Tracer-bullet first slice**
The smallest end-to-end slice that would prove the idea works and deliver real value. Phrased
so it could be dropped directly into `write-a-prd` as a starting brief.
