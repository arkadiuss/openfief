---
title: The owner grants, revokes and lists grants, each change with its audit entry in one transaction, and the core cuts over once and sweeps what outlived its record
status: active
sourceFile: internal/identity/app/grants_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T10:00:08.661Z'
id: TEST-identity-grants-app
type: test
---
Runs the grant use cases over in-memory fakes of the store (whose Atomically restores every table when its function fails, like a rolled back transaction), the directory of accounts, sleeves and plugins, and the authorization model. Create makes a grant of each kind of target for a script token and a view plugin, with an entry by the owner naming holder and target by ID, passes it to the model and makes it once however often it is asked; an unknown holder or target is not found, a malformed one or a person invalid. With the audit log down no grant or revocation is made. A revocation removes the relationship from the model first and changes nothing when the model fails; a revoked grant stays as history. A demo session, no session and a script token's secret are refused every call. Revoking a script token ends its grants in its own transaction; the core ends a plugin's and a sleeve's. The cutover gives the household to each token that may be used and each plugin with no grant, once, and waits when the plugins are unknown; the sweep ends grants of a revoked token, an uninstalled plugin and a deleted account or sleeve, and leaves what it can't tell about.