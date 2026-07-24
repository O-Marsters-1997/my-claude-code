# Flow friction audit — {app name / surface}

**Run goal:** {friction discovery | conversion | concept validation}
**Audited:** {url(s)} · {date}
**Flows walked:** {n} — {flow names}

---

## Findings

Ranked most severe first. Each ID is stable across re-runs — it describes the friction, not its position — so a later run can diff by ID to show what closed.

### {flow}-{what}  ·  {blocker | major | minor}

- **Flow:** {which flow, which step}
- **What happened:** {the element → expected → actual triple that exposed it}
- **Fails:** {rubric heuristic, e.g. "6 · Recognition rather than recall"}
- **Why it's friction:** {one line — the user's felt experience}

<!-- repeat per finding -->

---

## Evidence — step logs

The raw walk each finding is drawn from. Kept so the critique and any re-run can trace a finding back to what actually happened.

### Flow: {name}

| # | Element (ref · label) | Expected | Actual | Wait |
|---|-----------------------|----------|--------|------|
| 1 | e15 · "Continue" | advance to payment | advanced, no confirmation shown | a beat |
| 2 | … | … | … | … |

<!-- repeat per flow -->

---

## Handoff

- **Critique:** {what impeccable's critique mode returned against these findings}
- **Fix directions:** {the divergent directions impeccable's shape/redesign mode proposed, each with its low-fi wireframe — `![id](id.png)` when rendered via Open Design, ASCII block when it wasn't available}
