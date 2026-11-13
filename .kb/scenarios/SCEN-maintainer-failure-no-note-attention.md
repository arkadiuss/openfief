---
title: A refused or failed run leaves nothing but its audit entry, and two weekly failures raise one item
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:31.875Z'
id: SCEN-maintainer-failure-no-note-attention
type: scenario
---
Given a router that refuses, every model down, an answer that is no note or no figures to send, then the run ends refused or failed with its class and leaves no note or proposal; two weekly runs in a row ending so raise one Inbox item, an owner's failed run does not count, and a run with a note clears it.