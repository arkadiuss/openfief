---
title: Each identity change commits with its audit entry, and fails when the entry can't be written
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T19:17:42.847Z'
id: SCEN-identity-audit-every-change
type: scenario
---
Given an owner session, when the owner signs in, signs out, adds or removes a passkey, issues an enrollment link or creates or revokes a script token, then the change and its entry commit in one transaction; when the entry cannot be written, the change fails and nothing is left. A refused sign-in is recorded on its own and stays refused. Only the identity events are built; grants, settings, entries, decisions, plugin installs and model calls come with their contexts.