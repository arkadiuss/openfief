---
title: Token commands print a secret once and never list one
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T14:32:56.546Z'
id: SCEN-core-token-commands
type: scenario
---
Given a migrated database, when the owner runs core token issue with a name and lifetime, then the secret alone is the first line on stdout with the ID and expiry after it, and nothing is logged; when the owner runs core token list, then every token shows with its status and no secret; when core token revoke ID runs, then that token is refused from then on.