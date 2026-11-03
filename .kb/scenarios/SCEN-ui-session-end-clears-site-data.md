---
title: Sign-out and a rejected session clear the site's storage
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T16:30:41.125Z'
id: SCEN-ui-session-end-clears-site-data
type: scenario
---
Given a browser with the UI's session cookie, when the owner signs out, or the core rejects the session on a page or an htmx request, then the response drops the cookie and carries Clear-Site-Data "storage", so the browser deletes what the origin keeps; a page with a valid session carries none (D-151).