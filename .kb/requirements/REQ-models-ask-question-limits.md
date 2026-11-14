---
title: The Ask box's door counts each question that goes to a model, once and never for a fixed command, per session and per address in rolling hours and days, refuses a question past the configured limits with the moment the limit ends and counts nothing for it, checks the day's budget first, persists the counts so a restart and every replica agree and so simultaneous questions can't pass the last place, keeps an address only as a salted hash, counts an IPv6 address as its /64, and leaves the address limit out when the address is unknown; with no limit set every question is allowed and nothing is counted.
status: open
firmness: default
semantic_text: The Ask box's door counts each question that goes to a model, once and never for a fixed command, per session and per address in rolling hours and days, refuses a question past the configured limits with the moment the limit ends and counts nothing for it, checks the day's budget first, persists the counts so a restart and every replica agree and so simultaneous questions can't pass the last place, keeps an address only as a salted hash, counts an IPv6 address as its /64, and leaves the address limit out when the address is unknown; with no limit set every question is allowed and nothing is counted.
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T20:37:44.506Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 667b52bfdd04e9bf11bdb1ed05f80cdba52171cf44337c845f13553b334a3a62
semantic_inventory:
  - claim_key: CLAIM-C754D37AFCAE7198
    claim_text: The Ask box's door counts each question that goes to a model, once and never for a fixed command, per session and per address in rolling hours and days, refuses a question past the configured limits with the moment the limit ends and counts nothing for it, checks the day's budget first, persists the counts so a restart and every replica agree and so simultaneous questions can't pass the last place, keeps an address only as a salted hash, counts an IPv6 address as its /64, and leaves the address limit out when the address is unknown
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 537
  - claim_key: CLAIM-93BB3913ED7427C9
    claim_text: with no limit set every question is allowed and nothing is counted
    role: descriptive
    status: missing
    span:
      start: 539
      end: 605
logic_claims:
  - CLAIM-C754D37AFCAE7198
  - CLAIM-93BB3913ED7427C9
id: REQ-models-ask-question-limits
type: req
---
The Ask box's door counts each question that goes to a model, once and never for a fixed command, per session and per address in rolling hours and days, refuses a question past the configured limits with the moment the limit ends and counts nothing for it, checks the day's budget first, persists the counts so a restart and every replica agree and so simultaneous questions can't pass the last place, keeps an address only as a salted hash, counts an IPv6 address as its /64, and leaves the address limit out when the address is unknown; with no limit set every question is allowed and nothing is counted.

## Context

D-208's third guard: per-session and per-IP limits on model questions, 10 an hour and 30 a day, with the Ask box saying when it resets while the commands keep working. In memory a restart would give every visitor a fresh allowance and the core can run several replicas, so the counts are in the database. An IP address is personal data, so only a salted hash of it is stored and for a day at most.

## Source

> The design session's M5 corrections (2026-11-14): Per-session and per-IP limits on model questions (not on fixed commands): 10 questions an hour and 30 a day each, configurable; persisted or in-memory is to be stated and why; past a limit the box says so with when it resets, and the commands still work (M5 brief, 2026-11-14).