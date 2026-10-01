# Evidence, insight gate and critic

## Labels

- **Codebase basis** (ideas): `direct:` a `file:line` you read, `external:` a URL from the matrix,
  `reasoned:` a written argument. Strength order: direct > external > reasoned. An idea whose only
  basis is `reasoned:` can be at most **Speculative**.
- **Market claims**: Fact / Attributed / Inference / Assumption, as the market scan labelled them.
  Never upgrade a label.
- **Independent** means a different publisher with different underlying data. Two blogs repeating
  one vendor statistic count as one source.
- **Confidence** is a category, not a percentage: high (verified, fresh, independent), medium
  (single credible source, or an absence claim), low (stale, disputed or inferred).

## Insight gate

An insight is a claim that the market has something wrong. It needs the most evidence of any
claim, not the least. Write each candidate in this form:

> The market does **X**, assuming **A**. External evidence that A is false: **[cited]**. Our
> capability: **[file:line]**. Implication: **Y**.

The "external evidence" slot may contain only external sources. A code fact belongs in "our
capability" and never counts as evidence that A is false. The tier depends on what fills that slot:

| Tier | Requires |
|---|---|
| **Insight** | 2 or more independent external sources, at least one of them user voice (reviews, forum threads, issues), with the passages read, not snippets. The critic's disconfirmation pass ran and the claim survived. |
| **Hypothesis** | Plausible, with some support but below the Insight bar. Give the cheapest build-and-measure check that would settle it, ideally a tracer-bullet slice. |
| **Watch** | A single signal, or a vendor's own claim. Note it and revisit next run. |

**Absence claims** ("no competitor does X") are written "not found in N sources (listed)". Their
confidence is capped at medium. Before keeping one, consider the rival readings: it exists and the
search missed it; it was tried and abandoned; nobody wants it (a dead zone); a legal, terms-of-
service or cost barrier blocks it.

If nothing clears the bar, write: "None this run; conventional wisdom looks sound because …".
That is a normal result. Solid conventional wisdom raises the bar for anything that contradicts it.

## Critic brief

Spawn a fresh `general-purpose` agent. Give it only the shortlist (full idea write-ups), the
insight candidates, the inventory table, and the matrix with its URLs. Do not give it the raw
candidates, the cut table or your reasoning. It has a budget of roughly 10 web calls plus any
code reads it needs.

```text
You are a skeptic reviewing a product-ideation shortlist for this codebase. Assume each idea is
wrong and try to prove it, using the code and sources. Do not propose new ideas.

For each idea:
- Check its cited file:lines. Does the code say what the idea claims? Does the thing already
  exist (search for it)? Does an ADR or doc rule it out?
- Check its external basis. Does the URL actually support the claim?
- Verdict: holds / weakened / refuted, with evidence (file:line or URL). The same source rules
  apply to you: a roundup or review-aggregator blog is not evidence. Cite a primary page or user
  voice, or say "unverified".
- Pre-mortem, one line: "Six months on, nobody uses this, because …"

For each insight claim:
- Search for disconfirming evidence: someone already doing it, a dead zone, a barrier, rival
  explanations such as selection bias in the complaints.
- Verdict: survives (state the strongest counter-argument) / downgrade to Hypothesis or Watch /
  drop. Report what you searched for even if you found nothing.

Also list anything that looks well-reasoned, so doubt isn't manufactured.
```

Apply the verdicts as SKILL.md Phase 7 describes. Refuted ideas join the cut table with the
critic's evidence as the reason. The pre-mortem lines go into each idea's write-up.
