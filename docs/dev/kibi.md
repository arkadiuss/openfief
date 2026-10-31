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

## Evidence, as kibi counts it

kibi counts a requirement as *tested* only when an end-to-end test of one of its scenarios has a fresh, passing receipt from `kibi prove`, bound to the code it ran against and at most seven days old. That is the bar, and no shortcut lowers it:

- Every test entity declares its `verification_scope` (`unit`, `integration` or `end_to_end`), matching what the test really runs. A unit or integration test is never relabelled to count as end to end.
- Unit and integration tests prove the pieces and stay linked to the symbols they cover (`covered_by`); their receipts are recorded but don't make a requirement tested.
- End-to-end tests in `test/e2e` carry a `kibi.proof-contract.v1` for the end-to-end integration, which `kibi proof inspect` proposes and `.kb/proof/integrations.json` holds; nobody hand-writes either. CI's end-to-end job runs `kibi prove`, and the requirements report the site publishes from `main` shows that run's receipts.
- A behavior the owner relies on gets an end-to-end scenario. Each milestone's acceptance criteria (P-10 for M0) become scenarios verified end to end, so *tested* grows where it matters, not by count.

## Setup notes

- kibi and its Go symbol plugin are pinned npm dev dependencies, installed by `task setup`.
- `.mcp.json` registers kibi's MCP server for Claude Code; other agents start the same server.
- kibi's own git hooks stay uninstalled: lefthook runs the same checks through task targets.
- kibi's sync sometimes publishes a broken compiled store, about one run in ten under load: a relationship missing or an entity twice, although the same files compile correctly next time. Until kibi fixes it upstream, the repository carries a local patch that lets the compiling process finish before the store is published, `task kibi:check` recompiles once on exactly that defect and treats the second check as final, and the bug is reported to kibi's maintainers, with the author reviewing the report before it is filed. Any other failure is final at once.
