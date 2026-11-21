---
title: UI Client Events Limit Key
status: active
fact_kind: subject
subject_key: ui.client_events_limit_key
origin:
  kind: agent
  ref: Security review of the new surfaces before the alpha.2 cut, /client-events (the design session's exit criterion (c), 2026-11-21)
  recorded_at: '2026-11-21T18:15:32.102Z'
id: FACT-SUBJECT-UI-CLIENT-EVENTS-LIMIT-KEY
type: fact
---
What one browser's reports of client events are counted under in the built-in UI's limit per address: an IPv4 address as itself, an IPv4-mapped IPv6 address as the IPv4 address it maps, an IPv6 address as its /64 network, and an address that is unknown or no address as one shared key. It does not cover how many reports a key may make, the cap on every address together, or the Ask box's own limits.