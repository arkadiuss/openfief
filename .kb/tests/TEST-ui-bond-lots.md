---
title: The owner adds a bond lot on a bond account's page, and the UI hands it to the core as typed
status: active
sourceFile: internal/views/ui/bondlots_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T02:14:27.460Z'
id: TEST-ui-bond-lots
type: test
---
Handler tests of the UI against in-memory Connect fakes of the core: the owner sees the Add bond lot form on an account at the institution treasury-bonds, a read session and another institution's account don't; a posted lot reaches the fake BondLotService with the plugin's token, the owner's session and origin, the series trimmed and upper-cased, and redirects to the open lots; the core's refusal shows its reason with the lot as posted (400), an unreachable connector 503; a read session gets 403 and a malformed form 400, neither reaching the core. The page passes the existing no-inline-code template check.