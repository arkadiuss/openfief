# Requirements with kibi

[kibi](https://github.com/Looted/kibi) (by pfranczyk) holds the requirements: requirements, scenarios, tests, facts and code links in `.kb/`, checked by Prolog and exposed to agents over MCP. DESIGN.md says why; kibi says what, and enforces it. Requirements cite DESIGN.md's GR and D IDs.

## Rules

- **Never edit `.kb/` by hand.** Write only through kibi's tools. kibi never commits; stage and commit its files with your change.
- **Before changing behavior,** search kibi for the requirements that govern it and respect their firmness:
  - `firm` means the author stated it: ask before changing it, and change it only through `supersedes` with a reason.
  - `default` and `exploratory` may be revisited on evidence.
- **New behavior** gets a modeled requirement with a scenario, before or with the code. Each behavioral function is owned by a kibi `symbol` that `implements` the requirement and is `covered_by` the test that verifies its scenario. `// implements` comments are not the record.
- **Decisions and open questions:** a requirement imported from DESIGN.md is `firm` and cites the item, such as `design-decisions:D-18#2`. Open questions are observation facts tagged `review:open-question`. Proposals stay in docs/proposals/ until the author decides them.
- **Blocked commit:** the pre-commit gate blocks unowned or uncovered behavioral symbols. Fix the cause with kibi's own `kibi-commit-review` skill; never bypass the hook.
- **Learn the workflow on demand** from kibi's bundled skills (`kibi-usage`, `kibi-commit-review`). Don't preload them.
- **Before handing work back,** run the full check, and show kibi's signoff summary when requirements changed.

## Setup notes

- kibi and its Go symbol plugin are pinned npm dev dependencies, installed by `task setup`.
- `.mcp.json` registers kibi's MCP server for Claude Code; other agents start the same server.
- kibi's own git hooks stay uninstalled: lefthook runs the same checks through task targets.
