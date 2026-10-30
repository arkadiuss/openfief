---
title: The transfers page lists pairs and hands the owner's decisions to TransferService
status: active
sourceFile: internal/views/ui/transfers_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T06:32:27.923Z'
id: TEST-ui-transfers
type: test
---
Runs the UI against in-memory Connect fakes of the read API and TransferService: the page lists the synthetic suggestion with accounts, dates and amount in the number format and confirm and reject forms; the owner's confirmation and rejection reach the core with the session and the pair is then listed under that decision; a read session sees no buttons and is refused, the core's refusal is shown, a malformed form never reaches the core; the overview counts the suggestions and links the page.