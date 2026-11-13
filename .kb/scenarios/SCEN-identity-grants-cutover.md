---
title: At the start in which the core first answers from SpiceDB, each script token and view plugin that has no grant gets the household grant, once, each with its own audit entry
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T10:00:20.004Z'
id: SCEN-identity-grants-cutover
type: scenario
---
Given script tokens and installed view plugins that held read access under the policy, when the core starts with --authz=spicedb for the first time, then each token that may still be used and each view plugin that holds or asks for the view capability, and that has no grant, gets exactly one household grant with its own audit entry by the core, written to SpiceDB, so none loses access; a revoked or expired token, and a holder with a grant, get none; a token made afterwards gets none, and a second start writes nothing; and when the installed plugins can't be read yet the cutover waits and is not recorded as done. Under the policy nobody is given a grant. Assumes D-209.