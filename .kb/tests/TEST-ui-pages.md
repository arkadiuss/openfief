---
title: The UI's pages show the record from the read API with the session token
status: active
sourceFile: internal/views/ui/pages_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T19:48:30.057Z'
id: TEST-ui-pages
type: test
verification_scope: integration
---
Runs the UI against fakes of both core services over a small synthetic record. Asserts pages need a session (redirect, or HX-Redirect for htmx), an expired session drops the cookie, the overview shows the total, each account's valuation status, the allocation chart and table with its dimension switch and the value-over-time chart from go-live with 'TWR: not built', the account page shows holdings with values, lots, cash, checks and paged transactions, the flagged assertions page lists flagged checks, a read session sees no owner actions, and every read API call carried the session token as its bearer.