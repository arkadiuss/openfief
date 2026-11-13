---
title: The core ends the grants of an uninstalled plugin, a deleted sleeve or account and a revoked token, cuts over once at its start, and `core token` passes grants on
status: active
sourceFile: cmd/core/grants_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-13T10:00:10.743Z'
id: TEST-core-grants
type: test
---
Under task test:pg, over a fresh database with a recording SpiceDB. The directory answers whether an account, a sleeve and a view plugin exist, treats a plugin that asks for view as one that holds it, answers unknown before the plugin cache has synced, and knows no plugin outside a cluster. The start's cutover gives the household grant to every script token and view plugin that has no grant, once and each with the core's own entry, writes it to SpiceDB, gives nobody a grant under the policy, waits for the plugins and never marks itself done without them. The core ends the grants of an uninstalled plugin, a deleted sleeve and a deleted account with an entry each, and its sweep ends those of a plugin that went while it was down. core token issue --grant-household makes the household grant and core token revoke ends the token's grants, with their entries.