# Friction rubric — Nielsen heuristics as timed review questions

Use this to convert a raw step log into named findings. Each question is answerable from the `element → expected → actual` triples and the wait notes you recorded while driving. A "no" is a candidate finding; attach the triple that exposed it as evidence.

The point of phrasing these as *timed, concrete* questions is that "is this usable?" is unanswerable but "did the result of this action appear within ~1s?" is checkable from what you observed. Don't grade on taste — grade on whether the observed behaviour answers each question.

## The ten, as review questions

1. **Visibility of system status** — After each action, did the UI show what happened within ~1s (spinner, toast, state change, navigation)? A step you logged as *had-to-wait* or *stalled* with no feedback fails here.

2. **Match with the real world** — Did labels, icons, and ordering match what a first-time user would expect, or did the actual outcome contradict what the control's label promised? Every *mismatch* triple is a candidate here.

3. **User control & freedom** *(weighted — watch closely)* — At every state you landed in, was there a visible way out: undo, cancel, back, close, escape? Flag every point where you felt trapped, where "back" lost work, or where there was no way to correct a mistake without starting over.

4. **Consistency & standards** — Did the same action use the same word/control across steps and flows? Did the app follow platform conventions (a link looks like a link, primary action is where users expect)? Inconsistencies you had to re-learn mid-walk fail here.

5. **Error prevention** — Did the flow let you enter a bad state it could have prevented (submit an invalid email, double-submit a form, proceed with a required field empty)? A *retry* caused by an error the UI could have blocked up front is a finding.

6. **Recognition rather than recall** *(weighted — watch closely)* — At each decision point, was the information needed visible right there, or did you have to remember it from an earlier screen (an order number, a code, which option you picked two steps ago)? Every point where you had to hold something in your head is a candidate.

7. **Flexibility & efficiency** — Were there accelerators for a returning user (saved details, sensible defaults, skip-ahead), or did every run force the full slow path? Absence only matters if the run goal is conversion or repeat-use.

8. **Aesthetic & minimalist design** — Did any step bury the primary action under competing elements, so you had to hunt for it? A *retry* caused by "couldn't find where to click" fails here.

9. **Help users recognise, diagnose, recover from errors** — When an error did appear, was it in plain language, did it say what went wrong, and did it point to the fix? An error you logged as cryptic or dead-ended fails here.

10. **Help & documentation** — Where the flow genuinely needed guidance (a non-obvious step), was help available in context without leaving the flow? Only flag when a step actually stalled for lack of it — don't manufacture a need.

## Severity

Rate each finding so `impeccable` and the user can triage:

- **blocker** — the flow cannot be completed, or completion requires knowledge the UI never provides.
- **major** — the flow completes but with real confusion, lost work, or a retry most users would hit.
- **minor** — noticeable friction that most users push through; polish, not rescue.

Severity is about impact on completing the run's goal, not how ugly the step is — aesthetics are `impeccable`'s call, not the rubric's.
