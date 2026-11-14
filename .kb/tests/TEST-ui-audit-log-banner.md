---
title: The Audit log page shows the audit-log-behind item's words in a banner while it is open
status: active
sourceFile: internal/views/ui/audit_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:00.020Z'
id: TEST-ui-audit-log-banner
type: test
---
TestAuditLogBannerFollowsTheBehindItem checks no banner without the item, the banner with its count while it is open, and the page without a banner when the core can't list items; TestAuditLogBehindItemIsInTheInbox checks the Inbox's label and link; task ui:browser, over the UI's fakes, checks the banner at phone and desktop width without sideways scroll.