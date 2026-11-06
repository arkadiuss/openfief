---
title: The overview's places, the Ask box's answers and the Inbox's Ask column over the fakes
status: active
sourceFile: internal/views/ui/home_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T08:17:40.170Z'
id: TEST-ui-home
type: test
---
Serves the UI against in-memory Connect fakes of the read API: the overview shows the status sentence with the open items and the failing sync, the Ask box with its chips, the value card, the chart card's tabs as a page and fragment, Needs you, accounts by sleeve, allocation bars and recent activity; /ask answers each command with a card and an Open link and an unknown question with the commands; a failing core logs no question; the Inbox shows its Ask column in the open view only.