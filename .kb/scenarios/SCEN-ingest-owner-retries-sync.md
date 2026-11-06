---
title: The owner retries a failing sync from the Inbox
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T04:22:56.112Z'
id: SCEN-ingest-owner-retries-sync
type: scenario
---
Given a connector instance whose sync keeps failing, when the owner asks to retry it, then one sync of it starts at once in the background and is recorded like a scheduled one; when a sync of it runs already, here or under another core's claim, the retry is refused saying so; a read session, an owner the authorizer doesn't let change every bound account, an unknown instance and a core without a running scheduler are refused too.