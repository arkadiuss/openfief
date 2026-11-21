---
title: Core Database Session Timeouts
status: active
fact_kind: subject
subject_key: core.database_session_timeouts
origin:
  kind: agent
  ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
  recorded_at: '2026-11-20T12:51:08.000Z'
id: FACT-SUBJECT-C5C78A6962242D2B
type: fact
---
The statement timeout and idle-in-transaction timeout every connection of the core's database pool starts with, two minutes each by default and configurable through the environment.