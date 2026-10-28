---
title: The identity schema holds hashes and refuses malformed rows
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:15.601Z'
id: SCEN-identity-store-hash-only
type: scenario
---
Given a migrated identity schema, when tokens are saved, then each row holds the ID, name, scope, times and a 32-byte SHA-256 hash and no column holds the key; when a row has a short hash, an expiry not after its creation, an upper-case ID or a name with a space, then PostgreSQL refuses it.