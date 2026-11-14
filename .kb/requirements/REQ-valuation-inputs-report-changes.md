---
title: Storing prices, FX rates or bond terms must report how many rows were new or differed from the stored ones, and a row sent again unchanged must count for none.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:41.261Z'
semantic_text: Storing prices, FX rates or bond terms must report how many rows were new or differed from the stored ones, and a row sent again unchanged must count for none.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 57c5e8f131226995f7508294b88f783741c6d032ed04e26ae9a9db008fb2154c
semantic_inventory:
  - claim_key: CLAIM-3F09970B34AAB6A4
    claim_text: Storing prices, FX rates or bond terms must report how many rows were new or differed from the stored ones
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 106
  - claim_key: CLAIM-D7092F13A17FC6B2
    claim_text: a row sent again unchanged must count for none
    role: normative
    status: ontology_gap
    span:
      start: 112
      end: 158
logic_claims:
  - CLAIM-3F09970B34AAB6A4
  - CLAIM-D7092F13A17FC6B2
id: REQ-valuation-inputs-report-changes
type: req
---
Storing prices, FX rates or bond terms must report how many rows were new or differed from the stored ones, and a row sent again unchanged must count for none.

## Context

A sync's audit entry is written only when it changed state, and the stores upsert: without this count a connector that re-sends the same prices every sync would enter a change each time. Asked for by the author on 2026-11-14.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S9 brief, from D-211, 2026-11-14