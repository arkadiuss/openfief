---
title: The account page lists cash checkpoints and takes the owner's checkpoint and booking
status: active
sourceFile: internal/views/ui/cashcheckpoints_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T03:49:57.454Z'
id: TEST-ui-cash-checkpoints
type: test
---
Over the UI's fakes: the owner's account page lists the checkpoints the fake read API marks entered, with the shown, derived and difference amounts in the number format, a Book as a deposit button on the one that differs and nothing to book on the matching one, and the Enter cash checkpoint form dated today; a read session sees the list without either. Posting a checkpoint or a booking reaches the fake core and redirects to the checkpoints; a refusal shows the core's reason with the form as posted; a read session and malformed forms never reach the core (D-158).