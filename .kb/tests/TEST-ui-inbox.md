---
title: The Inbox's failing sync, redirects, navigation, account banner, in-place mapping and empty state
status: active
sourceFile: internal/views/ui/inbox_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T04:31:59.234Z'
id: TEST-ui-inbox
type: test
---
Runs the UI against in-memory Connect fakes of the core, with a fake SyncService and the read API's sync history: a failing sync's item shows its attempts, error classes and next try, the owner's Retry now reaches the core and comes back saying it started, a refused retry shows the core's reason and a read session gets no button; the five removed check pages answer 301 to the Inbox; the navigation has Overview, Accounts, Inbox with its open count and Settings, with sleeves and plugins under it; an account's page shows its own open items, a broken transfer's on its second account too; an unmapped broker account is mapped in place and back to the Inbox; nothing open shows Kasztan asleep.