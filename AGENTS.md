# OpenFief: agent guide

Self-hosted portfolio tracker and orchestrator for AI investment agents (docs/OVERVIEW.md). Status: design phase. M0 (tracker plus Polish imports) is next; there's no product code yet.

## Authority, highest first

1. Ground rules `[GR-n]` in docs/DESIGN.md. Never edit them.
2. Author decisions `[D-n]` in docs/DESIGN.md.
3. Requirements in kibi (`.kb/`). Change them only through kibi; `firm` ones need the author.

Proposals (P), research (R), tensions (T) and open questions (OQ) are undecided: don't treat them as rules.

## Read only when needed

| When you | Read first |
|---|---|
| Edit DESIGN.md, a proposal, research or the glossary | `.claude/skills/design-docs/SKILL.md` |
| Change behavior, or touch requirements and traceability | docs/dev/kibi.md |
| Write or move Go code | docs/dev/code-organization.md, docs/glossary.md |
| Write tests | docs/dev/testing.md |
| Write a docs page, or change code a page covers | docs/dev/docs.md |
| Record HTTP fixtures | `.claude/skills/hypert-recording/SKILL.md` |
| Hit a lint failure | `.golangci.yml`: its comments give the reason for each choice |
| Commit, open a PR, or work on CI, release or deploy | docs/dev/sdlc.md |
| Need product context | docs/DESIGN.md (decided), docs/proposals/<area>.md (not decided); read sections by ID |

## Hard rules

- No real financial data, account numbers or credentials anywhere: code, tests, recordings, logs, commits.
- Domain logic is pure functions with table tests. No `float64` for money.
- Use the glossary's domain terms exactly.
- Never weaken lint, tests or hooks to get green, and never use `--no-verify`. Fix the cause or ask.
- Never commit to `main`.
- Use `task` for everything: `task check` before handing work back (the pre-commit hook runs `task check:commit`), `task --list` for all targets, and `task setup` once per clone. Don't call tools directly when a target exists.
