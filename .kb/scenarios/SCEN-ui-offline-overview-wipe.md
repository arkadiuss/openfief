---
title: Sign-out, a rejected session and a passed session end wipe the kept overview
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T10:11:01.666Z'
id: SCEN-ui-offline-overview-wipe
type: scenario
---
Given a kept overview, when the owner signs out, then the page deletes it before the form goes, the worker deletes it and the response carries Clear-Site-Data "storage"; when the core rejects the session, the response sending the browser to the login page carries the same header and the login page deletes it; when the session's end has passed and the device is offline, the offline page deletes it unread and shows nothing of it.