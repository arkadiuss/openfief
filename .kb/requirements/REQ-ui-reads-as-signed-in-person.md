---
title: The built-in UI reads the record with the signed-in person's own session token, never with the plugin's identity, so the core decides what each person sees.
status: open
firmness: default
required_evidence: unverified
semantic_text: The built-in UI reads the record with the signed-in person's own session token, never with the plugin's identity, so the core decides what each person sees.
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 81b83ad34196c701f47f071c6bd94e862a0feca1b8623976fd828e4745a828d4
semantic_inventory:
  - claim_key: CLAIM-3DBBB68626F051C2
    claim_text: The built-in UI reads the record with the signed-in person's own session token, never with the plugin's identity, so the core decides what each person sees
    role: descriptive
    status: missing
    span:
      start: 0
      end: 155
logic_claims:
  - CLAIM-3DBBB68626F051C2
id: REQ-ui-reads-as-signed-in-person
type: req
---
The built-in UI reads the record with the signed-in person's own session token, never with the plugin's identity, so the core decides what each person sees.

## Context

Design-session ask: a test that fails if the UI reads with its plugin identity. Decisions: D-163, D-210.

## Source

> S6 task brief and the permissions plan's author and design-session answers (Q6, Q7, Q8), 2026-11-13.