---
title: An account's frame and tabs, the sync pill and the holdings behind an allocation bar
status: active
sourceFile: internal/views/ui/accounttabs_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T10:42:52.173Z'
id: TEST-ui-account-tabs
type: test
---
Serves the UI against in-memory Connect fakes: an account's breadcrumb, pills, four figure cards and six tabs, each a link with tab= and an htmx fragment, an unknown tab falling back to holdings; the settings tab with its connector instances; a table test of the sync pill; the holdings behind a currency and an asset class bar with links to instruments and the group's cash, and a refusal of other dimensions.