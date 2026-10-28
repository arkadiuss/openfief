---
title: Expired, revoked, unknown, mismatched and malformed tokens are refused
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:11.866Z'
id: SCEN-identity-verify-refused
type: scenario
---
Given an issued read token, when the secret is presented at or after the expiry, after revocation, for a scope it lacks, with another key, with an ID that is not stored, or malformed, then Verify fails with ErrInvalidToken and the reason, never including the secret; when the store fails, then Verify fails with another error so the API answers 500, not 401.