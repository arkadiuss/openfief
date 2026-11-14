---
title: The model router counts every call's cost against a daily budget in US dollars, priced from the model entries' prices per million tokens, refuses a call before sending it once the day's total has reached the configured cap, records the refusal as an audit entry of class budget, and withholds an answer whose cost it can't record; the day is the core's time zone's, the total is persisted so a restart and every replica see it, and with no cap set nothing is read or recorded.
status: open
firmness: default
semantic_text: The model router counts every call's cost against a daily budget in US dollars, priced from the model entries' prices per million tokens, refuses a call before sending it once the day's total has reached the configured cap, records the refusal as an audit entry of class budget, and withholds an answer whose cost it can't record; the day is the core's time zone's, the total is persisted so a restart and every replica see it, and with no cap set nothing is read or recorded.
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T20:37:42.605Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6e4b129eadf582e9e6ece36eacaa47b212bf8332950254141cf1f61836249425
semantic_inventory:
  - claim_key: CLAIM-004DBD465FC462EC
    claim_text: The model router counts every call's cost against a daily budget in US dollars, priced from the model entries' prices per million tokens, refuses a call before sending it once the day's total has reached the configured cap, records the refusal as an audit entry of class budget
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 277
  - claim_key: CLAIM-88774B46B9233BDA
    claim_text: withholds an answer whose cost it can't record
    role: normative
    status: ontology_gap
    span:
      start: 283
      end: 329
  - claim_key: CLAIM-F3841463A5ABFE49
    claim_text: the day is the core's time zone's, the total is persisted so a restart and every replica see it, and with no cap set nothing is read or recorded
    role: descriptive
    status: missing
    span:
      start: 331
      end: 475
logic_claims:
  - CLAIM-004DBD465FC462EC
  - CLAIM-88774B46B9233BDA
  - CLAIM-F3841463A5ABFE49
id: REQ-models-daily-budget
type: req
---
The model router counts every call's cost against a daily budget in US dollars, priced from the model entries' prices per million tokens, refuses a call before sending it once the day's total has reached the configured cap, records the refusal as an audit entry of class budget, and withholds an answer whose cost it can't record; the day is the core's time zone's, the total is persisted so a restart and every replica see it, and with no cap set nothing is read or recorded.

## Context

D-208's second guard: a daily token budget the core enforces before each call, priced from the model entries, after which the Ask box says the day's demo budget is used up. The cap is a flag of the core, off by default so a self-hoster is unaffected, and set by the demo's component. Failing closed on a usage that can't be read or written matches the audit log (D-211).

## Source

> The design session's M5 corrections (2026-11-14): Daily token budget in the core: priced from the model entries' prices (tokens times price), a configurable daily USD cap, counted across all model calls of the day, persisted; past it the router refuses before calling, the audit entry says so, and the Ask box says the day's budget is used up and keeps its commands (M5 brief, 2026-11-14).