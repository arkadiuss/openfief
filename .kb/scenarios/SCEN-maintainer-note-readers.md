---
title: The owner and a session that may view everything read the note
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:33.583Z'
id: SCEN-maintainer-note-readers
type: scenario
---
Given a stored note, when the owner or a session the authorizer lets view every account asks, then it reads the note, and only the owner also sees the last failure and the waiting proposals; a session that may view fewer accounts, or none, is refused; an authorizer that can't tell is a refusal.