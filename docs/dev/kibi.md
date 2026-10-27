# Requirements with kibi

[kibi](https://github.com/Looted/kibi) (by pfranczyk) is the requirements layer for agents: requirements, scenarios, tests, facts and code links in `.kb/`, checked by Prolog and exposed over MCP. Decided by the author: DESIGN.md keeps the *why* and the provenance; `.kb/` holds the *what*, enforced (D-33).

**State.** Bootstrapped 2026-10-27, from plan `f1bca4af…`, which the author approved:
- 67 `firm` M0 requirements, citing `design-ground-rules:GR-n#k` or `design-decisions:D-n#k`, each with the verbatim GR/D text as its excerpt.
- Review facts for OQ-21 to OQ-27 and the T-11 conflict.
- Scenarios, predicates and proof integration are authored per bounded context when M0 work on that context starts.
- GR-4 to GR-11 are authored at their milestones.
- Five `default` tooling requirements (docs check, vulnerability check, e2e guard), recorded by Claude Code under D-41, with scenarios and symbol links for the shell functions.

## Setup (once per clone)

- `task setup` installs the pinned kibi packages (`npm ci`, Node 22+), including `kibi-plugin-treesitter` for Go symbols, along with the other tools and the git hooks.
- `task kibi:doctor` checks the setup. kibi's own hooks are deliberately not installed; lefthook runs their commands through `task kibi:staged` and `task kibi:sync` ([sdlc.md](sdlc.md)).
- MCP: `.mcp.json` registers `kibi-mcp` for Claude Code. Other agents start the same server (`npx --no-install kibi-mcp`, stdio, repo root).
- The optional `kibi-claude` plugin is user-level; install it only if the owner asks.

## How to work

- **Learn the workflow on demand:** `npm exec -- kibi skills load kibi-usage --format markdown`. Don't preload it.
- **Before changing behavior:** `kb_search` (or `task kibi:search -- <topic>`) for the governing requirements. Respect their firmness:
  - `firm` means the author stated it: ask before changing it, and change it only through `supersedes` with a reason.
  - `default` and `exploratory` may be revisited on evidence.
- **New behavior:** model the requirement (`kb_model`, then `kb_upsert`) with a scenario, before or with the code. Own each behavioral function with a `symbol` entity that `implements` the requirement; kibi treats `// implements` comments as legacy, not as the record. Tests link the scenarios they verify.
- **Blocked commit:** if the pre-commit gate stops a commit, load `npm exec -- kibi skills load kibi-commit-review --format markdown` and fix the cause. Never bypass the hook.
- **Provenance mapping:**
  - GR and D items are `firm`, with origin `import` citing a stable reference such as `design-decisions:D-18#2` (the item, and which atomic part of it).
  - Open questions are observation facts tagged `review:open-question`.
  - Proposals (P, P-CC) stay in DESIGN.md until the author accepts them.
- **Never edit `.kb/` by hand.** Write through kibi tools only. kibi never commits; you stage and commit its files with your change.
- **Before handing work back:** `task check` (it includes `kibi check`), and `kibi signoff` when requirements changed.
- **SPEC.md §2 is frozen.** New or changed requirements go to kibi, not SPEC.md.
