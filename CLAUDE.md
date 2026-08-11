# homelab — agent instructions

## Commenting rules

Write comments for the code's enduring intent, contract, invariants, and
non-obvious operational constraints.

Do not write comments that narrate:

- Previous failed attempts or tool runs
- The sequence of changes made in this session
- "I changed this because…"
- Speculative explanations or temporary debugging context
- Information obvious from the code itself

Prefer a short comment explaining *why* a constraint exists over a long
procedural history.

If historical context is genuinely needed, put it in a git commit message, an
issue/ADR, or a concise reference such as `See ADR-012`.

Keep comments compact: usually 1–3 lines. For operational procedures, write a
short invariant near the code and put full runbooks in a task/script README or
`justfile` help text.

Before finishing, review every new or modified comment:

1. Does it explain a stable invariant or non-obvious reason?
2. Would it make sense without this conversation or the current diff?
3. Can it be shortened?
4. Is historical/debugging context better suited to the commit message or an ADR?

Delete comments that fail these checks.
