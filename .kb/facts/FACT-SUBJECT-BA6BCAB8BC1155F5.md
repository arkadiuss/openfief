---
title: Models Chat Client 429 Retry
status: active
fact_kind: subject
subject_key: models.chat_client_429_retry
origin:
  kind: agent
  ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
  recorded_at: '2026-11-20T12:51:08.000Z'
id: FACT-SUBJECT-BA6BCAB8BC1155F5
type: fact
---
The chat client's single retry of a 429 after the Retry-After wait capped at ten seconds, or one second when absent, and only when the wait ends before the call's deadline.