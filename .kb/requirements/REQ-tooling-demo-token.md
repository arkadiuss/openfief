---
title: The demo automation must read the Hetzner API token only from HCLOUD_TOKEN, must fail with an explanation when it is unset, and must never print the token or write it to disk.
status: open
text_ref: AGENTS.md#hard-rules
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:agents-md-hard-rules
  - lane:strict
semantic_text: The demo automation must read the Hetzner API token only from HCLOUD_TOKEN, must fail with an explanation when it is unset, and must never print the token or write it to disk.
logic_claims:
  - CLAIM-B3566993445FEEBC
semantic_clauses:
  - The demo automation must read the Hetzner API token only from HCLOUD_TOKEN, must fail with an explanation when it is unset, and must never print the token or write it to disk.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7265e68d4f2bc84120f3befdd419ce58ad65a4629695a35138a0baa083997d04
semantic_inventory:
  - claim_key: CLAIM-B3566993445FEEBC
    claim_text: The demo automation must read the Hetzner API token only from HCLOUD_TOKEN, must fail with an explanation when it is unset, and must never print the token or write it to disk
    role: normative
    status: modeled
    span:
      start: 0
      end: 174
origin:
  kind: agent
  recorded_at: '2026-10-28T09:00:53.466Z'
id: REQ-tooling-demo-token
type: req
---
The demo automation must read the Hetzner API token only from HCLOUD_TOKEN, must fail with an explanation when it is unset, and must never print the token or write it to disk.

## Context

The token can create and delete paid resources in the owner's project. The hard rules forbid credentials in code, logs and commits; the scripts also ignore stored hcloud CLI contexts so they never act on another project. Modeled by Claude Code for the demo machine automation, which the orchestrating agent specified from the author's request for good committed automation of the demo machine, easy to create and recreate, with spending limited to 10 EUR a month if possible; the author has not reviewed it yet.

## Source

> No real financial data, account numbers or credentials anywhere: code, tests, recordings, logs, commits.

AGENTS.md#hard-rules
