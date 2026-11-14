---
title: Entries stuck in the outbox raise an item that follows their count and clears with them
status: active
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:05.427Z'
id: SCEN-attention-audit-log-behind
type: scenario
---
Given audit-entry jobs in the outbox, when a job has failed five attempts and still waits, then the core raises the item audit-log-behind in the group of data coming in with the count of such jobs in its detail, jobs that failed fewer raise none, the count follows the jobs still stuck, and the item clears when none is.

## Context

D-211's outbox never drops an entry: River keeps retrying a job without limit. The count is of the jobs that failed five attempts and still wait, not of the whole backlog, so a burst of entries that only just queued raises nothing; the item clears when no job is that stuck, which is when the log has caught up. The detail changes with the count, so an acknowledged item opens again when it changes.

## Source

> AUDIT-LOG-BEHIND ATTENTION ITEM: after an audit-entry job has failed 5 attempts (keep River retrying indefinitely; never drop), raise attention item kind audit-log-behind in the data coming in group next to failing syncs (D-196). Cleared when the backlog is empty. Same state as a banner on Settings, Audit log.

Source: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18