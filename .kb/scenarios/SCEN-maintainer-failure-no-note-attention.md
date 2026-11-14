---
title: A refused or failed run leaves nothing but its audit entry, and a scheduled week whose every attempt failed raises one item
status: active
origin:
  kind: agent
  ref: 'author: D-198 pattern, 2026-11-13 (the M3a brief)'
  recorded_at: '2026-11-13T23:48:56.175Z'
id: SCEN-maintainer-failure-no-note-attention
type: scenario
---
Given a router that refuses, every model down, an answer that is no note or no figures to send, then the run ends refused or failed with its class and leaves no note or proposal; the scheduled week's run is tried three times, a quarter of an hour and then an hour apart, each attempt kept with its class; only when all three ended so is one Inbox item raised, an owner's failed run does not count, and a run with a note clears it.

## Context

A failed first attempt is a bad hour of a provider; a week whose every attempt failed is the owner's to look at. The author answered on 2026-11-13 that the weekly run retries like D-198's connector syncs.

## Source

> Weekly run retries within its window like D-198: River MaxAttempts 3 with backoff about 15 min then 1 h. Raise maintainer-failing when a scheduled week's run fails ALL its attempts (not after N weeks); clear it on the next success.

Source: the author's answers to the M3 follow-up (M3a), 2026-11-13
