---
title: The owner uploads a statement on an account's page and the UI streams it to the core
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T05:39:09.366Z'
id: SCEN-ui-statement-upload
type: scenario
---
Given an owner-scoped session on an account's page, when the owner posts the Upload statement form with a mode and a file, then the UI streams the file in chunks to the core's StatementService with the plugin's token, never storing it, and shows the import's counts and ingest problems, or the core's reason; a read session never sees the form and is refused, and a malformed form or a body over 64 MiB never imports.