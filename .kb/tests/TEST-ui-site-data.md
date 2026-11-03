---
title: A session's end clears the site's data, and htmx keeps no history
status: active
verification_scope: unit
sourceFile: internal/views/ui/sitedata_test.go
origin:
  kind: agent
  recorded_at: '2026-11-03T16:30:38.205Z'
id: TEST-ui-site-data
type: test
---
Serves sign-out, a page and an htmx fragment with a session the fake core rejects, and the overview with a valid session, through httptest recorders, and checks Clear-Site-Data is "storage" exactly on the first three; and parses every page's layout and checks its htmx config turns the history cache off with a refresh on a history miss. A page keeping the cache, or a valid session clearing data, fails it.