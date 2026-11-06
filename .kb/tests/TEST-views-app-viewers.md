---
title: Syncs, attempts and instance-only records follow the accounts a caller may view; a suggested transfer's item is tied to both accounts
status: active
sourceFile: internal/views/app/viewers_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T06:14:52.552Z'
id: TEST-views-app-viewers
type: test
---
Runs app.Views over in-memory fakes: connector bindings tie xtb-main to acc-1 and files to acc-2, nbp-fx to none. A script token sees every instance; a caller denied acc-1 loses xtb-main's status, attempts, failed attempts and instance-only batch; a viewer (a holder the core doesn't know yet) granted acc-1 sees xtb-main and never nbp-fx, and never an attention item tied to no account. A transfer-suggested item about a withdrawal in acc-1 and a deposit in acc-2 is tied to both, a decision, and hidden when either account is. A false pass would come from a fake authorizer that allowed everything; each case checks the hidden side too.