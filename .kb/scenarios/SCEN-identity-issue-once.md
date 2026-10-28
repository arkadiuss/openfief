---
title: A new script token's secret is shown once and only its hash is stored
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:08.008Z'
id: SCEN-identity-issue-once
type: scenario
---
Given the owner issues a script token named demo for 168 hours, when Issue runs, then it returns a read-scoped token expiring 168 hours later and a secret ofs_<id>_<key>, built from 40 bytes of crypto/rand, and the store keeps the token with the SHA-256 hash of the key and never the key or the secret.