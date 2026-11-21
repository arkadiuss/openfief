---
title: UI Client Events Shared Cap
status: active
fact_kind: subject
subject_key: ui.client_events_shared_cap
origin:
  kind: agent
  ref: Security review of the new surfaces before the alpha.2 cut, /client-events (the design session's exit criterion (c), 2026-11-21)
  recorded_at: '2026-11-21T18:15:34.157Z'
id: FACT-SUBJECT-UI-CLIENT-EVENTS-SHARED-CAP
type: fact
---
The most reports of client events the built-in UI takes in a minute from every address together, which bounds what many addresses can put into the log and the metrics, and which a report refused by its own address's limit does not count toward. It does not cover the limit of one address, what a key is, or the number of events in one report.