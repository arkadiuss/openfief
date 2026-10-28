---
title: Revoking a token stops it at once and keeps the first revocation
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:13.712Z'
id: SCEN-identity-revoke
type: scenario
---
Given an issued token, when the owner revokes it, then its secret is refused from then on; when it is revoked again later, then the first revocation time is kept; when an unknown or malformed ID is revoked, then Revoke fails; listing shows every token with its status.