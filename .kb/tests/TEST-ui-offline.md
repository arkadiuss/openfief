---
title: The overview page hands the installed app only what D-151 lets it keep, and sign-out or a rejected session tells the browser to wipe it
status: active
sourceFile: internal/views/ui/offline_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-02T10:10:51.309Z'
id: TEST-ui-offline
type: test
---
Table tests of buildOfflineOverview check the kept overview holds the total's three values, each account's name, institution, sleeve and three values, the allocation's dimension, value and shares, the date and the session's end in UTC, written in the number format, with the allocation's total standing in for a missing total, and never the account's ID; a walk over its JSON allows no other member. Handler tests over the in-memory fakes check that only the overview page carries data-offline-overview, every page is no-store, sign-out and a session the core rejects (page and htmx) answer Clear-Site-Data "storage", the login page marks itself for wiping, sessionEnd takes the earlier of the idle and absolute expiry, the worker's CONFIG and offline.js name the same database, offline.js generates its AES-GCM key with extractable false and stores the key object, no script touches localStorage, sessionStorage or cookies, every page loads offline.js and htmx's history cache is off. A false pass would be a script that behaves differently in a browser; the browser checks (TEST-ui-browser-checks) drive it in Chromium.