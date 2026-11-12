---
title: No entry can hold a token secret, an amount or a question's text
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T19:17:43.134Z'
id: SCEN-identity-audit-no-amount-secret-text
type: scenario
---
Given a script token created in an owner session, when the entry is written it names the token by its ID and the secret appears nowhere in the log; an attempt to put a secret, an amount or a sentence into any reference or the actor is refused by the domain, and an insert around it is refused by the table's checks. The shape of a passkey reference cannot tell a random key from a credential ID, so callers pass only the typed ID.