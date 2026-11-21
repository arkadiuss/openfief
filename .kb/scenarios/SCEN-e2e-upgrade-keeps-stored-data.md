---
title: An in-place upgrade from the published 2026.11.0-alpha.1 keeps what the owner stored, runs the grant cutover once, creates SpiceDB's credentials and leaves agents off
status: active
origin:
  kind: agent
  ref: the design session's exit criterion (a) of 2026.11.0-alpha.2, relayed by the coordinator, 2026-11-21
  recorded_at: '2026-11-21T12:00:00.000Z'
id: SCEN-e2e-upgrade-keeps-stored-data
type: scenario
---
Given an install of the published release 2026.11.0-alpha.1 on a local k3d or kind cluster, fetched and verified with cosign as the install guide says and applied with its commands, seeded with the synthetic demo dataset by that release's own demo-seed image, with the built-in UI installed by its CLI, one script token kept and one revoked, and its attention items raised, when the install is upgraded in place to the release candidate built from the checkout with the guide's Upgrade section, then every row the published release stored, accounts, transactions, lots, valuations (but for what the new valuation rules derive again) and attention items, is still there and unchanged, the household-grant cutover has run once and left exactly the household grant to the kept token and the installed view plugin, each with its audit entry, a restart of the core changes no grant, SpiceDB's credentials Job has made its two Secrets, the agentic switch is off and the MCP endpoint answers 404, the core is ready, the old token still lists accounts and the revoked one is refused, and the built-in UI, upgraded with the CLI, serves its login page.

Assumptions: the published release's demo-seed image and the install files are the ones its release workflow signed, which the fetch script verifies before anything reaches the cluster; the published release has no audit log, so the audit entries checked are those the upgrade itself makes; and the valuation rules' version changes on the first start of a core with newer rules, which revalues every account, so the columns those rules derive are not compared.
